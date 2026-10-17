package pass_test

import (
	"reflect"
	"testing"

	"github.com/mmcloughlin/avo/attr"
	"github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/pass"
	"github.com/mmcloughlin/avo/reg"
)

func TestZeroExtend32BitOutputs(t *testing.T) {
	collection := reg.NewCollection()
	v16 := collection.GP16()
	v32 := collection.GP32()

	i := &ir.Instruction{
		Outputs: []operand.Op{
			reg.R8B,
			reg.R9W,
			reg.R10L,
			reg.R11,
			v16,
			v32,
		},
	}

	err := pass.ZeroExtend32BitOutputs(i)
	if err != nil {
		t.Fatal(err)
	}

	got := i.Outputs
	expect := []reg.Register{
		reg.R8B,
		reg.R9W,
		reg.R10, // converted from R10L
		reg.R11,
		v16,
		v32.As64(), // converted from 32-bit
	}

	if len(expect) != len(got) {
		t.Fatal("length mismatch")
	}

	for j := range got {
		r, ok := got[j].(reg.Register)
		if !ok {
			t.Fatalf("expected register; got %s", got[j].Asm())
		}

		if !reg.Equal(expect[j], r) {
			t.Fatalf("got %s; expect %s", expect[j].Asm(), r.Asm())
		}
	}
}

func TestLivenessBasic(t *testing.T) {
	// Build: a = 1, b = 2, a = a+b
	ctx := build.NewContext()
	ctx.Function("add")
	a := ctx.GP64()
	b := ctx.GP64()
	ctx.MOVQ(operand.U64(1), a)
	ctx.MOVQ(operand.U64(2), b)
	ctx.ADDQ(a, b)

	AssertLiveness(t, ctx,
		[][]reg.Register{
			{},
			{a},
			{a, b},
		},
		[][]reg.Register{
			{a},
			{a, b},
			{},
		},
	)
}

func AssertLiveness(t *testing.T, ctx *build.Context, in, out [][]reg.Register) {
	t.Helper()
	fn := ConstructLiveness(t, ctx)
	is := fn.Instructions()

	if len(in) != len(is) || len(out) != len(is) {
		t.Fatalf("%d instructions: %d/%d in/out expectations", len(is), len(in), len(out))
	}

	for idx, i := range is {
		AssertRegistersMatchSet(t, in[idx], i.LiveIn)
		AssertRegistersMatchSet(t, out[idx], i.LiveOut)
	}
}

func AssertRegistersMatchSet(t *testing.T, rs []reg.Register, s reg.MaskSet) {
	t.Helper()
	if !s.Equals(reg.NewMaskSetFromRegisters(rs)) {
		t.Fatalf("register slice does not match set: %#v and %#v", rs, s)
	}
}

func ConstructLiveness(t *testing.T, ctx *build.Context) *ir.Function {
	t.Helper()
	return BuildFunction(t, ctx, pass.LabelTarget, pass.CFG, pass.Liveness)
}

func TestAllocateRegistersBasePointerDeprioritized(t *testing.T) {
	// Construct a function that requires n general-purpose registers all live
	// at once. Choose n to be the maximal possible number of registers without
	// touching the base pointer.
	n := 14

	ctx := build.NewContext()
	ctx.Function("sum")
	ctx.SignatureExpr("func() uint64")

	x := make([]reg.GPVirtual, n)
	for i := range n {
		x[i] = ctx.GP64()
		ctx.MOVQ(operand.U64(i), x[i])
	}

	for i := 1; i < n; i++ {
		ctx.ADDQ(x[i], x[0])
	}

	ctx.Store(x[0], ctx.ReturnIndex(0))
	ctx.RET()

	// Build and compile the function up to register allocation.
	fn := BuildFunction(t, ctx, pass.LabelTarget, pass.CFG, pass.Liveness, pass.AllocateRegisters, pass.BindRegisters)

	// Verify this function uses n registers, but not the base pointer.
	ps := map[reg.Physical]bool{}
	for _, i := range fn.Instructions() {
		for _, r := range i.OutputRegisters() {
			ps[reg.ToPhysical(r)] = true
		}
	}

	if len(ps) != n {
		t.Fatalf("expected function to require %d registers", n)
	}

	for p := range ps {
		if (p.Info() & reg.BasePointer) != 0 {
			t.Fatal("base pointer used")
		}
	}
}

