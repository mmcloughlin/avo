package load

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/opcodescsv"
	"github.com/mmcloughlin/avo/internal/opcodesextra"
	"github.com/mmcloughlin/avo/internal/opcodesxml"
)

// This file is a mess. Some of this complexity is unavoidable, since the state
// of x86 instruction databases is also a mess, especially when it comes to
// idiosyncrasies of the Go assembler implementation. Some of the complexity is
// probably avoidable by migrating to using Intel XED
// (https://github.com/mmcloughlin/avo/issues/23), but for now this is an unholy
// mix of PeachPy's Opcodes database and Go's x86 CSV file.
//
// The goal is simply to keep as much of the uglyness in this file as possible,
// producing a clean instruction database for the rest of avo to use. Any nasty
// logic here should be backed up with a test somewhere to ensure the result is
// correct, even if the code that produced it is awful.

// Expected data source filenames.
const (
	DefaultCSVName        = "x86.v0.2.csv"
	DefaultOpcodesXMLName = "x86_64.xml"
)

// Loader builds an instruction database from underlying datasources.
type Loader struct {
	X86CSVPath     string
	OpcodesXMLPath string

	alias map[opcodescsv.Alias]string
	order map[string]opcodescsv.OperandOrder
}

// NewLoaderFromDataDir constructs an instruction loader from datafiles in the
// given directory. The files themselves are expected to have the standard
// names: see Default*Name constants.
func NewLoaderFromDataDir(dir string) *Loader {
	return &Loader{
		X86CSVPath:     filepath.Join(dir, DefaultCSVName),
		OpcodesXMLPath: filepath.Join(dir, DefaultOpcodesXMLName),
	}
}

// Load performs instruction loading.
func (l *Loader) Load() ([]inst.Instruction, error) {
	if err := l.init(); err != nil {
		return nil, err
	}

	// Load Opcodes XML file.
	iset, err := opcodesxml.ReadFile(l.OpcodesXMLPath)
	if err != nil {
		return nil, err
	}

	// Load opcodes XML data, grouped by Go opcode.
	im := map[string]*inst.Instruction{}
	for _, i := range iset.Instructions {
		for _, f := range i.Forms {
			if !l.include(f) {
				continue
			}

			for _, opcode := range l.gonames(f) {
				if im[opcode] == nil {
					im[opcode] = &inst.Instruction{
						Opcode:  opcode,
						Summary: i.Summary,
					}
				}
				forms := l.forms(opcode, f)
				im[opcode].Forms = append(im[opcode].Forms, forms...)
			}
		}
	}

	// Add extras to our list.
	for _, e := range opcodesextra.Instructions() {
		im[e.Opcode] = e
	}

	// Merge aliased forms. This is primarily for MOVQ (issue #50).
	for _, a := range aliases {
		if existing, found := im[a.From]; found {
			im[a.To].Forms = append(im[a.To].Forms, existing.Forms...)
		}
	}

	// Apply list of aliases.
	for _, a := range aliases {
		cpy := *im[a.To]
		cpy.Opcode = a.From
		cpy.AliasOf = a.To
		im[a.From] = &cpy
	}

	// Dedupe forms.
	for _, i := range im {
		i.Forms = dedupe(i.Forms)
	}

	// Resolve forms that have VEX and EVEX encoded forms.
	for _, i := range im {
		i.Forms, err = vexevex(i.Forms)
		if err != nil {
			return nil, err
		}
	}

	// Convert to a slice. Sort instructions and forms for reproducibility.
	is := make([]inst.Instruction, 0, len(im))
	for _, i := range im {
		is = append(is, *i)
	}

	sort.Slice(is, func(i, j int) bool {
		return is[i].Opcode < is[j].Opcode
	})

	for _, i := range im {
		sortforms(i.Forms)
	}

	return is, nil
}

func (l *Loader) init() error {
	icsv, err := opcodescsv.ReadFile(l.X86CSVPath)
	if err != nil {
		return err
	}

	l.alias, err = opcodescsv.BuildAliasMap(icsv)
	if err != nil {
		return err
	}

	l.order = opcodescsv.BuildOrderMap(icsv)

	return nil
}

