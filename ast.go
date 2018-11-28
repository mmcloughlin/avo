package avo

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

// Instruction is a single instruction in a function.
type Instruction struct {
	Opcode   string
	Operands []Operand
}

// File represents an assembly file.
type File struct {
	functions []*Function
}

// Function represents an assembly function.
type Function struct {
	name   string
	params []Parameter
	inst   []Instruction
}

func NewFunction(name string) *Function {
	return &Function{
		name: name,
	}
}

func (f *Function) AddInstruction(i Instruction) {
	f.inst = append(f.inst, i)
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