func TestEnsureBasePointerCalleeSavedFrameless(t *testing.T) {
	// Construct a function that writes to the base pointer.
	ctx := build.NewContext()
	ctx.Function("clobber")
	ctx.ADDQ(reg.RAX, reg.RBP)

	// Build the function with the EnsureBasePointerCalleeSaved pass.
	fn := BuildFunction(t, ctx, pass.EnsureBasePointerCalleeSaved)

	// Since the function was frameless, expect that the pass would have
	expect := 8
	if fn.LocalSize != expect {
		t.Fatalf("expected frame size %d; got %d", expect, fn.LocalSize)
	}
}

func TestEnsureBasePointerCalleeSavedWithFrame(t *testing.T) {
	// Construct a function that writes to the base pointer, but already has a
	// stack frame.
	expect := 64

	ctx := build.NewContext()
	ctx.Function("clobber")
	ctx.AllocLocal(expect)
	ctx.ADDQ(reg.RAX, reg.RBP)

	// Build the function with the EnsureBasePointerCalleeSaved pass.
	fn := BuildFunction(t, ctx, pass.EnsureBasePointerCalleeSaved)

	// Expect that since the function already has a stack frame, there's no need to increase its size.
	if fn.LocalSize != expect {
		t.Fatalf("expected frame size %d; got %d", expect, fn.LocalSize)
	}
}

func TestEnsureBasePointerCalleeSavedNOFRAME(t *testing.T) {
	// Construct a NOFRAME function that writes to base pointer.
	ctx := build.NewContext()
	ctx.Function("clobber")
	ctx.Attributes(attr.NOFRAME)
	ctx.ADDQ(reg.RAX, reg.RBP)

	// Build the function.
	fn := BuildFunction(t, ctx)

	// Expect the pass to fail due to NOFRAME exception.
	if err := pass.EnsureBasePointerCalleeSaved(fn); err == nil {
		t.Fatal("expected error from NOFRAME function that clobbers base pointer")
	}
}

func TestIssue83RegisterAllocationStabilityAcrossRET(t *testing.T) {
	// Construct two functions: one with a dead jump preceding RET and one without.
	// Verifies that CFG liveness isolation across RET prevents register allocation churn.
	buildFn := func(withJump bool) *ir.Function {
		ctx := build.NewContext()
		ctx.Function("Issue83")
		ctx.SignatureExpr("func(cond, a, b uint64) uint64")
		cond := ctx.Load(ctx.Param("cond"), ctx.GP64())
		a := ctx.Load(ctx.Param("a"), ctx.GP64())
		b := ctx.Load(ctx.Param("b"), ctx.GP64())
		ctx.CMPQ(cond, operand.Imm(0))
		ctx.JE(operand.LabelRef("false"))
		ctx.Store(a, ctx.ReturnIndex(0))
		ctx.RET()
		if withJump {
			ctx.JMP(operand.LabelRef("false"))
		}
		ctx.Label("false")
		ctx.Store(b, ctx.ReturnIndex(0))
		ctx.RET()

		return BuildFunction(t, ctx,
			pass.PruneJumpToFollowingLabel,
			pass.PruneDanglingLabels,
			pass.LabelTarget,
			pass.CFG,
			pass.Liveness,
			pass.AllocateRegisters,
			pass.BindRegisters,
			pass.VerifyAllocation,
		)
	}

	fnWith := buildFn(true)
	fnWithout := buildFn(false)

	// Confirm no jump remains in fnWith after pruning.
	for _, i := range fnWith.Instructions() {
		if i.Opcode == "JMP" {
			t.Fatal("expected dead JMP instruction to be pruned")
		}
	}

	// Confirm RET instructions have empty successor sets.
	for _, fn := range []*ir.Function{fnWith, fnWithout} {
		for _, i := range fn.Instructions() {
			if i.IsTerminal && len(i.Succ) != 0 {
				t.Fatalf("terminal instruction %s has non-empty successor set", i.Opcode)
			}
		}
	}

	// Confirm register allocation stability: both allocations must match identically.
	if !reflect.DeepEqual(fnWith.Allocation, fnWithout.Allocation) {
		t.Fatalf("register allocation mismatch: with_jump=%v without_jump=%v", fnWith.Allocation, fnWithout.Allocation)
	}
}

