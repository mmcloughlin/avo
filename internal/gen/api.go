package gen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/inst"
)

type Enum struct {
	name   string
	doc    []string
	values []string
}

func NewEnum(name string) *Enum {
	return &Enum{name: name}
}

func (e *Enum) Name() string { return e.name }

func (e *Enum) Receiver() string {
	return strings.ToLower(e.name[:1])
}

func (e *Enum) SetDoc(doc ...string) {
	e.doc = doc
}

func (e *Enum) Doc() []string { return e.doc }

func (e *Enum) AddValue(value string) {
	e.values = append(e.values, value)
}

func (e *Enum) Values() []string { return e.values }

func (e *Enum) None() string {
	return e.ConstName("None")
}

func (e *Enum) ConstName(value string) string {
	return e.name + value
}

func (e *Enum) ConstNames() []string {
	var consts []string
	for _, v := range e.values {
		consts = append(consts, e.ConstName(v))
	}
	return consts
}

func (e *Enum) MaxName() string {
	return strings.ToLower(e.ConstName("max"))
}

func (e *Enum) Max() int {
	return len(e.values)
}

func (e *Enum) UnderlyingType() string {
	b := uint(8)
	for ; b < 64 && e.Max() > ((1<<b)-1); b <<= 1 {
	}
	return fmt.Sprintf("uint%d", b)
}

type Table struct {
	instructions []inst.Instruction

	operandType      *Enum
	implicitRegister *Enum
	suffix           *Enum
	suffixesClass    *Enum
	isas             *Enum
	opcode           *Enum
}

func NewTable(is []inst.Instruction) *Table {
	t := &Table{instructions: is}
	t.init()
	return t
}

func (t *Table) init() {
	// Operand type.
	t.operandType = NewEnum("OperandType")
	types := inst.OperandTypes(t.instructions)
	for _, typ := range types {
		t.operandType.AddValue(api.OperandTypeIdentifier(typ))
	}

	// Implicit register.
	registers := inst.ImplicitRegisters(t.instructions)
	t.implicitRegister = NewEnum("ImplicitRegister")
	for _, r := range registers {
		t.implicitRegister.AddValue(api.ImplicitRegisterIdentifier(r))
	}

	// Suffix.
	t.suffix = NewEnum("Suffix")
	for _, s := range inst.UniqueSuffixes(t.instructions) {
		t.suffix.AddValue(s.String())
	}

	// Suffixes class.
	t.suffixesClass = NewEnum("SuffixesClass")

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
	t.isas = NewEnum("ISAs")
	for _, isas := range inst.ISACombinations(t.instructions) {
		t.isas.AddValue(api.ISAsIdentifier(isas))
	}

	// Opcodes.
	t.opcode = NewEnum("Opcode")
	for _, i := range t.instructions {
		t.opcode.AddValue(i.Opcode)
	}
}

func (t *Table) OperandType() *Enum { return t.operandType }

func (t *Table) OperandTypeConst(typ string) string {
	return t.operandType.ConstName(api.OperandTypeIdentifier(typ))
}

func (t *Table) ImplicitRegister() *Enum { return t.implicitRegister }

func (t *Table) ImplicitRegisterConst(r string) string {
	return t.implicitRegister.ConstName(api.ImplicitRegisterIdentifier(r))
}

func (t *Table) Suffix() *Enum { return t.suffix }

func (t *Table) SuffixConst(s inst.Suffix) string { return t.suffix.ConstName(s.String()) }

func (t *Table) SuffixesConst(suffixes inst.Suffixes) string {
	var parts []string
	for _, suffix := range suffixes {
		parts = append(parts, t.SuffixConst(suffix))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func (t *Table) SuffixesClass() *Enum { return t.suffixesClass }

func (t *Table) SuffixesClassConst(key string) string {
	ident := api.SuffixesClassIdentifier(key)
	if ident == "" {
		return t.suffixesClass.None()
	}
	return t.suffixesClass.ConstName(ident)
}

func (t *Table) ISAs() *Enum { return t.isas }

func (t *Table) ISAsConst(isas []string) string {
	return t.isas.ConstName(api.ISAsIdentifier(isas))
}

func (t *Table) Opcode() *Enum { return t.opcode }

func (t *Table) OpcodeConst(opcode string) string {
	return t.opcode.ConstName(opcode)
}

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
			enabled = append(enabled, "Feature"+feature.Name)
		}
	}

	if len(enabled) == 0 {
		return "0"
	}
	return strings.Join(enabled, "|")
}

func Action(a inst.Action) string {
	c := strings.ToUpper(a.String())
	if c == "" {
		c = "None"
	}
	return "Action" + c
}
