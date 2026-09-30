package printer

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
	"github.com/mmcloughlin/avo/x86"
)

var (
	coalesceCheckPrograms = flag.Int("coalesce.programs", 300,
		"random programs TestCoalesceValidatorSound checks")
	coalesceCheckSeed = flag.Int64("coalesce.seed", 1, "seed for TestCoalesceValidatorSound")
)

// TestCoalesceValidatorSound is a bounded model check of validateCoalesce,
// the trusted half of stack-slot coalescing. It generates small programs over
// three register families and two promoted slots (copies, ALU operations,
// slot ALU operations, branches and loops), and for each hands the validator
// the planner's plan, mutations of it, and random plans. Whenever the
// validator accepts a plan, an abstract execution of the renamed program must
// agree with the original on every path up to a step bound: every read sees
// the value the original read. The planner's own plan must be accepted.
//
// The execution model is the one the validator relies on: an instruction
// reads the locations its operands are renamed to and writes a fresh value
// (or, for MOVQ, the value it read) to its destinations' locations. What the
// lowerings emit for an instruction is outside this check; the property tests
// in tests/arm64lower cover that on hardware.
func TestCoalesceValidatorSound(t *testing.T) {
	n := *coalesceCheckPrograms
	if testing.Short() {
		n /= 10
	}
	r := rand.New(rand.NewSource(*coalesceCheckSeed))
	var stats struct{ programs, plannerFolds, accepted, rejected int }
	for i := 0; i < n; i++ {
		fn, desc := randomSlotProgram(r)
		c, ok := coalesceCaseFor(fn)
		if !ok {
			continue
		}
		stats.programs++
		if err := validateCoalesce(c.nodes, c.slotReg, c.plan, c.excluded); err != nil {
			t.Fatalf("planner plan rejected: %v\n%s", err, desc)
		}
		if c.plan != nil {
			stats.plannerFolds++
		}
		if bad := checkPlan(c, c.plan); bad != "" {
			t.Fatalf("planner plan miscompiles: %s\nplan %v\n%s", bad, c.plan, desc)
		}
		for j := 0; j < 40; j++ {
			plan := randomPlan(r, c)
			if validateCoalesce(c.nodes, c.slotReg, plan, c.excluded) != nil {
				stats.rejected++
				continue
			}
			stats.accepted++
			if bad := checkPlan(c, plan); bad != "" {
				t.Fatalf("validator accepted a wrong plan: %s\nplan %v\n%s", bad, plan, desc)
			}
		}
	}
	t.Logf("%d programs (%d with planner folds); random plans: %d accepted and checked, %d rejected",
		stats.programs, stats.plannerFolds, stats.accepted, stats.rejected)
	if stats.accepted < stats.programs {
		t.Errorf("too few accepted random plans (%d) for the check to mean much", stats.accepted)
	}
}

var checkRegs = []reg.Physical{reg.RAX, reg.RCX, reg.RDX}

