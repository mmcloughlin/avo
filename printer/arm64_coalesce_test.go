package printer

import (
	"strings"
	"testing"

	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
	"github.com/mmcloughlin/avo/x86"
)

// Register families by x86 physical index.
const (
	famAX = 0
	famCX = 1
	famDX = 2
	famBX = 3
)

// coalesceCase builds a one-function program whose body starts with RAX
// defined and two promotable 8-byte slots, and returns what coalescing
// decides for it.
type coalesceCase struct {
	nodes    []ir.Node
	slotReg  map[int]string
	plan     coalescePlan
	excluded map[int]bool
}

// fnBuilder assembles a function from x86 constructors; the build package
// cannot be used here (it imports this one).
type fnBuilder struct {
	t *testing.T
	f *ir.Function
}

func (b *fnBuilder) add(i *ir.Instruction, err error) {
	b.t.Helper()
	if err != nil {
		b.t.Fatal(err)
	}
	b.f.AddInstruction(i)
}

func buildCoalesceCase(t *testing.T, body func(ctx *fnBuilder, s0, s8 operand.Mem)) coalesceCase {
	t.Helper()
	ctx := &fnBuilder{t, ir.NewFunction("coalesce")}
	local := ctx.f.AllocLocal(16)
	ctx.add(x86.MOVQ(operand.U32(7), reg.RAX))
	body(ctx, local, local.Offset(8))
	ctx.add(x86.RET())
	f := ctx.f
	p := &arm64{cfg: Config{ARM64PromoteStackSlots: true}}
	c := coalesceCase{nodes: f.Nodes, slotReg: p.stackSlotPromotions(f)}
	c.excluded = analyzeFunction(f.Nodes).foldGroups(f.Nodes)
	c.plan = coalesceSlots(c.nodes, c.slotReg, c.excluded)
	if err := validateCoalesce(c.nodes, c.slotReg, c.plan, c.excluded); err != nil {
		t.Fatalf("planner output failed validation: %v", err)
	}
	return c
}

// renamed reports whether family f is renamed anywhere in the plan.
func (c coalesceCase) renamed(f int) bool {
	for _, m := range c.plan {
		if _, ok := m[f]; ok {
			return true
		}
	}
	return false
}

