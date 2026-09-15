package printer

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/mmcloughlin/avo/buildtags"
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

// arm64 is an EXPERIMENTAL printer that lowers an avo file whose instructions
// were written against the amd64 ISA into Go arm64 (AArch64) assembly.
//
// It is deliberately NOT a general arm64 backend. avo's frontend, IR, register
// allocator and ABI handling are all amd64-oriented; this printer runs *after*
// pass.Compile and mechanically lowers the already-allocated x86 instruction
// stream to arm64, translating a single x86 instruction into one or more arm64
// instructions. The goal is correct, regenerable arm64 asm that beats Go's
// pure-Go fallback — not optimal hand-tuned code.
//
// Design:
//
//   - Registers: avo allocated physical x86 GP registers. We remap by physical
//     index to a fixed arm64 register (RAX->R0, RCX->R1, ...), reserving R15/R16
//     as scratch for address computation, immediate-to-memory stores and
//     memory-destination read-modify-write. Vector (XMM) registers map by index
//     to V registers. Pseudo registers (FP/SB/SP) pass through unchanged.
//
//   - Memory operands: x86 base+index*scale+disp has no arm64 equivalent, so an
//     indexed operand is lowered to "ADD scratch, base, index<<log2(scale)"
//     followed by a simple disp(scratch) access.
//
//   - Flags: avo's IR does not model EFLAGS. x86 sets flags implicitly and a
//     later Jcc/CMOVcc consumes them. We do NOT fuse; instead:
//
//   - comparisons (CMP*/TEST*) always emit an arm64 flag-setter (CMP/CMN/TST);
//
//   - arithmetic that an immediately-following branch/CMOV consumes is emitted
//     as the flag-setting S-variant (SUBS/ADDS/...);
//
//   - conditional branches and CMOVcc consume the live flags (Bcc / CSEL).
//     This is safe because every other lowering uses non-flag-setting arm64 ops,
//     so NZCV survives from producer to consumer.
//
// Unsupported opcodes panic loudly, so growing the supported set surfaces
// exactly what is missing.
type arm64 struct {
	cfg Config
	prnt.Generator

	// pending buffers lowered instructions (opcode, operands) so a block can be
	// flushed with operands column-aligned, matching the goasm printer (and thus
	// asmfmt). clear tracks whether a blank line is already present.
	pending [][2]string
	clear   bool

	// Set while lowering an instruction isFlagTransparent claims leaves NZCV
	// alone, so emitf() can hold the lowering to that claim.
	inTransparent bool
	transparentOp string

	// The mirror of the above for the producer side: set while lowering an
	// instruction flagProducers marked as supplying a consumer's flags, with a
	// count of the flag-writers its lowering emitted.
	inProducer  bool
	producerOp  string
	writerCount int

	// One-instruction constant-tracking window: constReg holds constVal when
	// constOK and the immediately preceding instruction was "MOVQ/MOVL $imm,
	// constReg". Any other instruction or a label invalidates it (comments are
	// transparent). Used to fold BMI2 register-control ops (BEXTR/BZHI/SHLX/
	// SHRX) whose control register is loaded with a constant right before use
	// -- the x86 encodings have no immediate forms for these, but the arm64
	// equivalents do (UBFX/AND/LSL/LSR), collapsing multi-instruction mask
	// builds into one instruction.
	constReg string
	constVal int64
	constOK  bool

	// The operand rewrites shiftFolds decided for the shift being lowered,
	// zero for every other instruction.
	shift shiftFold
}

// NewARM64Asm constructs a printer for writing Go arm64 assembly files by
// lowering an amd64 avo instruction stream. EXPERIMENTAL.
//
// Preprocessor directives come with a caveat the caller must satisfy. avo emits
// them as comments, so they are inert until something rewrites "\t// #" back to
// "#". This printer resolves GOAMD64 conditionals on the assumption that the
// rewrite happens to BOTH outputs: only the arm the arm64 build takes is
// emitted, and any other directive becomes a boundary the flag and constant
// analyses stop at.
//
// If a generator emits directives and the amd64 output is NOT rewritten, the two
// sides disagree -- amd64 runs both arms of a conditional this printer already
// resolved -- and nothing here can detect it. Note that klauspost/compress
// currently applies that rewrite only in s2, which generates amd64 only; zstd,
// the one generator using -arch amd64,arm64, has no rewrite step and emits no
// directives. So the assumption holds there by vacuity rather than by
// construction, and the first directive added to a zstd generator needs the
// rewrite added with it.
//
// An unsupported instruction or operand form aborts generation with a panic
// rather than emitting assembly that differs from the amd64 original, so a
// program that generates is a program whose two versions agree -- with one
// exception, which cannot be checked at generation time. BZHIQ and BEXTRQ with
// a runtime (non-constant) control assume the bit count is below 64, as it is
// wherever these are emitted today. x86 saturates instead: BZHI with count >= 64
// copies the source unchanged, and BEXTR with len >= 64 extracts everything
// above start. The lowering produces zero in both cases. If a program can reach
// a count of 64 or more, mask it before the instruction.
func NewARM64Asm(cfg Config) Printer { return &arm64{cfg: cfg} }

// armReg maps an x86 physical GP register index to an arm64 register name.
// x86 indices: 0=AX 1=CX 2=DX 3=BX 4=SP 5=BP 6=SI 7=DI 8..15=R8..R15.
// RSP (4) is never allocated by avo. All targets are arm64 caller-saved
// registers (R0-R17, excluding R18), so the lowered leaf functions need no
// callee-save prologue.
var armReg = map[reg.Index]string{
	0: "R0", 1: "R1", 2: "R2", 3: "R3", 5: "R4",
	6: "R5", 7: "R6",
	8: "R7", 9: "R8", 10: "R9", 11: "R10", 12: "R11", 13: "R12", 14: "R13", 15: "R14",
}

const (
	// Scratch registers reserved for lowering; never produced by the x86 map.
	scratchAddr = "R15" // effective-address computation for indexed/RMW operands
	scratchVal  = "R16" // immediate materialization / RMW value

	// Scratch vector register for lowerings that need one (POPCNT). x86 has
	// only XMM0-15, which map to V0-V15, so V31 is never allocated.
	scratchVec = "V31"
)

func (p *arm64) Print(f *ir.File) ([]byte, error) {
	p.header(f)
	twins := collectTwins(f)
	for _, s := range f.Sections {
		switch s := s.(type) {
		case *ir.Function:
			p.function(s, twins)
		case *ir.Global:
			p.global(s)
		default:
			panic("unknown section type")
		}
	}
	return p.Result()
}

// logicalName strips the trailing _amd64/_bmi2 arch+feature suffixes from a
// generated function name, yielding the architecture-independent base. The _safe
// marker and everything before it are preserved, so a generic variant and its
// BMI2 twin reduce to the same base (e.g. both foo_safe_amd64 and foo_safe_bmi2
// yield foo_safe).
func logicalName(name string) string {
	for {
		switch {
		case strings.HasSuffix(name, "_bmi2"):
			name = strings.TrimSuffix(name, "_bmi2")
		case strings.HasSuffix(name, "_amd64"):
			name = strings.TrimSuffix(name, "_amd64")
		default:
			return name
		}
	}
}

// arm64Name is the symbol emitted for a lowered function. A name carrying an
// arch/feature suffix is renamed to <base>_arm64; a name with no such suffix is
// architecture-independent and reused verbatim (the arm64 file selects it via a
// build tag, as with the hand-written differential tests).
func arm64Name(name string) string {
	if !strings.Contains(name, "_amd64") && !strings.Contains(name, "bmi2") {
		return name
	}
	return logicalName(name) + "_arm64"
}

// twinPair records, for a logical function name, which of the generic/BMI2
// variants are present.
type twinPair struct {
	generic, bmi2 bool
}

// collectTwins maps each logical function name to which variants exist, so
// function() can tell a real generic/BMI2 pair (a choice to make, per
// Config.ARM64PreferBMI2) apart from a function with only one variant (lower
// it regardless -- there is no alternative).
func collectTwins(f *ir.File) map[string]twinPair {
	twins := make(map[string]twinPair)
	for _, s := range f.Sections {
		fn, ok := s.(*ir.Function)
		if !ok {
			continue
		}
		name := logicalName(fn.Name)
		t := twins[name]
		if strings.Contains(fn.Name, "bmi2") {
			t.bmi2 = true
		} else {
			t.generic = true
		}
		twins[name] = t
	}
	return twins
}

func (p *arm64) header(f *ir.File) {
	p.Comment(p.cfg.GeneratedWarning())
	p.Comment("EXPERIMENTAL arm64 output lowered from an amd64 avo program.")
	// This printer only ever emits arm64; require GOARCH=arm64 so the file does
	// not clash with the amd64 implementation of the same symbols.
	if len(f.Constraints) > 0 {
		constraints, err := buildtags.Format(f.Constraints)
		if err != nil {
			p.AddError(err)
		}
		// Parenthesized because && binds tighter than ||: rewriting
		// "//go:build !noasm || purego" without brackets yields
		// "arm64 && !noasm || purego", which parses as
		// "(arm64 && !noasm) || purego" and pulls the arm64 file into non-arm64
		// builds. The two agree only while the expression is pure AND.
		constraints = strings.Replace(constraints, "//go:build ", "//go:build arm64 && (", 1)
		constraints = strings.TrimRight(constraints, "\n") + ")\n"
		p.NL()
		p.Printf("%s", constraints)
	} else {
		p.NL()
		p.Printf("//go:build arm64\n")
	}
	if len(f.Includes) > 0 {
		p.NL()
		for _, path := range f.Includes {
			checkSingleLine("include path", path)
			p.Printf("#include \"%s\"\n", path)
		}
	}
}

func (p *arm64) global(g *ir.Global) {
	checkSingleLine("global symbol", g.Symbol.Name)
	p.NL()
	for _, d := range g.Data {
		a := operand.NewDataAddr(g.Symbol, d.Offset)
		p.Printf("DATA %s/%d, %s\n", a.Asm(), d.Value.Bytes(), d.Value.Asm())
	}
	p.Printf("GLOBL %s(SB), %s, $%d\n", g.Symbol, g.Attributes.Asm(), g.Size)
}

func (p *arm64) function(f *ir.Function, twins map[string]twinPair) {
	// Before anything else, and before the twin-skip return below: the flag
	// analyses read straight through comments on the premise that comments
	// carry no code, and a skipped twin is never walked at all. Both would
	// otherwise trust a guard that had not run.
	for _, n := range f.Nodes {
		if c, ok := n.(*ir.Comment); ok {
			checkNoDirective(c)
		}
	}
	// On arm64 we emit one implementation per logical function. Where the
	// generator produced both a generic ("_amd64") and a BMI2 ("_bmi2") variant,
	// Config.ARM64PreferBMI2 picks which one; the other is skipped. arm64 has
	// native equivalents for the BMI2 idioms, but they are not reliably faster
	// than lowering the generic path -- BMI2 x86 code is tuned for x86 (e.g.
	// BEXTR packs into one instruction what arm64 needs two UBFX to unpack), so
	// this is a measure-and-choose knob, not a default preference. A function
	// with only one variant (no twin to choose between) is always lowered.
	isBMI2 := strings.Contains(f.Name, "bmi2")
	if t := twins[logicalName(f.Name)]; t.generic && t.bmi2 && isBMI2 != p.cfg.ARM64PreferBMI2 {
		p.NL()
		label := "generic"
		if p.cfg.ARM64PreferBMI2 {
			label = "BMI2"
		}
		p.Comment(fmt.Sprintf("skipped %s (%s twin preferred on arm64)", f.Name, label))
		return
	}
	name := arm64Name(f.Name)

	p.NL()
	p.Comment(f.Stub())
	// The BMI2 requirement does not survive lowering (the ops become native arm64
	// instructions), so do not carry a misleading "Requires: BMI2" over.
	if len(f.ISA) > 0 && !isBMI2 {
		p.Comment("Requires: " + strings.Join(f.ISA, ", "))
	}
	checkSingleLine("function name", name)
	p.Printf("TEXT %s%s(SB)", dot, name)
	if f.Attributes != 0 {
		p.Printf(", %s", f.Attributes.Asm())
	}
	p.Printf(", %s\n", textsize(f))

	p.clear = true
	nodes := f.Nodes
	setflags := flagProducers(nodes)
	subwordSafe := subwordSafeEqNe(nodes)
	bt := btPairs(nodes)
	// The branch half of each fused pair emits nothing of its own: the TBNZ
	// stands in for both. Skipping it by index rather than by advancing past
	// it keeps any comments between the two in the output.
	fusedBranch := make(map[int]bool, len(bt))
	for _, f := range bt {
		fusedBranch[f.branch] = true
	}
	byteFold := shiftExtractFold(nodes)
	shifts, dropped := shiftFolds(nodes)
	setFull := setccFolds(nodes, setflags, dropped)
	p.constOK = false
	for idx := 0; idx < len(nodes); idx++ {
		if byteFold[idx] {
			p.emitShiftExtractFold(nodes[idx].(*ir.Instruction), nodes[idx+1].(*ir.Instruction), nodes[idx+2].(*ir.Instruction))
			idx += 2
			continue
		}
		switch n := nodes[idx].(type) {
		case ir.Label:
			// The only other text from the IR that reaches the file verbatim.
			// Not exploitable today -- a payload label emits a trailing colon
			// that both assemblers reject -- but it is the same shape as the
			// comment injection, and the check costs nothing.
			checkSingleLine("label", string(n))
			p.constOK = false
			p.flush()
			p.ensureclear()
			p.Printf("%s:\n", n)
		case *ir.Comment:
			p.flush()
			p.ensureclear()
			checkNoDirective(n)
			for _, line := range n.Lines {
				p.Printf("\t// %s\n", line)
			}
		case *ir.Instruction:
			if fusedBranch[idx] {
				continue
			}
			if dropped[idx] {
				// A register copy some later instruction absorbed (see
				// shiftFolds and setccFolds). Nothing is emitted; the constant
				// window closes as it would for any register move.
				p.constOK = false
				continue
			}
			if len(n.Suffixes) != 0 {
				// The dispatch keys on the opcode alone, so a suffix's meaning
				// (zeroing, broadcast, rounding) would simply be discarded. Only
				// EVEX forms carry them and all of those are unsupported anyway,
				// but that is an argument about avo's instruction database, not
				// something this file enforces.
				panic(fmt.Sprintf("arm64: %s carries suffixes %v, which this lowering ignores",
					n.Opcode, n.Suffixes))
			}
			p.inTransparent, p.transparentOp = isFlagTransparent(n.Opcode), n.Opcode
			p.inProducer, p.producerOp, p.writerCount = setflags[idx], n.Opcode, 0
			p.shift = shifts[idx]
			switch {
			case n.Opcode == "JMP":
				// Only a label target is translatable. A register or memory
				// operand is rendered in x86 syntax and passed through, and the
				// arm64 assembler accepts "JMP (R12)" as a branch through a
				// register: the memory load x86 would do is silently dropped, and
				// x86's R12 is a different physical register after renaming, so it
				// branches through the wrong one. Only bases outside R8-R15 fail
				// loudly, so this has to be rejected here.
				if _, ok := n.Operands[0].(operand.LabelRef); !ok {
					panic(fmt.Sprintf("arm64: JMP to a non-label target (%s) is not supported", n.Operands[0].Asm()))
				}
				p.emitf("JMP %s", n.Operands[0].Asm())
			case bt[idx].mnemonic != "":
				// BTL + adjacent carry branch, emitted as one test-and-branch.
				// TBNZ/TBZ have a 14-bit branch range where B.cond has 19;
				// Go's arm64 assembler rewrites an out-of-range one into an
				// inverted skip over an unconditional branch during its span
				// pass, so a long function needs no handling here. A non-Go
				// assembler would reject it outright rather than mis-branch.
				f := bt[idx]
				p.emitf("%s %s, %s, %s", f.mnemonic,
					n.Operands[0].Asm(),
					operandReg(n.Operands[1]),
					nodes[f.branch].(*ir.Instruction).Operands[0].Asm())
			case strings.HasPrefix(n.Opcode, "CMOV"):
				p.lowerCMOV(n)
			case strings.HasPrefix(n.Opcode, "SET"):
				p.lowerSET(n, setFull[idx])
			case isConditionalBranch(n):
				// Same reasoning as JMP: a relative target is a byte offset, and
				// a byte offset cannot mean the same thing on two instruction
				// sets. Both assemblers happen to reject the rendering, but that
				// is not something to depend on.
				if _, ok := n.Operands[0].(operand.LabelRef); !ok {
					panic(fmt.Sprintf("arm64: %s to a non-label target (%s) is not supported",
						n.Opcode, n.Operands[0].Asm()))
				}
				p.emitf("%s %s", branchMnemonic(n.Opcode), n.Operands[0].Asm())
			default:
				p.lower(n, setflags[idx], subwordSafe[idx])
			}
			if p.inProducer && p.writerCount != 1 {
				// The mirror of the transparency check. A producer the analysis
				// marked must actually supply the flags its consumer reads: none
				// leaves the branch running on whatever survived, which is the
				// worst failure available here, and more than one means a second
				// writer clobbered the first.
				panic(fmt.Sprintf("arm64: %s was marked as a flag producer but its lowering emitted "+
					"%d flag-writing instructions, not 1", p.producerOp, p.writerCount))
			}
			p.inTransparent, p.transparentOp = false, ""
			p.inProducer, p.producerOp, p.writerCount = false, "", 0
			p.shift = shiftFold{}
			p.trackConst(n)
			if n.IsTerminal || n.IsUnconditionalBranch() {
				p.flush()
			}
		default:
			panic("unexpected node type")
		}
	}
	p.flush()
}

