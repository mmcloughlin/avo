package pass_test

import (
	"reflect"
	"testing"

	"github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/pass"
	"github.com/mmcloughlin/avo/reg"
)

func TestPruneSelfMoves(t *testing.T) {
	// Construct a function containing a self-move.
	ctx := build.NewContext()
	ctx.Function("add")
	ctx.MOVQ(operand.U64(1), reg.RAX)
	ctx.MOVQ(operand.U64(2), reg.RCX)
	ctx.MOVQ(reg.RAX, reg.RAX) // self move
	ctx.MOVQ(reg.RCX, reg.R8)
	ctx.ADDQ(reg.R8, reg.RAX)

	// Build the function without the pass and save the nodes.
	fn := BuildFunction(t, ctx)
	pre := append([]ir.Node{}, fn.Nodes...)

	// Apply the pass.
	pass.PruneSelfMoves(fn)

	// Confirm the self-move was removed and everything else was untouched.
	expect := []ir.Node{}
	for i, n := range pre {
		if i != 2 {
			expect = append(expect, n)
		}
	}

	if !reflect.DeepEqual(fn.Nodes, expect) {
		t.Fatal("unexpected result from self-move pruning")
	}
}
