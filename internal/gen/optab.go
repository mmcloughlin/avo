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

	// Operand types and implicit registers.
	t.operandTypes(is)
	t.implicitRegisters(is)

	// Opcodes table.
	t.opcodes(is)

	return t.Result()
}

func (t *optab) operandTypes(is []inst.Instruction) {
	e := &enum{name: "OperandType"}
	for _, t := range inst.OperandTypes(is) {
		e.values = append(e.values, api.OperandTypeIdentifier(t))
	}
	e.Print(&t.Generator)
}

func (t *optab) implicitRegisters(is []inst.Instruction) {
	e := &enum{name: "ImplicitRegister"}
	for _, r := range inst.ImplicitRegisters(is) {
		e.values = append(e.values, api.ImplicitRegisterIdentifier(r))
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
