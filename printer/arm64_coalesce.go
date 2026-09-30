package printer

import (
	"sort"

	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

// Stack-slot coalescing.
//
// Promotion alone (stackSlotPromotions) turns each access to a promoted slot
// into a register move, which saves the memory round trip but not the
// instruction: "MOVQ 40(SP), R8; DECQ R8; MOVQ R8, 40(SP)" becomes three
// instructions still. Coalescing removes the moves. It gives the temporary's
// value -- the web of R8 definitions and uses that the load starts -- the
// slot's register for its whole lifetime, so the loop counter above lowers to
// "SUB $1, R20, R20" and both moves become self-moves, which are not emitted.
//
// This is copy coalescing in the classic register-allocation sense, with the
// promoted slot as one more variable:
//
//   - A web is a set of definitions of one x86 register family joined by the
//     uses they reach (reaching definitions over the function's control-flow
//     graph). An instruction that reads and writes the family, or writes only
//     part of it, joins the web it reads, since its lowering names one register
//     for both roles. Each register mention in the function belongs to
//     exactly one web, so renaming a web renames every mention of that value
//     and nothing else.
//   - Two variables interfere when one is live out of an instruction that
//     defines the other, unless that instruction is a MOVQ copy from the one
//     to the other (after which both hold the same value). A web joins a
//     slot's class only if it interferes with neither the slot nor any web
//     already in the class, so the register holds at most one live value at
//     every point, which is what the slot's memory held.
//   - Two class members may not both appear in one instruction, except as the
//     two sides of a MOVQ copy: lowerings assume distinct x86 registers are
//     distinct arm64 registers, and may write one before reading the other.
//
// A web is left alone if it could be live on entry to the function (its value
// would come from the caller), if any instruction touches its family
// implicitly (MULQ's RDX:RAX, say: renaming reaches only explicit operands),
// if it names a high-byte register, or if it touches an instruction that one
// of the cross-instruction folds (shiftFolds, shiftExtractFold, setccFolds)
// rewrites: those fix register names at analysis time, before any renaming.
// Control flow the graph cannot follow (a branch to anything but a local label)
// disables coalescing for the whole function.
//
// A definition nothing reads may vanish: a web whose only instruction is a load
// from its slot is coalesced, and the load is a self-move. Registers carry no
// results out of an ABI0 function, so no caller can observe the difference.

// coalescePlan maps a node index to the renamed register families at that
// instruction: family (x86 physical index) to arm64 register.
type coalescePlan map[int]map[int]string

// numFamilies is the number of x86 general-purpose register families.
const numFamilies = 16

// noSlot stands for "not a slot" where a variable is either a web or a slot.
const noSlot = -1

// web is a set of definitions of one register family joined by the uses they
// reach (see coalesceSlots).
type web struct {
	f        int
	mentions []int // instructions naming the family in this web
	defs     []int // instructions defining it
	ok       bool  // eligible for coalescing
}

// slotAccess is one instruction's access to a promoted slot.
type slotAccess struct {
	use, def bool
	copyWeb  *web // the other side of a MOVQ copy; nil otherwise
}

// famDef is one definition of a register family: instruction k, or the
// function entry for k == -1.
type famDef struct{ k, f int }

// coalescer holds the analyses coalesceSlots builds for one function.
type coalescer struct {
	g       cfg
	n       int
	slotReg map[int]string

	// Per instruction: families read, written, and mentioned in a way that
	// makes their web ineligible.
	uses, defs, bad []uint16

	// Definitions, numbered globally; per family, instruction -> def id, and
	// the id of the entry definition.
	allDefs []famDef
	defID   []map[int]int
	entryID []int
	uf      unionFind

	useWeb     []map[int]int  // instruction -> family -> a def id in the web the use reads
	webs       map[int]*web   // union-find root -> web
	mentionWeb []map[int]*web // instruction -> family -> web
	access     []map[int]slotAccess

	webLive  map[*web][]bool
	slotLive map[int][]bool
}

// coalesceSlots computes the coalescing plan for one function. slotReg is the
// promotion map; excluded marks node indices a cross-instruction fold rewrites.
func coalesceSlots(nodes []ir.Node, slotReg map[int]string, excluded map[int]bool) coalescePlan {
	if len(slotReg) == 0 {
		return nil
	}
	g, ok := buildCFG(nodes)
	if !ok {
		return nil
	}
	c := &coalescer{
		g: g, n: len(g.ins), slotReg: slotReg,
		webLive: map[*web][]bool{}, slotLive: map[int][]bool{},
	}
	c.effects(excluded)
	c.numberDefs()
	for f := 0; f < numFamilies; f++ {
		c.joinFamily(f)
	}
	c.collectWebs()
	c.markIneligible()
	c.collectAccesses()
	return c.assign(c.candidates())
}

// effects records each instruction's family effects.
func (c *coalescer) effects(excluded map[int]bool) {
	c.uses = make([]uint16, c.n)
	c.defs = make([]uint16, c.n)
	c.bad = make([]uint16, c.n)
	for k, ins := range c.g.ins {
		e := familyEffects(ins)
		c.uses[k], c.defs[k] = e.uses, e.defs
		c.bad[k] = e.implicit | e.high
		if excluded[c.g.node[k]] {
			c.bad[k] |= e.explicit
		}
	}
}

// numberDefs numbers every definition, family by family, each family's entry
// definition first.
func (c *coalescer) numberDefs() {
	c.defID = make([]map[int]int, numFamilies)
	c.entryID = make([]int, numFamilies)
	for f := 0; f < numFamilies; f++ {
		c.defID[f] = map[int]int{}
		c.entryID[f] = len(c.allDefs)
		c.allDefs = append(c.allDefs, famDef{-1, f})
		for k := 0; k < c.n; k++ {
			if c.defs[k]&(1<<f) != 0 {
				c.defID[f][k] = len(c.allDefs)
				c.allDefs = append(c.allDefs, famDef{k, f})
			}
		}
	}
	c.uf = newUnionFind(len(c.allDefs))
	c.useWeb = make([]map[int]int, c.n)
}

// joinFamily solves reaching definitions for family f and joins, for every
// use, the definitions that reach it (and the instruction's own definition,
// if it also writes f) into one web.
func (c *coalescer) joinFamily(f int) {
	local := []int{c.entryID[f]}
	for k := 0; k < c.n; k++ {
		if id, ok := c.defID[f][k]; ok {
			local = append(local, id)
		}
	}
	in := reachingDefs(c.g, len(local), func(k int) (int, bool) {
		if c.defs[k]&(1<<f) == 0 {
			return 0, false
		}
		return sort.SearchInts(local, c.defID[f][k]), true
	})
	for k := 0; k < c.n; k++ {
		if c.uses[k]&(1<<f) == 0 {
			continue
		}
		first := c.joinReaching(in[k], local)
		if first < 0 {
			continue // unreachable: no definition reaches the use
		}
		if id, ok := c.defID[f][k]; ok {
			c.uf.union(first, id)
		}
		if c.useWeb[k] == nil {
			c.useWeb[k] = map[int]int{}
		}
		c.useWeb[k][f] = first
	}
}

// joinReaching unions the definitions in reach (indices into local) and
// returns the first one's id, or -1 if none reach.
func (c *coalescer) joinReaching(reach bitset, local []int) int {
	first := -1
	reach.each(func(i int) {
		if first < 0 {
			first = local[i]
		} else {
			c.uf.union(first, local[i])
		}
	})
	return first
}

// webOf returns the web holding definition id, creating it on first sight.
func (c *coalescer) webOf(id int) *web {
	root := c.uf.find(id)
	w := c.webs[root]
	if w == nil {
		w = &web{f: c.allDefs[id].f, ok: true}
		c.webs[root] = w
	}
	return w
}

// mentionAt returns the web of family f's mention at instruction k, if any.
func (c *coalescer) mentionAt(k, f int) (*web, bool) {
	if id, ok := c.useWeb[k][f]; ok {
		return c.webOf(id), true
	}
	if id, ok := c.defID[f][k]; ok {
		return c.webOf(id), true
	}
	return nil, false
}

// collectWebs assigns every mention to its web and records the web's mentions,
// definitions and eligibility.
func (c *coalescer) collectWebs() {
	c.webs = map[int]*web{}
	c.mentionWeb = make([]map[int]*web, c.n)
	for k := 0; k < c.n; k++ {
		for f := 0; f < numFamilies; f++ {
			w, ok := c.mentionAt(k, f)
			if !ok {
				continue
			}
			if c.mentionWeb[k] == nil {
				c.mentionWeb[k] = map[int]*web{}
			}
			c.mentionWeb[k][f] = w
			w.mentions = append(w.mentions, k)
			if c.defs[k]&(1<<f) != 0 {
				w.defs = append(w.defs, k)
			}
			if c.bad[k]&(1<<f) != 0 {
				w.ok = false
			}
		}
	}
}

// markIneligible rules out webs that may carry a value from the caller and
// webs that touch unreachable code. Unreachable code is never renamed, so the
// validator, which only walks reachable instructions, checks everything that
// is.
func (c *coalescer) markIneligible() {
	for f := 0; f < numFamilies; f++ {
		if w := c.webs[c.uf.find(c.entryID[f])]; w != nil {
			w.ok = false
		}
	}
	reach := c.g.reachable()
	for _, w := range c.webs {
		for _, k := range w.mentions {
			if !reach[k] {
				w.ok = false
			}
		}
	}
}

// collectAccesses records the promoted-slot accesses. Promotion guarantees
// every access to a promoted slot is a slotEffect: a clean MOVQ load or store
// (a copy between the slot and a register, or an immediate store) or an ALU
// operation on the slot.
func (c *coalescer) collectAccesses() {
	c.access = make([]map[int]slotAccess, c.n)
	for k, ins := range c.g.ins {
		s, ok := slotEffect(ins)
		if !ok {
			continue
		}
		if _, promoted := c.slotReg[s.disp]; !promoted {
			continue
		}
		a := slotAccess{use: s.use, def: s.def}
		if ins.Opcode == "MOVQ" {
			src, dst := ins.Operands[0], ins.Operands[1]
			if s.use {
				a.copyWeb = c.mentionWeb[k][regFamily(dst)]
			} else if f := regFamily(src); f >= 0 {
				a.copyWeb = c.mentionWeb[k][f]
			}
		}
		c.access[k] = map[int]slotAccess{s.disp: a}
	}
}

// liveOfWeb returns whether web w is live out of each instruction.
func (c *coalescer) liveOfWeb(w *web) []bool {
	if l, ok := c.webLive[w]; ok {
		return l
	}
	l := liveOut(c.g, func(k int) (use, kill bool) {
		use = c.uses[k]&(1<<w.f) != 0 && c.mentionWeb[k][w.f] == w
		return use, c.defs[k]&(1<<w.f) != 0
	})
	c.webLive[w] = l
	return l
}

// liveOfSlot returns whether slot d is live out of each instruction.
func (c *coalescer) liveOfSlot(d int) []bool {
	if l, ok := c.slotLive[d]; ok {
		return l
	}
	l := liveOut(c.g, func(k int) (use, kill bool) {
		a := c.access[k][d]
		return a.use, a.def
	})
	c.slotLive[d] = l
	return l
}

// copySide reports whether op, an operand of instruction k, is web w (when w
// is set) or slot s.
func (c *coalescer) copySide(k int, op operand.Op, w *web, s int) bool {
	if w != nil {
		f := regFamily(op)
		return f >= 0 && isFullGP(op) && c.mentionWeb[k][f] == w
	}
	d, ok := frameSlot(op)
	return ok && d == s
}

// isCopy reports whether instruction k is a MOVQ from web or slot "from" into
// web or slot "to" (exactly one of each pair is set).
func (c *coalescer) isCopy(k int, fromW *web, fromS int, toW *web, toS int) bool {
	ins := c.g.ins[k]
	if ins.Opcode != "MOVQ" || len(ins.Operands) != 2 {
		return false
	}
	return c.copySide(k, ins.Operands[0], fromW, fromS) && c.copySide(k, ins.Operands[1], toW, toS)
}

// webSlotInterfere reports whether web w and slot s interfere.
func (c *coalescer) webSlotInterfere(w *web, s int) bool {
	sl := c.liveOfSlot(s)
	for _, k := range w.defs {
		if sl[k] && !c.isCopy(k, nil, s, w, noSlot) {
			return true
		}
	}
	wl := c.liveOfWeb(w)
	for k := 0; k < c.n; k++ {
		if a, ok := c.access[k][s]; ok && a.def && wl[k] && !c.isCopy(k, w, noSlot, nil, s) {
			return true
		}
	}
	// An ALU operation on the slot naming the web too would name one
	// register twice.
	for _, k := range w.mentions {
		if a, ok := c.access[k][s]; ok && a.copyWeb == nil {
			return true
		}
	}
	return false
}

// webWebInterfere reports whether webs a and b interfere.
func (c *coalescer) webWebInterfere(a, b *web) bool {
	la, lb := c.liveOfWeb(a), c.liveOfWeb(b)
	for _, k := range a.defs {
		if lb[k] && !c.isCopy(k, b, noSlot, a, noSlot) {
			return true
		}
	}
	for _, k := range b.defs {
		if la[k] && !c.isCopy(k, a, noSlot, b, noSlot) {
			return true
		}
	}
	// Both named by one instruction: only a copy between them is safe.
	for _, k := range a.mentions {
		if c.mentionWeb[k][b.f] == b && !c.isCopy(k, a, noSlot, b, noSlot) && !c.isCopy(k, b, noSlot, a, noSlot) {
			return true
		}
	}
	return false
}

// candidate is an eligible web and a slot it is copied to or from, scored by
// the moves coalescing would remove.
type candidate struct {
	w     *web
	slot  int
	score int
	first int
}

// candidates returns the candidates, best first.
func (c *coalescer) candidates() []candidate {
	score := map[*web]map[int]int{}
	for k := 0; k < c.n; k++ {
		for d, a := range c.access[k] {
			if a.copyWeb == nil || !a.copyWeb.ok {
				continue
			}
			if score[a.copyWeb] == nil {
				score[a.copyWeb] = map[int]int{}
			}
			score[a.copyWeb][d]++
		}
	}
	var cands []candidate
	for w, m := range score {
		for d, n := range m {
			cands = append(cands, candidate{w, d, n, w.mentions[0]*numFamilies + w.f})
		}
	}
	sort.Slice(cands, func(i, j int) bool {
		a, b := cands[i], cands[j]
		if a.score != b.score {
			return a.score > b.score
		}
		if a.first != b.first {
			return a.first < b.first
		}
		return a.slot < b.slot
	})
	return cands
}

// assign coalesces candidates greedily, best first, into their slot's class
// when they interfere with neither the slot nor any web already in the class.
func (c *coalescer) assign(cands []candidate) coalescePlan {
	class := map[int][]*web{}
	taken := map[*web]bool{}
	plan := coalescePlan{}
	for _, cand := range cands {
		if taken[cand.w] || c.webSlotInterfere(cand.w, cand.slot) || c.clashesWithClass(cand.w, class[cand.slot]) {
			continue
		}
		taken[cand.w] = true
		class[cand.slot] = append(class[cand.slot], cand.w)
		for _, k := range cand.w.mentions {
			idx := c.g.node[k]
			if plan[idx] == nil {
				plan[idx] = map[int]string{}
			}
			plan[idx][cand.w.f] = c.slotReg[cand.slot]
		}
	}
	if len(plan) == 0 {
		return nil
	}
	return plan
}

// clashesWithClass reports whether w interferes with any web in class.
func (c *coalescer) clashesWithClass(w *web, class []*web) bool {
	for _, o := range class {
		if c.webWebInterfere(w, o) {
			return true
		}
	}
	return false
}

// famEffects are the register families an instruction reads and writes, as
// bitmasks over x86 physical indices.
type famEffects struct {
	uses, defs uint16
	explicit   uint16 // named by an operand, so renaming reaches them
	implicit   uint16 // read or written without being named
	high       uint16 // named as a high-byte register
}

// familyEffects models an instruction's register effects for coalescing and
// its validation. A write narrower than 32 bits keeps the rest of the
// register, and a conditional write keeps all of it when it does not happen:
// both count as reads of the old value too. An explicit register the
// instruction database lists as neither read nor written is taken as read,
// the conservative reading.
func familyEffects(ins *ir.Instruction) famEffects {
	var e famEffects
	for _, op := range ins.Operands {
		for _, r := range operand.Registers(op) {
			if f := coalesceFamily(r); f >= 0 {
				e.explicit |= 1 << f
				if isHighByte(r) {
					e.high |= 1 << f
				}
			}
		}
	}
	for _, r := range ins.InputRegisters() {
		if f := coalesceFamily(r); f >= 0 {
			e.uses |= 1 << f
		}
	}
	for _, r := range ins.OutputRegisters() {
		f := coalesceFamily(r)
		if f < 0 {
			continue
		}
		e.defs |= 1 << f
		if r.Size() < 4 || conditionalWrite(ins.Opcode) {
			e.uses |= 1 << f
		}
	}
	e.uses |= e.explicit &^ (e.uses | e.defs)
	e.implicit = (e.uses | e.defs) &^ e.explicit
	return e
}

// coalesceFamily is regFamily for the families coalescing may rename: every
// general-purpose register except the stack pointer.
func coalesceFamily(r reg.Register) int {
	f := regFamily(r)
	if f == int(reg.RSP.PhysicalIndex()) {
		return -1
	}
	return f
}

// cfg is a function's instructions with their control-flow successors.
type cfg struct {
	ins  []*ir.Instruction
	node []int   // instruction position -> node index
	succ [][]int // instruction position -> successor positions
	pred [][]int
}

// buildCFG builds the control-flow graph of a function's instruction stream.
// ok is false if some branch target is not a label of the function.
func buildCFG(nodes []ir.Node) (g cfg, ok bool) {
	labels := map[string]int{}
	for idx, nd := range nodes {
		switch v := nd.(type) {
		case ir.Label:
			labels[string(v)] = len(g.ins)
		case *ir.Instruction:
			g.ins = append(g.ins, v)
			g.node = append(g.node, idx)
		}
	}
	n := len(g.ins)
	g.succ = make([][]int, n)
	g.pred = make([][]int, n)
	add := func(from, to int) {
		if to < n {
			g.succ[from] = append(g.succ[from], to)
			g.pred[to] = append(g.pred[to], from)
		}
	}
	for k, ins := range g.ins {
		if ins.IsBranch || ins.Opcode == "JMP" || isConditionalBranch(ins) {
			ref, isLabel := ins.Operands[0].(operand.LabelRef)
			if !isLabel {
				return cfg{}, false
			}
			t, found := labels[string(ref)]
			if !found {
				return cfg{}, false
			}
			add(k, t)
			if isConditionalBranch(ins) {
				add(k, k+1)
			}
			continue
		}
		if !ins.IsTerminal {
			add(k, k+1)
		}
	}
	return g, true
}

// reachable reports which instructions control can reach from the entry.
func (g cfg) reachable() []bool {
	seen := make([]bool, len(g.ins))
	if len(g.ins) == 0 {
		return seen
	}
	stack := []int{0}
	seen[0] = true
	for len(stack) > 0 {
		k := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, s := range g.succ[k] {
			if !seen[s] {
				seen[s] = true
				stack = append(stack, s)
			}
		}
	}
	return seen
}

// reachingDefs solves reaching definitions for one variable whose definitions
// are numbered 0..ndefs-1, with 0 the function entry. gen reports the
// definition an instruction makes, if any. The result is the set reaching the
// start of each instruction.
func reachingDefs(g cfg, ndefs int, gen func(k int) (int, bool)) []bitset {
	n := len(g.ins)
	in := make([]bitset, n)
	out := make([]bitset, n)
	for k := range in {
		in[k] = newBitset(ndefs)
		out[k] = newBitset(ndefs)
	}
	entry := newBitset(ndefs)
	entry.set(0)
	for changed := true; changed; {
		changed = false
		for k := 0; k < n; k++ {
			if k == 0 {
				in[k].or(entry)
			}
			for _, p := range g.pred[k] {
				in[k].or(out[p])
			}
			next := in[k]
			if d, ok := gen(k); ok {
				next = newBitset(ndefs)
				next.set(d)
			}
			if out[k].or(next) {
				changed = true
			}
		}
	}
	return in
}

// liveOut solves backward liveness for one variable: at each instruction,
// whether it is read (use) and whether it is overwritten (kill; a use in the
// same instruction reads the old value first). The result is whether the
// variable is live out of each instruction.
func liveOut(g cfg, at func(k int) (use, kill bool)) []bool {
	n := len(g.ins)
	in := make([]bool, n)
	out := make([]bool, n)
	for changed := true; changed; {
		changed = false
		for k := n - 1; k >= 0; k-- {
			o := false
			for _, s := range g.succ[k] {
				o = o || in[s]
			}
			use, kill := at(k)
			i := use || (o && !kill)
			if o != out[k] || i != in[k] {
				out[k], in[k] = o, i
				changed = true
			}
		}
	}
	return out
}

// bitset is a fixed-size set of small integers.
type bitset []uint64

func newBitset(n int) bitset { return make(bitset, (n+63)/64) }

func (b bitset) set(i int) { b[i/64] |= 1 << (i % 64) }

// or adds o's members to b and reports whether b grew.
func (b bitset) or(o bitset) bool {
	grew := false
	for i := range b {
		if v := b[i] | o[i]; v != b[i] {
			b[i] = v
			grew = true
		}
	}
	return grew
}

func (b bitset) each(fn func(int)) {
	for i, w := range b {
		for j := 0; j < 64; j++ {
			if w&(1<<j) != 0 {
				fn(i*64 + j)
			}
		}
	}
}

type unionFind []int

func newUnionFind(n int) unionFind {
	u := make(unionFind, n)
	for i := range u {
		u[i] = i
	}
	return u
}

func (u unionFind) find(i int) int {
	for u[i] != i {
		u[i] = u[u[i]]
		i = u[i]
	}
	return i
}

func (u unionFind) union(a, b int) { u[u.find(a)] = u.find(b) }