// randomSlotProgram builds a function of 4-12 random instructions with
// labels between them, returning it and a listing.
func randomSlotProgram(r *rand.Rand) (*ir.Function, string) {
	fn := ir.NewFunction("check")
	local := fn.AllocLocal(16)
	slots := []operand.Mem{local, local.Offset(8)}
	var lines []string
	add := func(i *ir.Instruction, err error) {
		if err != nil {
			panic(err)
		}
		fn.AddInstruction(i)
		var ops []string
		for _, op := range i.Operands {
			ops = append(ops, op.Asm())
		}
		lines = append(lines, "\t"+i.Opcode+" "+strings.Join(ops, ", "))
	}
	g := func() reg.Physical { return checkRegs[r.Intn(len(checkRegs))] }
	s := func() operand.Mem { return slots[r.Intn(len(slots))] }
	// Usually define everything first; sometimes leave values to come from
	// the caller.
	if r.Intn(4) != 0 {
		for i, x := range checkRegs {
			add(x86.MOVQ(operand.U32(uint64(i+1)), x))
		}
	}
	if r.Intn(3) != 0 {
		for i, m := range slots {
			add(x86.MOVQ(operand.U32(uint64(i+10)), m))
		}
	}
	nlabels := 1 + r.Intn(3)
	length := 4 + r.Intn(9)
	labelAt := map[int]string{}
	for i := 0; i < nlabels; i++ {
		labelAt[r.Intn(length+1)] = fmt.Sprintf("L%d", i)
	}
	var labels []string
	for _, l := range labelAt {
		labels = append(labels, l)
	}
	lbl := func() operand.LabelRef { return operand.LabelRef(labels[r.Intn(len(labels))]) }
	for k := 0; k <= length; k++ {
		if l, ok := labelAt[k]; ok {
			fn.AddLabel(ir.Label(l))
			lines = append(lines, l+":")
		}
		if k == length {
			break
		}
		switch r.Intn(13) {
		case 0:
			add(x86.MOVQ(g(), g()))
		case 1, 2:
			add(x86.MOVQ(s(), g()))
		case 3, 4:
			add(x86.MOVQ(g(), s()))
		case 5:
			add(x86.MOVQ(operand.U32(uint64(r.Intn(5))), s()))
		case 6:
			add(x86.ADDQ(g(), g()))
		case 7:
			add(x86.ADDQ(operand.U32(1), g()))
		case 8:
			add(x86.ADDQ(g(), s()))
		case 9:
			add(x86.CMPQ(g(), s()))
		case 10:
			add(x86.DECQ(s()))
		case 11:
			add(x86.CMPQ(g(), g()))
			add(x86.JNE(lbl()))
		case 12:
			if r.Intn(3) == 0 {
				add(x86.JMP(lbl()))
			} else {
				add(x86.XORQ(g(), g()))
			}
		}
	}
	// Observe everything at the end.
	for _, x := range checkRegs {
		add(x86.ADDQ(x, reg.RBX))
	}
	for _, m := range slots {
		add(x86.ADDQ(m, reg.RBX))
	}
	add(x86.RET())
	return fn, strings.Join(lines, "\n")
}

func coalesceCaseFor(fn *ir.Function) (coalesceCase, bool) {
	p := &arm64{cfg: Config{ARM64PromoteStackSlots: true}}
	c := coalesceCase{nodes: fn.Nodes, slotReg: p.stackSlotPromotions(fn)}
	if len(c.slotReg) == 0 {
		return c, false
	}
	c.excluded = analyzeFunction(fn.Nodes).foldGroups(fn.Nodes)
	c.plan = coalesceSlots(c.nodes, c.slotReg, c.excluded)
	return c, true
}

// randomPlan returns the planner's plan with random entries added or
// removed, or an entirely random one, renaming into the slot registers and
// occasionally into ordinary ones (which the validator must refuse).
func randomPlan(r *rand.Rand, c coalesceCase) coalescePlan {
	targets := []string{"R0", "R1", "R2"}
	for _, v := range c.slotReg {
		targets = append(targets, v, v, v)
	}
	var instrs []int
	for idx, n := range c.nodes {
		if _, ok := n.(*ir.Instruction); ok {
			instrs = append(instrs, idx)
		}
	}
	p := coalescePlan{}
	if r.Intn(2) == 0 {
		p = clonePlan(c.plan)
	}
	switch r.Intn(3) {
	case 0: // a family renamed over a random range of instructions
		f := r.Intn(len(checkRegs))
		reg := targets[r.Intn(len(targets))]
		a := r.Intn(len(instrs))
		b := a + r.Intn(len(instrs)-a)
		for _, idx := range instrs[a : b+1] {
			if p[idx] == nil {
				p[idx] = map[int]string{}
			}
			p[idx][f] = reg
		}
	case 1: // a few random single entries
		for i := r.Intn(4); i >= 0; i-- {
			idx := instrs[r.Intn(len(instrs))]
			if p[idx] == nil {
				p[idx] = map[int]string{}
			}
			p[idx][r.Intn(len(checkRegs))] = targets[r.Intn(len(targets))]
		}
	case 2: // drop random entries
		for idx, m := range p {
			for f := range m {
				if r.Intn(3) == 0 {
					delete(m, f)
				}
			}
			if len(m) == 0 {
				delete(p, idx)
			}
		}
	}
	return p
}

