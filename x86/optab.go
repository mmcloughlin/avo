package x86

type Form struct {
	Opcode        Opcode
	SuffixesClass SuffixesClass
	Features      Feature
	ISAs          ISAs
	Arity         uint8
	Operands      Operands
}

type Suffixes [MaxSuffixes]Suffix

type Feature uint8

const (
	FeatureTerminal Feature = 1 << iota
	FeatureBranch
	FeatureConditionalBranch
	FeatureCancellingInputs
)

type Operands [MaxOperands]Operand

type Operand struct {
	Type     uint8
	Implicit bool
	Action   Action
}

type Action uint8

const (
	ActionNone Action = 0
	ActionR    Action = 1
	ActionW    Action = 2
	ActionRW   Action = ActionR | ActionW
)
