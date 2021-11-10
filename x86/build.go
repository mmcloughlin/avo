package x86

import (
	"errors"

	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
)

// BuildInstruction constructs an instruction object from a list of acceptable
// forms, and given input operands and suffixes.
func BuildInstruction(forms []Form, suffixes Suffixes, ops []operand.Op) (*ir.Instruction, error) {
	for i := range forms {
		f := &forms[i]
		if match(f, suffixes, ops) {
			return build(f, suffixes, ops)
		}
	}
	return nil, errors.New("bad operands")
}

func match(f *Form, suffixes Suffixes, ops []operand.Op) bool {
	// Match suffix.
	accept := f.SuffixesClass.SuffixesSet()
	if !accept[suffixes] {
		return false
	}

	// Match operands.
	if len(ops) != int(f.Arity) {
		return false
	}

	for i, op := range ops {
		t := OperandType(f.Operands[i].Type)
		if !t.Match(op) {
			return false
		}
	}

	return true
}

func build(f *Form, suffixes Suffixes, ops []operand.Op) (*ir.Instruction, error) {
	// Base instruction properties.
	i := &ir.Instruction{
		Opcode:           f.Opcode.String(),
		Suffixes:         suffixes.Strings(),
		Operands:         ops,
		IsTerminal:       (f.Features & FeatureTerminal) != 0,
		IsBranch:         (f.Features & FeatureBranch) != 0,
		IsConditional:    (f.Features & FeatureConditionalBranch) != 0,
		CancellingInputs: (f.Features & FeatureCancellingInputs) != 0,
		ISA:              f.ISAs.List(),
	}

	// Input/output operands.
	for _, spec := range f.Operands {
		if spec.Type == 0 {
			break
		}

		var op operand.Op
		if spec.Implicit {
			op = ImplicitRegister(spec.Type).Register()
		} else {
			op, ops = ops[0], ops[1:]
		}

		if spec.Action.Read() {
			i.Inputs = append(i.Inputs, op)
		}
		if spec.Action.Write() {
			i.Outputs = append(i.Outputs, op)
		}
	}

	return i, nil
}