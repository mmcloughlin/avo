package printer

import (
	"sort"
	"strings"

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
	n := len(g.ins)

	// Per-instruction family effects.
	const nfam = 16
	uses := make([]uint16, n)
	defs := make([]uint16, n)
	bad := make([]uint16, n) // mentions that make the family's web ineligible
	for k, ins := range g.ins {
		e := familyEffects(ins)
		uses[k], defs[k] = e.uses, e.defs
		bad[k] = e.implicit | e.high
		if excluded[g.node[k]] {
			bad[k] |= e.explicit
		}
	}

	// Reaching definitions, one family at a time. Definition 0 of each family
	// is the function entry.
	type def struct{ k, f int }
	var allDefs []def
	defID := make([]map[int]int, nfam) // family -> instruction -> global def id
	entryID := make([]int, nfam)
	for f := 0; f < nfam; f++ {
		defID[f] = map[int]int{}
		entryID[f] = len(allDefs)
		allDefs = append(allDefs, def{-1, f})
		for k := 0; k < n; k++ {
			if defs[k]&(1<<f) != 0 {
				defID[f][k] = len(allDefs)
				allDefs = append(allDefs, def{k, f})
			}
		}
	}
	uf := newUnionFind(len(allDefs))
	useWeb := make([]map[int]int, n) // instruction -> family -> a def id in the web the use reads
	for f := 0; f < nfam; f++ {
		local := []int{entryID[f]}
		for k := 0; k < n; k++ {
			if id, ok := defID[f][k]; ok {
				local = append(local, id)
			}
		}
		in := reachingDefs(g, len(local), func(k int) (int, bool) {
			if defs[k]&(1<<f) == 0 {
				return 0, false
			}
			return sort.SearchInts(local, defID[f][k]), true
		})
		for k := 0; k < n; k++ {
			if uses[k]&(1<<f) == 0 {
				continue
			}
			first := -1
			in[k].each(func(i int) {
				if first < 0 {
					first = local[i]
				} else {
					uf.union(first, local[i])
				}
			})
			if first < 0 {
				continue // unreachable: no definition reaches the use
			}
			if id, ok := defID[f][k]; ok {
				uf.union(first, id)
			}
			if useWeb[k] == nil {
				useWeb[k] = map[int]int{}
			}
			useWeb[k][f] = first
		}
	}

	// Collect webs: their mentions, and whether they are eligible.
	type web struct {
		f        int
		mentions []int // instructions naming the family in this web
		defs     []int // instructions defining it
		ok       bool
	}
	webs := map[int]*web{}
	webOf := func(id int) *web {
		root := uf.find(id)
		w := webs[root]
		if w == nil {
			w = &web{f: allDefs[id].f, ok: true}
			webs[root] = w
		}
		return w
	}
	mentionWeb := make([]map[int]*web, n)
	for k := 0; k < n; k++ {
		for f := 0; f < nfam; f++ {
			var w *web
			if id, ok := useWeb[k][f]; ok {
				w = webOf(id)
			} else if id, ok := defID[f][k]; ok {
				w = webOf(id)
			} else {
				continue
			}
			if mentionWeb[k] == nil {
				mentionWeb[k] = map[int]*web{}
			}
			mentionWeb[k][f] = w
			w.mentions = append(w.mentions, k)
			if defs[k]&(1<<f) != 0 {
				w.defs = append(w.defs, k)
			}
			if bad[k]&(1<<f) != 0 {
				w.ok = false
			}
		}
	}
	for f := 0; f < nfam; f++ {
		if w := webs[uf.find(entryID[f])]; w != nil {
			w.ok = false // may carry a value from the caller
		}
	}

	// Slot accesses. Promotion guarantees every access to a promoted slot is a
	// slotEffect: a clean MOVQ load or store (a copy between the slot and a
	// register, or an immediate store) or an ALU operation on the slot.
	type slotAccess struct {
		use, def bool
		copyWeb  *web // the other side of a MOVQ copy; nil otherwise
	}
	access := make([]map[int]slotAccess, n) // instruction -> slot disp -> access
	for k, ins := range g.ins {
		d, use, def, ok := slotEffect(ins)
		if !ok {
			continue
		}
		if _, promoted := slotReg[d]; !promoted {
			continue
		}
		a := slotAccess{use: use, def: def}
		if ins.Opcode == "MOVQ" {
			src, dst := ins.Operands[0], ins.Operands[1]
			if use {
				a.copyWeb = mentionWeb[k][regFamily(dst)]
			} else if f := regFamily(src); f >= 0 {
				a.copyWeb = mentionWeb[k][f]
			}
		}
		access[k] = map[int]slotAccess{d: a}
	}

	// Liveness, computed on demand per variable. A variable is a web or a
	// slot; live[k] is whether it is live out of instruction k.
	webLive := map[*web][]bool{}
	liveOfWeb := func(w *web) []bool {
		if l, ok := webLive[w]; ok {
			return l
		}
		l := liveOut(g, func(k int) (use, kill bool) {
			use = uses[k]&(1<<w.f) != 0 && mentionWeb[k][w.f] == w
			return use, defs[k]&(1<<w.f) != 0
		})
		webLive[w] = l
		return l
	}
	slotLive := map[int][]bool{}
	liveOfSlot := func(d int) []bool {
		if l, ok := slotLive[d]; ok {
			return l
		}
		l := liveOut(g, func(k int) (use, kill bool) {
			a := access[k][d]
			return a.use, a.def
		})
		slotLive[d] = l
		return l
	}
	// isCopy reports whether instruction k is a MOVQ from web or slot "from"
	// into web or slot "to" (exactly one of each pair is set).
	isCopy := func(k int, fromW *web, fromS int, toW *web, toS int) bool {
		ins := g.ins[k]
		if ins.Opcode != "MOVQ" || len(ins.Operands) != 2 {
			return false
		}
		side := func(op operand.Op, w *web, s int) bool {
			if w != nil {
				f := regFamily(op)
				return f >= 0 && isFullGP(op) && mentionWeb[k][f] == w
			}
			d, ok := frameSlot(op)
			return ok && d == s
		}
		return side(ins.Operands[0], fromW, fromS) && side(ins.Operands[1], toW, toS)
	}
	const noSlot = -1
	webSlotInterfere := func(w *web, s int) bool {
		sl := liveOfSlot(s)
		for _, k := range w.defs {
			if sl[k] && !isCopy(k, nil, s, w, noSlot) {
				return true
			}
		}
		wl := liveOfWeb(w)
		for k := 0; k < n; k++ {
			if a, ok := access[k][s]; ok && a.def && wl[k] && !isCopy(k, w, noSlot, nil, s) {
				return true
			}
		}
		// An ALU operation on the slot naming the web too would name one
		// register twice.
		for _, k := range w.mentions {
			if a, ok := access[k][s]; ok && a.copyWeb == nil {
				return true
			}
		}
		return false
	}
	webWebInterfere := func(a, b *web) bool {
		la, lb := liveOfWeb(a), liveOfWeb(b)
		for _, k := range a.defs {
			if lb[k] && !isCopy(k, b, noSlot, a, noSlot) {
				return true
			}
		}
		for _, k := range b.defs {
			if la[k] && !isCopy(k, a, noSlot, b, noSlot) {
				return true
			}
		}
		// Both named by one instruction: only a copy between them is safe.
		for _, k := range a.mentions {
			if mentionWeb[k][b.f] == b && !isCopy(k, a, noSlot, b, noSlot) && !isCopy(k, b, noSlot, a, noSlot) {
				return true
			}
		}
		return false
	}

	// Candidates: an eligible web and a slot it is copied to or from, scored by
	// the moves coalescing would remove.
	type cand struct {
		w     *web
		slot  int
		score int
		first int
	}
	var cands []cand
	score := map[*web]map[int]int{}
	for k := 0; k < n; k++ {
		for d, a := range access[k] {
			if a.copyWeb == nil || !a.copyWeb.ok {
				continue
			}
			if score[a.copyWeb] == nil {
				score[a.copyWeb] = map[int]int{}
			}
			score[a.copyWeb][d]++
		}
	}
	for w, m := range score {
		for d, c := range m {
			cands = append(cands, cand{w, d, c, w.mentions[0]*nfam + w.f})
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

	class := map[int][]*web{}
	taken := map[*web]bool{}
	plan := coalescePlan{}
	for _, c := range cands {
		if taken[c.w] || webSlotInterfere(c.w, c.slot) {
			continue
		}
		clash := false
		for _, o := range class[c.slot] {
			if webWebInterfere(c.w, o) {
				clash = true
				break
			}
		}
		if clash {
			continue
		}
		taken[c.w] = true
		class[c.slot] = append(class[c.slot], c.w)
		for _, k := range c.w.mentions {
			idx := g.node[k]
			if plan[idx] == nil {
				plan[idx] = map[int]string{}
			}
			plan[idx][c.w.f] = slotReg[c.slot]
		}
	}
	if len(plan) == 0 {
		return nil
	}
	return plan
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

// conditionalWrite reports whether opcode may leave its destination register
// unchanged, so that its result depends on the old value even at full width.
func conditionalWrite(opcode string) bool {
	return strings.HasPrefix(opcode, "CMOV") || strings.HasPrefix(opcode, "BSF") ||
		strings.HasPrefix(opcode, "BSR")
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
