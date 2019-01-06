// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("Sum", "func(xs []uint64) uint64")
	Doc("Sum returns the sum of the elements in xs.")
	ptr := Load(Param("xs").Base(), GP64())
	n := Load(Param("xs").Len(), GP64())
	s := GP64()
	XORQ(s, s)
	LABEL("loop")
	CMPQ(n, Imm(0))
	JE(LabelRef("done"))
	ADDQ(Mem{Base: ptr}, s)
	ADDQ(Imm(8), ptr)
	DECQ(n)
	JMP(LabelRef("loop"))
	LABEL("done")
	Store(s, ReturnIndex(0))
	RET()
	Generate()
}