// include decides whether to include the instruction form in the avo listing.
// This discards some opcodes that are not supported in Go.
func (l Loader) include(f opcodesxml.Form) bool {
	// Exclude certain ISAs simply not present in Go
	for _, isa := range f.ISA {
		switch isa.ID {
		// AMD-only.
		case "TBM", "CLZERO", "FMA4", "XOP", "SSE4A", "3dnow!", "3dnow!+":
			return false
		// AVX512PF doesn't work without some special case handling, and is only on Knights Landing/Knights Mill.
		case "AVX512PF":
			return false
		// Incomplete support for some prefetching instructions.
		case "PREFETCH", "PREFETCHW", "PREFETCHWT1", "CLWB":
			return false
		// Remaining oddities.
		case "MONITORX", "FEMMS":
			return false
		}
	}

	// Go appears to have skeleton support for MMX instructions. See the many TODO lines in the testcases:
	// Reference: https://github.com/golang/go/blob/649b89377e91ad6dbe710784f9e662082d31a1ff/src/cmd/asm/internal/asm/testdata/amd64enc.s#L3310-L3312
	//
	//		//TODO: PALIGNR $7, (BX), M2            // 0f3a0f1307
	//		//TODO: PALIGNR $7, (R11), M2           // 410f3a0f1307
	//		//TODO: PALIGNR $7, M2, M2              // 0f3a0fd207
	//
	if f.MMXMode == "MMX" {
		return false
	}

	// x86 csv contains a number of CMOV* instructions which are actually not valid
	// Go instructions. The valid Go forms should have different opcodes from GNU.
	// Therefore a decent "heuristic" is CMOV* instructions that do not have
	// aliases.
	if strings.HasPrefix(f.GASName, "cmov") && l.lookupAlias(f) == "" {
		return false
	}

	// Some specific exclusions.
	switch f.GASName {
	// Certain branch instructions appear to not be supported.
	//
	// Reference: https://github.com/golang/go/blob/649b89377e91ad6dbe710784f9e662082d31a1ff/src/cmd/asm/internal/asm/testdata/amd64enc.s#L757
	//
	//		//TODO: CALLQ* (BX)                     // ff13
	//
	// Reference: https://github.com/golang/go/blob/649b89377e91ad6dbe710784f9e662082d31a1ff/src/cmd/asm/internal/asm/testdata/amd64enc.s#L2108
	//
	//		//TODO: LJMPL* (R11)                    // 41ff2b
	//
	case "callq", "jmpl":
		return false
	// MOVABS doesn't seem to be supported either.
	//
	// Reference: https://github.com/golang/go/blob/1ac84999b93876bb06887e483ae45b27e03d7423/src/cmd/asm/internal/asm/testdata/amd64enc.s#L2354
	//
	//		//TODO: MOVABSB 0x123456789abcdef1, AL  // a0f1debc9a78563412
	//
	case "movabs":
		return false
	// Only one XLAT form is supported.
	//
	// Reference: https://github.com/golang/arch/blob/b19384d3c130858bb31a343ea8fce26be71b5998/x86/x86.v0.2.csv#L2221-L2222
	//
	//	"XLATB","XLAT","xlat","D7","V","V","","ignoreREXW","","",""
	//	"XLATB","XLAT","xlat","REX.W D7","N.E.","V","","pseudo","","",""
	//
	// Reference: https://github.com/golang/go/blob/b3294d9491b898396e134bad5412d85337c37b64/src/cmd/internal/obj/x86/asm6.go#L1519
	//
	//		{AXLAT, ynone, Px, opBytes{0xd7}},
	//
	// TODO(mbm): confirm the Px prefix means non REX mode
	case "xlatb":
		return f.Encoding.REX == nil
	}

	return true
}

func (l Loader) lookupAlias(f opcodesxml.Form) string {
	// Attempt lookup with datasize.
	k := opcodescsv.Alias{
		Opcode:      f.GASName,
		DataSize:    datasize(f),
		NumOperands: len(f.Operands),
	}
	if a := l.alias[k]; a != "" {
		return a
	}

	// Fallback to unknown datasize.
	k.DataSize = 0
	return l.alias[k]
}