// arm64WritesNZCV classifies every mnemonic this printer can emit as to whether
// it writes the condition flags. emitf() panics on a mnemonic missing from this
// table, so it cannot silently fall behind the lowerings.
//
// It exists to make one hand-maintained property mechanical. isFlagTransparent
// classifies by x86 OPCODE, but the property actually relied on is that the
// EMITTED arm64 sequence leaves NZCV alone -- including the instructions a
// lowering emits incidentally, like the scratch-address ADD behind an indexed
// operand. Those two were kept in sync by inspection, and the XORL bug (a
// 64-bit TST synthesized where a 32-bit one was needed) is what that costs.
//
// This checks only the arm64 half. The x86 half -- that the opcode really is
// flag-neutral on x86, or the backward scan attributes the wrong producer --
// cannot be checked from here, so isFlagTransparent's list is the thing to
// review against the Intel manual when adding to it.
var arm64WritesNZCV = map[string]bool{
	// Flag setters, emitted only from producer-classified lowerings.
	"ADDS": true, "ADDSW": true, "SUBS": true, "SUBSW": true,
	"ANDS": true, "ANDSW": true,
	"CMP": true, "CMPW": true, "CMN": true, "CMNW": true,
	"TST": true, "TSTW": true,

	// Data movement and arithmetic: no flag effects.
	"ADD": false, "ADDW": false, "SUB": false, "SUBW": false,
	"AND": false, "ANDW": false, "ORR": false, "ORRW": false,
	"EOR": false, "EORW": false,
	"MUL": false, "MULW": false, "UMULH": false, "SMULH": false,
	"LSL": false, "LSLW": false, "LSR": false, "LSRW": false,
	"ASR": false, "ASRW": false, "ROR": false, "RORW": false,
	"MVN": false, "MVNW": false, "NEG": false, "NEGW": false,
	"BIC":  false, // BICS is the flag-setting form; this is not it
	"MOVD": false, "MOVW": false, "MOVWU": false, "MOVH": false,
	"MOVHU": false, "MOVB": false, "MOVBU": false,
	"PRFM":  false, // a hint: no register, memory or flag effects
	"FMOVD": false, "FMOVQ": false, "VMOV": false, "VEOR": false,
	"VCNT": false, "VUADDLV": false,
	// Test-and-branch: reads one bit of a register directly and writes no
	// flags at all. See btPairs for why that is sound here.
	"TBNZ": false, "TBZ": false,
	"BFI": false, "UBFX": false, "SBFX": false,
	"RBIT": false, "CLZ": false, "REVW": false,

	// Conditional select/set read the flags and never write them.
	"CSEL": false, "CSELW": false, "CSET": false, "CSINC": false,

	// Control flow. The conditional branches read NZCV; none writes it.
	"RET": false, "JMP": false,
	"BEQ": false, "BNE": false, "BLT": false, "BLE": false,
	"BGT": false, "BGE": false, "BLO": false, "BHS": false,
	"BHI": false, "BLS": false, "BMI": false, "BPL": false,
	"BVS": false, "BVC": false,
}

// checkSingleLine refuses text that would span more than one output line.
//
// Everything downstream assumes one emitted item becomes one line: the
// alignment pass, and every analysis that reasons about instruction positions.
// A newline reaching the file is also how text becomes live code -- see
// checkNoDirective, where a comment carrying one injected an instruction into
// both outputs.
func checkSingleLine(what, text string) {
	if strings.ContainsAny(text, "\n\r") {
		panic(fmt.Sprintf("arm64: %s %q spans more than one line", what, text))
	}
}

// checkNoDirective refuses a preprocessor directive in the instruction stream.
//
// avo emits directives as comments, inert until a postprocessing step rewrites
// "\t// #" back to "#". This printer used to evaluate the GOAMD64 conditionals
// among them and treat the rest as analysis boundaries. That machinery produced
// roughly a third of the miscompiles found in review -- every one an
// external-coupling bug, where this printer's model of when a directive is live
// disagreed with the assembler's, the rewrite's, or a header's -- and it served
// no generator that uses this lowering: zstd and huff0 emit no directives at
// all, and s2, which does, is amd64-only.
//
// Refusing them outright removes that entire class. BMI2 and other GOAMD64
// variants are still supported through twin FUNCTIONS, which is what the
// generators here actually use; only inline conditional arms are gone. The
// evaluator, with every fix review found for it, is preserved on the
// lizf.arm64-goamd64-directives branch if it is ever wanted back.
//
// The file-level #include lines (f.Includes, e.g. textflag.h) are not policed
// here: they are emitted by the printer itself, identically on both sides, and
// never come from the instruction stream.
func checkNoDirective(c *ir.Comment) {
	for _, line := range c.Lines {
		// Embedded newlines first, and unconditionally. The emitter writes
		// "\t// %s\n", so a comment line containing a newline produces a SECOND
		// physical line with no comment prefix at all -- live text at column 0,
		// in both outputs. It does not have to look like a directive to do
		// damage: an instruction there is assembled by both sides, and since the
		// two architectures rename registers differently, the same text reads
		// and writes different values on each. It is also invisible to every
		// analysis here, all of which assume a comment carries no code.
		//
		// Checking what follows the newline was the wrong fix, twice over. The
		// invariant is simply that one comment line becomes one prefixed output
		// line, so anything that breaks it is refused.
		if strings.ContainsAny(line, "\n\r") {
			panic(fmt.Sprintf("arm64: comment line %q contains a newline; "+
				"only the first physical line would be commented, leaving the rest live", line))
		}
		if d := strings.TrimSpace(line); strings.HasPrefix(d, "#") {
			panic(fmt.Sprintf("arm64: comment line %q would be a preprocessor directive; "+
				"use twin functions for GOAMD64 variants, and avoid a leading '#' in prose", d))
		}
	}
}

// zeroSelf lowers x86's "XOR r, r" zeroing idiom. The register is cleared with
// a move, but x86 also sets ZF here, so when a consumer reads those flags the
// move must be followed by a test -- the arithmetic lowering that would
// otherwise have supplied them is skipped by this shortcut.
func (p *arm64) zeroSelf(r, test string, flags bool) {
	p.emitf("MOVD $0, %s", r)
	if flags {
		p.emitf("%s %s, %s", test, r, r)
	}
}

// emit buffers a single lowered arm64 instruction. The block is column-aligned
// and written by flush(), matching the goasm printer's layout (and asmfmt).
func (p *arm64) emitf(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	// One emit, one output line. Everything downstream -- the alignment pass,
	// the analyses that reason about instruction positions -- assumes that, and
	// a newline reaching the file un-prefixed is how text becomes live code.
	// Asserting it here closes the class rather than each site that could
	// reintroduce it.
	checkSingleLine("emitted text", line)
	op, operands := line, ""
	if i := strings.IndexByte(line, ' '); i >= 0 {
		op, operands = line[:i], line[i+1:]
	}
	writes, known := arm64WritesNZCV[op]
	if !known {
		panic(fmt.Sprintf("arm64: emitted %q is not classified in arm64WritesNZCV; "+
			"add it, deciding whether it writes the condition flags", op))
	}
	if writes && p.inTransparent {
		panic(fmt.Sprintf("arm64: lowering %s emitted %s, which writes the condition flags, "+
			"but isFlagTransparent classifies it as leaving them alone", p.transparentOp, op))
	}
	if writes && p.inProducer {
		p.writerCount++
	}
	p.pending = append(p.pending, [2]string{op, operands})
	p.clear = false
}

// flush writes the buffered instructions with operands aligned to a common
// column (width of the widest opcode in the block), like the goasm printer.
func (p *arm64) flush() {
	if len(p.pending) == 0 {
		return
	}
	width := 0
	for _, in := range p.pending {
		if in[1] != "" && len(in[0]) > width {
			width = len(in[0])
		}
	}
	for _, in := range p.pending {
		if in[1] != "" {
			p.Printf("\t%-*s%s\n", width+1, in[0], in[1])
		} else {
			p.Printf("\t%s\n", in[0])
		}
	}
	p.pending = nil
}

// ensureclear emits a blank separator line unless one is already present.
func (p *arm64) ensureclear() {
	if !p.clear {
		p.NL()
		p.clear = true
	}
}

// rename returns the arm64 syntax for a register operand.
func rename(r reg.Register) string {
	switch r.Kind() {
	case reg.KindGP:
		if ph, ok := r.(reg.Physical); ok {
			if name, ok := armReg[ph.PhysicalIndex()]; ok {
				return name
			}
			panic(fmt.Sprintf("arm64: unmapped x86 GP register index %d", ph.PhysicalIndex()))
		}
		panic("arm64: non-physical register reached printer (run pass.Compile first)")
	case reg.KindVector:
		if ph, ok := r.(reg.Physical); ok {
			return fmt.Sprintf("V%d", ph.PhysicalIndex())
		}
		panic("arm64: non-physical vector register reached printer")
	case reg.KindPseudo, reg.KindOpmask:
		// Falls through to the generic Asm()-based renaming below.
	}
	// Pseudo registers. FP/SB/PC share their tokens with arm64; the amd64
	// hardware stack pointer "SP" becomes the arm64 hardware stack pointer
	// "RSP" (bare "SP" is the pseudo frame-relative register in Go arm64 asm).
	if a := r.Asm(); a == "SP" {
		return "RSP"
	}
	return r.Asm()
}

// isHighByte reports whether op is an x86 high-byte register (AH/BH/CH/DH),
// which has no arm64 equivalent and must be read with a bitfield extract.
func isHighByte(op operand.Op) bool {
	r, ok := op.(reg.Register)
	if !ok {
		return false
	}
	switch r.Asm() {
	case "AH", "BH", "CH", "DH":
		return true
	}
	return false
}

func operandReg(op operand.Op) string {
	r, ok := op.(reg.Register)
	if !ok {
		panic(fmt.Sprintf("arm64: expected register operand, got %T", op))
	}
	return rename(r)
}

func immAsm(op operand.Op) (string, bool) {
	if c, ok := op.(operand.Constant); ok {
		return c.Asm(), true
	}
	return "", false
}

// trackConst maintains the one-instruction constant window: after a plain
// "MOVQ/MOVL $imm, reg" the register's value is known while lowering the next
// instruction. Every other instruction closes the window (labels do too, in
// the dispatch loop). MOVL immediates are <=32-bit and x86 zero-extends the
// destination, so the tracked 64-bit value is the same for both widths.
func (p *arm64) trackConst(n *ir.Instruction) {
	p.constOK = false
	if n.Opcode != "MOVQ" && n.Opcode != "MOVL" || len(n.Operands) != 2 {
		return
	}
	v, ok := immVal(n.Operands[0])
	if !ok {
		return
	}
	if n.Opcode == "MOVL" {
		v = int64(uint32(v)) // the register holds the zero-extended 32-bit value
	}
	r, ok := n.Operands[1].(reg.Register)
	if !ok {
		return
	}
	p.constReg = rename(r)
	p.constVal = v
	p.constOK = true
}

// knownConst reports the constant held by register operand op, if the
// immediately preceding instruction loaded it with an immediate.
func (p *arm64) knownConst(op operand.Op) (int64, bool) {
	r, ok := op.(reg.Register)
	if !ok || !p.constOK || rename(r) != p.constReg {
		return 0, false
	}
	return p.constVal, true
}

// immAsmQ renders an immediate destined for a 64-bit operand slot.
//
// x86-64 has no 64-bit immediate for the ALU, compare, test and store forms: it
// encodes imm32 and the CPU SIGN-EXTENDS it. So "ANDQ $0x80000000, AX" does not
// mask with 0x80000000, it masks with 0xffffffff80000000 -- and Go's assembler
// accepts the unsigned spelling without complaint, quietly encoding the
// sign-extended value. arm64 reads the same text literally, so one instruction
// computes two different things. Rendering the sign-extended value explicitly
// makes them agree.
//
// This is deliberately not folded into immAsm or regOrImm. Those are shared
// with the 32-bit lowerings, where imm32 is the whole operand and nothing is
// extended, and with MOVQ-to-register, which the assembler narrows to a
// zero-extending MOVL; in both, the literal reading is the correct one.
func immAsmQ(op operand.Op) (string, bool) {
	s, ok := immAsm(op)
	if !ok {
		return "", false
	}
	v, ok := immVal(op)
	if !ok {
		// Unparsable means the sign-extension question cannot be answered, and
		// passing the text through would be a guess about which way it goes.
		panic(fmt.Sprintf("arm64: cannot interpret 64-bit immediate %s", s))
	}
	if v >= 1<<31 && v < 1<<32 {
		return fmt.Sprintf("$%d", int64(int32(uint32(v)))), true
	}
	return s, true
}

// regOrImmQ is regOrImm for a 64-bit operand slot.
func (p *arm64) regOrImmQ(op operand.Op) string {
	if imm, ok := immAsmQ(op); ok {
		return imm
	}
	return operandReg(op)
}

func immVal(op operand.Op) (int64, bool) {
	c, ok := immAsm(op)
	if !ok {
		return 0, false
	}
	t := strings.TrimPrefix(c, "$")
	if v, err := strconv.ParseInt(t, 0, 64); err == nil {
		return v, true
	}
	// A 64-bit pattern with the top bit set is legal written unsigned
	// ("$0xffffffffffffffff"), which overflows a signed parse. amd64 accepts it
	// and wraps to the same bit pattern, so interpret it the same way rather
	// than refusing a program the other architecture assembles.
	if u, err := strconv.ParseUint(t, 0, 64); err == nil {
		return int64(u), true
	}
	return 0, false
}

func log2scale(s uint8) int {
	switch s {
	case 1:
		return 0
	case 2:
		return 1
	case 4:
		return 2
	case 8:
		return 3
	}
	panic(fmt.Sprintf("arm64: bad scale %d", s))
}

// memAsm lowers a memory operand to a simple base+disp arm64 operand string,
// emitting an ADD into scratchAddr first for indexed operands.
func (p *arm64) memAsm(m operand.Mem) string { return p.memAsmW(m, 0) }

// pseudoSPLocalOffset is where a function's locals begin, relative to the arm64
// hardware stack pointer.
//
// The two architectures store the return address in different places. x86's
// CALL pushes it before the callee runs, so it sits above the frame and local 0
// is at SP+0. arm64's BL leaves it in the link register and the callee spills it
// to the BOTTOM of its own frame, so 0(RSP) is the saved link register and
// locals begin at 8(RSP). avo's AllocLocal hands out x86-relative
// displacements, so they need shifting.
const pseudoSPLocalOffset = 8

// frameAdjust rewrites a pseudo-SP-based memory operand for arm64's frame
// layout. Without it every local lands 8 bytes low and local 0 lands on the
// saved link register. Nothing computes a wrong value -- all locals shift
// uniformly, so loads and stores still agree -- but any stack unwind (GC
// scanning, panic, profiling) then reads a local as the return address.
func frameAdjust(m operand.Mem) operand.Mem {
	if m.Base != nil && m.Base.Asm() == "SP" {
		m.Disp += pseudoSPLocalOffset
	}
	return m
}