// TestCoalesceRules pins which temporaries coalescing folds into a slot's
// register, including the refusals no lowering in the property tests happens
// to miscompile (a value from the caller, two class members in one
// instruction), which only the planner and the validator enforce.
func TestCoalesceRules(t *testing.T) {
	cases := []struct {
		name   string
		body   func(ctx *fnBuilder, s0, s8 operand.Mem)
		folded []int
		kept   []int
	}{
		{"read-modify-write through a temporary", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(reg.RAX, s0))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.ADDQ(operand.U32(1), reg.RCX))
			ctx.add(x86.MOVQ(reg.RCX, s0))
			ctx.add(x86.MOVQ(s0, reg.RDX))
			ctx.add(x86.ADDQ(reg.RDX, reg.RAX))
		}, []int{famCX, famDX}, nil},
		{"value from the caller", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(reg.RBX, s0))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.ADDQ(reg.RCX, reg.RAX))
		}, []int{famCX}, []int{famBX}},
		{"live across a store to the slot", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(reg.RAX, s0))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.MOVQ(operand.U32(3), s0))
			ctx.add(x86.ADDQ(reg.RCX, reg.RAX))
			ctx.add(x86.MOVQ(s0, reg.RDX))
			ctx.add(x86.ADDQ(reg.RDX, reg.RAX))
		}, []int{famDX}, []int{famCX}},
		{"two class members in one instruction", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(reg.RAX, s0))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.MOVQ(reg.RCX, reg.RDX))
			ctx.add(x86.ADDQ(reg.RCX, reg.RDX))
			ctx.add(x86.MOVQ(reg.RDX, s0))
			ctx.add(x86.MOVQ(s0, reg.RBX))
			ctx.add(x86.ADDQ(reg.RBX, reg.RAX))
		}, nil, nil},
		{"implicit operand", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(reg.RAX, s0))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.MOVQ(reg.RCX, reg.RAX))
			ctx.add(x86.MULQ(reg.RCX))
			ctx.add(x86.MOVQ(reg.RAX, s8))
			ctx.add(x86.MOVQ(s8, reg.RBX))
			ctx.add(x86.ADDQ(reg.RBX, reg.RDX))
			ctx.add(x86.MOVQ(reg.RDX, s0))
		}, nil, nil}, // the implicit-operand check below applies
		{"copy absorbed by a shift", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(operand.U32(3), s0))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.MOVQ(reg.RCX, reg.RDX))
			ctx.add(x86.SHLQ(operand.U8(5), reg.RDX))
			ctx.add(x86.ADDQ(reg.RDX, reg.RAX))
			ctx.add(x86.ADDQ(reg.RCX, reg.RAX))
		}, nil, []int{famCX}},
		{"copy added back into its slot", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(operand.U32(3), s0))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.ADDQ(reg.RCX, s0))
			ctx.add(x86.MOVQ(s0, reg.RDX))
			ctx.add(x86.ADDQ(reg.RDX, reg.RAX))
		}, []int{famDX}, []int{famCX}},
		{"unreachable code", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(reg.RAX, s0))
			ctx.add(x86.JMP(operand.LabelRef("done")))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.ADDQ(operand.U32(9), reg.RCX))
			ctx.add(x86.MOVQ(reg.RCX, s0))
			ctx.f.AddLabel("done")
			ctx.add(x86.MOVQ(s0, reg.RDX))
			ctx.add(x86.ADDQ(reg.RDX, reg.RAX))
		}, []int{famDX}, []int{famCX}},
		{"live around a back edge", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(reg.RAX, s0))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.MOVQ(operand.U32(2), reg.RDX))
			ctx.f.AddLabel("loop")
			ctx.add(x86.ADDQ(reg.RCX, reg.RAX))
			ctx.add(x86.MOVQ(reg.RAX, s0))
			ctx.add(x86.SUBQ(operand.U32(1), reg.RDX))
			ctx.add(x86.JNE(operand.LabelRef("loop")))
			ctx.add(x86.MOVQ(s0, reg.RBX))
			ctx.add(x86.ADDQ(reg.RBX, reg.RAX))
		}, nil, []int{famCX}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cc := buildCoalesceCase(t, c.body)
			for _, f := range c.folded {
				if !cc.renamed(f) {
					t.Errorf("family %d not coalesced", f)
				}
			}
			for _, f := range c.kept {
				if cc.renamed(f) {
					t.Errorf("family %d coalesced, want it kept", f)
				}
			}
			// No family is renamed where it is also an implicit operand, and
			// no instruction names two families by one register unless it is
			// a copy between them.
			for idx, m := range cc.plan {
				ins := cc.nodes[idx].(*ir.Instruction)
				for f := range m {
					if familyEffects(ins).implicit&(1<<f) != 0 {
						t.Errorf("%s renames family %d, an implicit operand", ins.Opcode, f)
					}
				}
				if ins.Opcode == "MOVQ" {
					continue
				}
				seen := map[string]bool{}
				for _, r := range m {
					if seen[r] {
						t.Errorf("%s names two families as %s", ins.Opcode, r)
					}
					seen[r] = true
				}
			}
		})
	}
}

