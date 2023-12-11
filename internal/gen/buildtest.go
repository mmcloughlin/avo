package gen

import (
	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/printer"
)

type buildtest struct {
	cfg printer.Config
	printer.Generator
}

// NewBuildTest autogenerates tests for instruction methods on the build
// context.
func NewBuildTest(cfg printer.Config) Interface {
	return GoFmt(&buildtest{cfg: cfg})
}

func (b *buildtest) Generate(is []inst.Instruction) ([]byte, error) {
	b.Printf("// %s\n\n", b.cfg.GeneratedWarning())
	b.BuildConstraint("!integration")
	b.NL()
	b.Printf("package build\n\n")
	b.Printf("import (\n")
	b.Printf("\t\"math\"\n")
	b.Printf("\t\"testing\"\n")
	b.NL()
	b.Printf("\t%q\n", api.ImportPath(api.OperandPackage))
	b.Printf("\t%q\n", api.ImportPath(api.RegisterPackage))
	b.Printf(")\n\n")

	DeclareTestArguments(&b.Generator)

	b.Printf("func TestContextInstructions(t *testing.T) {")
	b.Printf("ctx := NewContext()\n")
	b.Printf("ctx.Function(\"Instructions\")\n")

	fns := api.InstructionsFunctions(is)
	for _, fn := range fns {
		f := fn.Forms[0]
		s := TestSignature(f)
		b.Printf("ctx.%s(%s)\n", fn.Name(), s.Arguments())
	}

	b.Printf("if _, err := ctx.Result(); err != nil { t.Fatal(err) }\n")
	b.Printf("}\n\n")

	return b.Result()
}
