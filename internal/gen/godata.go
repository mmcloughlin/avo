package gen

import (
	"github.com/mmcloughlin/avo/internal/inst"
)

type godata struct {
	cfg Config
}

func NewGoData(cfg Config) Interface {
	return GoFmt(godata{cfg: cfg})
}

func (g godata) Generate(is []inst.Instruction) ([]byte, error) {
	p := &printer{}

	p.Printf("// %s\n\n", g.cfg.GeneratedWarning())
	p.Printf("package inst\n\n")

	p.Printf("var Instructions = []Instruction{\n")

	for _, i := range is {
		p.Printf("{\n")

		p.Printf("Opcode: %#v,\n", i.Opcode)
		if i.AliasOf != "" {
			p.Printf("AliasOf: %#v,\n", i.AliasOf)
		}
		p.Printf("Summary: %#v,\n", i.Summary)

		p.Printf("Forms: []Form{\n")
		for _, f := range i.Forms {
			p.Printf("{\n")

			if f.ISA != nil {
				p.Printf("ISA: %#v,\n", f.ISA)
			}

			if f.Operands != nil {
				p.Printf("Operands: []Operand{\n")
				for _, op := range f.Operands {
					p.Printf("{Type: %#v, Action: %#v},\n", op.Type, op.Action)
				}
				p.Printf("},\n")
			}

			if f.ImplicitOperands != nil {
				p.Printf("ImplicitOperands: []ImplicitOperand{\n")
				for _, op := range f.ImplicitOperands {
					p.Printf("{Register: %#v, Action: %#v},\n", op.Register, op.Action)
				}
				p.Printf("},\n")
			}

			p.Printf("},\n")
		}
		p.Printf("},\n")

		p.Printf("},\n")
	}

	p.Printf("}\n")

	return p.Result()
}

type godatatest struct {
	cfg Config
}

func NewGoDataTest(cfg Config) Interface {
	return GoFmt(godatatest{cfg: cfg})
}

func (g godatatest) Generate(is []inst.Instruction) ([]byte, error) {
	p := &printer{}

	p.Printf("// %s\n\n", g.cfg.GeneratedWarning())
	p.Printf("package inst_test\n\n")

	p.Printf(`import (
		"reflect"
		"testing"

		"%s/internal/inst"
	)
	`, pkg)

	p.Printf("var raw = %#v\n\n", is)

	p.Printf(`func TestVerifyInstructionsList(t *testing.T) {
		if !reflect.DeepEqual(raw, inst.Instructions) {
			t.Fatal("bad code generation for instructions list")
		}
	}
	`)

	return p.Result()
}
