package inst

import (
	"sort"
	"strings"
)

// Instruction represents an x86 instruction.
type Instruction struct {
	Opcode  string // Golang assembly mnemonic
	AliasOf string // Opcode of instruction that this is an alias for
	Summary string // Description of the instruction
	Forms          // Accepted operand forms
}

// IsTerminal reports whether the instruction exits a function.
func (i Instruction) IsTerminal() bool {
	// TODO(mbm): how about the RETF* instructions
	return i.Opcode == "RET"
}

// IsBranch reports whether the instruction is a branch; that is, if it can
// cause control flow to jump to another location.
func (i Instruction) IsBranch() bool {
	if i.Opcode == "CALL" {
		return false
	}
	for _, f := range i.Forms {
		for _, op := range f.Operands {
			if strings.HasPrefix(op.Type, "rel") {
				return true
			}
		}
	}
	return false
}

// IsConditionalBranch reports whether the instruction branches dependent on some condition.
func (i Instruction) IsConditionalBranch() bool {
	return i.IsBranch() && i.Opcode != "JMP"
}

// Forms is a collection of instruction forms.
type Forms []Form

// Arities returns the unique arities among the instruction forms.
func (fs Forms) Arities() []int {
	s := map[int]bool{}
	for _, f := range fs {
		s[f.Arity()] = true
	}
	a := make([]int, 0, len(s))
	for n := range s {
		a = append(a, n)
	}
	sort.Ints(a)
	return a
}

// Arity is a convenience for returning the unique instruction arity when you
// know it is not variadic. Panics for a variadic instruction.
func (fs Forms) Arity() int {
	if fs.IsVariadic() {
		panic("variadic")
	}
	a := fs.Arities()
	return a[0]
}

// IsVariadic reports whether the instruction has more than one arity.
func (fs Forms) IsVariadic() bool {
	return len(fs.Arities()) > 1
}

// IsNiladic reports whether the instruction takes no operands.
func (fs Forms) IsNiladic() bool {
	a := fs.Arities()
	return len(a) == 1 && a[0] == 0
}

// Form specifies one accepted set of operands for an instruction.
type Form struct {
	// Instruction sets this instruction form requires.
	ISA []string

	// Operands required for this form.
	Operands []Operand

	// Registers read or written but not explicitly passed to the instruction.
	ImplicitOperands []ImplicitOperand

	// Encoding type required for this instruction form.
	EncodingType EncodingType

	// CancellingInputs indicates this instruction form has no dependency on the
	// input operands when they refer to the same register. The classic example of
	// this is "XORQ RAX, RAX", in which case the output has no dependence on the
	// value of RAX. Instruction forms with cancelling inputs have only two input
	// operands, which have the same register type.
	CancellingInputs bool

	// Zeroing indicates whether the instruction form uses AVX-512 zeroing. This
	// is the .Z suffix in Go, usually indicated with {z} operand suffix in
	// Intel manuals.
	Zeroing bool

	// EmbeddedRounding indicates whether the instruction form uses AVX-512
	// embedded rounding. This is the RN_SAE, RZ_SAE, RD_SAE and RU_SAE suffixes
	// in Go, usually indicated with {er} in Intel manuals.
	EmbeddedRounding bool

	// SuppressAllExceptions indicates whether the instruction form uses AVX-512
	// "suppress all exceptions". This is the SAE suffix in Go, usually
	// indicated with {sae} in Intel manuals.
	SuppressAllExceptions bool

	// Broadcast indicates whether the instruction form uses AVX-512
	// broadcast. This is the BCST suffix in Go, usually indicated by operand
	// types like "m64bcst" in Intel manuals.
	Broadcast bool
}

// Arity returns the number of operands this form expects.
func (f Form) Arity() int {
	return len(f.Operands)
}

// Signature returns the list of operand types.
func (f Form) Signature() []string {
	s := make([]string, f.Arity())
	for i, op := range f.Operands {
		s[i] = op.Type
	}
	return s
}

// Clone the instruction form.
func (f Form) Clone() Form {
	c := f
	c.ISA = append([]string(nil), f.ISA...)
	c.Operands = append([]Operand(nil), f.Operands...)
	c.ImplicitOperands = append([]ImplicitOperand(nil), f.ImplicitOperands...)
	return c
}

// AcceptsSuffixes reports whether this form takes any opcode suffixes.
func (f Form) AcceptsSuffixes() bool {
	return f.Broadcast || f.EmbeddedRounding || f.SuppressAllExceptions || f.Zeroing
}

