package gen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/inst"
)

// Enum is a generated enumeration type. This assists with mapping between the
// conceptual values of the enum, and it's materialization as Go code.
type Enum struct {
	name   string
	doc    []string
	values []string
}

// NewEnum initializes an empty enum type with the given name.
func NewEnum(name string) *Enum {
	return &Enum{name: name}
}

// Name returns the type name.
func (e *Enum) Name() string { return e.name }

// Receiver returns the receiver variable name.
func (e *Enum) Receiver() string {
	return strings.ToLower(e.name[:1])
}

// SetDoc sets type documentation, as a list of lines.
func (e *Enum) SetDoc(doc ...string) {
	e.doc = doc
}

// Doc returns the type documentation.
func (e *Enum) Doc() []string { return e.doc }

// AddValue adds a named enumerator.
func (e *Enum) AddValue(value string) {
	e.values = append(e.values, value)
}

// Values returns all enumerators.
func (e *Enum) Values() []string { return e.values }

// None returns the name of the "unset" constant of this enumeration.
func (e *Enum) None() string {
	return e.ConstName("None")
}

// ConstName returns the constant name that refers to the given enumerator
// value.
func (e *Enum) ConstName(value string) string {
	return e.name + value
}

// ConstNames returns the constant names for all enumerator values.
func (e *Enum) ConstNames() []string {
	var consts []string
	for _, v := range e.values {
		consts = append(consts, e.ConstName(v))
	}
	return consts
}

// MaxName returns the name of the constant that represents the maximum
// enumerator. This value is placed at the very end of the enum, so all values
// will be between the None and Max enumerators.
func (e *Enum) MaxName() string {
	return strings.ToLower(e.ConstName("max"))
}

// Max returns the value of the maximum enumerator.
func (e *Enum) Max() int {
	return len(e.values)
}

// UnderlyingType returns the underlying unsigned integer type used for the
// enumeration. This will be the smallest type that can represent all the
// values.
func (e *Enum) UnderlyingType() string {
	b := uint(8)
	for ; b < 64 && e.Max() > ((1<<b)-1); b <<= 1 {
	}
	return fmt.Sprintf("uint%d", b)
}

// Table represents all the types required to represent the instruction
// operation table (optab).
type Table struct {
	instructions []inst.Instruction

	operandType      *Enum
	implicitRegister *Enum
	suffix           *Enum
	suffixesClass    *Enum
	isas             *Enum
	opcode           *Enum
}

// NewTable builds optab types to represent the given instructions.
func NewTable(is []inst.Instruction) *Table {
	t := &Table{instructions: is}
	t.init()
	return t
}

func (t *Table) init() {
	// Operand type.
	t.operandType = NewEnum("oprndtype")
	types := inst.OperandTypes(t.instructions)
	for _, typ := range types {
		t.operandType.AddValue(api.OperandTypeIdentifier(typ))
	}

	// Implicit register.
	registers := inst.ImplicitRegisters(t.instructions)
	t.implicitRegister = NewEnum("implreg")
	for _, r := range registers {
		t.implicitRegister.AddValue(api.ImplicitRegisterIdentifier(r))
	}

	// Suffix.
	t.suffix = NewEnum("sffx")
	for _, s := range inst.UniqueSuffixes(t.instructions) {
		t.suffix.AddValue(s.String())
	}

	// Suffixes class.
	t.suffixesClass = NewEnum("sffxscls")

	classes := inst.SuffixesClasses(t.instructions)
	keys := make([]string, 0, len(classes))
	for key := range classes {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		t.suffixesClass.AddValue(api.SuffixesClassIdentifier(key))
	}

	// ISAs.
	t.isas = NewEnum("isas")
	for _, isas := range inst.ISACombinations(t.instructions) {
		t.isas.AddValue(api.ISAsIdentifier(isas))
	}

	// Opcodes.
	t.opcode = NewEnum("opc")
	for _, i := range t.instructions {
		t.opcode.AddValue(i.Opcode)
	}
}

// OperandType returns the enumeration representing all possible operand types.
func (t *Table) OperandType() *Enum { return t.operandType }

// OperandTypeConst returns the constant name for the given operand type.
func (t *Table) OperandTypeConst(typ string) string {
	return t.operandType.ConstName(api.OperandTypeIdentifier(typ))
}

// ImplicitRegister returns the enumeration representing all possible operand
// types.
func (t *Table) ImplicitRegister() *Enum { return t.implicitRegister }

// ImplicitRegisterConst returns the constant name for the given register.
func (t *Table) ImplicitRegisterConst(r string) string {
	return t.implicitRegister.ConstName(api.ImplicitRegisterIdentifier(r))
}

// Suffix returns the enumeration representing instruction suffixes.
func (t *Table) Suffix() *Enum { return t.suffix }

// SuffixConst returns the constant name for the given instruction suffix.
func (t *Table) SuffixConst(s inst.Suffix) string { return t.suffix.ConstName(s.String()) }

// SuffixesTypeName returns the name of the array type for a list of suffixes.
func (t *Table) SuffixesTypeName() string {
	return t.Suffix().Name() + "s"
}

// SuffixesConst returns the constant for a list of suffixes. Suffixes is a
// generated array type, so the list is a value not slice type.
func (t *Table) SuffixesConst(suffixes inst.Suffixes) string {
	return t.SuffixesTypeName() + t.SuffixesList(suffixes)
}

// SuffixesList returns the constant literal for a list of suffixes, type name
// not included. Use SuffxesConst if the type is required.
func (t *Table) SuffixesList(suffixes inst.Suffixes) string {
	var parts []string
	for _, suffix := range suffixes {
		parts = append(parts, t.SuffixConst(suffix))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// SuffixesClass returns the enumeration representing all suffixes classes.
func (t *Table) SuffixesClass() *Enum { return t.suffixesClass }

// SuffixesClassConst returns the constant name for a given suffixes class. The
// key is as returned from inst.SuffixesClasses() function.
func (t *Table) SuffixesClassConst(key string) string {
	ident := api.SuffixesClassIdentifier(key)
	if ident == "" {
		return t.suffixesClass.None()
	}
	return t.suffixesClass.ConstName(ident)
}

// ISAs returns the enumeration for all possible ISA combinations.
func (t *Table) ISAs() *Enum { return t.isas }

// ISAsConst returns the constant name for the given ISA combination.
func (t *Table) ISAsConst(isas []string) string {
	return t.isas.ConstName(api.ISAsIdentifier(isas))
}

// Opcode returns the opcode enumeration.
func (t *Table) Opcode() *Enum { return t.opcode }

// OpcodeConst returns the constant name for the given opcode.
func (t *Table) OpcodeConst(opcode string) string {
	return t.opcode.ConstName(opcode)
}

// Features returns code for the features constant describing the features of
// the given instruction form.
func Features(i inst.Instruction, f inst.Form) string {
	var enabled []string
	for _, feature := range []struct {
		Name    string
		Enabled bool
	}{
		{"Terminal", i.IsTerminal()},
		{"Branch", i.IsBranch()},
		{"ConditionalBranch", i.IsConditionalBranch()},
		{"CancellingInputs", f.CancellingInputs},
	} {
		if feature.Enabled {
			enabled = append(enabled, "feature"+feature.Name)
		}
	}

	if len(enabled) == 0 {
		return "0"
	}
	return strings.Join(enabled, "|")
}

// Action returns code representing the instruction action.
func Action(a inst.Action) string {
	c := strings.ToUpper(a.String())
	if c == "" {
		c = "N"
	}
	return "action" + c
}