func (l Loader) gonames(f opcodesxml.Form) []string {
	s := datasize(f)

	// Suspect a "bug" in x86 CSV for the CVTTSD2SQ instruction, as CVTTSD2SL appears twice.
	//
	// Reference: https://github.com/golang/arch/blob/b19384d3c130858bb31a343ea8fce26be71b5998/x86/x86.v0.2.csv#L345-L346
	//
	//	"CVTTSD2SI r32, xmm2/m64","CVTTSD2SL xmm2/m64, r32","cvttsd2si xmm2/m64, r32","F2 0F 2C /r","V","V","SSE2","operand16,operand32","w,r","Y","32"
	//	"CVTTSD2SI r64, xmm2/m64","CVTTSD2SL xmm2/m64, r64","cvttsd2siq xmm2/m64, r64","F2 REX.W 0F 2C /r","N.E.","V","SSE2","","w,r","Y","64"
	//
	// Reference: https://github.com/golang/go/blob/048c9164a0c5572df18325e377473e7893dbfb07/src/cmd/internal/obj/x86/asm6.go#L1073-L1074
	//
	//		{ACVTTSD2SL, yxcvfl, Pf2, opBytes{0x2c}},
	//		{ACVTTSD2SQ, yxcvfq, Pw, opBytes{Pf2, 0x2c}},
	//
	if f.GASName == "cvttsd2si" && s == 64 {
		return []string{"CVTTSD2SQ"}
	}

	// Return alias if available.
	if a := l.lookupAlias(f); a != "" {
		return []string{a}
	}

	// Some odd special cases.
	// TODO(mbm): can this be handled by processing csv entries with slashes /
	if f.GoName == "RET" && len(f.Operands) == 1 {
		return []string{"RETFW", "RETFL", "RETFQ"}
	}

	// IMUL 3-operand forms are not recorded correctly in either x86 CSV or opcodes. They are supposed to be called IMUL3{W,L,Q}
	//
	// Reference: https://github.com/golang/go/blob/649b89377e91ad6dbe710784f9e662082d31a1ff/src/cmd/internal/obj/x86/asm6.go#L1112-L1114
	//
	//		{AIMUL3W, yimul3, Pe, opBytes{0x6b, 00, 0x69, 00}},
	//		{AIMUL3L, yimul3, Px, opBytes{0x6b, 00, 0x69, 00}},
	//		{AIMUL3Q, yimul3, Pw, opBytes{0x6b, 00, 0x69, 00}},
	//
	// Reference: https://github.com/golang/arch/blob/b19384d3c130858bb31a343ea8fce26be71b5998/x86/x86.v0.2.csv#L549
	//
	//	"IMUL r32, r/m32, imm32","IMULL imm32, r/m32, r32","imull imm32, r/m32, r32","69 /r id","V","V","","operand32","rw,r,r","Y","32"
	//
	if strings.HasPrefix(f.GASName, "imul") && len(f.Operands) == 3 {
		return []string{strings.ToUpper(f.GASName[:4] + "3" + f.GASName[4:])}
	}

	// Use go opcode from Opcodes XML where available.
	if f.GoName != "" {
		return []string{f.GoName}
	}

	// Fallback to GAS name.
	n := strings.ToUpper(f.GASName)

	// Some need data sizes added to them.
	n += sizesuffix(n, f)

	return []string{n}
}