// SuffixesClass returns a key representing the class of instruction suffixes it
// accepts. All instructions sharing a suffix class accept the same suffixes.
func (f Form) SuffixesClass() string {
	if !f.AcceptsSuffixes() {
		return "nil"
	}
	var parts []string
	for _, flag := range []struct {
		Name    string
		Enabled bool
	}{
		{"er", f.EmbeddedRounding},
		{"sae", f.SuppressAllExceptions},
		{"bcst", f.Broadcast},
		{"z", f.Zeroing},
	} {
		if flag.Enabled {
			parts = append(parts, flag.Name)
		}
	}
	return strings.Join(parts, "_")
}

// SupportedSuffixes returns the list of all possible suffix combinations
// supported by this instruction form.
func (f Form) SupportedSuffixes() []Suffixes {
	suffixes := []Suffixes{
		{},
	}

	add := func(ss ...Suffix) {
		var exts []Suffixes
		for _, s := range ss {
			for _, suffix := range suffixes {
				ext := append(Suffixes(nil), suffix...)
				ext = append(ext, s)
				exts = append(exts, ext)
			}
		}
		suffixes = exts
	}

	if f.Broadcast {
		add(BCST)
	}

	if f.EmbeddedRounding {
		add(RN_SAE, RZ_SAE, RD_SAE, RU_SAE)
	}

	if f.SuppressAllExceptions {
		add(SAE)
	}

	if f.Zeroing {
		add(Z)
	}

	return suffixes
}

// Suffix is an opcode suffix.
type Suffix string

// Supported opcode suffixes in x86 assembly.
const (
	BCST   Suffix = "BCST"
	RN_SAE Suffix = "RN_SAE"
	RZ_SAE Suffix = "RZ_SAE"
	RD_SAE Suffix = "RD_SAE"
	RU_SAE Suffix = "RU_SAE"
	SAE    Suffix = "SAE"
	Z      Suffix = "Z"
)

func (s Suffix) String() string {
	return string(s)
}

// Summary of the opcode suffix, for documentation purposes.
func (s Suffix) Summary() string {
	return suffixsummary[s]
}

var suffixsummary = map[Suffix]string{
	BCST:   "Broadcast",
	RN_SAE: "Round Towards Nearest",
	RZ_SAE: "Round Towards Zero",
	RD_SAE: "Round Towards Negative Infinity",
	RU_SAE: "Round Towards Positive Infinity",
	SAE:    "Suppress All Exceptions",
	Z:      "Zeroing Masking",
}

// Suffixes is a list of opcode suffixes.
type Suffixes []Suffix

// String returns the dot-separated suffixes.
func (s Suffixes) String() string { return s.Join(".") }

// Join suffixes with the given separator.
func (s Suffixes) Join(sep string) string {
	return strings.Join(s.Strings(), sep)
}

// Strings returns the suffixes as strings.
func (s Suffixes) Strings() []string {
	var ss []string
	for _, suffix := range s {
		ss = append(ss, suffix.String())
	}
	return ss
}

// Summaries returns all the suffix summaries.
func (s Suffixes) Summaries() []string {
	var summaries []string
	for _, suffix := range s {
		summaries = append(summaries, suffix.Summary())
	}
	return summaries
}

// Operand is an operand to an instruction, describing the expected type and read/write action.
type Operand struct {
	Type   string
	Action Action
}

// ImplicitOperand describes a register that is implicitly read/written by an instruction.
type ImplicitOperand struct {
	Register string
	Action   Action
}

// Action specifies the read/write operation of an instruction on an operand.
type Action uint8

// Possible Action types.
const (
	R Action = 1 << iota // Read
	W                    // Write

	RW Action = R | W // Read-Write
)

// ActionFromReadWrite builds an Action from boolean flags.
func ActionFromReadWrite(r, w bool) Action {
	var a Action
	if r {
		a |= R
	}
	if w {
		a |= W
	}
	return a
}

// ContainsAll reports whether a supports all actions in s.
func (a Action) ContainsAll(s Action) bool {
	return (a & s) == s
}

// ContainsAny reports whether a supports any actions in s.
func (a Action) ContainsAny(s Action) bool {
	return (a & s) != 0
}

// Read reports whether a supports read.
func (a Action) Read() bool {
	return a.ContainsAll(R)
}

// Write reports whether a supports write.
func (a Action) Write() bool {
	return a.ContainsAll(W)
}

// String represents a as a human-readable string.
func (a Action) String() string {
	s := ""
	if a.Read() {
		s += "r"
	}
	if a.Write() {
		s += "w"
	}
	return s
}

// EncodingType specifies a category of encoding types.
type EncodingType uint8

// Supported encoding types.
const (
	EncodingTypeLegacy EncodingType = 1 + iota
	EncodingTypeREX
	EncodingTypeVEX
	EncodingTypeEVEX
)
