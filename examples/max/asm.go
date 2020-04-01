// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("Max", NOSPLIT, "func(x, y []uint8) uint64")
	Doc("Max sets x to the pairwise max of x and y.")
	x := Load(Param("x").Base(), GP64())
	y := Load(Param("y").Base(), GP64())
	n := Load(Param("x").Len(), GP64())

	idx := GP64()
	XORQ(idx, idx)

	Label("loop")
	Comment("Loop until zero bytes remain.")
	CMPQ(idx, n)
	JE(LabelRef("done"))

	Comment("Load.")
	xptr := Mem{Base: x, Index: idx, Scale: 1}
	yptr := Mem{Base: y, Index: idx, Scale: 1}
	t := XMM()
	VMOVDQU(yptr, t)
	PMAXUB(xptr, t)
	VMOVDQU(t, xptr)

	Comment("Advance index.")
	ADDQ(Imm(16), idx)
	JMP(LabelRef("loop"))

	Label("done")
	RET()
	Generate()
}
