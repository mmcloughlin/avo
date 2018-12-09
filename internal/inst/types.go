package inst

import (
	"sort"
	"strings"
)

type Instruction struct {
	Opcode  string
	AliasOf string
	Summary string
	Forms   []Form
}

func (i Instruction) IsTerminal() bool {
	// TODO(mbm): how about the RETF* instructions
	return i.Opcode == "RET"
}

func (i Instruction) IsBranch() bool {
	if i.Opcode == "CALL" {
		return false
	}
	for _, f := range i.Forms {
		for _, op := range f.Operands {
			if strings.HasPrefix(op.Type, "rel") {
				return true
			}
		}
	}
	return false
}

func (i Instruction) IsConditionalBranch() bool {
	return i.IsBranch() && i.Opcode != "JMP"
}

func (i Instruction) Arities() []int {
	s := map[int]bool{}
	for _, f := range i.Forms {
		s[f.Arity()] = true
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

func (f Form) Arity() int {
	return len(f.Operands)
}

func (f Form) Signature() []string {
	s := make([]string, f.Arity())
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

func (a Action) Contains(s Action) bool {
	return (a & s) == s
}

func (a Action) Read() bool {
	return a.Contains(R)
}

func (a Action) Write() bool {
	return a.Contains(W)
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
