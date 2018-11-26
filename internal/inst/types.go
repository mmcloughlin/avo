package inst

type Instruction struct {
	Opcode  string
	AliasOf string
	Summary string
	Forms   []Form
}

type Form struct {
	ISA              []string
	Operands         []Operand
	ImplicitOperands []ImplicitOperand
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
