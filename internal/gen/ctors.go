package gen

import (
	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/printer"
)

type ctors struct {
	cfg printer.Config
	prnt.Generator
}

// NewCtors will build instruction constructors. Each constructor will check
// that the provided operands match one of the allowed instruction forms. If so
// it will return an Instruction object that can be added to an avo Function.

func NewCtors(cfg printer.Config) Interface {
	return GoFmt(&ctors{cfg: cfg})
}

func (c *ctors) Generate(is []inst.Instruction) ([]byte, error) {
	c.Printf("// %s\n\n", c.cfg.GeneratedWarning())
	c.Printf("package x86\n\n")
	c.Printf("import (\n")
	c.Printf("\tintrep %q\n", api.ImportPath(api.IRPackage))
	c.Printf("\t%q\n", api.ImportPath(api.OperandPackage))
	c.Printf(")\n\n")

	fns := api.InstructionsFunctions(is)
	table := NewTable(is)
	for _, fn := range fns {
		c.function(fn, table)
	}

	return c.Result()
}

func (c *ctors) function(fn *api.Function, table *Table) {
	c.Comment(fn.Doc()...)

	s := fn.Signature()

	c.Printf("func %s(%s) (*intrep.Instruction, error) {\n", fn.Name(), s.ParameterList())
	c.Printf(
		"return BuildInstruction(%s.Forms(), %s, %s)\n",
		table.OpcodeConst(fn.Instruction.Opcode),
		table.SuffixesConst(fn.Suffixes),
		s.ParameterSlice(),
	)
	c.Printf("}\n\n")
}
