package gen

import (
	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/printer"
)

type optab struct {
	cfg printer.Config
	prnt.Generator
}

func NewOptab(cfg printer.Config) Interface {
	return GoFmt(&optab{cfg: cfg})
}

func (t *optab) Generate(is []inst.Instruction) ([]byte, error) {
	t.Printf("// %s\n\n", t.cfg.GeneratedWarning())
	t.Printf("package x86\n\n")
	t.Printf("import (\n")
	t.Printf("\t%q\n", api.ImportPath("operand"))
	t.Printf(")\n\n")

	// Arity.
	t.arity(is)

	// Operand types and implicit registers.
	t.operandTypes(is)
	t.implicitRegisters(is)

	// Suffixes.
	t.suffixes(is)

	// Opcodes table.
	t.opcodes(is)

	return t.Result()
}

func (t *optab) arity(is []inst.Instruction) {
	max := 0
	for _, i := range inst.Instructions {
		for _, f := range i.Forms {
			a := len(f.Operands) + len(f.ImplicitOperands)
			if a > max {
				max = a
			}
		}
	}

	t.Printf("const MaxArity = %d\n", max)
}

func (t *optab) operandTypes(is []inst.Instruction) {
	types := inst.OperandTypes(is)

	// Operand type enum.
	e := &enum{name: "OperandType"}
	for _, t := range types {
		e.values = append(e.values, api.OperandTypeIdentifier(t))
	}
	e.Print(&t.Generator)

	// Operand match function.
	t.Printf("func (%s %s) Match(op %s) bool {\n", e.Receiver(), e.Name(), api.OperandType)
	t.Printf("\tswitch %s {\n", e.Receiver())
	t.Printf("\t\tdefault: return false\n")
	for _, typ := range types {
		t.Printf("\t\tcase %s: return %s(op)\n", e.ConstName(api.OperandTypeIdentifier(typ)), api.CheckerName(typ))
	}
	t.Printf("\t}\n")
	t.Printf("}\n\n")
}

func (t *optab) implicitRegisters(is []inst.Instruction) {
	e := &enum{name: "ImplicitRegister"}
	for _, r := range inst.ImplicitRegisters(is) {
		e.values = append(e.values, api.ImplicitRegisterIdentifier(r))
	}
	e.Print(&t.Generator)
}

func (t *optab) suffixes(is []inst.Instruction) {
	e := &enum{name: "Suffix"}
	for _, s := range inst.UniqueSuffixes(is) {
		e.values = append(e.values, s.String())
	}
	e.Print(&t.Generator)
}

func (t *optab) opcodes(is []inst.Instruction) {
	e := &enum{name: "Opcode"}
	for _, i := range is {
		e.values = append(e.values, i.Opcode)
	}
	e.Print(&t.Generator)
	e.StringMethod(&t.Generator)
}
