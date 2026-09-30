package printer

import (
	"fmt"
	"sort"

	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

// validateCoalesce checks a function's promotion and coalescing result
// independently of how coalesceSlots derived it, as a second line of defence:
// the planner reasons about webs, liveness and interference; this reasons only
// about values, so a mistake in one is unlikely to be repeated in the other.
//
// It runs a forward must-equality analysis over the lowered program. For each
// arm64 register the lowering can write for a program variable (the x86
// register families' own registers and the promoted slots' registers) it
// tracks the set of variables -- x86 register families and promoted slots --
// whose current value that register provably holds, intersecting at control
// flow joins. Every read the lowering performs is then checked: a register
// family read through the register this instruction names for it, and a slot
// read through the slot's register, must find the variable in that register's
// set. At entry each family's own register holds the family and each slot
// register holds its slot (both undefined, and equally so).
//
// It also rejects renaming at an instruction a cross-instruction fold rewrites,
// renaming to anything but a promoted slot's register, renaming a family the
// instruction does not name, and the two renamings that equal values do not make safe: two
// families named by one register in a non-copy instruction (lowerings assume
// distinct x86 registers are distinct arm64 registers), and a renamed family
// the instruction also touches implicitly (renaming reaches only explicit
// operands).
func validateCoalesce(nodes []ir.Node, slotReg map[int]string, plan coalescePlan, excluded map[int]bool) error {
	for idx := range plan {
		if excluded[idx] {
			// A cross-instruction fold reads registers at another instruction
			// than the one naming them, which the per-instruction check below
			// does not see; renaming never reaches those instructions.
			return fmt.Errorf("%s (instruction %d): renamed inside a cross-instruction fold",
				nodes[idx].(*ir.Instruction).Opcode, idx)
		}
	}
	if len(slotReg) == 0 {
		if len(plan) != 0 {
			return fmt.Errorf("coalescing planned with no promoted slots")
		}
		return nil
	}
	slotRegs := map[string]bool{}
	for _, r := range slotReg {
		slotRegs[r] = true
	}
	for idx, m := range plan {
		ins, ok := nodes[idx].(*ir.Instruction)
		if !ok {
			return fmt.Errorf("node %d: renamed, but not an instruction", idx)
		}
		e := familyEffects(ins)
		for f, r := range m {
			if !slotRegs[r] {
				return fmt.Errorf("%s (instruction %d): family %d renamed to %s, not a promoted slot's register",
					ins.Opcode, idx, f, r)
			}
			if (e.uses|e.defs)&(1<<f) == 0 {
				// Nothing below would model it, and emission may use an
				// unmentioned family for its own purposes (slotAsRegister).
				return fmt.Errorf("%s (instruction %d): renames family %d, which it does not name",
					ins.Opcode, idx, f)
			}
		}
	}
	g, ok := buildCFG(nodes)
	if !ok {
		if plan != nil {
			return fmt.Errorf("coalescing planned for a function whose control flow cannot be followed")
		}
		return nil // move-only promotion; nothing here to check
	}
	n := len(g.ins)

	// Variables: families 0-15, then promoted slots.
	const nfam = 16
	slotVar := map[int]int{}
	var locs []string
	locIdx := map[string]int{}
	loc := func(name string) int {
		i, ok := locIdx[name]
		if !ok {
			i = len(locs)
			locIdx[name] = i
			locs = append(locs, name)
		}
		return i
	}
	for f := 0; f < nfam; f++ {
		if name, ok := armReg[reg.Index(f)]; ok {
			loc(name)
		}
	}
	var disps []int
	for d := range slotReg {
		disps = append(disps, d)
	}
	sort.Ints(disps)
	for i, d := range disps {
		slotVar[d] = nfam + i
		loc(slotReg[d])
	}
	nl := len(locs)
	type state = []uint64 // location -> variable set
	init := make(state, nl)
	for f := 0; f < nfam; f++ {
		if name, ok := armReg[reg.Index(f)]; ok {
			init[locIdx[name]] |= 1 << f
		}
	}
	for d, v := range slotVar {
		init[locIdx[slotReg[d]]] |= 1 << v
	}

	famLoc := func(k, f int) (int, error) {
		if name, ok := plan[g.node[k]][f]; ok {
			return loc(name), nil
		}
		name, ok := armReg[reg.Index(f)]
		if !ok {
			return 0, fmt.Errorf("family %d has no arm64 register", f)
		}
		return locIdx[name], nil
	}

	// transfer applies instruction k to in, returning the out state, or an
	// error for a read that would see the wrong value.
	transfer := func(k int, in state) (state, error) {
		ins := g.ins[k]
		e := familyEffects(ins)
		out := append(state(nil), in...)
		renamed := plan[g.node[k]]
		for f := range renamed {
			if e.implicit&(1<<f) != 0 {
				return nil, fmt.Errorf("renamed family %d is also an implicit operand", f)
			}
		}

		// A MOVQ between two tracked variables (whole registers or promoted
		// slots) is a copy: its destination then holds whatever its source did.
		isCopy := false
		var srcVar, srcLoc int
		tracked := func(op operand.Op) (v, l int, ok bool, err error) {
			if isFullGP(op) {
				f := regFamily(op)
				l, err := famLoc(k, f)
				return f, l, err == nil, err
			}
			if d, isSlot := frameSlot(op); isSlot {
				if v, promoted := slotVar[d]; promoted {
					return v, locIdx[slotReg[d]], true, nil
				}
			}
			return 0, 0, false, nil
		}
		if ins.Opcode == "MOVQ" && len(ins.Operands) == 2 {
			sv, sl, sok, err := tracked(ins.Operands[0])
			if err != nil {
				return nil, err
			}
			_, _, dok, err := tracked(ins.Operands[1])
			if err != nil {
				return nil, err
			}
			isCopy, srcVar, srcLoc = sok && dok, sv, sl
		}

		// Reads.
		seen := map[int]int{} // location -> family, for the aliasing check
		for f := 0; f < nfam; f++ {
			if (e.uses|e.defs)&(1<<f) == 0 {
				continue
			}
			l, err := famLoc(k, f)
			if err != nil {
				return nil, err
			}
			if o, dup := seen[l]; dup && !isCopy {
				return nil, fmt.Errorf("families %d and %d both named %s", o, f, locs[l])
			}
			seen[l] = f
			if e.uses&(1<<f) != 0 && in[l]&(1<<f) == 0 {
				return nil, fmt.Errorf("reads family %d from %s, which does not hold it", f, locs[l])
			}
		}
		var slotDef = -1
		if d, use, def, ok := slotEffect(ins); ok {
			if v, promoted := slotVar[d]; promoted {
				l := locIdx[slotReg[d]]
				if use && in[l]&(1<<v) == 0 {
					return nil, fmt.Errorf("reads slot %d from %s, which does not hold it", d, slotReg[d])
				}
				if o, dup := seen[l]; dup && !isCopy {
					return nil, fmt.Errorf("family %d and slot %d both named %s", o, d, locs[l])
				}
				if def {
					slotDef = d
				}
			}
		}

		// Writes: every written variable leaves every set, then each written
		// location holds exactly what was written to it.
		var written uint64
		for f := 0; f < nfam; f++ {
			if e.defs&(1<<f) != 0 {
				written |= 1 << f
			}
		}
		if slotDef >= 0 {
			written |= 1 << slotVar[slotDef]
		}
		for l := range out {
			out[l] &^= written
		}
		put := func(l, v int) {
			if isCopy {
				out[l] = in[srcLoc]&^written | 1<<srcVar | 1<<v
			} else {
				out[l] = 1 << v
			}
		}
		for f := 0; f < nfam; f++ {
			if e.defs&(1<<f) != 0 {
				l, _ := famLoc(k, f)
				put(l, f)
			}
		}
		if slotDef >= 0 {
			put(locIdx[slotReg[slotDef]], slotVar[slotDef])
		}
		return out, nil
	}

	// Forward must-analysis to a fixed point. nil is "not yet reached" (top).
	in := make([]state, n)
	out := make([]state, n)
	for changed := true; changed; {
		changed = false
		for k := 0; k < n; k++ {
			var cur state
			meet := func(s state) {
				if s == nil {
					return
				}
				if cur == nil {
					cur = append(state(nil), s...)
					return
				}
				for l := range cur {
					cur[l] &= s[l]
				}
			}
			if k == 0 {
				meet(init)
			}
			for _, p := range g.pred[k] {
				meet(out[p])
			}
			if cur == nil {
				continue
			}
			in[k] = cur
			o, err := transfer(k, cur)
			if err != nil {
				// A read that fails in an intermediate state may succeed
				// once the analysis settles only if sets grew, and a
				// must-analysis from top only shrinks them; so this is final.
				return fmt.Errorf("%s (instruction %d): %v", g.ins[k].Opcode, g.node[k], err)
			}
			if !equalState(o, out[k]) {
				out[k] = o
				changed = true
			}
		}
	}
	return nil
}

func equalState(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