// memAsmW renders a memory operand for a scalar access of the given width in
// bytes (0 when the caller cannot use arm64's folded base+index form).
func (p *arm64) memAsmW(m operand.Mem, width int) string {
	m = frameAdjust(m)
	if m.Symbol.Name != "" {
		s := m.Symbol.String() + fmt.Sprintf("%+d", m.Disp)
		if m.Base != nil {
			s += fmt.Sprintf("(%s)", rename(m.Base))
		}
		if m.Index != nil {
			panic("arm64: indexed symbol operand not supported")
		}
		return s
	}
	if m.Index == nil || m.Scale == 0 {
		if m.Disp != 0 {
			return fmt.Sprintf("%d(%s)", m.Disp, rename(m.Base))
		}
		return fmt.Sprintf("(%s)", rename(m.Base))
	}
	sh := log2scale(m.Scale)
	base, index := rename(m.Base), rename(m.Index)
	// arm64 scalar loads/stores can fold base+index addressing into the
	// instruction, saving the separate effective-address ADD that x86's richer
	// addressing otherwise costs us on every access. The architecture allows it
	// only when the index is unscaled or scaled by exactly the access width, and
	// never alongside a displacement, so fall through to scratchAddr otherwise.
	// width == 0 means the caller has no folded form available (FMOVQ), so it
	// must always get a plain base(+disp) operand.
	if width > 0 && m.Disp == 0 && (m.Scale == 1 || int(m.Scale) == width) {
		if sh == 0 {
			return fmt.Sprintf("(%s)(%s)", base, index)
		}
		return fmt.Sprintf("(%s)(%s<<%d)", base, index, sh)
	}

	// Otherwise the effective address is computed into scratchAddr, fresh for
	// each access. An earlier version cached it across instructions so a second
	// access to the same base and index could reuse it. That saved two
	// instructions across the whole of zstd and huff0 -- most accesses fold into
	// the instruction and never materialize an address at all -- and cost two
	// silent miscompiles, because the cache had to be invalidated at every
	// boundary this printer does not otherwise model.
	if sh == 0 {
		p.emitf("ADD %s, %s, %s", index, base, scratchAddr)
	} else {
		p.emitf("ADD %s<<%d, %s, %s", index, sh, base, scratchAddr)
	}
	if m.Disp != 0 {
		return fmt.Sprintf("%d(%s)", m.Disp, scratchAddr)
	}
	return fmt.Sprintf("(%s)", scratchAddr)
}

// prefetchHints maps x86's temporal-locality prefetch hints to PRFM's. The
// x86 hint names the nearest cache level the line should land in (T0 = all
// levels, T1 = L2 and outward, T2 = L3 and outward, NTA = non-temporal);
// arm64's PLDL<n>KEEP and PLDL1STRM say the same thing.
var prefetchHints = map[string]string{
	"PREFETCHT0":  "PLDL1KEEP",
	"PREFETCHT1":  "PLDL2KEEP",
	"PREFETCHT2":  "PLDL3KEEP",
	"PREFETCHNTA": "PLDL1STRM",
}

// lowerPrefetch emits PRFM for a software prefetch hint. A prefetch never
// faults, never writes a register and never touches NZCV, so nothing about
// the address it names is a correctness hazard; the only constraint is what
// Go's arm64 assembler accepts as the operand. Probed with Go 1.27.1 at the
// boundaries: a base register with a displacement of 0 through 255 (the
// unscaled 9-bit form) or a multiple of 8 from 256 through 32760 (the
// 8-byte-scaled 12-bit form) assembles; 257, 260, 4095, 32761, any negative
// displacement and any index register are rejected. Anything outside the
// folded form goes through scratchAddr the way memAsmW does for an indexed
// load, at the cost of one ADD -- generator-side code that cares should keep
// prefetch operands to base or base+small-aligned-disp.
func (p *arm64) lowerPrefetch(op string, src operand.Op) {
	m, ok := src.(operand.Mem)
	if !ok {
		panic(fmt.Sprintf("arm64: %s expects a memory operand, got %s", op, src.Asm()))
	}
	m = frameAdjust(m)
	if m.Symbol.Name != "" {
		panic(fmt.Sprintf("arm64: %s of a symbol operand (%s) is not supported", op, src.Asm()))
	}
	base := rename(m.Base)
	if m.Index != nil && m.Scale != 0 {
		if sh := log2scale(m.Scale); sh == 0 {
			p.emitf("ADD %s, %s, %s", rename(m.Index), base, scratchAddr)
		} else {
			p.emitf("ADD %s<<%d, %s, %s", rename(m.Index), sh, base, scratchAddr)
		}
		base = scratchAddr
	}
	if m.Disp < 0 || m.Disp > 32760 || (m.Disp%8 != 0 && m.Disp >= 256) {
		p.emitf("ADD $%d, %s, %s", m.Disp, base, scratchAddr)
		base, m.Disp = scratchAddr, 0
	}
	if m.Disp != 0 {
		p.emitf("PRFM %d(%s), %s", m.Disp, base, prefetchHints[op])
	} else {
		p.emitf("PRFM (%s), %s", base, prefetchHints[op])
	}
}

func (p *arm64) lower(i *ir.Instruction, flags, subwordEqNeSafe bool) {
	ops := i.Operands
	switch i.Opcode {
	case "RET":
		p.emitf("RET")

	// ---- moves and loads ----
	case "MOVQ":
		p.lowerMove("MOVD", ops[0], ops[1])
	case "PREFETCHT0", "PREFETCHT1", "PREFETCHT2", "PREFETCHNTA":
		p.lowerPrefetch(i.Opcode, ops[0])
	case "MOVL":
		p.lowerMOVL(ops[0], ops[1])
	case "MOVW":
		p.lowerMOVW(ops[0], ops[1])
	case "MOVB":
		p.lowerMOVB(ops[0], ops[1])
	case "MOVWQSX": // load/extend int16, sign-extend (mem or reg source)
		p.emitf("MOVH %s, %s", p.srcAsmW(ops[0], 2), operandReg(ops[1]))
	case "MOVWQZX", "MOVWLZX": // load/extend uint16, zero-extend
		// The two differ only in named destination width, not in result: x86
		// zeroes bits 63:32 on any 32-bit destination write, so the L form
		// leaves the same fully zero-extended register the Q form does, and
		// MOVHU zero-extends to the whole register either way. This is the
		// same reasoning that lets MOVBQZX and MOVBLZX share a lowering below.
		// No high-byte case exists at this width, so none is checked.
		p.emitf("MOVHU %s, %s", p.srcAsmW(ops[0], 2), operandReg(ops[1]))
	case "MOVBQZX", "MOVBQSX", "MOVBLSX", "MOVBLZX":
		// A high-byte source (AH/BH/CH/DH) names bits 15:8 but renames to the
		// same arm64 register as the low byte, so a plain load would read 7:0.
		// Extracting the right field is only correct where x86 can express the
		// same thing, which is narrower than it looks: any encoding carrying a
		// REX prefix redefines that register slot as SPL, so the amd64 side
		// silently reads the stack pointer's low byte instead. Refuse those
		// rather than have the two architectures compute different values from
		// one avo program.
		if isHighByte(ops[0]) {
			if i.Opcode == "MOVBQZX" || i.Opcode == "MOVBQSX" {
				panic(fmt.Sprintf("arm64: %s from a high-byte register cannot be expressed on x86-64: "+
					"the 64-bit destination forces a REX prefix, which renames that operand to SPL", i.Opcode))
			}
			checkHighByteEncodable(ops[0], ops[1])
			ext := "UBFX"
			if i.Opcode == "MOVBLSX" {
				ext = "SBFX"
			}
			d := operandReg(ops[1])
			p.emitf("%s $8, %s, $8, %s", ext, rename(ops[0].(reg.Register)), d)
			if i.Opcode == "MOVBLSX" {
				p.emitf("MOVWU %s, %s", d, d) // sign-extended: clear the upper half
			}
			return
		}
		switch i.Opcode {
		case "MOVBQZX", "MOVBLZX":
			// Zero-extending a byte already leaves the upper bits clear, so the
			// 32-bit form needs no extra truncation.
			p.emitf("MOVBU %s, %s", p.srcAsmW(ops[0], 1), operandReg(ops[1]))
		case "MOVBQSX":
			p.emitf("MOVB %s, %s", p.srcAsmW(ops[0], 1), operandReg(ops[1]))
		default: // MOVBLSX
			d := operandReg(ops[1])
			p.emitf("MOVB %s, %s", p.srcAsmW(ops[0], 1), d)
			p.emitf("MOVWU %s, %s", d, d)
		}

	case "MOVLQSX": // load/extend int32, sign-extend
		p.emitf("MOVW %s, %s", p.srcAsmW(ops[0], 4), operandReg(ops[1]))
	case "MOVLQZX": // load/extend uint32, zero-extend
		p.emitf("MOVWU %s, %s", p.srcAsmW(ops[0], 4), operandReg(ops[1]))
	case "MOVWLSX":
		// Sign-extend a halfword into a 32-bit destination: x86 leaves the upper
		// 32 bits zeroed, so extend to 64 and then clear the top half.
		d := operandReg(ops[1])
		p.emitf("MOVH %s, %s", p.srcAsmW(ops[0], 2), d)
		p.emitf("MOVWU %s, %s", d, d)
	case "MOVUPS", "MOVOU", "MOVOA":
		// 128-bit SSE moves. arm64 NEON loads/stores have no alignment
		// requirement, so the aligned (MOVOA) and unaligned (MOVOU) forms lower
		// identically.
		p.lowerMOVUPS(ops[0], ops[1])
	case "PXOR":
		// 128-bit bitwise XOR of two vector registers.
		a, b := operandReg(ops[0]), operandReg(ops[1])
		p.emitf("VEOR %s.B16, %s.B16, %s.B16", a, b, b)

	// ---- arithmetic / logic (dst is last operand) ----
	case "ADDQ":
		p.lowerArith("ADD", "ADDS", ops[0], ops[1], flags)
	case "SUBQ":
		p.lowerArith("SUB", "SUBS", ops[0], ops[1], flags)
	case "ANDQ":
		p.lowerArith("AND", "ANDS", ops[0], ops[1], flags)
	case "ORQ":
		p.lowerArith("ORR", "", ops[0], ops[1], flags)

	// 32-bit ALU. x86 zero-extends a 32-bit register destination to 64 bits,
	// which the arm64 W-forms do as well, so these map directly.
	case "ADDL":
		p.lowerArithW("ADDW", "ADDSW", ops[0], ops[1], flags)
	case "SUBL":
		p.lowerArithW("SUBW", "SUBSW", ops[0], ops[1], flags)
	case "ANDL":
		p.lowerArithW("ANDW", "ANDSW", ops[0], ops[1], flags)
	case "ORL":
		p.lowerArithW("ORRW", "", ops[0], ops[1], flags)
	case "XORQ":
		if ra, ok := ops[0].(reg.Register); ok {
			if rb, ok := ops[1].(reg.Register); ok && rename(ra) == rename(rb) {
				p.zeroSelf(rename(rb), "TST", flags)
				return
			}
		}
		p.lowerArith("EOR", "", ops[0], ops[1], flags)
	case "XORL":
		if ra, ok := ops[0].(reg.Register); ok {
			if rb, ok := ops[1].(reg.Register); ok && rename(ra) == rename(rb) {
				p.zeroSelf(rename(rb), "TSTW", flags)
				return
			}
		}
		// Note the W: XORL is a 32-bit operation, so it needs the 32-bit lowering
		// like its ADDL/SUBL/ANDL/ORL neighbours. Going through the 64-bit path
		// synthesized a 64-bit TST, whose N bit reads bit 63 -- always zero after
		// a zero-extending EORW -- so a following JS/JLT/JGT/JLE tested the wrong
		// bit, and a memory destination got an 8-byte read-modify-write.
		p.lowerArithW("EORW", "", ops[0], ops[1], flags)
	case "ADDB":
		checkHighByteEncodable(ops[0], ops[1])
		checkHighByteEncodable(ops[1], ops[0])
		// x86 ADDB replaces only the destination's addressed byte with the byte
		// sum (wrapping mod 256). Compute the sum in scratch — addition carries
		// travel upward only, so garbage above bit 7 of either input cannot
		// affect bits 7:0 — and insert exactly those 8 bits.
		if isHighByte(ops[1]) {
			v := p.byteVal(ops[0])
			d := operandReg(ops[1])
			p.emitf("UBFX $8, %s, $8, %s", d, scratchAddr)
			p.emitf("ADD %s, %s, %s", v, scratchAddr, scratchAddr)
			p.emitf("BFI $8, %s, $8, %s", scratchAddr, d)
			return
		}
		v := p.byteVal(ops[0])
		d := operandReg(ops[1])
		p.emitf("ADD %s, %s, %s", v, d, scratchAddr)
		p.emitf("BFI $0, %s, $8, %s", scratchAddr, d)
	case "ADCB":
		// Narrow lowering of the carry-accumulate idiom "CMPQ x, y; ADCB $0, dst":
		// dst's low byte += x86 CF, where CF after a compare is the unsigned
		// borrow (x < y). arm64 inverts carry for subtraction, so borrow is the
		// LO condition and no-borrow is HS: CSINC yields dst on HS and dst+1
		// otherwise; only bits 7:0 of the result are inserted, as on x86.
		// Assumes NZCV comes from a compare (flagProducers verifies the producer).
		if v, ok := immVal(ops[0]); !ok || v != 0 {
			panic("arm64: ADCB only supported with a $0 immediate source")
		}
		if isHighByte(ops[1]) {
			panic("arm64: ADCB high-byte destination not supported")
		}
		d := operandReg(ops[1])
		p.emitf("CSINC HS, %s, %s, %s", d, d, scratchVal)
		p.emitf("BFI $0, %s, $8, %s", scratchVal, d)
	case "ADCQ":
		// The full-width form of the same idiom: dst += CF over all 64 bits,
		// so the CSINC writes the destination directly and nothing needs
		// inserting. Same producer assumption as ADCB (flagProducers checks).
		if v, ok := immVal(ops[0]); !ok || v != 0 {
			panic("arm64: ADCQ only supported with a $0 immediate source")
		}
		d := operandReg(ops[1])
		p.emitf("CSINC HS, %s, %s, %s", d, d, d)
	case "BSWAPL":
		// Byte-reverse the low 32 bits; the 32-bit result zero-extends, as on x86.
		r := operandReg(ops[0])
		p.emitf("REVW %s, %s", r, r)
	case "INCQ":
		p.lowerIncDec("ADD", "ADDS", ops[0], 8, flags)
	case "INCL":
		p.lowerIncDec("ADDW", "ADDSW", ops[0], 4, flags)
	case "DECQ":
		p.lowerIncDec("SUB", "SUBS", ops[0], 8, flags)
	case "DECL":
		p.lowerIncDec("SUBW", "SUBSW", ops[0], 4, flags)
	case "NEGQ":
		p.emitf("NEG %s, %s", operandReg(ops[0]), operandReg(ops[0]))
	case "NEGL":
		p.emitf("NEGW %s, %s", operandReg(ops[0]), operandReg(ops[0]))
	case "NOTQ":
		p.emitf("MVN %s, %s", operandReg(ops[0]), operandReg(ops[0]))
	case "NOTL":
		p.emitf("MVNW %s, %s", operandReg(ops[0]), operandReg(ops[0]))
	case "SHRQ":
		checkNotDoubleShift(i)
		p.lowerShift("LSR", ops[0], ops[1], 64)
	case "SHLQ":
		checkNotDoubleShift(i)
		p.lowerShift("LSL", ops[0], ops[1], 64)
	case "SARQ":
		p.lowerShift("ASR", ops[0], ops[1], 64)
	case "SHRL":
		checkNotDoubleShift(i)
		p.lowerShift("LSRW", ops[0], ops[1], 32)
	case "SHLL":
		checkNotDoubleShift(i)
		p.lowerShift("LSLW", ops[0], ops[1], 32)
	case "SARL":
		p.lowerShift("ASRW", ops[0], ops[1], 32)
	case "SHLB":
		// x86 SHLB shifts only the destination's low byte and leaves the rest of
		// the register untouched. Shift in scratch and insert bits 7:0, matching
		// the partial-register semantics (see lowerMOVB).
		if isHighByte(ops[1]) {
			panic("arm64: SHLB high-byte destination not supported")
		}
		d := operandReg(ops[1])
		p.emitf("LSLW %s, %s, %s", p.regOrImm(ops[0]), d, scratchVal)
		p.emitf("BFI $0, %s, $8, %s", scratchVal, d)
	case "ROLQ":
		p.lowerROL(ops[0], ops[1], 64)
	case "ROLL":
		p.lowerROL(ops[0], ops[1], 32)
	case "RORQ":
		p.lowerShift("ROR", ops[0], ops[1], 64)
	case "RORL":
		p.lowerShift("RORW", ops[0], ops[1], 32)

	// ---- BMI2 flag-free shifts/rotate: SHIFTX count, src, dst ----
	// arm64 register shifts are already flag-free and take an arbitrary count
	// register, so these map one-to-one (the count is masked mod 64, matching x86).
	case "SHLXQ":
		p.lowerShiftX("LSL", ops[0], ops[1], ops[2])
	case "SHRXQ":
		p.lowerShiftX("LSR", ops[0], ops[1], ops[2])
	case "SARXQ":
		p.lowerShiftX("ASR", ops[0], ops[1], ops[2])
	case "RORXQ":
		p.lowerRORX(ops[0], ops[1], ops[2])

	case "LEAQ":
		p.lowerLEA(ops[0].(operand.Mem), operandReg(ops[1]))
	case "LEAL":
		// 32-bit LEA: the address arithmetic is computed modulo 2^32 and the
		// result zero-extends into the 64-bit destination.
		d := operandReg(ops[1])
		p.lowerLEA(ops[0].(operand.Mem), d)
		p.emitf("MOVWU %s, %s", d, d)

	case "IMULL":
		// 32-bit two-operand multiply; the W-form zeroes the upper half as x86 does.
		p.emitf("MULW %s, %s, %s", p.valRegW(ops[0], 4), operandReg(ops[1]), operandReg(ops[1]))

	case "POPCNTQ":
		// arm64 has no scalar population count: move to a vector register, count
		// bits per byte, then sum the bytes across the vector.
		src := p.valReg(ops[0])
		dst := operandReg(ops[1])
		f := "F" + scratchVec[1:]
		p.emitf("FMOVD %s, %s", src, f)
		p.emitf("VCNT %s.B8, %s.B8", scratchVec, scratchVec)
		p.emitf("VUADDLV %s.B8, %s", scratchVec, scratchVec)
		p.emitf("FMOVD %s, %s", f, dst)

	case "XCHGQ":
		// Register swap. x86's memory form is implicitly atomic, which this
		// lowering would not reproduce, so it is not accepted.
		if _, ok := ops[0].(operand.Mem); ok {
			panic("arm64: XCHGQ with a memory operand is atomic on x86 and is not supported")
		}
		if _, ok := ops[1].(operand.Mem); ok {
			panic("arm64: XCHGQ with a memory operand is atomic on x86 and is not supported")
		}
		a, b := operandReg(ops[0]), operandReg(ops[1])
		if a != b {
			p.emitf("MOVD %s, %s", a, scratchVal)
			p.emitf("MOVD %s, %s", b, a)
			p.emitf("MOVD %s, %s", scratchVal, b)
		}

	case "BTRQ", "BTCQ":
		// Clear/complement bit n. x86 also reports the previous bit in CF; no
		// consumer can read that here because flagSetter rejects these opcodes.
		op := "BIC"
		if i.Opcode == "BTCQ" {
			op = "EOR"
		}
		a := operandReg(ops[0])
		b := operandReg(ops[1])
		p.emitf("MOVD $1, %s", scratchVal)
		p.emitf("LSL %s, %s, %s", a, scratchVal, scratchVal)
		p.emitf("%s %s, %s, %s", op, scratchVal, b, b)

	case "BTSQ":
		a := operandReg(ops[0])
		b := operandReg(ops[1])
		p.emitf("MOVD $1, %s", scratchVal)
		p.emitf("LSL %s, %s, %s", a, scratchVal, scratchVal)
		p.emitf("ORR %s, %s, %s", scratchVal, b, b)

	case "BSFQ", "TZCNTQ":
		// Index of the lowest set bit == count of trailing zeros, which arm64
		// spells as reverse-then-count-leading-zeros. For a zero input this
		// yields 64, matching TZCNT exactly; x86 BSF leaves the destination
		// undefined there, so returning 64 is a valid refinement. The generators
		// emit both mnemonics guarded by #ifdef GOAMD64_v3, and both lower here
		// to the same sequence.
		src := operandReg(ops[0])
		dst := operandReg(ops[1])
		p.emitf("RBIT %s, %s", src, dst)
		p.emitf("CLZ %s, %s", dst, dst)

	case "BSRQ":
		// Index of the highest set bit. x86 leaves the destination architecturally
		// undefined for a zero input (real hardware leaves it unchanged); this
		// yields -1, since CLZ gives 64 and the subtraction runs past zero. That
		// is a refinement of undefined, like the BSF case above -- but it means a
		// differential test must not feed BSR a zero input and expect agreement.
		src := operandReg(ops[0])
		dst := operandReg(ops[1])
		p.emitf("CLZ %s, %s", src, scratchVal)
		p.emitf("MOVD $63, %s", dst)
		p.emitf("SUB %s, %s, %s", scratchVal, dst, dst)

	// ---- BMI2 bit-field ops: BZHI/BEXTR (counts assumed < 64, as in zstd) ----
	case "BZHIQ":
		p.lowerBZHI(ops[0], ops[1], ops[2])
	case "BEXTRQ":
		p.lowerBEXTR(ops[0], ops[1], ops[2])

	// ---- multiplication ----
	case "MULXQ":
		p.lowerMULX(ops[0], ops[1], ops[2])
	case "IMUL3Q":
		p.lowerIMUL3(ops[0], ops[1], ops[2])
	case "IMULQ":
		p.lowerIMUL(ops)
	case "MULQ":
		p.lowerWideMul("UMULH", ops[0])

	// ---- comparisons: always emit an arm64 flag-setter ----
	// The compare must run at the operand width: a 64-bit CMP of registers whose
	// upper bits are not provably zero sets flags from the wrong bits. arm64 has
	// a native 32-bit form (CMPW/TSTW); sub-32-bit widths would generally need the
	// operands extended for the consuming condition's signedness, which is not
	// modelled. The one exception is when every consumer is EQ/NE (checked by
	// subwordEqNeSafe, from subwordSafeEqNe): equality doesn't depend on sign, so
	// zero-extending both operands to the compared width before a full-width
	// compare is correct unconditionally, with no need to prove the operands were
	// already clean above that width. Anything else fails loudly rather than
	// silently comparing full 64-bit registers.
	case "CMPQ":
		p.lowerCompare("CMP", "CMN", ops[0], ops[1], 8)
	case "CMPL":
		p.lowerCompare("CMPW", "CMNW", ops[0], ops[1], 4)
	case "CMPW", "CMPB":
		if !subwordEqNeSafe {
			panic(fmt.Sprintf("arm64: %s not supported (sub-32-bit compare needs width- and sign-correct operand extension unless every consumer is EQ/NE)", i.Opcode))
		}
		bits := 16
		if i.Opcode == "CMPB" {
			bits = 8
		}
		p.lowerSubwordCompareEqNe(bits, "CMP", ops[0], ops[1])
	case "TESTQ":
		p.lowerTest("TST", ops[0], ops[1], 8)
	case "TESTL":
		p.lowerTest("TSTW", ops[0], ops[1], 4)
	case "TESTW", "TESTB":
		if !subwordEqNeSafe {
			panic(fmt.Sprintf("arm64: %s not supported (sub-32-bit test needs width-correct operands unless every consumer is EQ/NE)", i.Opcode))
		}
		bits := 16
		if i.Opcode == "TESTB" {
			bits = 8
		}
		p.lowerSubwordTestEqNe(bits, ops[0], ops[1])

	default:
		panic(fmt.Sprintf("arm64: unsupported opcode %q (operands: %s)", i.Opcode, joinOperands(ops)))
	}
}

