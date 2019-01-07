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

	// Initialize sum register to zero.
	s := GP64()
	XORQ(s, s)

	// Loop until zero bytes remain.
	Label("loop")
	CMPQ(n, Imm(0))
	JE(LabelRef("done"))

	// Load from pointer and add to running sum.
	ADDQ(Mem{Base: ptr}, s)

	// Advance pointer, decrement byte count.
	ADDQ(Imm(8), ptr)
	DECQ(n)
	JMP(LabelRef("loop"))

	// Store sum to return value.
	Label("done")
	Store(s, ReturnIndex(0))
	RET()
	Generate()
}
