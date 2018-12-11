package gen

import (
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/printer"
)

type godata struct {
	cfg printer.Config
	prnt.Generator
}

func NewGoData(cfg printer.Config) Interface {
	return GoFmt(&godata{cfg: cfg})
}

func (g *godata) Generate(is []inst.Instruction) ([]byte, error) {
	g.Printf("// %s\n\n", g.cfg.GeneratedWarning())
	g.Printf("package inst\n\n")

	g.Printf("var Instructions = []Instruction{\n")

	for _, i := range is {
		g.Printf("{\n")

		g.Printf("Opcode: %#v,\n", i.Opcode)
		if i.AliasOf != "" {
			g.Printf("AliasOf: %#v,\n", i.AliasOf)
		}
		g.Printf("Summary: %#v,\n", i.Summary)

		g.Printf("Forms: []Form{\n")
		for _, f := range i.Forms {
			g.Printf("{\n")

			if f.ISA != nil {
				g.Printf("ISA: %#v,\n", f.ISA)
			}

			if f.Operands != nil {
				g.Printf("Operands: []Operand{\n")
				for _, op := range f.Operands {
					g.Printf("{Type: %#v, Action: %#v},\n", op.Type, op.Action)
				}
				g.Printf("},\n")
			}

			if f.ImplicitOperands != nil {
				g.Printf("ImplicitOperands: []ImplicitOperand{\n")
				for _, op := range f.ImplicitOperands {
					g.Printf("{Register: %#v, Action: %#v},\n", op.Register, op.Action)
				}
				g.Printf("},\n")
			}

			g.Printf("},\n")
		}
		g.Printf("},\n")

		g.Printf("},\n")
	}

	g.Printf("}\n")

	return g.Result()
}

type godatatest struct {
	cfg printer.Config
	prnt.Generator
}

func NewGoDataTest(cfg printer.Config) Interface {
	return GoFmt(&godatatest{cfg: cfg})
}

func (g *godatatest) Generate(is []inst.Instruction) ([]byte, error) {
	g.Printf("// %s\n\n", g.cfg.GeneratedWarning())
	g.Printf("package inst_test\n\n")

	g.Printf(`import (
		"reflect"
		"testing"

		"%s/internal/inst"
	)
	`, pkg)

	g.Printf("var raw = %#v\n\n", is)

	g.Printf(`func TestVerifyInstructionsList(t *testing.T) {
		if !reflect.DeepEqual(raw, inst.Instructions) {
			t.Fatal("bad code generation for instructions list")
		}
	}
	`)

	return g.Result()
}