func (p *arm64) regOrImm(op operand.Op) string {
	if imm, ok := immAsm(op); ok {
		return imm
	}
	return operandReg(op)
}

// srcAsmW renders a source operand that may be a memory reference, immediate,
// or register, for an access of the given width in bytes. The extend loads
// (MOVBQZX, MOVWQZX, MOVLQSX and friends) know exactly how many bytes they
// touch, so they can use arm64's folded base+index form where a width-0
// caution would force the address through a scratch register instead. That
// matters: huff0's table lookup is one of these, and unfolding it adds an ADD
// to the innermost decode loop.
func (p *arm64) srcAsmW(op operand.Op, width int) string {
	if m, ok := op.(operand.Mem); ok {
		return p.memAsmW(m, width)
	}
	return p.regOrImm(op)
}

// valReg returns a register name holding op's value, loading a memory operand
// into scratchVal first. op must not be an immediate. x86 permits at most one
// memory operand per instruction, so callers never contend for scratchVal.
func (p *arm64) valReg(op operand.Op) string { return p.valRegW(op, 8) }

// valRegW is valReg for an access of the given width in bytes. Loading more
// than the operand's width would read past it, which is harmless for the value
// on a little-endian machine but can fault when the operand ends at a page
// boundary -- reachable for a 32-bit compare against the tail of a buffer.
func (p *arm64) valRegW(op operand.Op, width int) string {
	if m, ok := op.(operand.Mem); ok {
		ld := "MOVD"
		switch width {
		case 4:
			ld = "MOVWU"
		case 2:
			ld = "MOVHU"
		case 1:
			ld = "MOVBU"
		}
		p.emitf("%s %s, %s", ld, p.memAsmW(m, width), scratchVal)
		return scratchVal
	}
	return operandReg(op)
}

// lowerMove handles MOV* of reg/imm/mem to reg/mem.
func (p *arm64) lowerMove(op string, src, dst operand.Op) {
	w := accessWidth(op)
	dmem, dstIsMem := dst.(operand.Mem)
	smem, srcIsMem := src.(operand.Mem)
	switch {
	case dstIsMem:
		// A 64-bit store takes imm32 sign-extended; narrower stores take the
		// immediate literally, and only the low bytes are written anyway.
		immOf := immAsm
		if w == 8 {
			immOf = immAsmQ
		}
		if imm, ok := immOf(src); ok {
			p.emitf("MOVD %s, %s", imm, scratchVal)
			p.emitf("%s %s, %s", op, scratchVal, p.memAsmW(dmem, w))
			return
		}
		p.emitf("%s %s, %s", op, operandReg(src), p.memAsmW(dmem, w))
	case srcIsMem:
		p.emitf("%s %s, %s", op, p.memAsmW(smem, w), operandReg(dst))
	default:
		if imm, ok := immAsm(src); ok {
			p.emitf("%s %s, %s", op, imm, operandReg(dst))
			return
		}
		p.emitf("%s %s, %s", op, operandReg(src), operandReg(dst))
	}
}

// byteVal materializes the byte value of a MOVB/ADDB-style source operand in a
// register whose bits 7:0 hold the value (higher bits may be garbage unless the
// operand was a high-byte register, an immediate, or memory, which are cleanly
// extracted/loaded into scratch). Callers must consume only bits 7:0.
func (p *arm64) byteVal(op operand.Op) string {
	if isHighByte(op) {
		p.emitf("UBFX $8, %s, $8, %s", rename(op.(reg.Register)), scratchVal)
		return scratchVal
	}
	if imm, ok := immAsm(op); ok {
		p.emitf("MOVD %s, %s", imm, scratchVal)
		return scratchVal
	}
	if m, ok := op.(operand.Mem); ok {
		p.emitf("MOVBU %s, %s", p.memAsmW(m, 1), scratchVal)
		return scratchVal
	}
	return operandReg(op)
}

// lowerMOVB lowers x86 MOVB with exact partial-register semantics: a byte store
// writes one byte of memory; a register destination has only its addressed byte
// (bits 7:0, or 15:8 for AH/BH/CH/DH) replaced, via a bit-field insert, with
// every other bit preserved.
func (p *arm64) lowerMOVB(src, dst operand.Op) {
	// Either operand may be the high-byte one, and either may be what forces the
	// REX prefix that makes the other unreachable -- including a memory operand
	// based on an extended register.
	checkHighByteEncodable(src, dst)
	checkHighByteEncodable(dst, src)
	if dmem, ok := dst.(operand.Mem); ok {
		v := p.byteVal(src)
		p.emitf("MOVB %s, %s", v, p.memAsmW(dmem, 1))
		return
	}
	v := p.byteVal(src)
	lsb := 0
	if isHighByte(dst) {
		lsb = 8
	}
	p.emitf("BFI $%d, %s, $8, %s", lsb, v, operandReg(dst))
}

// lowerMOVW lowers a 16-bit move. Every form writes exactly bits 15:0 of its
// destination and leaves the rest alone, as on x86: stores touch two bytes,
// and loads, register moves and immediates all insert through BFI. Loads are
// the subtle one -- a bare MOVHU would be shorter, but it zeroes the upper 48
// bits of the destination, which x86 preserves.
func (p *arm64) lowerMOVW(src, dst operand.Op) {
	if dmem, ok := dst.(operand.Mem); ok {
		if imm, ok := immAsm(src); ok {
			p.emitf("MOVD %s, %s", imm, scratchVal)
			p.emitf("MOVH %s, %s", scratchVal, p.memAsmW(dmem, 2))
			return
		}
		p.emitf("MOVH %s, %s", operandReg(src), p.memAsmW(dmem, 2))
		return
	}
	if smem, ok := src.(operand.Mem); ok {
		p.emitf("MOVHU %s, %s", p.memAsmW(smem, 2), scratchVal)
		p.emitf("BFI $0, %s, $16, %s", scratchVal, operandReg(dst))
		return
	}
	v := operandReg(dst)
	if imm, ok := immAsm(src); ok {
		p.emitf("MOVD %s, %s", imm, scratchVal)
		p.emitf("BFI $0, %s, $16, %s", scratchVal, v)
		return
	}
	p.emitf("BFI $0, %s, $16, %s", operandReg(src), v)
}

// accessWidth is the width in bytes a Go arm64 load/store mnemonic touches,
// which determines whether an index may be scaled in a folded address operand.
func accessWidth(op string) int {
	switch op {
	case "MOVD":
		return 8
	case "MOVW", "MOVWU":
		return 4
	case "MOVH", "MOVHU":
		return 2
	case "MOVB", "MOVBU":
		return 1
	}
	return 0
}

// lowerMOVL lowers a 32-bit move. x86 MOVL zero-extends a register destination
// to 64 bits, so loads and register-to-register moves use MOVWU (zero-extend);
// Go arm64 MOVW would sign-extend. Stores write the low 32 bits.
func (p *arm64) lowerMOVL(src, dst operand.Op) {
	if dmem, ok := dst.(operand.Mem); ok {
		if imm, ok := immAsm(src); ok {
			p.emitf("MOVD %s, %s", imm, scratchVal)
			p.emitf("MOVW %s, %s", scratchVal, p.memAsmW(dmem, 4))
			return
		}
		p.emitf("MOVW %s, %s", operandReg(src), p.memAsmW(dmem, 4))
		return
	}
	if smem, ok := src.(operand.Mem); ok {
		p.emitf("MOVWU %s, %s", p.memAsmW(smem, 4), operandReg(dst))
		return
	}
	if _, isImm := immAsm(src); isImm {
		// A 32-bit move zero-extends, so the immediate must be materialized as
		// its unsigned 32-bit value: "MOVL $-1, r32" leaves 0x00000000ffffffff
		// on x86, where a bare "MOVD $-1" would set all 64 bits.
		v, ok := immVal(src)
		if !ok {
			panic("arm64: bad MOVL immediate")
		}
		// Rendered in avo's own hex style so an unaffected immediate keeps its
		// existing spelling and regeneration stays byte-for-byte stable.
		p.emitf("MOVD $0x%08x, %s", uint32(v), operandReg(dst))
		return
	}
	p.emitf("MOVWU %s, %s", operandReg(src), operandReg(dst))
}

// lowerMOVUPS lowers a 16-byte unaligned move between memory and a vector reg.
func (p *arm64) lowerMOVUPS(src, dst operand.Op) {
	// FMOVQ moves 128 bits and, unlike VLD1/VST1, takes a base+displacement
	// operand directly. That keeps the common "MOVOU disp(base), X" a single
	// instruction instead of materializing the address into a scratch register
	// first, which matters in the literal-copy paths where 16-byte moves are
	// dense. memAsm still folds an indexed operand into scratchAddr, and FMOVQ
	// accepts that plain "(reg)" form too. F<n> and V<n> name the same physical
	// register, so this interoperates with the VEOR/VMOV forms used elsewhere.
	if dmem, ok := dst.(operand.Mem); ok {
		p.emitf("FMOVQ %s, %s", vecAsF(src), p.memAsm(dmem))
		return
	}
	if smem, ok := src.(operand.Mem); ok {
		p.emitf("FMOVQ %s, %s", p.memAsm(smem), vecAsF(dst))
		return
	}
	p.emitf("VMOV %s.B16, %s.B16", operandReg(src), operandReg(dst))
}

// vecAsF renders a vector register under its F name, which the scalar FMOVQ
// form expects (F<n> and V<n> are the same physical register).
func vecAsF(op operand.Op) string {
	name := operandReg(op)
	if len(name) > 1 && name[0] == 'V' {
		return "F" + name[1:]
	}
	panic(fmt.Sprintf("arm64: expected a vector register, got %q", name))
}

