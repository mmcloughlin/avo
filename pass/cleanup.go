package pass

import (
	"errors"

	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
)

// PruneJumpToFollowingLabel removes jump instructions that target an
// immediately following label.
func PruneJumpToFollowingLabel(fn *ir.Function) error {
	for i := 0; i+1 < len(fn.Nodes); i++ {
		node := fn.Nodes[i]
		next := fn.Nodes[i+1]

		// This node is an unconditional jump.
		inst, ok := node.(*ir.Instruction)
		if !ok || !inst.IsBranch || inst.IsConditional {
			continue
		}

		target := inst.TargetLabel()
		if target == nil {
			return errors.New("no label for branch instruction")
		}

		// And the jump target is the immediately following node.
		lbl, ok := next.(ir.Label)
		if !ok || lbl != *target {
			continue
		}

		// Then the jump is unnecessary and can be removed.
		fn.Nodes = deletenode(fn.Nodes, i)
		i--
	}

	return nil
}

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

		fn.Nodes = deletenode(fn.Nodes, i)
	}

	return nil
}

// deletenode deletes node i from nodes and returns the resulting slice.
func deletenode(nodes []ir.Node, i int) []ir.Node {
	n := len(nodes)
	copy(nodes[i:], nodes[i+1:])
	nodes[n-1] = nil
	return nodes[:n-1]
}

// invalidatecfg clears CFG structures.
func invalidatecfg(fn *ir.Function) {
	fn.LabelTarget = nil
	for _, i := range fn.Instructions() {
		i.Pred = nil
		i.Succ = nil
	}
}
