//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("Triangle", NOSPLIT, "func(n uint64) uint64")
	Doc("Triangle computes the nth triangle number.")
	n := Load(Param("n"), GP64())

	Comment("Initialize sum register to zero.")
	s := GP64()
	XORQ(s, s)

	// Use two labels for the top of the loop.
	Label("loop_even")
	Label("loop_odd")
	Comment("Loop until n is zero.")
	CMPQ(n, Imm(0))
	JE(LabelRef("done"))

	Comment("Add n to sum.")
	ADDQ(n, s)

	Comment("Decrement n.")
	DECQ(n)

	Comment("Jump to one of the loop labels depending on parity.")
	TESTQ(U32(1), n)
	JZ(LabelRef("loop_even"))
	JMP(LabelRef("loop_odd"))

	Label("done")
	Comment("Store sum to return value.")
	Store(s, ReturnIndex(0))
	RET()
	Generate()
}