// lowerArith lowers "OP src, dst" (dst op= src). dst may be a register or memory
// (read-modify-write via scratch). If flags is set, the flag-setting variant
// (sop) is used so a following branch/CMOV can consume NZCV.
func (p *arm64) lowerArith(op, sop string, src, dst operand.Op, flags bool) {
	mnem := op
	synth := ""
	if flags {
		if sop == "" {
			// ORR/EOR have no arm64 flag-setting form. x86's logical ops clear
			// CF and OF and set only ZF/SF meaningfully, so a following TST of
			// the result reproduces every condition that can legitimately be
			// read here; flagProducers rejects any other consumer.
			synth = "TST"
		} else {
			mnem = sop
		}
	}
	if dmem, ok := dst.(operand.Mem); ok {
		// Read-modify-write; src is a register or immediate (never memory too).
		s := p.regOrImmQ(src)
		m := p.memAsm(dmem)
		p.emitf("MOVD %s, %s", m, scratchVal)
		p.emitf("%s %s, %s, %s", mnem, s, scratchVal, scratchVal)
		p.emitf("MOVD %s, %s", scratchVal, m)
		if synth != "" {
			p.emitf("TST %s, %s", scratchVal, scratchVal)
		}
		return
	}
	d := operandReg(dst)
	var s string
	if imm, ok := immAsmQ(src); ok {
		s = imm
	} else {
		s = p.valReg(src) // loads memory source into scratch if needed
	}
	p.emitf("%s %s, %s, %s", mnem, s, d, d)
	if synth != "" {
		p.emitf("TST %s, %s", d, d)
	}
}

// lowerArithW is lowerArith for 32-bit operations. Register destinations use
// the arm64 W-forms, which zero the upper 32 bits exactly as x86 does; memory
// destinations are read-modify-written at 32-bit width.
func (p *arm64) lowerArithW(op, sop string, src, dst operand.Op, flags bool) {
	mnem := op
	synth := ""
	if flags {
		if sop == "" {
			synth = "TSTW" // see lowerArith
		} else {
			mnem = sop
		}
	}
	if dmem, ok := dst.(operand.Mem); ok {
		s := p.regOrImm(src)
		m := p.memAsm(dmem)
		p.emitf("MOVWU %s, %s", m, scratchVal)
		p.emitf("%s %s, %s, %s", mnem, s, scratchVal, scratchVal)
		p.emitf("MOVW %s, %s", scratchVal, m)
		if synth != "" {
			p.emitf("TSTW %s, %s", scratchVal, scratchVal)
		}
		return
	}
	d := operandReg(dst)
	var s string
	if imm, ok := immAsm(src); ok {
		s = imm
	} else if m, ok := src.(operand.Mem); ok {
		p.emitf("MOVWU %s, %s", p.memAsm(m), scratchVal)
		s = scratchVal
	} else {
		s = operandReg(src)
	}
	p.emitf("%s %s, %s, %s", mnem, s, d, d)
	if synth != "" {
		p.emitf("TSTW %s, %s", d, d)
	}
}

func (p *arm64) lowerIncDec(op, sop string, dst operand.Op, width int, flags bool) {
	mnem := op
	if flags {
		mnem = sop
	}
	if dmem, ok := dst.(operand.Mem); ok {
		// The read-modify-write has to run at the destination's width. A 64-bit
		// load and store around a 32-bit increment reads and writes four bytes
		// past the operand, and because the W-form arithmetic zeroes bits 63:32
		// of the scratch register, the store clears them: "INCL (AX)" would
		// increment the target dword and zero the one after it. The wide access
		// can also fault where the operand ends at a page boundary.
		ld, st := "MOVD", "MOVD"
		if width == 4 {
			ld, st = "MOVWU", "MOVW"
		}
		m := p.memAsmW(dmem, width)
		p.emitf("%s %s, %s", ld, m, scratchVal)
		p.emitf("%s $1, %s, %s", mnem, scratchVal, scratchVal)
		p.emitf("%s %s, %s", st, scratchVal, m)
		return
	}
	d := operandReg(dst)
	p.emitf("%s $1, %s, %s", mnem, d, d)
}

// checkNotDoubleShift rejects the three-operand form of SHL/SHR, which is a
// different x86 instruction than its two-operand namesake.
//
// "SHLQ $8, DX, AX" does not shift AX by 8: Go's assembler encodes it as SHLDQ,
// the double-precision shift, which fills AX's vacated bits from DX. The
// two-operand lowering would read the wrong operand as its destination and
// shift the fill register in place, leaving the real destination untouched --
// plausible-looking assembly that computes something else entirely. SARx, ROLx
// and RORx have no three-operand form, so only these four need the guard.
func checkNotDoubleShift(i *ir.Instruction) {
	if len(i.Operands) > 2 {
		panic(fmt.Sprintf("arm64: %s with %d operands is a double-precision shift (x86 %sD), which this lowering does not implement",
			i.Opcode, len(i.Operands), strings.TrimSuffix(i.Opcode, i.Opcode[len(i.Opcode)-1:])))
	}
}

// lowerShift lowers "SHIFT count, dst" (count imm or register).
func (p *arm64) lowerShift(op string, count, dst operand.Op, width int) {
	d := operandReg(dst)
	// shiftFolds may have absorbed the copy that fed this shift's source or
	// its count, in which case the arm64 three-operand form reads the
	// original registers directly (see shiftFolds).
	s := d
	if p.shift.src != "" {
		s = p.shift.src
	}
	// x86 masks the shift count to the low 6 bits (64-bit) or 5 (32-bit), so
	// "SHRQ $64" is a no-op and "SHRQ $65" shifts by one. arm64's register form
	// masks identically, but its immediate form rejects a count at or above the
	// width outright, which would turn a legal x86 program into an assembly
	// error. Mask here so the immediate matches what x86 would have done.
	if n, ok := immVal(count); ok && n&int64(width-1) != n {
		// Only rewrite when the mask actually changes the count. Reformatting an
		// in-range immediate would churn the generated assembly for no reason:
		// avo renders these in hex, and "$0x02" and "$2" assemble identically.
		p.emitf("%s $%d, %s, %s", op, n&int64(width-1), s, d)
		return
	}
	p.emitf("%s %s, %s, %s", op, p.shiftCount(count), s, d)
}

// shiftCount renders a shift's count operand, substituting the register
// shiftFolds chose when it absorbed the copy into CL.
func (p *arm64) shiftCount(count operand.Op) string {
	if p.shift.count != "" {
		if _, isImm := immAsm(count); isImm {
			panic("arm64: shiftFolds rewrote the count of an immediate shift")
		}
		return p.shift.count
	}
	return p.regOrImm(count)
}

// lowerROL lowers a rotate-left of the given width as arm64's rotate-right by
// the complementary amount, since arm64 has no rotate-left. ROR by (width-count)
// equals ROL by count, and the rotate amount is taken modulo the width.
func (p *arm64) lowerROL(count, dst operand.Op, width int) {
	ror, neg := "ROR", "NEG"
	if width == 32 {
		ror, neg = "RORW", "NEGW"
	}
	d := operandReg(dst)
	s := d // see lowerShift
	if p.shift.src != "" {
		s = p.shift.src
	}
	if _, isImm := immAsm(count); isImm {
		// immVal parses with a base-aware conversion; avo renders immediates in
		// hex, which a "$%d" scan would silently truncate to 0.
		n, ok := immVal(count)
		if !ok {
			panic("arm64: bad ROL immediate")
		}
		p.emitf("%s $%d, %s, %s", ror, (int64(width)-n)&int64(width-1), s, d)
		return
	}
	p.emitf("%s %s, %s", neg, p.shiftCount(count), scratchVal)
	p.emitf("%s %s, %s, %s", ror, scratchVal, s, d)
}

// srcRegInto returns a register name holding op's value, loading a memory operand
// into the given scratch register first. Callers pass a scratch that is free for
// the remainder of the lowering.
func (p *arm64) srcRegInto(op operand.Op, scratch string) string {
	if m, ok := op.(operand.Mem); ok {
		p.emitf("MOVD %s, %s", p.memAsm(m), scratch)
		return scratch
	}
	return operandReg(op)
}

// lowerShiftX lowers a BMI2 flag-free shift "SHIFTX count, src, dst":
// dst = src <shift> count. The count register is masked mod 64, as on x86.
// A count register just loaded with a constant folds to an immediate shift.
func (p *arm64) lowerShiftX(op string, count, src, dst operand.Op) {
	s := p.srcRegInto(src, scratchVal)
	if n, ok := p.knownConst(count); ok {
		p.emitf("%s $%d, %s, %s", op, n&63, s, operandReg(dst))
		return
	}
	p.emitf("%s %s, %s, %s", op, operandReg(count), s, operandReg(dst))
}

// lowerRORX lowers "RORXQ imm, src, dst": dst = ror(src, imm) (flag-free).
func (p *arm64) lowerRORX(imm, src, dst operand.Op) {
	c, ok := immAsm(imm)
	if !ok {
		panic("arm64: RORXQ requires an immediate rotate")
	}
	n, err := strconv.ParseInt(strings.TrimPrefix(c, "$"), 0, 64)
	if err != nil {
		panic("arm64: bad RORX immediate " + c)
	}
	s := p.srcRegInto(src, scratchVal)
	p.emitf("ROR $%d, %s, %s", n&63, s, operandReg(dst))
}

// lowerBZHI lowers "BZHIQ count, src, dst": dst = src & ((1<<count)-1). count is
// a register bit-count; zstd only uses counts < 64, for which the mask is exact
// (arm64 shifts mask the amount mod 64, so count == 64 is not handled). When the
// count register was just loaded with a constant, the mask is folded to an
// immediate AND (with x86's full n==0 / n>=64 semantics, which are exact).
func (p *arm64) lowerBZHI(count, src, dst operand.Op) {
	d := operandReg(dst)
	if n, ok := p.knownConst(count); ok {
		// x86 reads the bit count from the low 8 bits of the control operand and
		// ignores the rest, so the whole constant is the wrong thing to classify:
		// $256 means zero bits (clear the destination), not "at least 64" (copy
		// it), and $-1 means 255 (copy) rather than a negative count (clear).
		// Both of those get the answer exactly backwards. lowerBEXTR already
		// masks; this is the same rule.
		n &= 0xff
		s := p.srcRegInto(src, d)
		switch {
		case n == 0:
			p.emitf("MOVD $0, %s", d)
		case n >= 64:
			if s != d {
				p.emitf("MOVD %s, %s", s, d)
			}
		default:
			p.emitf("AND $%d, %s, %s", (int64(1)<<uint(n))-1, s, d)
		}
		return
	}
	s := p.srcRegInto(src, scratchAddr)
	n := operandReg(count)
	p.emitf("MOVD $1, %s", scratchVal)
	p.emitf("LSL %s, %s, %s", n, scratchVal, scratchVal) // 1 << count
	p.emitf("SUB $1, %s, %s", scratchVal, scratchVal)    // mask = (1<<count)-1
	p.emitf("AND %s, %s, %s", scratchVal, s, d)
}

// lowerBEXTR lowers "BEXTRQ ctrl, src, dst": with ctrl[7:0]=start and
// ctrl[15:8]=len, dst = (src >> start) & ((1<<len)-1). start and len come from the
// ctrl register at run time; zstd's control values keep both < 64. start and len
// are extracted before any destination write so ctrl may alias dst. When the
// ctrl register was just loaded with a constant, the extract folds to a single
// UBFX (or LSR when the field reaches bit 63, or a zero move when empty).
func (p *arm64) lowerBEXTR(ctrl, src, dst operand.Op) {
	d := operandReg(dst)
	if c, ok := p.knownConst(ctrl); ok {
		start := c & 0xff
		length := (c >> 8) & 0xff
		s := p.srcRegInto(src, d)
		switch {
		case length == 0 || start >= 64:
			p.emitf("MOVD $0, %s", d)
		case start+length >= 64:
			p.emitf("LSR $%d, %s, %s", start, s, d)
		default:
			p.emitf("UBFX $%d, %s, $%d, %s", start, s, length, d)
		}
		return
	}
	c := operandReg(ctrl)
	// Staging order matters here. The value is materialized first, because an
	// indexed memory source computes its address through scratchAddr and would
	// otherwise destroy a field already staged there. ctrl is then read twice,
	// and only after its last read is dst written, so ctrl may alias dst; and
	// only scratchAddr is reused between the two fields, after the first is
	// consumed. Every operand therefore survives until it is no longer needed.
	if m, ok := src.(operand.Mem); ok {
		p.emitf("MOVD %s, %s", p.memAsm(m), scratchVal)
	} else {
		p.emitf("MOVD %s, %s", operandReg(src), scratchVal)
	}
	p.emitf("UBFX $0, %s, $8, %s", c, scratchAddr)                 // start = ctrl[7:0]
	p.emitf("LSR %s, %s, %s", scratchAddr, scratchVal, scratchVal) // value >>= start
	p.emitf("UBFX $8, %s, $8, %s", c, scratchAddr)                 // len = ctrl[15:8]
	p.emitf("MOVD $1, %s", d)                                      // ctrl is dead; dst free
	p.emitf("LSL %s, %s, %s", scratchAddr, d, d)                   // 1 << len
	p.emitf("SUB $1, %s, %s", d, d)                                // (1<<len)-1
	p.emitf("AND %s, %s, %s", d, scratchVal, d)                    // dst = value & mask
}

// lowerMULX lowers "MULXQ src, lo, hi" (BMI2, flag-free): the 128-bit product
// src * RDX has its low half written to lo and its high half to hi. The low half
// is staged in scratch so lo/hi may alias src or RDX.
func (p *arm64) lowerMULX(src, lo, hi operand.Op) {
	dx := rename(reg.RDX)
	s := p.srcRegInto(src, scratchAddr)
	p.emitf("MUL %s, %s, %s", s, dx, scratchVal)       // low  -> scratch
	p.emitf("UMULH %s, %s, %s", s, dx, operandReg(hi)) // high -> hi (s, dx intact)
	if rename(lo.(reg.Register)) == rename(hi.(reg.Register)) {
		// x86 allows the two destinations to name the same register and defines
		// the high half as written second, so the high half is what survives.
		// Copying the staged low half here would leave the wrong one.
		return
	}
	p.emitf("MOVD %s, %s", scratchVal, operandReg(lo)) // low  -> lo
}

// lowerIMUL3 lowers "IMUL3Q imm, src, dst": dst = src * imm. x86 sign-extends the
// imm8/imm32 multiplier to 64 bits, so replicate that — otherwise a constant with
// bit 31 set (e.g. the 0x9E3779B1 golden-ratio multiplier) would differ from the
// amd64 result.
func (p *arm64) lowerIMUL3(imm, src, dst operand.Op) {
	c, ok := immAsm(imm)
	if !ok {
		panic("arm64: IMUL3Q requires an immediate multiplier")
	}
	v, err := strconv.ParseInt(strings.TrimPrefix(c, "$"), 0, 64)
	if err != nil {
		panic("arm64: bad IMUL3Q immediate " + c)
	}
	s := p.srcRegInto(src, scratchAddr)
	p.emitf("MOVD $%d, %s", int64(int32(v)), scratchVal) // materialize sign-extended imm32
	p.emitf("MUL %s, %s, %s", scratchVal, s, operandReg(dst))
}

// lowerIMUL lowers IMULQ in its 2-operand ("src, dst": dst *= src) and 1-operand
// ("src": RDX:RAX = RAX * src, signed) forms.
func (p *arm64) lowerIMUL(ops []operand.Op) {
	switch len(ops) {
	case 2:
		s := p.valReg(ops[0])
		d := operandReg(ops[1])
		p.emitf("MUL %s, %s, %s", s, d, d)
	case 1:
		p.lowerWideMul("SMULH", ops[0])
	default:
		panic(fmt.Sprintf("arm64: IMULQ with %d operands not supported", len(ops)))
	}
}

// lowerWideMul lowers the single-operand MULQ/IMULQ: RDX:RAX = RAX * src, with
// hiOp = UMULH (unsigned) or SMULH (signed). The source is staged when it aliases
// RDX so writing the result cannot clobber it.
func (p *arm64) lowerWideMul(hiOp string, src operand.Op) {
	rax := rename(reg.RAX)
	rdx := rename(reg.RDX)
	s := p.srcRegInto(src, scratchVal)
	if s == rdx {
		p.emitf("MOVD %s, %s", s, scratchVal)
		s = scratchVal
	}
	p.emitf("%s %s, %s, %s", hiOp, s, rax, rdx) // RDX = high(RAX*src)
	p.emitf("MUL %s, %s, %s", s, rax, rax)      // RAX = low(RAX*src)
}

