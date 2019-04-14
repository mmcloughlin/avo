package pass

import (
	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
)

// PruneSelfMoves removes move instructions from one register to itself.
func PruneSelfMoves(fn *ir.Function) error {
	return removeinstructions(fn, func(i *ir.Instruction) bool {
		switch i.Opcode {
		case "MOVB", "MOVW", "MOVL", "MOVQ":
		default:
			return false
		}

		return operand.IsRegister(i.Operands[0]) && operand.IsRegister(i.Operands[1]) && i.Operands[0] == i.Operands[1]
	})
}

// removeinstructions deletes instructions from the given function which match predicate.
func removeinstructions(fn *ir.Function, predicate func(*ir.Instruction) bool) error {
	// Removal of instructions has the potential to invalidate CFG structures.
	// Clear them to prevent accidental use of stale structures after this pass.
	invalidatecfg(fn)

	for i := 0; i < len(fn.Nodes); i++ {
		n := fn.Nodes[i]

		inst, ok := n.(*ir.Instruction)
		if !ok || !predicate(inst) {
			continue
		}

		copy(fn.Nodes[i:], fn.Nodes[i+1:])
		fn.Nodes[len(fn.Nodes)-1] = nil
		fn.Nodes = fn.Nodes[:len(fn.Nodes)-1]
	}

	return nil
}

// invalidatecfg clears CFG structures.
func invalidatecfg(fn *ir.Function) {
	fn.LabelTarget = nil
	for _, i := range fn.Instructions() {
		i.Pred = nil
		i.Succ = nil
	}
}
