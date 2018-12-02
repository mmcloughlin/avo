package gen

import "github.com/mmcloughlin/avo/internal/inst"

type build struct {
	cfg Config
	printer
}

func NewBuild(cfg Config) Interface {
	return GoFmt(&build{cfg: cfg})
}

func (b *build) Generate(is []inst.Instruction) ([]byte, error) {
	b.Printf("// %s\n\n", b.cfg.GeneratedWarning())
	b.Printf("package build\n\n")

	b.Printf("import (\n")
	b.Printf("\t\"%s/operand\"\n", pkg)
	b.Printf("\t\"%s/x86\"\n", pkg)
	b.Printf(")\n\n")

	for _, i := range is {
		b.instruction(i)
	}

	return b.Result()
}

func (b *build) instruction(i inst.Instruction) {
	s := params(i)

	// Context method.
	b.Printf("func (c *Context) %s(%s) {\n", i.Opcode, s.ParameterList())
	b.Printf("if inst, err := x86.%s(%s); err == nil", i.Opcode, s.Arguments())
	b.Printf(" { c.Instruction(*inst) }")
	b.Printf(" else { c.AddError(err) }\n")
	b.Printf("}\n")

	// Global version.
	b.Printf("func %s(%s) { ctx.%s(%s) }\n\n", i.Opcode, s.ParameterList(), i.Opcode, s.Arguments())
}