func (l Loader) forms(opcode string, f opcodesxml.Form) []inst.Form {
	// Map operands to avo format and ensure correct order.
	ops := operands(f.Operands)

	switch l.order[opcode] {
	case opcodescsv.IntelOrder:
		// Nothing to do.
	case opcodescsv.CMP3Order:
		ops[0], ops[1] = ops[1], ops[0]
	case opcodescsv.UnknownOrder:
		// Instructions not in x86 CSV are assumed to have reverse intel order.
		fallthrough
	case opcodescsv.ReverseIntelOrder:
		for l, r := 0, len(ops)-1; l < r; l, r = l+1, r-1 {
			ops[l], ops[r] = ops[r], ops[l]
		}
	}

	// Handle some exceptions.
	// TODO(mbm): consider if there's some nicer way to handle the list of special cases.
	switch opcode {
	// Go assembler has an internal Yu2 operand type for unsigned 2-bit immediates.
	//
	// Reference: https://github.com/golang/go/blob/6d5caf38e37bf9aeba3291f1f0b0081f934b1187/src/cmd/internal/obj/x86/asm6.go#L109
	//
	//		Yu2 // $x, x fits in uint2
	//
	// Reference: https://github.com/golang/go/blob/6d5caf38e37bf9aeba3291f1f0b0081f934b1187/src/cmd/internal/obj/x86/asm6.go#L858-L864
	//
	//	var yextractps = []ytab{
	//		{Zibr_m, 2, argList{Yu2, Yxr, Yml}},
	//	}
	//
	//	var ysha1rnds4 = []ytab{
	//		{Zibm_r, 2, argList{Yu2, Yxm, Yxr}},
	//	}
	//
	case "SHA1RNDS4", "EXTRACTPS":
		ops[0].Type = "imm2u"
	}

	// Extract implicit operands.
	var implicits []inst.ImplicitOperand
	for _, implicit := range f.ImplicitOperands {
		implicits = append(implicits, inst.ImplicitOperand{
			Register: implicit.ID,
			Action:   inst.ActionFromReadWrite(implicit.Input, implicit.Output),
		})
	}

	// Extract ISA flags.
	var isas []string
	for _, isa := range f.ISA {
		isas = append(isas, isa.ID)
	}
	sort.Strings(isas)

	// Initialize form.
	form := inst.Form{
		ISA:              isas,
		Operands:         ops,
		ImplicitOperands: implicits,
		EncodingType:     enctype(f),
		CancellingInputs: f.CancellingInputs,
	}

	// Apply modification stages to produce final list of forms.
	stages := []func(string, inst.Form) []inst.Form{
		avx512rounding,
		avx512sae,
		avx512bcst,
		avx512masking,
		avx512zeroing,
	}

	forms := []inst.Form{form}
	for _, stage := range stages {
		var next []inst.Form
		for _, f := range forms {
			next = append(next, stage(opcode, f)...)
		}
		forms = next
	}

	return forms
}

// operands maps Opcodes XML operands to avo format. Returned in Intel order.
func operands(ops []opcodesxml.Operand) []inst.Operand {
	n := len(ops)
	r := make([]inst.Operand, n)
	for i, op := range ops {
		r[i] = operand(op)
	}
	return r
}

// operand maps an Opcodes XML operand to avo format.
func operand(op opcodesxml.Operand) inst.Operand {
	return inst.Operand{
		Type:   op.Type,
		Action: inst.ActionFromReadWrite(op.Input, op.Output),
	}
}

// avx512rounding handles AVX-512 embedded rounding. Opcodes database represents
// these as {er} operands, whereas Go uses instruction suffixes. Remove the
// operand if present and set the corresponding flag.
func avx512rounding(opcode string, f inst.Form) []inst.Form {
	i, found := findoperand(f.Operands, "{er}")
	if !found {
		return []inst.Form{f}
	}

	// Delete the {er} operand.
	f.Operands = append(f.Operands[:i], f.Operands[i+1:]...)

	// Create a second form with the rounding flag.
	er := f.Clone()
	er.EmbeddedRounding = true

	return []inst.Form{f, er}
}

// avx512sae handles AVX-512 "suppress all exceptions". Opcodes database
// represents these as {sae} operands, whereas Go uses instruction suffixes.
// Remove the operand if present and set the corresponding flag.
func avx512sae(opcode string, f inst.Form) []inst.Form {
	i, found := findoperand(f.Operands, "{sae}")
	if !found {
		return []inst.Form{f}
	}

	// Delete the {sae} operand.
	f.Operands = append(f.Operands[:i], f.Operands[i+1:]...)

	// Create a second form with the rounding flag.
	sae := f.Clone()
	sae.SuppressAllExceptions = true

	return []inst.Form{f, sae}
}

// avx512bcst handles AVX-512 broadcast. Opcodes database uses operands like
// "m512/m64bcst" to indicate broadcast. Go uses the BCST suffix to enable it.
// Split the form into two, the regular and broadcast versions.
func avx512bcst(opcode string, f inst.Form) []inst.Form {
	// Look for broadcast operand.
	idx := -1
	for i, op := range f.Operands {
		if bcstrx.MatchString(op.Type) {
			idx = i
			break
		}
	}

	if idx < 0 {
		return []inst.Form{f}
	}

	// Create two forms.
	match := bcstrx.FindStringSubmatch(f.Operands[idx].Type)

	mem := f.Clone()
	mem.Operands[idx].Type = match[1]

	bcst := f.Clone()
	bcst.Broadcast = true
	bcst.Operands[idx].Type = match[2]

	return []inst.Form{mem, bcst}
}