// checkPlan executes the original and renamed programs side by side over
// every path of at most maxSteps instructions, returning a description of the
// first read that disagrees, or "".
func checkPlan(c coalesceCase, plan coalescePlan) string {
	g, ok := buildCFG(c.nodes)
	if !ok {
		return ""
	}
	const maxSteps = 40
	var disps []int
	for d := range c.slotReg {
		disps = append(disps, d)
	}
	famLoc := func(k, f int) string {
		if name, ok := plan[g.node[k]][f]; ok {
			return name
		}
		return armReg[reg.Index(f)]
	}
	type world struct {
		vars map[string]uint64 // original: "f<n>" or "s<disp>"
		locs map[string]uint64 // renamed: arm64 register
	}
	init := world{vars: map[string]uint64{}, locs: map[string]uint64{}}
	for f := 0; f < 16; f++ {
		if name, ok := armReg[reg.Index(f)]; ok {
			v := uint64(1000 + f)
			init.vars[fmt.Sprintf("f%d", f)] = v
			init.locs[name] = v
		}
	}
	for _, d := range disps {
		v := uint64(2000 + d)
		init.vars[fmt.Sprintf("s%d", d)] = v
		init.locs[c.slotReg[d]] = v
	}
	var bad string
	budget := 20000 // instructions executed over all paths
	var walk func(k, steps int, w world)
	walk = func(k, steps int, w world) {
		if bad != "" || k >= len(g.ins) || steps > maxSteps || budget == 0 {
			return
		}
		budget--
		ins := g.ins[k]
		e := familyEffects(ins)
		type ref struct{ v, l string }
		var reads, writes []ref
		for f := 0; f < 16; f++ {
			if e.uses&(1<<f) != 0 {
				reads = append(reads, ref{fmt.Sprintf("f%d", f), famLoc(k, f)})
			}
			if e.defs&(1<<f) != 0 {
				writes = append(writes, ref{fmt.Sprintf("f%d", f), famLoc(k, f)})
			}
		}
		if d, use, def, ok := slotEffect(ins); ok {
			if l, promoted := c.slotReg[d]; promoted {
				if use {
					reads = append(reads, ref{fmt.Sprintf("s%d", d), l})
				}
				if def {
					writes = append(writes, ref{fmt.Sprintf("s%d", d), l})
				}
			}
		}
		// Distinct variables named by one location in a non-copy
		// instruction: the lowering may write one before reading the other.
		if ins.Opcode != "MOVQ" {
			byLoc := map[string]string{}
			for _, x := range append(append([]ref{}, reads...), writes...) {
				if o, dup := byLoc[x.l]; dup && o != x.v {
					bad = fmt.Sprintf("instruction %d (%s) names %s and %s as %s", k, ins.Opcode, o, x.v, x.l)
					return
				}
				byLoc[x.l] = x.v
			}
		}
		h := uint64(k+1) * 0x9e3779b97f4a7c15
		for _, x := range reads {
			ov, rv := w.vars[x.v], w.locs[x.l]
			if ov != rv {
				bad = fmt.Sprintf("instruction %d (%s) reads %s=%d from %s, which holds %d", k, ins.Opcode, x.v, ov, x.l, rv)
				return
			}
			h = (h ^ ov) * 0x100000001b3
		}
		nw := world{vars: map[string]uint64{}, locs: map[string]uint64{}}
		for kk, v := range w.vars {
			nw.vars[kk] = v
		}
		for kk, v := range w.locs {
			nw.locs[kk] = v
		}
		for _, x := range writes {
			v := h ^ uint64(len(x.v))
			if ins.Opcode == "MOVQ" && len(reads) == 1 {
				v = w.vars[reads[0].v]
			}
			nw.vars[x.v] = v
			nw.locs[x.l] = v
		}
		for _, s := range g.succ[k] {
			walk(s, steps+1, nw)
		}
	}
	walk(0, 0, init)
	return bad
}
