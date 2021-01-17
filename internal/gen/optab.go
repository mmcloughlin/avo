package gen

import (
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
	t.opcodes(is)
	return t.Result()
}

func (t *optab) opcodes(is []inst.Instruction) {
	e := &enum{name: "Opcode"}
	for _, i := range is {
		e.values = append(e.values, i.Opcode)
	}
	e.Print(&t.Generator)
}
