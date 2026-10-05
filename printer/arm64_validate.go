package printer

import (
	"errors"
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
// instruction does not name, renaming in unreachable code (which the analysis
// never visits), and the two renamings that equal values do not make safe: two
// families named by one register in a non-copy instruction (lowerings assume
// distinct x86 registers are distinct arm64 registers), and a renamed family
// the instruction also touches implicitly (renaming reaches only explicit
// operands).
func validateCoalesce(nodes []ir.Node, slotReg map[int]string, plan coalescePlan, excluded map[int]bool) error {
	if err := checkPlanShape(nodes, slotReg, plan, excluded); err != nil {
		return err
	}
	if len(slotReg) == 0 {
		return nil
	}
	g, ok := buildCFG(nodes)
	if !ok {
		if plan != nil {
			return errors.New("coalescing planned for a function whose control flow cannot be followed")
		}
		return nil // move-only promotion; nothing here to check
	}
	reach := g.reachable()
	for k := range g.ins {
		if !reach[k] && plan[g.node[k]] != nil {
			return fmt.Errorf("%s (instruction %d): renamed in unreachable code, which is not checked",
				g.ins[k].Opcode, g.node[k])
		}
	}
	v := newValidator(g, slotReg, plan)
	return v.run()
}

// checkPlanShape rejects plans that rename where renaming is never allowed:
// inside a cross-instruction fold, to anything but a promoted slot's register,
// or a family the instruction does not name.
func checkPlanShape(nodes []ir.Node, slotReg map[int]string, plan coalescePlan, excluded map[int]bool) error {
	for idx := range plan {
		if excluded[idx] {
			// A cross-instruction fold reads registers at another instruction
			// than the one naming them, which the per-instruction check below
			// does not see; renaming never reaches those instructions.
			op := "?"
			if ins, ok := nodes[idx].(*ir.Instruction); ok {
				op = ins.Opcode
			}
			return fmt.Errorf("%s (instruction %d): renamed inside a cross-instruction fold", op, idx)
		}
	}
	if len(slotReg) == 0 {
		if len(plan) != 0 {
			return errors.New("coalescing planned with no promoted slots")
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
		if err := checkRenames(ins, idx, m, slotRegs); err != nil {
			return err
		}
	}
	return nil
}

// checkRenames checks one instruction's renames against the promoted slot
// registers and the families the instruction names.
func checkRenames(ins *ir.Instruction, idx int, m map[int]string, slotRegs map[string]bool) error {
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
	return nil
}

// valState is, per location, the set of variables whose value it holds.
type valState = []uint64

// validator is the must-equality analysis of validateCoalesce. Variables are
// register families 0-15, then promoted slots; locations are arm64 registers.
type validator struct {
	g       cfg
	slotReg map[int]string
	plan    coalescePlan
	slotVar map[int]int // slot displacement -> variable
	locs    []string
	locIdx  map[string]int
	init    valState
}

func newValidator(g cfg, slotReg map[int]string, plan coalescePlan) *validator {
	v := &validator{g: g, slotReg: slotReg, plan: plan, slotVar: map[int]int{}, locIdx: map[string]int{}}
	for f := 0; f < numFamilies; f++ {
		if name, ok := armReg[reg.Index(f)]; ok {
			v.loc(name)
		}
	}
	var disps []int
	for d := range slotReg {
		disps = append(disps, d)
	}
	sort.Ints(disps)
	for i, d := range disps {
		v.slotVar[d] = numFamilies + i
		v.loc(slotReg[d])
	}
	v.init = make(valState, len(v.locs))
	for f := 0; f < numFamilies; f++ {
		if name, ok := armReg[reg.Index(f)]; ok {
			v.init[v.locIdx[name]] |= 1 << f
		}
	}
	for d, sv := range v.slotVar {
		v.init[v.locIdx[slotReg[d]]] |= 1 << sv
	}
	return v
}

// loc returns the index of location name, adding it if new.
func (v *validator) loc(name string) int {
	i, ok := v.locIdx[name]
	if !ok {
		i = len(v.locs)
		v.locIdx[name] = i
		v.locs = append(v.locs, name)
	}
	return i
}

// famLoc is the location instruction k names for family f.
func (v *validator) famLoc(k, f int) (int, error) {
	if name, ok := v.plan[v.g.node[k]][f]; ok {
		return v.loc(name), nil
	}
	name, ok := armReg[reg.Index(f)]
	if !ok {
		return 0, fmt.Errorf("family %d has no arm64 register", f)
	}
	return v.locIdx[name], nil
}

// varLoc is a variable and the location holding it.
type varLoc struct{ v, l int }

// tracked reports the variable and location of op, an operand of instruction
// k, when it is a whole register or a promoted slot.
func (v *validator) tracked(k int, op operand.Op) (varLoc, bool, error) {
	if isFullGP(op) {
		f := regFamily(op)
		l, err := v.famLoc(k, f)
		return varLoc{f, l}, err == nil, err
	}
	if d, isSlot := frameSlot(op); isSlot {
		if sv, promoted := v.slotVar[d]; promoted {
			return varLoc{sv, v.locIdx[v.slotReg[d]]}, true, nil
		}
	}
	return varLoc{}, false, nil
}

// copyOf describes instruction k as a copy, if it is a MOVQ between two
// tracked variables: its destination then holds whatever its source did.
type copyOf struct {
	is       bool
	src, loc int // source variable and location
}

func (v *validator) copyAt(k int) (copyOf, error) {
	ins := v.g.ins[k]
	if ins.Opcode != "MOVQ" || len(ins.Operands) != 2 {
		return copyOf{}, nil
	}
	src, sok, err := v.tracked(k, ins.Operands[0])
	if err != nil {
		return copyOf{}, err
	}
	_, dok, err := v.tracked(k, ins.Operands[1])
	if err != nil {
		return copyOf{}, err
	}
	return copyOf{sok && dok, src.v, src.l}, nil
}

// readFamilies checks every family instruction k reads is held by the
// location it is read from, and that no two families share a location unless
// k is a copy. It returns the location -> family map it used for the check.
func (v *validator) readFamilies(k int, e famEffects, in valState, cp copyOf) (map[int]int, error) {
	seen := map[int]int{}
	for f := 0; f < numFamilies; f++ {
		if (e.uses|e.defs)&(1<<f) == 0 {
			continue
		}
		l, err := v.famLoc(k, f)
		if err != nil {
			return nil, err
		}
		if o, dup := seen[l]; dup && !cp.is {
			return nil, fmt.Errorf("families %d and %d both named %s", o, f, v.locs[l])
		}
		seen[l] = f
		if e.uses&(1<<f) != 0 && in[l]&(1<<f) == 0 {
			return nil, fmt.Errorf("reads family %d from %s, which does not hold it", f, v.locs[l])
		}
	}
	return seen, nil
}

// readSlot checks a promoted slot instruction k reads is held by its register
// and that its register is not also named for a family. It returns the slot
// the instruction writes, or -1.
func (v *validator) readSlot(ins *ir.Instruction, in valState, seen map[int]int, cp copyOf) (int, error) {
	su, ok := slotEffect(ins)
	if !ok {
		return -1, nil
	}
	sv, promoted := v.slotVar[su.disp]
	if !promoted {
		return -1, nil
	}
	l := v.locIdx[v.slotReg[su.disp]]
	if su.use && in[l]&(1<<sv) == 0 {
		return -1, fmt.Errorf("reads slot %d from %s, which does not hold it", su.disp, v.slotReg[su.disp])
	}
	if o, dup := seen[l]; dup && !cp.is {
		return -1, fmt.Errorf("family %d and slot %d both named %s", o, su.disp, v.locs[l])
	}
	if su.def {
		return su.disp, nil
	}
	return -1, nil
}

// write applies instruction k's writes to in: every written variable leaves
// every set, then each written location holds exactly what was written to it.
func (v *validator) write(k int, e famEffects, in valState, slotDef int, cp copyOf) valState {
	var written uint64
	for f := 0; f < numFamilies; f++ {
		if e.defs&(1<<f) != 0 {
			written |= 1 << f
		}
	}
	if slotDef >= 0 {
		written |= 1 << v.slotVar[slotDef]
	}
	out := append(valState(nil), in...)
	for l := range out {
		out[l] &^= written
	}
	put := func(l, vr int) {
		if cp.is {
			out[l] = in[cp.loc]&^written | 1<<cp.src | 1<<vr
		} else {
			out[l] = 1 << vr
		}
	}
	for f := 0; f < numFamilies; f++ {
		if e.defs&(1<<f) != 0 {
			l, _ := v.famLoc(k, f)
			put(l, f)
		}
	}
	if slotDef >= 0 {
		put(v.locIdx[v.slotReg[slotDef]], v.slotVar[slotDef])
	}
	return out
}

// transfer applies instruction k to in, returning the out state, or an error
// for a read that would see the wrong value.
func (v *validator) transfer(k int, in valState) (valState, error) {
	ins := v.g.ins[k]
	e := familyEffects(ins)
	for f := range v.plan[v.g.node[k]] {
		if e.implicit&(1<<f) != 0 {
			return nil, fmt.Errorf("renamed family %d is also an implicit operand", f)
		}
	}
	cp, err := v.copyAt(k)
	if err != nil {
		return nil, err
	}
	seen, err := v.readFamilies(k, e, in, cp)
	if err != nil {
		return nil, err
	}
	slotDef, err := v.readSlot(ins, in, seen, cp)
	if err != nil {
		return nil, err
	}
	return v.write(k, e, in, slotDef, cp), nil
}

// meet intersects s into cur; nil is "not yet reached" (top).
func meet(cur, s valState) valState {
	if s == nil {
		return cur
	}
	if cur == nil {
		return append(valState(nil), s...)
	}
	for l := range cur {
		cur[l] &= s[l]
	}
	return cur
}

// run solves the forward must-analysis to a fixed point, checking every read.
func (v *validator) run() error {
	n := len(v.g.ins)
	out := make([]valState, n)
	for changed := true; changed; {
		changed = false
		for k := 0; k < n; k++ {
			var cur valState
			if k == 0 {
				cur = meet(cur, v.init)
			}
			for _, p := range v.g.pred[k] {
				cur = meet(cur, out[p])
			}
			if cur == nil {
				continue
			}
			o, err := v.transfer(k, cur)
			if err != nil {
				// A read that fails in an intermediate state may succeed
				// once the analysis settles only if sets grew, and a
				// must-analysis from top only shrinks them; so this is final.
				return fmt.Errorf("%s (instruction %d): %w", v.g.ins[k].Opcode, v.g.node[k], err)
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
