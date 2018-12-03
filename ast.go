package avo

import (
	"github.com/mmcloughlin/avo/operand"
)

type Asm interface {
	Asm() string
}

// GoType represents a Golang type.
type GoType interface{}

// Parameter represents a parameter to an assembly function.
type Parameter struct {
	Name string
	Type GoType
}

type Operand interface {
	Asm
}

type Node interface {
	node()
}

type Label string

func (l Label) node() {}

// Instruction is a single instruction in a function.
type Instruction struct {
	Opcode   string
	Operands []operand.Op

	Inputs  []operand.Op
	Outputs []operand.Op

	IsTerminal    bool
	IsBranch      bool
	IsConditional bool

	// CFG.
	Pred []*Instruction
	Succ []*Instruction
}

func (i Instruction) TargetLabel() *Label {
	if !i.IsBranch {
		return nil
	}
	if len(i.Operands) == 0 {
		return nil
	}
	if ref, ok := i.Operands[0].(operand.LabelRef); ok {
		lbl := Label(ref)
		return &lbl
	}
	return nil
}

func (i Instruction) node() {}

// File represents an assembly file.
type File struct {
	Functions []*Function
}

func NewFile() *File {
	return &File{}
}

// Function represents an assembly function.
type Function struct {
	name   string
	params []Parameter
	Nodes  []Node

	// LabelTarget maps from label name to the following instruction.
	LabelTarget map[Label]*Instruction
}

func NewFunction(name string) *Function {
	return &Function{
		name: name,
	}
}

func (f *Function) AddInstruction(i *Instruction) {
	f.AddNode(i)
}

func (f *Function) AddLabel(l Label) {
	f.AddNode(l)
}

func (f *Function) AddNode(n Node) {
	f.Nodes = append(f.Nodes, n)
}

// Instructions returns just the list of instruction nodes.
func (f *Function) Instructions() []*Instruction {
	var is []*Instruction
	for _, n := range f.Nodes {
		i, ok := n.(*Instruction)
		if ok {
			is = append(is, i)
		}
	}
	return is
}

// Name returns the function name.
func (f *Function) Name() string { return f.name }

// FrameBytes returns the size of the stack frame in bytes.
func (f *Function) FrameBytes() int {
	// TODO(mbm): implement
	return 0
}

// ArgumentBytes returns the size of the arguments in bytes.
func (f *Function) ArgumentBytes() int {
	// TODO(mbm): implement
	return 0
}
