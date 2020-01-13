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
	var empty ir.Label
	pending := empty
	for _, node := range fn.Nodes {
		switch n := node.(type) {
		case ir.Label:
			if pending != empty {
				return fmt.Errorf("expected instruction following label %q", pending)
			}
			pending = n
			if _, found := target[pending]; found {
				return fmt.Errorf("duplicate label \"%s\"", pending)
			}
		case *ir.Instruction:
			if pending != empty {
				target[pending] = n
				pending = empty
			}
		}
	}
	if pending != empty {
		return errors.New("function ends with label")
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
				return fmt.Errorf("unknown label %q", *lbl)
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