func (p *arm64) lowerLEA(m operand.Mem, dst string) {
	m = frameAdjust(m)
	if m.Symbol.Name != "" {
		panic("arm64: LEA of symbol not supported")
	}
	base := rename(m.Base)
	if m.Index != nil && m.Scale != 0 {
		sh := log2scale(m.Scale)
		if sh == 0 {
			p.emitf("ADD %s, %s, %s", rename(m.Index), base, dst)
		} else {
			p.emitf("ADD %s<<%d, %s, %s", rename(m.Index), sh, base, dst)
		}
		base = dst
	}
	if m.Disp != 0 {
		if m.Disp > 0 {
			p.emitf("ADD $%d, %s, %s", m.Disp, base, dst)
		} else {
			p.emitf("SUB $%d, %s, %s", -m.Disp, base, dst)
		}
		return
	}
	if base != dst {
		p.emitf("MOVD %s, %s", base, dst)
	}
}

// lowerCompare emits an arm64 flag-setter for "CMP a, b" (flags from a - b),
// using the given compare mnemonic (cmp) and its negated-immediate counterpart
// (cmn), so callers select the operand width: CMP/CMN for 64-bit, CMPW/CMNW for
// 32-bit. At most one of a, b is a memory operand (loaded into scratchVal).
func (p *arm64) lowerCompare(cmp, cmn string, a, b operand.Op, width int) {
	// The sign-extension rewrite has to happen before negImm, so a value like
	// $0xfffffff0 becomes $-16 and takes the (equivalent) CMN path rather than
	// comparing against four billion.
	immOf := immAsm
	if width == 8 {
		immOf = immAsmQ
	}
	if imm, ok := immOf(b); ok {
		aReg := p.valRegW(a, width)
		// "CMP a, -b" becomes "CMN a, b" because subtracting a negative is
		// adding its magnitude -- but only while that magnitude is still
		// representable as a positive number at the compare width. At the
		// signed minimum it is not: negating INT32_MIN gives 2^31, which as a
		// 32-bit addend is INT32_MIN again, so arm64 would add -2^31 where x86
		// subtracts it. N, Z and C survive that (same result bits, nonzero
		// addend) but V is computed from a different true value and comes out
		// inverted on every input, flipping all four signed conditions. Emit the
		// literal compare instead; the assembler materializes the immediate and
		// uses SUBS, which is exact.
		if neg, val := negImm(imm); neg && !(width == 4 && val >= 1<<31) {
			p.emitf("%s $%d, %s", cmn, val, aReg)
			return
		}
		p.emitf("%s %s, %s", cmp, imm, aReg)
		return
	}
	// arm64 CMP Rm, Rn computes Rn - Rm; we want a - b, so Rn=a, Rm=b.
	if _, ok := a.(operand.Mem); ok {
		aReg := p.valRegW(a, width)
		p.emitf("%s %s, %s", cmp, operandReg(b), aReg)
		return
	}
	p.emitf("%s %s, %s", cmp, p.valRegW(b, width), operandReg(a))
}

// lowerTest emits an arm64 bitwise-test flag-setter for "TEST a, b" using the
// given mnemonic (TST for 64-bit, TSTW for 32-bit).
func (p *arm64) lowerTest(op string, a, b operand.Op, width int) {
	// x86 writes the immediate first ("TESTQ $0x10, AX"), which is the common
	// form and the only one avo produces. AND is commutative and the result is
	// discarded, so testing in either order sets the same flags.
	if _, ok := immAsm(a); ok {
		a, b = b, a
	}
	if _, ok := b.(operand.Mem); ok {
		// x86 allows TEST m64, r64, but this lowering stages only one operand
		// through a scratch register and would need a second for the load.
		panic(fmt.Sprintf("arm64: %s with the memory operand second is not supported; "+
			"swap the operands", op))
	}
	aReg := p.valRegW(a, width)
	rhs := p.regOrImm(b)
	if width == 8 {
		rhs = p.regOrImmQ(b)
	}
	p.emitf("%s %s, %s", op, rhs, aReg)
}

// lowerSubwordCompareEqNe lowers a sub-32-bit CMP (bits is 8 or 16) whose only
// consumers are EQ/NE (see subwordSafeEqNe), by zero-extending both operands
// to the compared width before a full-width compare. Correct unconditionally:
// equality doesn't depend on sign, so no proof that the operands were already
// clean above that width is needed, unlike the general sub-word compare case.
func (p *arm64) lowerSubwordCompareEqNe(bits int, cmp string, a, b operand.Op) {
	ra, rb := p.materializeEqNe(bits, a, b)
	// arm64 CMP Rm, Rn computes Rn - Rm; we only need Z, so the operand order
	// does not matter here, but keep it consistent with lowerCompare (a - b).
	p.emitf("%s %s, %s", cmp, rb, ra)
}

// lowerSubwordTestEqNe lowers a sub-32-bit TEST (bits is 8 or 16) whose only
// consumers are EQ/NE (i.e. only the Z flag is read), by zero-extending both
// operands to the tested width before a full-width TST. Correct
// unconditionally: with both operands clean above that width, the AND's upper
// bits are always zero, so Z exactly reflects whether the low bits are all
// zero, which is what EQ/NE tests.
func (p *arm64) lowerSubwordTestEqNe(bits int, a, b operand.Op) {
	// "TESTB/TESTW dst, dst" -- the same x86 register named on both sides -- is
	// the idiom for "is dst zero?". AND(dst,dst) is dst, so the general path's
	// two independent zero-extended copies are redundant: an immediate-masked
	// TST directly against the original register reads the same Z bit. This
	// needs no register materialization at all, unlike the two-operand case,
	// where a and b may differ and each must be pinned to a specific scratch
	// register before the compare.
	if ra, ok := a.(reg.Register); ok {
		if rb, ok := b.(reg.Register); ok && ra.Asm() == rb.Asm() {
			mask := uint64(1)<<uint(bits) - 1
			if isHighByte(a) {
				// AH/BH/CH/DH occupy bits 15:8 of the arm64 register the low byte
				// renames to (see isHighByte); shift the mask to match.
				p.emitf("TST $0x%x, %s", mask<<8, rename(ra))
				return
			}
			p.emitf("TST $0x%x, %s", mask, operandReg(a))
			return
		}
	}
	ra, rb := p.materializeEqNe(bits, a, b)
	p.emitf("TST %s, %s", rb, ra)
}

// zeroExtendEqNe materializes op, zero-extended to the given bit width (8 or
// 16), into the given scratch register and returns its name.
//
// A memory operand needs no masking: the zero-extending load of that exact
// width already produces the value. Callers must materialize a memory operand
// before any other operand, because computing an indexed address clobbers
// scratchAddr (see lowerSubwordCompareEqNe).
func (p *arm64) zeroExtendEqNe(bits int, op operand.Op, scratch string) string {
	mask := fmt.Sprintf("$0x%x", uint64(1)<<uint(bits)-1)
	if imm, ok := immAsm(op); ok {
		p.emitf("MOVD %s, %s", imm, scratch)
		p.emitf("AND %s, %s, %s", mask, scratch, scratch)
		return scratch
	}
	if m, ok := op.(operand.Mem); ok {
		ld := "MOVBU"
		if bits == 16 {
			ld = "MOVHU"
		}
		p.emitf("%s %s, %s", ld, p.memAsm(m), scratch)
		return scratch
	}
	if isHighByte(op) {
		// AH/BH/CH/DH occupy bits 15:8 of the same arm64 register AL/BL/CL/DL
		// rename to, so masking the renamed register would compare the low byte
		// instead -- and wrongly in both directions, since either byte can be
		// the larger.
		p.emitf("UBFX $8, %s, $8, %s", rename(op.(reg.Register)), scratch)
		return scratch
	}
	p.emitf("AND %s, %s, %s", mask, operandReg(op), scratch)
	return scratch
}

// forcesREX reports whether an operand can only be encoded with a REX prefix.
// That is true of the extended registers R8-R15, of the byte registers
// SPL/BPL/SIL/DIL (whose encodings are the AH/CH/DH/BH slots, reachable only by
// adding REX), and of a memory operand whose base or index is such a register.
func forcesREX(op operand.Op) bool {
	switch o := op.(type) {
	case operand.Mem:
		return forcesREX(o.Base) || (o.Index != nil && forcesREX(o.Index))
	case reg.Physical:
		if o.PhysicalIndex() >= 8 {
			return true
		}
		// The high-byte registers occupy indices 0-3 (AH, CH, DH, BH), so a byte
		// register at index 4 or above is SPL/BPL/SIL/DIL -- names that exist
		// only under REX, since without it those encodings mean the high bytes.
		return o.Size() == 1 && o.PhysicalIndex() >= 4
	}
	return false
}

// checkHighByteEncodable panics when high is a high-byte register (AH/BH/CH/DH)
// and other forces a REX prefix, a pairing x86-64 cannot express: REX redefines
// the high-byte operand's slot as SPL.
//
// This guard is load-bearing, not a diagnostic. Go's assembler does NOT reject
// the combination -- it silently encodes the stack-pointer byte instead, so
// "MOVB AH, (R8)" assembles to 41 88 20 and stores SPL, and "MOVB R8B, AH"
// overwrites RSP's low byte. Without this check the amd64 side would quietly
// read or write the wrong register while the arm64 lowering faithfully used AH,
// which is exactly the divergence this printer exists to prevent.
func checkHighByteEncodable(high, other operand.Op) {
	if !isHighByte(high) || !forcesREX(other) {
		return
	}
	panic(fmt.Sprintf("arm64: high-byte operand %s paired with %s cannot be encoded on x86-64: "+
		"the second operand forces a REX prefix, which renames the high-byte operand to SPL",
		high.Asm(), other.Asm()))
}

// materializeEqNe zero-extends both operands of a sub-word compare/test into
// the two scratch registers. x86 permits at most one memory operand, and that
// one is materialized first: its address computation may use scratchAddr, which
// would otherwise clobber a value already placed there.
func (p *arm64) materializeEqNe(bits int, a, b operand.Op) (ra, rb string) {
	// The same REX trap the byte-extend path guards: a high-byte operand is only
	// reachable on x86-64 from an encoding with no REX prefix, and anything that
	// forces one -- an extended register, a byte register in the SPL/BPL/SIL/DIL
	// group, or a memory operand based on either -- renames the high-byte
	// operand to SPL. The assembler does not object; it encodes SPL and moves
	// on, so the amd64 side would compare the stack pointer's low byte while the
	// arm64 side compared AH.
	checkHighByteEncodable(a, b)
	checkHighByteEncodable(b, a)
	if _, aIsMem := a.(operand.Mem); aIsMem {
		ra = p.zeroExtendEqNe(bits, a, scratchAddr)
		rb = p.zeroExtendEqNe(bits, b, scratchVal)
		return ra, rb
	}
	rb = p.zeroExtendEqNe(bits, b, scratchVal)
	ra = p.zeroExtendEqNe(bits, a, scratchAddr)
	return ra, rb
}

// lowerSET lowers "SETcc dst". x86 SETcc writes only the low byte of dst, so
// the 0/1 is materialized with CSET in scratch and inserted into bits 7:0.
// When setccFolds proved the rest of the register is zero (full), the
// insert is the whole register and CSET writes it directly.
func (p *arm64) lowerSET(i *ir.Instruction, full bool) {
	cc, ok := armCond(strings.TrimPrefix(i.Opcode, "SET"))
	if !ok {
		panic(fmt.Sprintf("arm64: unsupported SETcc %q", i.Opcode))
	}
	if isHighByte(i.Operands[0]) {
		panic("arm64: SETcc high-byte destination not supported")
	}
	if full {
		p.emitf("CSET %s, %s", cc, operandReg(i.Operands[0]))
		return
	}
	p.emitf("CSET %s, %s", cc, scratchVal)
	p.emitf("BFI $0, %s, $8, %s", scratchVal, operandReg(i.Operands[0]))
}

// lowerCMOV lowers "CMOVcc src, dst" to "CSEL cc, src, dst, dst".
func (p *arm64) lowerCMOV(i *ir.Instruction) {
	cond := cmovCond(i.Opcode)
	src := operandReg(i.Operands[0])
	dst := operandReg(i.Operands[1])
	switch {
	case strings.HasPrefix(i.Opcode, "CMOVW"):
		// x86 CMOVW inserts 16 bits, leaving the upper 48 untouched on both
		// condition outcomes. CSEL has no half-word form to express that.
		panic("arm64: CMOVW is not supported (no 16-bit conditional select)")
	case strings.HasPrefix(i.Opcode, "CMOVL"):
		// x86 writes a 32-bit CMOV destination whichever way the condition goes,
		// zero-extending it. The W-form select does the same; the 64-bit one
		// would leave stale upper bits when the condition is false.
		p.emitf("CSELW %s, %s, %s, %s", cond, src, dst, dst)
	default:
		p.emitf("CSEL %s, %s, %s, %s", cond, src, dst, dst)
	}
}

func negImm(imm string) (bool, int) {
	var v int
	if _, err := fmt.Sscanf(imm, "$%d", &v); err != nil {
		return false, 0
	}
	if v < 0 {
		if v == math.MinInt64 {
			// Negating this yields itself, so the caller would rewrite a compare
			// against the minimum into a compare against the minimum with the
			// opposite sense. Unreachable through avo's imm32 forms, but the
			// value is cheap to refuse and expensive to get wrong.
			return false, 0
		}
		return true, -v
	}
	return false, 0
}

func isConditionalBranch(i *ir.Instruction) bool {
	return i.Opcode != "JMP" && strings.HasPrefix(i.Opcode, "J")
}

// armCond maps an x86/Go condition suffix to the arm64 condition mnemonic shared
// by B.cond and CSEL. Mapping is by meaning: x86 and arm64 use opposite
// carry-flag conventions for subtraction, but the named conditions encode the
// intent, not the raw flag, so the same-named arm64 condition is correct.
func armCond(cc string) (string, bool) {
	switch cc {
	case "EQ", "E", "Z":
		return "EQ", true
	case "NE", "NZ":
		return "NE", true
	case "LT", "L":
		return "LT", true // signed <
	case "LE":
		return "LE", true // signed <=
	case "GT", "G":
		return "GT", true // signed >
	case "GE":
		return "GE", true // signed >=
	case "CS", "B", "LO", "NAE":
		return "LO", true // unsigned <
	case "CC", "AE", "HS", "NB":
		return "HS", true // unsigned >=
	case "HI", "A", "NBE":
		return "HI", true // unsigned >
	case "LS", "BE", "NA":
		return "LS", true // unsigned <=
	case "MI", "S":
		return "MI", true // negative
	case "PL", "NS":
		return "PL", true // non-negative
	case "OS":
		return "VS", true // overflow
	case "OC":
		return "VC", true // no overflow
	}
	return "", false
}

// branchMnemonic maps an x86 conditional-jump opcode (Go-canonical name or Intel
// alias) to the arm64 conditional branch with the same meaning.
func branchMnemonic(op string) string {
	if cc, ok := armCond(strings.TrimPrefix(op, "J")); ok {
		return "B" + cc
	}
	panic(fmt.Sprintf("arm64: unsupported conditional branch %q", op))
}

// cmovCond maps an x86 CMOVcc opcode to its arm64 CSEL condition, for the Q/L/W
// operand-size prefixes.
func cmovCond(op string) string {
	cc := op
	for _, prefix := range []string{"CMOVQ", "CMOVL", "CMOVW"} {
		if strings.HasPrefix(op, prefix) {
			cc = op[len(prefix):]
			break
		}
	}
	if a, ok := armCond(cc); ok {
		return a
	}
	panic(fmt.Sprintf("arm64: unsupported CMOV %q", op))
}

// flagProducers identifies, for every flag consumer (a conditional branch or a
// CMOVcc), the instruction that produces the NZCV it reads, and returns the set
// of node indices whose lowering must therefore emit a flag-setting variant.
//
// avo's IR does not model EFLAGS, so the producer/consumer link is recovered
// structurally: scanning back from a consumer, comments and flag-transparent
// instructions (moves, address computations, and other branches/CMOVs — which
// read flags but never write them) are skipped, and the first flag-affecting
// instruction is the producer. CMP*/TEST* already emit an unconditional
// flag-setter and need no mark; a producer whose lowering cannot carry flags
// (e.g. ORQ/XORQ/shifts) is unsupported and panics rather than let a branch run
// on stale flags. Reaching a label, an unevaluated preprocessor directive, or
// the start of the function means the flags cross an edge this printer does not
// model, and all three panic: a consumer left silently unmarked would branch on
// whatever NZCV happened to survive.
//
// A single-instruction lookahead is insufficient because x86 permits
// flag-transparent instructions (a MOV, an LEA) between a producer and the
// branch that consumes it; those must be skipped, not treated as the producer.
// btFusion records a BTL and its consuming branch, collapsed into one arm64
// test-and-branch: branch is the node index of the branch being absorbed, and
// mnemonic is the instruction that replaces the pair.
type btFusion struct {
	branch   int
	mnemonic string
}

