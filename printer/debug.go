package printer

import (
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/ir"
)

type debug struct {
	cfg Config
	prnt.Generator
}

// NewDebug constructs a printer for writing debug output.
func NewDebug(cfg Config) Printer {
	return &debug{cfg: cfg}
}

func (p *debug) Print(f *ir.File) ([]byte, error) {
	for _, s := range f.Sections {
		switch s := s.(type) {
		case *ir.Function:
			p.function(s)
		}
	}
	return p.Result()
}

func (p *debug) function(f *ir.Function) {
	p.enter("function")

	p.Linef("name       = %q", f.Name)
	p.Linef("attributes = 0x%04x %s", f.Attributes, f.Attributes.Asm())
	p.Linef("doc        = %#v", f.Doc)
	p.Linef("signature  = %s", f.Signature)
	p.Linef("local size = %d", f.LocalSize)

	// // LabelTarget maps from label name to the following instruction.
	// LabelTarget map[Label]*Instruction

	// // Register allocation.
	// Allocation reg.Allocation

	p.nodes(f.Nodes)

	p.leave()
}

func (p *debug) nodes(nodes []ir.Node) {
	p.enter("nodes")

	for _, node := range nodes {
		switch n := node.(type) {
		case *ir.Instruction:
			p.instruction(n)
		case ir.Label:
		case *ir.Comment:
		default:
			panic("unexpected node type")
		}
	}

	p.leave()
}

func (p *debug) instruction(i *ir.Instruction) {
	p.enter("instruction")

	p.Linef("addr        = 0x%p", i)
	p.Linef("opcode      = %q", i.Opcode)
	p.Linef("terminal    = %t", i.IsTerminal)
	p.Linef("branch      = %t", i.IsBranch)
	p.Linef("conditional = %t", i.IsConditional)

	p.enter("operands")
	for idx, op := range i.Operands {
		p.Linef("%d: %s", idx, op.Asm())
	}
	p.leave()

	p.enter("inputs")
	for idx, op := range i.Inputs {
		p.Linef("%d: %s", idx, op.Asm())
	}
	p.leave()

	p.enter("outputs")
	for idx, op := range i.Outputs {
		p.Linef("%d: %s", idx, op.Asm())
	}
	p.leave()

	p.enter("pred")
	for _, pred := range i.Pred {
		p.Linef("0x%p", pred)
	}
	p.leave()

	p.enter("succ")
	for _, succ := range i.Succ {
		p.Linef("0x%p", succ)
	}
	p.leave()

	// // LiveIn/LiveOut are sets of live register IDs pre/post execution.
	// LiveIn  reg.Set
	// LiveOut reg.Set

	p.leave()
}

func (p *debug) enter(name string) {
	p.Linef("%s {", name)
	p.Indent()
}

func (p *debug) leave() {
	p.Dedent()
	p.Linef("}")
}
