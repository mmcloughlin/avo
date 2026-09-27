package pass

import (
	"errors"
	"fmt"

	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
)

// Verify pass validates an avo file.
var Verify = Concat(
	InstructionPass(VerifyMemOperands),
)

// VerifyNoLabelOnPCALIGN rejects a PCALIGN placed directly after a label, so
// that jumps to the label target the PCALIGN itself. The amd64 assembler
// mishandles such jumps: before Go 1.26 it resolves them to the wrong place
// (golang/go#74648), and from Go 1.26 it loops forever if it has to widen any
// branch in the function. Emit the PCALIGN before the label instead, so the
// label marks the aligned instruction.
//
// It runs after PruneDanglingLabels, so only labels something jumps to count.
func VerifyNoLabelOnPCALIGN(fn *ir.Function) error {
	var label ir.Label
	for _, n := range fn.Nodes {
		switch n := n.(type) {
		case ir.Label:
			label = n
		case *ir.Instruction:
			if n.Opcode == "PCALIGN" && label != "" {
				return fmt.Errorf("function %s: label %s is directly on a PCALIGN; emit the PCALIGN before the label", fn.Name, label)
			}
			label = ""
		}
	}
	return nil
}

// VerifyMemOperands checks the instruction's memory operands.
func VerifyMemOperands(i *ir.Instruction) error {
	for _, op := range i.Operands {
		m, ok := op.(operand.Mem)
		if !ok {
			continue
		}

		if m.Base == nil {
			return errors.New("bad memory operand: missing base register")
		}

		if m.Index != nil && m.Scale == 0 {
			return errors.New("bad memory operand: index register with scale 0")
		}
	}
	return nil
}