// btPairs finds each BTL that is immediately consumed by a carry branch and
// maps its node index to that pairing, so the two can be emitted as a single
// arm64 TBNZ (carry set) or TBZ (carry clear). It panics on any BTL outside
// that exact shape.
//
// Why fuse rather than lower BTL on its own. x86 BT copies the selected bit
// into CF; Intel documents OF, SF, AF and PF as UNDEFINED afterwards, and AMD
// has printed ZF as undefined too, so CF is the only flag that means anything
// across vendors. Materializing CF and letting the generic branch path handle
// the branch would also be wrong in a way worth naming: armCond maps CS to LO
// because after a compare the two architectures use opposite borrow senses,
// but BT's CF is a raw bit with no borrow about it, so that mapping inverts the
// test. TBNZ/TBZ sidestep both -- they read the bit directly and write no flags.
//
// The safety of leaving BTL out of flagSetter and isFlagTransparent is what
// makes this narrow rather than merely small. Any *other* consumer that reaches
// a BTL by the backward scan -- a later branch reading the CF that x86 does
// leave live, a SETCS, an ADC, or a JEQ reading flags BT left undefined --
// lands on the unknown-producer panic in flagProducers instead of being lowered
// wrong. Only the one branch fused here is exempted from that scan.
//
// Adjacency is required for correctness, not caution. With an instruction in
// between, emitting the test-and-branch at the BTL would skip it on the taken
// path, and emitting it at the branch would test a register that instruction
// may have rewritten. Comments carry no code, so they may sit between.
func btPairs(nodes []ir.Node) map[int]btFusion {
	pairs := make(map[int]btFusion)
	for j, n := range nodes {
		ins, ok := n.(*ir.Instruction)
		if !ok || !strings.HasPrefix(ins.Opcode, "BT") || strings.HasPrefix(ins.Opcode, "BTC") ||
			strings.HasPrefix(ins.Opcode, "BTR") || strings.HasPrefix(ins.Opcode, "BTS") {
			// BTC/BTR/BTS modify the bit as well as testing it and have their
			// own lowerings; only the pure test lands here.
			continue
		}
		if ins.Opcode != "BTL" {
			panic(fmt.Sprintf("arm64: %s is not supported; only BTL, whose bit index this checks is "+
				"below the operand width, has a lowering", ins.Opcode))
		}
		bit, isImm := immVal(ins.Operands[0])
		if !isImm {
			panic("arm64: BTL with a register bit index is not supported; " +
				"x86 masks that index modulo the operand size and TBNZ cannot express it")
		}
		if bit < 0 || bit >= 32 {
			panic(fmt.Sprintf("arm64: BTL bit index %d is not below the 32-bit operand width; "+
				"x86 would mask it modulo 32 rather than test that bit", bit))
		}
		if _, isReg := ins.Operands[1].(reg.Register); !isReg {
			panic("arm64: BTL against a memory operand is not supported; " +
				"x86 addresses the bit string beyond the addressed word and TBNZ reads one register")
		}
		k := j + 1
		for k < len(nodes) {
			if _, isComment := nodes[k].(*ir.Comment); !isComment {
				break
			}
			k++
		}
		var branch *ir.Instruction
		isInstr := false
		if k < len(nodes) {
			branch, isInstr = nodes[k].(*ir.Instruction)
		}
		// Naming what actually follows matters here: a label between the pair
		// is the interesting rejection (it would let a jump reach the branch
		// without executing the BTL), and reporting it as "end of function"
		// sends whoever hits this looking in the wrong place.
		next := "end of function"
		if k < len(nodes) {
			switch n := nodes[k].(type) {
			case *ir.Instruction:
				next = n.Opcode
			case ir.Label:
				next = fmt.Sprintf("label %q", string(n))
			default:
				next = fmt.Sprintf("%T", n)
			}
		}
		// Only a carry branch says anything about the bit BT selected; every
		// other condition reads a flag x86 leaves undefined afterwards. JC and
		// JCS are one instruction under two names, as are JNC and JCC, and avo
		// keeps whichever spelling the generator wrote. The comparison-named
		// aliases for those same encodings (JB, JNAE, JAE, JNB) are left to the
		// panic: they mean the same thing here, but a borrow name after a BT
		// reads like the compare-derived carry this lowering deliberately
		// refuses, and no generator has needed them.
		var mnemonic string
		switch {
		case isInstr && (branch.Opcode == "JC" || branch.Opcode == "JCS"):
			mnemonic = "TBNZ"
		case isInstr && (branch.Opcode == "JNC" || branch.Opcode == "JCC"):
			mnemonic = "TBZ"
		default:
			panic(fmt.Sprintf("arm64: BTL is followed by %s, but is only supported when the next "+
				"instruction is a carry branch (JC/JCS or JNC/JCC), so the two can be fused into one "+
				"test-and-branch; on its own BT sets a carry flag this lowering cannot represent", next))
		}
		if _, ok := branch.Operands[0].(operand.LabelRef); !ok {
			panic(fmt.Sprintf("arm64: %s fused with BTL targets %s, not a label",
				branch.Opcode, branch.Operands[0].Asm()))
		}
		pairs[j] = btFusion{branch: k, mnemonic: mnemonic}
	}
	return pairs
}

func flagProducers(nodes []ir.Node) map[int]bool {
	setflags := make(map[int]bool)
	// A branch fused into a TBNZ reads its bit straight from the register, so
	// it has no flag producer to find. Left in the scan below it would walk
	// back to its BTL and panic there.
	fused := make(map[int]bool)
	for _, f := range btPairs(nodes) {
		fused[f.branch] = true
	}
	for j, n := range nodes {
		ins, ok := n.(*ir.Instruction)
		if !ok || !(strings.HasPrefix(ins.Opcode, "CMOV") || strings.HasPrefix(ins.Opcode, "SET") ||
			strings.HasPrefix(ins.Opcode, "ADC") || isConditionalBranch(ins)) {
			continue
		}
		if fused[j] {
			continue
		}
		found := false
		for k := j - 1; k >= 0; k-- {
			if _, isComment := nodes[k].(*ir.Comment); isComment {
				// Comments carry no code and no directives -- those are refused
				// before they reach any analysis -- so nothing here affects flags.
				continue
			}
			prev, isInstr := nodes[k].(*ir.Instruction)
			if !isInstr {
				// A label: the flags this consumer reads were produced on some
				// predecessor edge, which this printer does not model. Leaving
				// the producer unmarked would emit a non-flag-setting op and let
				// the branch run on whatever NZCV happened to survive --
				// plausible-looking but wrong assembly -- so fail generation.
				panic(fmt.Sprintf("arm64: %s reads flags produced across a label; "+
					"the lowering tracks flags only within a straight-line run", ins.Opcode))
			}
			if isFlagTransparent(prev.Opcode) {
				continue
			}
			mark, ok := flagSetter(prev.Opcode)
			if !ok {
				panic(fmt.Sprintf("arm64: %s consumes flags from %s, which the lowering cannot emit as a flag-setter", ins.Opcode, prev.Opcode))
			}
			cond, isCondConsumer := consumerCondition(ins.Opcode)
			if isLogicalFlagOp(prev.Opcode) && isCondConsumer {
				// x86's logical ops force CF=0 and OF=0, and their arm64
				// counterparts force C=0 too. The raw bits agree, so armCond's
				// meaning-inverting map (right for CMP/SUB, where the two ISAs
				// use opposite borrow conventions) is wrong here: "JB" after a
				// TEST is never taken on x86 but "BLO" always is. Only the Z/N
				// conditions carry information after these, plus the signed ones
				// since V is zero on both sides.
				switch cond {
				case "EQ", "NE", "MI", "PL", "LT", "LE", "GT", "GE":
				default:
					panic(fmt.Sprintf("arm64: %s consumes condition %s from %s, but after a logical op only the ZF/SF conditions mean the same thing on both architectures", ins.Opcode, cond, prev.Opcode))
				}
			}
			if strings.HasPrefix(ins.Opcode, "ADC") && !isBorrowProducer(prev.Opcode) {
				// The ADC lowering (CSINC on HS) hardcodes the borrow convention,
				// which only matches when the producer is a compare or subtract.
				// After an addition both ISAs use the same carry-out convention,
				// so the same CSINC would increment on exactly the wrong input.
				panic(fmt.Sprintf("arm64: %s reads carry from %s; the lowering assumes a borrow-producing compare or subtract", ins.Opcode, prev.Opcode))
			}
			if isCondConsumer && carryCondition(cond) && !isBorrowProducer(prev.Opcode) &&
				!isCarryPreservingOp(prev.Opcode) && !isLogicalFlagOp(prev.Opcode) {
				// armCond maps carry conditions by their post-compare meaning,
				// where x86 and arm64 use opposite borrow conventions. After an
				// addition both set carry-out with the SAME sense, so that map
				// inverts the test; and "above"/"below or equal" combine carry
				// with zero in a way no single arm64 condition expresses. Only a
				// compare or subtract can be translated.
				panic(fmt.Sprintf("arm64: %s reads carry condition %s from %s; only a compare or subtract produces the borrow sense this lowering assumes", ins.Opcode, cond, prev.Opcode))
			}
			if isCarryPreservingOp(prev.Opcode) && isCondConsumer && carryCondition(cond) {
				// x86 INC/DEC deliberately preserve CF so they can appear inside
				// a carry-driven loop; arm64 has no such form, and the ADDS/SUBS
				// this lowers to overwrites C with the increment's own carry.
				panic(fmt.Sprintf("arm64: %s reads carry condition %s across %s, which preserves CF on x86 but not once lowered", ins.Opcode, cond, prev.Opcode))
			}
			if mark {
				setflags[k] = true
			}
			found = true
			break
		}
		if !found {
			// The scan ran off the start of the function without finding a
			// producer. amd64's entry EFLAGS is undefined too, so a program in
			// this shape is already broken there -- but it is the one remaining
			// way to leave a consumer unmarked, and silently emitting a branch on
			// entry NZCV is not how the other unreachable-producer cases behave.
			panic(fmt.Sprintf("arm64: %s reads flags with no producer in this function; "+
				"the flags would come from the caller, which neither architecture defines", ins.Opcode))
		}
	}
	return setflags
}

// isFlagTransparent reports whether an opcode's lowering leaves NZCV unchanged.
// Moves and address computations never touch flags; conditional branches and
// CMOVcc read flags but do not write them, so they too are transparent to an
// earlier producer's flags.
func isFlagTransparent(op string) bool {
	switch op {
	// BMI2 flag-free variants (shifts and wide multiply) deliberately leave flags
	// untouched, unlike their non-BMI2 counterparts; their arm64 lowerings use
	// non-flag-setting ops too, so NZCV survives a producer across them.
	case "SHLXQ", "SHRXQ", "SARXQ", "RORXQ", "MULXQ":
		return true
	// x86 leaves EFLAGS alone for these, and so do their arm64 lowerings (MVN,
	// REVW, a three-move register swap, VEOR). Omitting them does not produce
	// wrong code -- an unrecognized instruction between a producer and its
	// consumer aborts generation -- but it rejects programs that are perfectly
	// well defined. Note the neighbours that deliberately stay out: NEG, POPCNT,
	// BSF/BSR and TZCNT all write flags on x86, so a consumer reading them has
	// no arm64 equivalent and must keep failing loudly.
	case "NOTQ", "NOTL", "BSWAPL", "XCHGQ", "PXOR":
		return true
	}
	switch {
	case strings.HasPrefix(op, "MOV"),
		strings.HasPrefix(op, "LEA"),
		strings.HasPrefix(op, "PREFETCH"), // hints; no architectural effect at all
		strings.HasPrefix(op, "CMOV"),
		strings.HasPrefix(op, "SET"), // reads flags, never writes them
		strings.HasPrefix(op, "J"):   // JMP and the Jcc family
		return true
	}
	return false
}

// flagSetter classifies a flag-affecting opcode. mark is true when its lowering
// must be switched to a flag-setting variant; false when it already emits a
// setter unconditionally (CMP*/TEST*). ok is false for flag-affecting opcodes
// this printer cannot lower as a flag-setter, so callers fail loudly instead of
// silently branching on stale flags.
func flagSetter(op string) (mark, ok bool) {
	switch op {
	case "ADDQ", "SUBQ", "ANDQ", "INCQ", "DECQ",
		"ADDL", "SUBL", "ANDL", "INCL", "DECL":
		return true, true
	case "ORQ", "XORQ", "ORL", "XORL":
		// No arm64 ORRS/EORS; lowerArith appends a TST instead. Only valid for
		// Z/N-based consumers, which flagProducers verifies.
		return true, true
	case "CMPQ", "CMPL", "CMPW", "CMPB", "TESTQ", "TESTL", "TESTW", "TESTB":
		return false, true
	default:
		return false, false
	}
}

// isLogicalFlagOp reports whether an opcode is a bitwise OR/XOR, whose arm64
// lowering has no flag-setting form and instead appends a TST (see lowerArith).
func isLogicalFlagOp(op string) bool {
	switch op {
	case "ORQ", "XORQ", "ORL", "XORL",
		"ANDQ", "ANDL",
		"TESTQ", "TESTL", "TESTW", "TESTB":
		return true
	}
	return false
}

// isCarryPreservingOp reports whether an x86 opcode leaves CF untouched. INC and
// DEC are defined that way, so a carry-condition consumer after one is really
// reading an earlier producer's carry -- something this lowering cannot express,
// because its arm64 counterpart (ADDS/SUBS) does write C.
func isCarryPreservingOp(op string) bool {
	switch op {
	case "INCQ", "INCL", "DECQ", "DECL":
		return true
	}
	return false
}

// isBorrowProducer reports whether an opcode's carry flag means "borrow" in the
// subtraction sense, which is the convention ADC's lowering assumes.
func isBorrowProducer(op string) bool {
	return strings.HasPrefix(op, "CMP") || strings.HasPrefix(op, "SUB")
}

// carryCondition reports whether an arm64 condition reads C.
func carryCondition(cond string) bool {
	switch cond {
	case "LO", "HS", "HI", "LS":
		return true
	}
	return false
}

// isSubwordCompare reports whether opcode is a sub-32-bit CMP/TEST, the only
// class this printer treats as conditionally supported (see subwordSafeEqNe).
func isSubwordCompare(opcode string) bool {
	switch opcode {
	case "CMPW", "CMPB", "TESTW", "TESTB":
		return true
	}
	return false
}

// consumerCondition returns the arm64 condition an instruction consumes flags
// under, if it is a CMOVcc, SETcc, ADC (carry), or a conditional branch.
func consumerCondition(opcode string) (string, bool) {
	if strings.HasPrefix(opcode, "CMOV") {
		cc := opcode
		for _, prefix := range []string{"CMOVQ", "CMOVL", "CMOVW"} {
			if strings.HasPrefix(opcode, prefix) {
				cc = opcode[len(prefix):]
				break
			}
		}
		return armCond(cc)
	}
	if strings.HasPrefix(opcode, "SET") {
		return armCond(strings.TrimPrefix(opcode, "SET"))
	}
	if strings.HasPrefix(opcode, "ADC") {
		return "LO", true // consumes the carry (x86 borrow after a compare)
	}
	if opcode != "JMP" && strings.HasPrefix(opcode, "J") {
		return armCond(strings.TrimPrefix(opcode, "J"))
	}
	return "", false
}

