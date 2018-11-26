package inst

import "sort"

type Instruction struct {
	Opcode  string
	AliasOf string
	Summary string
	Forms   []Form
}

func (i Instruction) Arities() []int {
	s := map[int]bool{}
	for _, f := range i.Forms {
		s[len(f.Operands)] = true
	}
	a := make([]int, 0, len(s))
	for n := range s {
		a = append(a, n)
	}
	sort.Ints(a)
	return a
}

func (i Instruction) Arity() int {
	if i.IsVariadic() {
		panic("variadic")
	}
	a := i.Arities()
	return a[0]
}

func (i Instruction) IsVariadic() bool {
	return len(i.Arities()) > 1
}

func (i Instruction) IsNiladic() bool {
	a := i.Arities()
	return len(a) == 1 && a[0] == 0
}

type Form struct {
	ISA              []string
	Operands         []Operand
	ImplicitOperands []ImplicitOperand
}

func (f Form) Signature() []string {
	s := make([]string, len(f.Operands))
	for i, op := range f.Operands {
		s[i] = op.Type
	}
	return s
}

type Operand struct {
	Type   string
	Action Action
}

type ImplicitOperand struct {
	Register string
	Action   Action
}

type Action uint8

const (
	R  Action = 0x1
	W  Action = 0x2
	RW Action = R | W
)

func ActionFromReadWrite(r, w bool) Action {
	var a Action
	if r {
		a |= R
	}
	if w {
		a |= W
	}
	return a
}

func (a Action) Read() bool {
	return (a & R) != 0
}

func (a Action) Write() bool {
	return (a & W) != 0
}

func (a Action) String() string {
	s := ""
	if a.Read() {
		s += "r"
	}
	if a.Write() {
		s += "w"
	}
	return s
}
