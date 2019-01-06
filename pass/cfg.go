package pass

import (
	"errors"
	"fmt"

	"github.com/mmcloughlin/avo/ir"
)

// LabelTarget populates the LabelTarget of the given function. This maps from
// label name to the following instruction.
func LabelTarget(fn *ir.Function) error {
	target := map[ir.Label]*ir.Instruction{}
	for idx := 0; idx < len(fn.Nodes); idx++ {
		// Is this a label?
		lbl, ok := fn.Nodes[idx].(ir.Label)
		if !ok {
			continue
		}
		// Check for a duplicate label.
		if _, found := target[lbl]; found {
			return fmt.Errorf("duplicate label \"%s\"", lbl)
		}
		// Advance to next node.
		if idx == len(fn.Nodes)-1 {
			return errors.New("function ends with label")
		}
		idx++
		// Should be an instruction.
		i, ok := fn.Nodes[idx].(*ir.Instruction)
		if !ok {
			return errors.New("instruction should follow a label")
		}
		target[lbl] = i
	}
	fn.LabelTarget = target
	return nil
}

// CFG constructs the call-flow-graph for the function.
func CFG(fn *ir.Function) error {
	is := fn.Instructions()
	n := len(is)

	// Populate successors.
	for i := 0; i < n; i++ {
		cur := is[i]
		var nxt *ir.Instruction
		if i+1 < n {
			nxt = is[i+1]
		}

		// If it's a branch, locate the target.
		if cur.IsBranch {
			lbl := cur.TargetLabel()
			if lbl == nil {
				return errors.New("no label for branch instruction")
			}
			target, found := fn.LabelTarget[*lbl]
			if !found {
				return errors.New("unknown label")
			}
			cur.Succ = append(cur.Succ, target)
		}

		// Otherwise, could continue to the following instruction.
		switch {
		case cur.IsTerminal:
		case cur.IsBranch && !cur.IsConditional:
		default:
			cur.Succ = append(cur.Succ, nxt)
		}
	}

	// Populate predecessors.
	for _, i := range is {
		for _, s := range i.Succ {
			if s != nil {
				s.Pred = append(s.Pred, i)
			}
		}
	}

	return nil
}
