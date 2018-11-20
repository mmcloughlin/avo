package inst

type Instruction struct {
	Opcode string
	Forms  []Form
}

type Form struct {
	Operands []Operand
	CPUID    []string
}

type Operand struct {
	Type   string
	Action Action
}

type Action uint8

const (
	R  Action = 0x1
	W  Action = 0x2
	RW Action = R | W
)