var bcstrx = regexp.MustCompile(`^(m\d+)/(m\d+)bcst$`)

// avx512masking handles AVX-512 masking forms.
func avx512masking(opcode string, f inst.Form) []inst.Form {
	// In order to support implicit masking (with K0), Go has two instruction
	// forms, one with the mask and one without. The mask register precedes the
	// output register. The Opcodes database (similar to Intel manuals)
	// represents masking with the {k} operand suffix, possibly with {z} for
	// zeroing.

	// Look for masking with possible zeroing. Zeroing is handled by a later
	// processing stage, but we need to be sure to notice and preserve it here.
	masking := false
	zeroing := false
	idx := -1
	for i := range f.Operands {
		op := &f.Operands[i]
		if strings.HasSuffix(op.Type, "{z}") {
			zeroing = true
			op.Type = strings.TrimSuffix(op.Type, "{z}")
		}
		if strings.HasSuffix(op.Type, "{k}") {
			masking = true
			idx = i
			op.Type = strings.TrimSuffix(op.Type, "{k}")
			break
		}
	}

	// Bail if no masking.
	if !masking {
		return []inst.Form{f}
	}

	// Unmasked variant.
	unmasked := f.Clone()

	// Masked form has "k" operand inserted.
	masked := f.Clone()
	mask := inst.Operand{Type: "k", Action: inst.R}
	ops := append([]inst.Operand(nil), masked.Operands[:idx]...)
	ops = append(ops, mask)
	ops = append(ops, masked.Operands[idx:]...)
	masked.Operands = ops

	// Restore zeroing suffix, so it can he handled later.
	if zeroing {
		masked.Operands[idx+1].Type += "{z}"
	}

	// Almost all instructions take an optional mask, apart from a few
	// special cases.
	if maskrequired[opcode] {
		return []inst.Form{masked}
	}
	return []inst.Form{unmasked, masked}
}

// avx512zeroing handles AVX-512 zeroing forms.
func avx512zeroing(opcode string, f inst.Form) []inst.Form {
	// Zeroing in Go is handled with the Z opcode suffix. Note that zeroing has
	// an important effect on the instruction form, since the merge masking form
	// has an input dependency for the output register, and the zeroing form
	// does not.

	// Look for zeroing operand.
	idx := -1
	for i := range f.Operands {
		op := &f.Operands[i]
		if strings.HasSuffix(op.Type, "{z}") {
			idx = i
			op.Type = strings.TrimSuffix(op.Type, "{z}")
		}
	}

	if idx < 0 {
		return []inst.Form{f}
	}

	// Duplicate into two forms for merging and zeroing.
	merging := f.Clone()
	merging.Operands[idx].Action |= inst.R

	zeroing := f.Clone()
	zeroing.Zeroing = true

	return []inst.Form{merging, zeroing}
}

// findoperand looks for an operand type and returns its index, if found.
func findoperand(ops []inst.Operand, t string) (int, bool) {
	for i, op := range ops {
		if op.Type == t {
			return i, true
		}
	}
	return 0, false
}

// enctype selects the encoding type for the instruction form.
func enctype(f opcodesxml.Form) inst.EncodingType {
	switch {
	case f.Encoding.EVEX != nil:
		return inst.EncodingTypeEVEX
	case f.Encoding.VEX != nil:
		return inst.EncodingTypeVEX
	case f.Encoding.REX != nil:
		return inst.EncodingTypeREX
	default:
		return inst.EncodingTypeLegacy
	}
}

// datasize (intelligently) guesses the datasize of an instruction form.
func datasize(f opcodesxml.Form) int {
	// Determine from encoding bits.
	e := f.Encoding
	if e.VEX != nil && e.VEX.W == nil {
		return 128 << e.VEX.L
	}

	// Guess from operand types.
	size := 0
	for _, op := range f.Operands {
		s := operandsize(op)
		if s != 0 && (size == 0 || op.Output) {
			size = s
		}
	}

	return size
}

func operandsize(op opcodesxml.Operand) int {
	for s := 256; s >= 8; s /= 2 {
		if strings.HasSuffix(op.Type, strconv.Itoa(s)) {
			return s
		}
	}
	return 0
}

