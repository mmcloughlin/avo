package printer

import (
	"strings"

	"github.com/mmcloughlin/avo"
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/operand"
)

// dot is the pesky unicode dot used in Go assembly.
const dot = "\u00b7"

type goasm struct {
	cfg Config
	prnt.Generator
}

func NewGoAsm(cfg Config) Printer {
	return &goasm{cfg: cfg}
}

func (p *goasm) Print(f *avo.File) ([]byte, error) {
	p.header()
	for _, fn := range f.Functions {
		p.function(fn)
	}
	return p.Result()
}

func (p *goasm) header() {
	p.NL()
	p.include("textflag.h")
	p.NL()
}

func (p *goasm) include(path string) {
	p.Printf("#include \"%s\"\n", path)
}

func (p *goasm) function(f *avo.Function) {
	p.Comment(f.Stub())
	p.Printf("TEXT %s%s(SB),0,$%d-%d\n", dot, f.Name, f.FrameBytes(), f.ArgumentBytes())

	for _, node := range f.Nodes {
		switch n := node.(type) {
		case *avo.Instruction:
			p.Printf("\t%s\t%s\n", n.Opcode, joinOperands(n.Operands))
		case avo.Label:
			p.Printf("%s:\n", n)
		default:
			panic("unexpected node type")
		}
	}
}

func joinOperands(operands []operand.Op) string {
	asm := make([]string, len(operands))
	for i, op := range operands {
		asm[i] = op.Asm()
	}
	return strings.Join(asm, ", ")
}