// subwordSafeEqNe identifies, for every straight-line CMPW/CMPB/TESTW/TESTB
// producer, whether every consumer that reads its flags uses only EQ/NE.
// Sub-word CMP/TEST are otherwise unsupported (see lower()) because ordering
// conditions need sign-aware operand extension this printer does not model;
// equality doesn't depend on sign, so this narrow case can be lowered safely
// regardless of the surrounding dataflow (see lowerSubwordCompareEqNe).
//
// The scan mirrors flagProducers but runs forward: from each producer, walk
// over flag-transparent instructions (CMOVcc/Jcc chain onto the same flags)
// collecting every consumer found, until a non-flag-transparent instruction,
// a label, or the end of the block. A producer with multiple consumers (e.g.
// two chained conditional branches) requires ALL of them to be EQ/NE.
func subwordSafeEqNe(nodes []ir.Node) map[int]bool {
	safe := make(map[int]bool)
	for j, n := range nodes {
		ins, ok := n.(*ir.Instruction)
		if !ok || !isSubwordCompare(ins.Opcode) {
			continue
		}
		eqne, hasConsumer := true, false
	consumers:
		for k := j + 1; k < len(nodes); k++ {
			switch next := nodes[k].(type) {
			case *ir.Comment:
				continue
			case *ir.Instruction:
				if cond, isConsumer := consumerCondition(next.Opcode); isConsumer {
					hasConsumer = true
					if cond != "EQ" && cond != "NE" {
						eqne = false
					}
					// Chained consumers may follow, so keep scanning. True of
					// Jcc/SETcc/CMOVcc, which read NZCV and never write it. ADC is
					// the exception -- on x86 it rewrites all of EFLAGS -- and
					// treating it as non-writing here would let a later consumer be
					// attributed to this producer. Two things stop that today: an
					// ADC consumer forces eqne false (its condition is LO, not
					// EQ/NE), and any consumer after an ADC panics in the backward
					// scan because flagSetter rejects ADCB. Both are load-bearing;
					// give ADC an explicit case here if either changes.
					continue
				}
				if isFlagTransparent(next.Opcode) {
					// Neither reads nor writes flags (a MOV, an LEA). The backward
					// scan in flagProducers skips these; this one must too, or a
					// producer whose consumer sits past one is wrongly blessed as
					// EQ/NE-only and a signed sub-word compare lowers to unsigned.
					continue
				}
				hasConsumer = true // a genuine flag writer: this producer's flags are dead
				break consumers
			default:
				break consumers // label: producer/consumer link does not cross blocks
			}
		}
		safe[j] = hasConsumer && eqne
	}
	return safe
}

// shiftExtractFold identifies straight-line "MOVQ/MOVL src, tmp; SHRQ/SHRL
// $n, tmp; MOVBQZX/MOVBLZX tmp, tmp" triples and returns the set of node
// indices where the triple starts. This is x86's two-address-register idiom
// for extracting one byte at a fixed bit offset -- needed because x86 has no
// single "extract this byte" instruction -- which arm64's UBFX does in one
// step (see emitShiftExtractFold).
//
// The match is deliberately narrow: all three instructions must be strictly
// adjacent (no comment, label, or other instruction between them -- unlike
// subwordSafeEqNe/flagProducers above, this does not skip comments, so a
// program that puts one there simply is not folded), and the extend must
// write back into the EXACT register the shift shifted (tmp == its own
// destination). That last restriction is what makes the fold provably safe
// with no liveness scan: nothing between the three instructions ever reads
// tmp (they are adjacent), and nothing after them can observe the shifted
// intermediate value either, because the extend is the last write to that
// register in the window and folding preserves that final value exactly --
// UBFX computes the same bits the three instructions compute, just without
// materializing the intermediate. A triple whose extend targets a DIFFERENT
// register is refused outright rather than folded some other way: proving
// that shape safe would require showing nothing downstream reads tmp's
// shifted value, which this pass does not attempt.
//
// Flags are not reasoned about here because they do not need to be: SHRQ/SHRL
// is not in flagSetter's table, so flagProducers already panics at generation
// time if anything downstream ever consumed its flags, before this fold's
// output is even considered. A program that reaches this pass without
// panicking is one where that never happens.
func shiftExtractFold(nodes []ir.Node) map[int]bool {
	fold := make(map[int]bool)
	for j := 0; j+2 < len(nodes); j++ {
		mov, ok := nodes[j].(*ir.Instruction)
		if !ok || (mov.Opcode != "MOVQ" && mov.Opcode != "MOVL") || len(mov.Operands) != 2 {
			continue
		}
		movSrc, srcOK := mov.Operands[0].(reg.Register)
		movTmp, tmpOK := mov.Operands[1].(reg.Register)
		if !srcOK || !tmpOK || isHighByte(movSrc) || isHighByte(movTmp) {
			continue
		}
		wantShr, wantExt, width := "SHRQ", "MOVBQZX", int64(64)
		if mov.Opcode == "MOVL" {
			wantShr, wantExt, width = "SHRL", "MOVBLZX", 32
		}
		shr, ok := nodes[j+1].(*ir.Instruction)
		if !ok || shr.Opcode != wantShr || len(shr.Operands) != 2 {
			continue
		}
		shiftAmt, ok := immVal(shr.Operands[0])
		if !ok || shiftAmt < 0 || shiftAmt > width-8 {
			continue
		}
		shrDst, ok := shr.Operands[1].(reg.Register)
		if !ok || isHighByte(shrDst) || rename(shrDst) != rename(movTmp) {
			continue
		}
		ext, ok := nodes[j+2].(*ir.Instruction)
		if !ok || ext.Opcode != wantExt || len(ext.Operands) != 2 {
			continue
		}
		extSrc, ok := ext.Operands[0].(reg.Register)
		if !ok || isHighByte(extSrc) || rename(extSrc) != rename(movTmp) {
			continue
		}
		extDst, ok := ext.Operands[1].(reg.Register)
		if !ok || rename(extDst) != rename(movTmp) {
			continue // only the self-referential form is folded; see doc comment
		}
		fold[j] = true
	}
	return fold
}

// emitShiftExtractFold emits the single-instruction collapse for a triple
// shiftExtractFold matched at some index j: mov, shr and ext are
// nodes[j], nodes[j+1] and nodes[j+2].
func (p *arm64) emitShiftExtractFold(mov, shr, ext *ir.Instruction) {
	shiftAmt, _ := immVal(shr.Operands[0]) // shiftExtractFold already validated this
	// Neither UBFX nor any instruction it could plausibly be confused with
	// writes NZCV (see arm64WritesNZCV), and shiftExtractFold's doc comment
	// covers why nothing in this window can be a live flag producer, so there
	// is nothing to set p.inProducer/p.inTransparent to but their neutral
	// values.
	p.inTransparent, p.transparentOp = false, ""
	p.inProducer, p.producerOp, p.writerCount = false, "", 0
	p.emitf("UBFX $%d, %s, $8, %s", shiftAmt, operandReg(mov.Operands[0]), operandReg(ext.Operands[1]))
	p.constOK = false
}

// ---- straight-line register-copy folds ----
//
// x86 has two-address shifts whose variable count must sit in CL, so avo
// programs are full of "MOVQ n, CX; MOVQ x, y; SHLQ CL, y": copy the count
// into CL, copy the source so it survives, shift in place. arm64 shifts are
// three-operand and take the count from any register, so both copies are
// redundant there. The folds below absorb them where a straight-line scan
// can prove nothing else observes the copy.
//
// The reasoning is deliberately local. A scan runs forward from the copy
// over instructions only -- never across a label, which is where another
// path could join, and never past a branch or return, beyond which the
// copy's value might be read on the path not taken. Within that run every
// instruction's register reads and writes come from the IR's Inputs and
// Outputs, which include implicit operands (CL for a shift, RAX/RDX for a
// widening multiply) and the base and index of memory operands, and
// registers are compared as families, so a write to CL or ECX counts as a
// write to RCX.

// shiftFold is the operand rewrite shiftFolds decided for one shift: the arm64
// register to read as the count and/or as the source instead of the copy the
// x86 program made. Either may be empty.
type shiftFold struct {
	count string
	src   string
}

// regFamily returns the physical index of a general-purpose register operand,
// which identifies the register regardless of the width the operand names (AL,
// AH, EAX and RAX are all index 0), or -1 for anything else: a pseudo register,
// a vector register, an immediate or a memory operand.
func regFamily(op operand.Op) int {
	r, ok := op.(reg.Register)
	if !ok || r.Kind() != reg.KindGP {
		return -1
	}
	ph, ok := r.(reg.Physical)
	if !ok {
		return -1
	}
	return int(ph.PhysicalIndex())
}

// isFullGP reports whether op names a whole 64-bit general-purpose register.
func isFullGP(op operand.Op) bool {
	r, ok := op.(reg.Register)
	return ok && regFamily(op) >= 0 && r.Size() == 8
}

// readsFamily reports whether ins reads any register of the given family,
// including through a memory operand's base or index and through implicit
// operands.
func readsFamily(ins *ir.Instruction, family int) bool {
	for _, r := range ins.InputRegisters() {
		if regFamily(r) == family {
			return true
		}
	}
	return false
}

// writesFamily reports whether ins writes any register of the given family,
// at any width.
func writesFamily(ins *ir.Instruction, family int) bool {
	for _, r := range ins.OutputRegisters() {
		if regFamily(r) == family {
			return true
		}
	}
	return false
}

// endsBlock reports whether control may leave the straight-line run at ins:
// a return, an unconditional jump or a conditional branch. Nothing about the
// register state on the other side of one of these is known here.
func endsBlock(ins *ir.Instruction) bool {
	return ins.IsTerminal || ins.IsUnconditionalBranch() || isConditionalBranch(ins)
}

// isTwoOperandShift reports whether ins is one of the shifts and rotates that
// lowerShift/lowerROL handle: "OP count, dst" with a register destination. The
// three-operand double shifts (SHLQ $n, src, dst is x86 SHLD) are a different
// instruction and are excluded by the operand count.
func isTwoOperandShift(ins *ir.Instruction) bool {
	switch ins.Opcode {
	case "SHLQ", "SHRQ", "SARQ", "SHLL", "SHRL", "SARL",
		"ROLQ", "ROLL", "RORQ", "RORL":
		return len(ins.Operands) == 2 && regFamily(ins.Operands[1]) >= 0
	}
	return false
}

// rcx is the family index of RCX, whose low byte CL is the only register an
// x86 shift can take its variable count from.
const rcx = 1

// shiftFolds finds the copies that feed shifts and that the arm64 three-operand
// forms make redundant, returning the rewrite for each shift and the set of
// copy nodes to drop.
//
// Count fold: "MOVQ Rs, CX" followed, in the same straight-line run, by a
// shift whose count is CL, where RCX is not otherwise referenced before the
// shift, Rs is not written before the shift, and the next reference to RCX
// after the shift is an instruction that writes it without reading it. The
// shift then takes its count from Rs and the copy is dropped. Rs must be
// unchanged so the count is the value x86 would have read; RCX must be dead
// after the shift or its stale value would be observed; and the next reference
// must be a pure definition so the copy's value is provably never read again
// on this path. A shift whose destination is RCX itself also reads the copied
// value as data, and is refused rather than paired with the source fold.
//
// Source fold: "MOVQ Ra, Rb" followed, in the same run, by a shift with
// destination Rb, where Rb is not referenced and Ra is not written in between.
// The shift then reads Ra as its source and writes Rb, and the copy is
// dropped: the shift was the only reader of the copy, and it overwrites Rb
// with the same value x86 computes. When the shift's count is CL and Rb is
// RCX the count would alias the dropped copy, so that shape is refused.
//
// Both folds only ever remove a MOVQ and change which register a shift reads;
// nothing is reordered, so the flag analyses, which key on node index, are
// unaffected, and a MOVQ never produces flags.
func shiftFolds(nodes []ir.Node) (map[int]shiftFold, map[int]bool) {
	folds := make(map[int]shiftFold)
	dropped := make(map[int]bool)
	// next returns the index of the next instruction after k in the same
	// straight-line run, or -1 at a label or the end of the function.
	next := func(k int) int {
		for k++; k < len(nodes); k++ {
			switch nodes[k].(type) {
			case *ir.Comment:
				continue
			case *ir.Instruction:
				return k
			default:
				return -1
			}
		}
		return -1
	}
	for j, n := range nodes {
		mov, ok := n.(*ir.Instruction)
		if !ok || mov.Opcode != "MOVQ" || len(mov.Operands) != 2 ||
			!isFullGP(mov.Operands[0]) || !isFullGP(mov.Operands[1]) {
			continue
		}
		src, dst := regFamily(mov.Operands[0]), regFamily(mov.Operands[1])
		if src == dst {
			continue
		}
		if dst == rcx {
			if k, cnt := countFold(nodes, next, j, src); k >= 0 {
				f := folds[k]
				f.count = cnt
				folds[k] = f
				dropped[j] = true
				continue
			}
			// Not a count copy; it may still be the source of a shift with
			// an immediate count, which the source fold handles.
		}
		if k, s := sourceFold(nodes, next, j, src, dst); k >= 0 {
			f := folds[k]
			f.src = s
			folds[k] = f
			dropped[j] = true
		}
	}
	return folds, dropped
}

// countFold checks the count-fold conditions from the "MOVQ Rs, CX" at node j
// (see shiftFolds) and returns the shift's node index and the arm64 name of
// Rs, or -1.
func countFold(nodes []ir.Node, next func(int) int, j, src int) (int, string) {
	k := next(j)
	for ; k >= 0; k = next(k) {
		ins, ok := nodes[k].(*ir.Instruction)
		if !ok {
			panic("arm64: next() guarantees an instruction index")
		}
		if endsBlock(ins) {
			return -1, ""
		}
		if readsFamily(ins, rcx) || writesFamily(ins, rcx) {
			break
		}
		if writesFamily(ins, src) {
			return -1, ""
		}
	}
	if k < 0 {
		return -1, ""
	}
	shift, ok := nodes[k].(*ir.Instruction)
	if !ok {
		panic("arm64: next() guarantees an instruction index")
	}
	if !isTwoOperandShift(shift) || regFamily(shift.Operands[0]) != rcx ||
		regFamily(shift.Operands[1]) == rcx {
		return -1, ""
	}
	for m := next(k); m >= 0; m = next(m) {
		ins, ok := nodes[m].(*ir.Instruction)
		if !ok {
			panic("arm64: next() guarantees an instruction index")
		}
		if endsBlock(ins) {
			return -1, ""
		}
		if writesFamily(ins, rcx) && !readsFamily(ins, rcx) {
			return k, rename(nodes[j].(*ir.Instruction).Operands[0].(reg.Register))
		}
		if readsFamily(ins, rcx) || writesFamily(ins, rcx) {
			return -1, ""
		}
	}
	return -1, ""
}

// sourceFold checks the source-fold conditions from the "MOVQ Ra, Rb" at node
// j (see shiftFolds) and returns the shift's node index and the arm64 name of
// Ra, or -1.
func sourceFold(nodes []ir.Node, next func(int) int, j, src, dst int) (int, string) {
	for k := next(j); k >= 0; k = next(k) {
		ins, ok := nodes[k].(*ir.Instruction)
		if !ok {
			panic("arm64: next() guarantees an instruction index")
		}
		if endsBlock(ins) {
			return -1, ""
		}
		if isTwoOperandShift(ins) && regFamily(ins.Operands[1]) == dst {
			if regFamily(ins.Operands[0]) == dst {
				return -1, "" // count in CL, destination RCX: the count aliases the copy
			}
			return k, rename(nodes[j].(*ir.Instruction).Operands[0].(reg.Register))
		}
		if readsFamily(ins, dst) || writesFamily(ins, dst) || writesFamily(ins, src) {
			return -1, ""
		}
	}
	return -1, ""
}

// setccFolds finds "zero r; ...; SETcc r8" runs where the SETcc's full-width
// result is known: r is zeroed (XORQ/XORL r, r or MOVQ/MOVL $0, r), nothing
// before the SETcc in the same straight-line run references r, and the
// zeroing's own flags are not consumed (a zeroing XOR sets ZF on x86, and
// zeroSelf would have to emit a TST for a consumer). The SETcc then writes the
// whole register as 0 or 1, which is exactly the state x86 leaves, and the
// zeroing is dropped. The zeroing MOVQ $0 would also have opened the constant
// window; a consumer of that window reads r, which the scan refuses.
func setccFolds(nodes []ir.Node, setflags map[int]bool, dropped map[int]bool) map[int]bool {
	full := make(map[int]bool)
	for j, n := range nodes {
		zero, ok := n.(*ir.Instruction)
		if !ok || len(zero.Operands) != 2 || setflags[j] {
			continue
		}
		r := regFamily(zero.Operands[1])
		if r < 0 {
			continue
		}
		switch zero.Opcode {
		case "XORQ", "XORL":
			if regFamily(zero.Operands[0]) != r {
				continue
			}
		case "MOVQ", "MOVL":
			if v, isImm := immVal(zero.Operands[0]); !isImm || v != 0 {
				continue
			}
		default:
			continue
		}
		for k := j + 1; k < len(nodes); k++ {
			ins, isInstr := nodes[k].(*ir.Instruction)
			if _, isComment := nodes[k].(*ir.Comment); isComment {
				continue
			}
			if !isInstr || endsBlock(ins) {
				break
			}
			if !readsFamily(ins, r) && !writesFamily(ins, r) {
				continue
			}
			if strings.HasPrefix(ins.Opcode, "SET") && !isHighByte(ins.Operands[0]) &&
				regFamily(ins.Operands[0]) == r && !readsFamily(ins, r) {
				full[k] = true
				dropped[j] = true
			}
			break
		}
	}
	return full
}