// sizesuffix returns an optional size suffix to be added to the opcode name.
func sizesuffix(n string, f opcodesxml.Form) string {
	// Reference: https://github.com/golang/arch/blob/5de9028c2478e6cb4e1c1b1f4386f3f0a93e383a/x86/x86avxgen/main.go#L275-L322
	//
	//	func addGoSuffixes(ctx *context) {
	//		var opcodeSuffixMatchers map[string][]string
	//		{
	//			opXY := []string{"VL=0", "X", "VL=1", "Y"}
	//			opXYZ := []string{"VL=0", "X", "VL=1", "Y", "VL=2", "Z"}
	//			opQ := []string{"REXW=1", "Q"}
	//			opLQ := []string{"REXW=0", "L", "REXW=1", "Q"}
	//
	//			opcodeSuffixMatchers = map[string][]string{
	//				"VCVTPD2DQ":   opXY,
	//				"VCVTPD2PS":   opXY,
	//				"VCVTTPD2DQ":  opXY,
	//				"VCVTQQ2PS":   opXY,
	//				"VCVTUQQ2PS":  opXY,
	//				"VCVTPD2UDQ":  opXY,
	//				"VCVTTPD2UDQ": opXY,
	//
	//				"VFPCLASSPD": opXYZ,
	//				"VFPCLASSPS": opXYZ,
	//
	//				"VCVTSD2SI":  opQ,
	//				"VCVTTSD2SI": opQ,
	//				"VCVTTSS2SI": opQ,
	//				"VCVTSS2SI":  opQ,
	//
	//				"VCVTSD2USI":  opLQ,
	//				"VCVTSS2USI":  opLQ,
	//				"VCVTTSD2USI": opLQ,
	//				"VCVTTSS2USI": opLQ,
	//				"VCVTUSI2SD":  opLQ,
	//				"VCVTUSI2SS":  opLQ,
	//				"VCVTSI2SD":   opLQ,
	//				"VCVTSI2SS":   opLQ,
	//				"ANDN":        opLQ,
	//				"BEXTR":       opLQ,
	//				"BLSI":        opLQ,
	//				"BLSMSK":      opLQ,
	//				"BLSR":        opLQ,
	//				"BZHI":        opLQ,
	//				"MULX":        opLQ,
	//				"PDEP":        opLQ,
	//				"PEXT":        opLQ,
	//				"RORX":        opLQ,
	//				"SARX":        opLQ,
	//				"SHLX":        opLQ,
	//				"SHRX":        opLQ,
	//			}
	//		}
	//

	type rule struct {
		Size   func(opcodesxml.Form) int
		Suffix map[int]string
	}

	var (
		XY  = rule{evexLLsize, map[int]string{128: "X", 256: "Y"}}
		XYZ = rule{evexLLsize, map[int]string{128: "X", 256: "Y", 512: "Z"}}
		Q   = rule{rexWsize, map[int]string{64: "Q"}}
		LQ  = rule{rexWsize, map[int]string{32: "L", 64: "Q"}}
		WLQ = rule{datasize, map[int]string{16: "W", 32: "L", 64: "Q"}}
	)

	rules := map[string]rule{
		"VCVTPD2DQ":   XY,
		"VCVTPD2PS":   XY,
		"VCVTTPD2DQ":  XY,
		"VCVTQQ2PS":   XY,
		"VCVTUQQ2PS":  XY,
		"VCVTPD2UDQ":  XY,
		"VCVTTPD2UDQ": XY,

		"VFPCLASSPD": XYZ,
		"VFPCLASSPS": XYZ,

		"VCVTSD2SI":  Q,
		"VCVTTSD2SI": Q,
		"VCVTTSS2SI": Q,
		"VCVTSS2SI":  Q,

		"VCVTSD2USI":  LQ,
		"VCVTSS2USI":  LQ,
		"VCVTTSD2USI": LQ,
		"VCVTTSS2USI": LQ,
		"VCVTUSI2SD":  LQ,
		"VCVTUSI2SS":  LQ,
		"VCVTSI2SD":   LQ,
		"VCVTSI2SS":   LQ,
		"ANDN":        LQ,
		"BEXTR":       LQ,
		"BLSI":        LQ,
		"BLSMSK":      LQ,
		"BLSR":        LQ,
		"BZHI":        LQ,
		"MULX":        LQ,
		"PDEP":        LQ,
		"PEXT":        LQ,
		"RORX":        LQ,
		"SARX":        LQ,
		"SHLX":        LQ,
		"SHRX":        LQ,

		"RDRAND": LQ,
		"RDSEED": LQ,

		// MOVEBE* instructions seem to be inconsistent with x86 CSV.
		//
		// Reference: https://github.com/golang/arch/blob/b19384d3c130858bb31a343ea8fce26be71b5998/x86/x86spec/format.go#L282-L287
		//
		//		"MOVBE r16, m16": "movbeww",
		//		"MOVBE m16, r16": "movbeww",
		//		"MOVBE m32, r32": "movbell",
		//		"MOVBE r32, m32": "movbell",
		//		"MOVBE m64, r64": "movbeqq",
		//		"MOVBE r64, m64": "movbeqq",
		//
		"MOVBEW": WLQ,
		"MOVBEL": WLQ,
		"MOVBEQ": WLQ,
	}

	r, ok := rules[n]
	if !ok {
		return ""
	}

	s := r.Size(f)
	return r.Suffix[s]
}

