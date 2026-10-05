//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("Sum", NOSPLIT, "func(n uint64) uint64")
	Doc("Sum returns the sum of integers in [0, n), using a loop aligned with PCALIGN.")
	// Alignment of 1024 requires the imm16 form, and also raises the alignment
	// of the function symbol itself.
	PCALIGN(Imm(1024))
	n := Load(Param("n"), GP64())
	s := GP64()
	XORQ(s, s)
	TESTQ(n, n)
	JZ(LabelRef("done"))

	// Alignment of 32 uses the imm8 form.
	PCALIGN(Imm(32))
	Label("loop")
	DECQ(n)
	ADDQ(n, s)
	TESTQ(n, n)
	JNZ(LabelRef("loop"))

	Label("done")
	Store(s, ReturnIndex(0))
	RET()

	TEXT("SumAddr", NOSPLIT, "func() uintptr")
	Doc("SumAddr returns the address of the assembly implementation of Sum.")
	addr := GP64()
	LEAQ(NewDataAddr(Symbol{Name: "·Sum"}, 0), addr)
	Store(addr, ReturnIndex(0))
	RET()

	Generate()
}
