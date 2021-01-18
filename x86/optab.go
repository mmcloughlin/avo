package x86

type Operand struct {
	Type     uint8
	Implicit bool
	Action   uint8
}

type Form struct {
	Opcode   Opcode
	Operands [MaxArity]Operand
}