func rexWsize(f opcodesxml.Form) int {
	e := f.Encoding
	switch {
	case e.EVEX != nil && e.EVEX.W != nil:
		return 32 << *e.EVEX.W
	default:
		return 32
	}
}

func evexLLsize(f opcodesxml.Form) int {
	e := f.Encoding
	if e.EVEX == nil {
		return 0
	}
	size := map[string]int{"00": 128, "01": 256, "10": 512}
	return size[e.EVEX.LL]
}

// vexevex fixes instructions that have both VEX and EVEX encoded forms with the
// same operand types. Go uses the VEX encoded form unless EVEX-only features
// are used. This function will only keep the VEX encoded version in the case
// where both exist.
//
// Note this is somewhat of a hack. There are real reasons to use the EVEX
// encoded version even when both exist. The main reason to use the EVEX version
// rather than VEX is to use the registers Z16, Z17, ... and up. However, avo
// does not implement the logic to distinguish between the two halfs of the
// vector registers. So in its current state the only reason to need the EVEX
// version is to encode suffixes, and these are represented by other instruction
// forms.
//
// TODO(mbm): restrict use of vector registers https://github.com/mmcloughlin/avo/issues/146
func vexevex(fs []inst.Form) ([]inst.Form, error) {
	// Group forms by deduping ID.
	byid := map[string][]inst.Form{}
	for _, f := range fs {
		id := fmt.Sprintf(
			"%s {%t,%t,%t,%t}",
			strings.Join(f.Signature(), "_"),
			f.Zeroing,
			f.EmbeddedRounding,
			f.SuppressAllExceptions,
			f.Broadcast,
		)
		byid[id] = append(byid[id], f)
	}

	// Resolve overlaps.
	var results []inst.Form
	for id, group := range byid {
		if len(group) < 2 {
			results = append(results, group...)
			continue
		}

		// We expect these conflicts are caused by VEX/EVEX pairs. Bail if it's
		// something else.
		if len(group) > 2 {
			return nil, fmt.Errorf("more than two forms of type %q", id)
		}

		if group[0].EncodingType == inst.EncodingTypeEVEX {
			group[0], group[1] = group[1], group[0]
		}

		if group[0].EncodingType != inst.EncodingTypeVEX || group[1].EncodingType != inst.EncodingTypeEVEX {
			fmt.Println(group)
			return nil, errors.New("expected pair of VEX/EVEX encoded forms")
		}

		vex := group[0]

		// In this case we only keep the VEX encoded form.
		results = append(results, vex)
	}

	return results, nil
}

// dedupe a list of forms.
func dedupe(fs []inst.Form) []inst.Form {
	uniq := make([]inst.Form, 0, len(fs))
	for _, f := range fs {
		have := false
		for _, u := range uniq {
			if reflect.DeepEqual(u, f) {
				have = true
				break
			}
		}
		if !have {
			uniq = append(uniq, f)
		}
	}
	return uniq
}

// sortforms sorts a list of forms.
func sortforms(fs []inst.Form) {
	sort.Slice(fs, func(i, j int) bool {
		return sortkey(fs[i]) < sortkey(fs[j])
	})
}

func sortkey(f inst.Form) string {
	return fmt.Sprintf("%d %v %v", f.EncodingType, f.ISA, f)
}