// TestValidateCoalesceRejects feeds the validator plans the planner would
// never make and checks each is refused, so the second line of defence has
// teeth of its own.
func TestValidateCoalesceRejects(t *testing.T) {
	rmw := func(ctx *fnBuilder, s0, s8 operand.Mem) {
		ctx.add(x86.MOVQ(reg.RAX, s0))
		ctx.add(x86.MOVQ(s0, reg.RCX))
		ctx.add(x86.ADDQ(operand.U32(1), reg.RCX))
		ctx.add(x86.MOVQ(reg.RCX, s0))
		ctx.add(x86.MOVQ(s0, reg.RDX))
		ctx.add(x86.ADDQ(reg.RDX, reg.RAX))
	}
	// Instructions of rmw by position in the function (node 0 is the MOVQ
	// $7, AX that buildCoalesceCase emits first).
	const (
		storeAX = 1 + iota
		loadCX
		addCX
		storeCX
		loadDX
		addDX
	)
	cases := []struct {
		name   string
		body   func(ctx *fnBuilder, s0, s8 operand.Mem)
		mutate func(c coalesceCase) coalescePlan
		want   string
	}{
		{"half a web renamed", rmw, func(c coalesceCase) coalescePlan {
			p := clonePlan(c.plan)
			delete(p[c.at(addCX)], famCX)
			return p
		}, "does not hold it"},
		{"value from before the function", rmw, func(c coalesceCase) coalescePlan {
			p := clonePlan(c.plan)
			p[c.at(storeAX)] = map[int]string{famAX: c.slotReg[0]}
			return p
		}, "does not hold it"},
		{"slot clobbered while its value is live", rmw, func(c coalesceCase) coalescePlan {
			p := clonePlan(c.plan)
			// Keep AX's final add reading through the slot register while the
			// slot's register holds DX's value, both RAX and RDX renamed.
			p[c.at(addDX)] = map[int]string{famAX: c.slotReg[0], famDX: c.slotReg[0]}
			return p
		}, "does not hold it"},
		{"renamed inside a fold", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(reg.RAX, s0))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.MOVQ(reg.RCX, reg.RDX))
			ctx.add(x86.SHLQ(operand.U8(5), reg.RDX))
			ctx.add(x86.ADDQ(reg.RDX, reg.RAX))
			ctx.add(x86.ADDQ(reg.RCX, reg.RAX))
		}, func(c coalesceCase) coalescePlan {
			p := clonePlan(c.plan)
			for idx := range c.excluded {
				p[idx] = map[int]string{famCX: c.slotReg[0]}
			}
			return p
		}, "cross-instruction fold"},
		{"family and slot share a register", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(operand.U32(3), s0))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.ADDQ(reg.RCX, s0))
			ctx.add(x86.ADDQ(reg.RCX, reg.RAX))
		}, func(c coalesceCase) coalescePlan {
			p := clonePlan(c.plan)
			for _, k := range []int{2, 3, 4} {
				if p[c.at(k)] == nil {
					p[c.at(k)] = map[int]string{}
				}
				p[c.at(k)][famCX] = c.slotReg[0]
			}
			return p
		}, "both named"},
		{"family the instruction does not name", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(reg.RAX, s0))
			ctx.add(x86.MOVQ(reg.RAX, s8))
			ctx.add(x86.ADDQ(reg.RBX, s8))
			ctx.add(x86.MOVQ(s8, reg.RDX))
			ctx.add(x86.ADDQ(reg.RDX, reg.RAX))
		}, func(c coalesceCase) coalescePlan {
			p := clonePlan(c.plan)
			if p[c.at(3)] == nil {
				p[c.at(3)] = map[int]string{}
			}
			p[c.at(3)][famAX] = c.slotReg[0]
			return p
		}, "does not name"},
		{"register outside the slot registers", rmw, func(c coalesceCase) coalescePlan {
			p := clonePlan(c.plan)
			p[c.at(loadCX)] = map[int]string{famCX: "R16"}
			return p
		}, "not a promoted slot's register"},
		{"renamed in unreachable code", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(reg.RAX, s0))
			ctx.add(x86.JMP(operand.LabelRef("done")))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.MOVQ(reg.RCX, s0))
			ctx.f.AddLabel("done")
			ctx.add(x86.RET())
		}, func(c coalesceCase) coalescePlan {
			p := clonePlan(c.plan)
			p[c.at(3)] = map[int]string{famCX: c.slotReg[0]}
			return p
		}, "unreachable"},
		{"implicit operand renamed", func(ctx *fnBuilder, s0, s8 operand.Mem) {
			ctx.add(x86.MOVQ(reg.RAX, s0))
			ctx.add(x86.MOVQ(s0, reg.RCX))
			ctx.add(x86.MOVQ(reg.RCX, reg.RAX))
			ctx.add(x86.MULQ(reg.RCX))
			ctx.add(x86.MOVQ(reg.RAX, s8))
		}, func(c coalesceCase) coalescePlan {
			p := clonePlan(c.plan)
			p[c.at(4)] = map[int]string{famAX: c.slotReg[8]}
			return p
		}, "implicit operand"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cc := buildCoalesceCase(t, c.body)
			err := validateCoalesce(cc.nodes, cc.slotReg, c.mutate(cc), cc.excluded)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("validateCoalesce = %v, want an error containing %q", err, c.want)
			}
		})
	}
}

// at returns the node index of the k'th instruction.
func (c coalesceCase) at(k int) int {
	for idx, n := range c.nodes {
		if _, ok := n.(*ir.Instruction); ok {
			if k == 0 {
				return idx
			}
			k--
		}
	}
	panic("no such instruction")
}

func clonePlan(p coalescePlan) coalescePlan {
	out := coalescePlan{}
	for idx, m := range p {
		out[idx] = map[int]string{}
		for f, r := range m {
			out[idx][f] = r
		}
	}
	return out
}
