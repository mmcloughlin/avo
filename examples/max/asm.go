// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

const unroll = 4

func main() {
	TEXT("Max", NOSPLIT, "func(x, y []uint8)")
	Doc("Max sets x to the pairwise max of x and y.")
	x := Load(Param("x").Base(), GP64())
	y := Load(Param("y").Base(), GP64())
	n := Load(Param("x").Len(), GP64())

	idx := GP64()
	XORQ(idx, idx)

	// Temporaries.
	t := make([]Register, unroll)
	for i := 0; i < unroll; i++ {
		t[i] = XMM()
	}

	// Loop Header.
	Label("loop")
	Comment("Loop until zero bytes remain.")
	CMPQ(idx, n)
	JE(LabelRef("done"))

	Comment("Load.")
	for i := 0; i < unroll; i++ {
		MOVUPD(Mem{Base: y, Index: idx, Scale: 1, Disp: 16 * i}, t[i])
	}

	for i := 0; i < unroll; i++ {
		PMAXUB(Mem{Base: x, Index: idx, Scale: 1, Disp: 16 * i}, t[i])
	}

	for i := 0; i < unroll; i++ {
		MOVUPD(t[i], Mem{Base: x, Index: idx, Scale: 1, Disp: 16 * i})
	}

	Comment("Advance index.")
	ADDQ(Imm(16*unroll), idx)
	JMP(LabelRef("loop"))

	Label("done")
	RET()
	Generate()
}
