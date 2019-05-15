// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("Issue83", NOSPLIT, "func(cond, a, b uint64) uint64")
	cond := Load(Param("cond"), GP64())
	a := Load(Param("a"), GP64())
	b := Load(Param("b"), GP64())
	CMPQ(cond, Imm(0))
	JE(LabelRef("false"))
	Store(a, ReturnIndex(0))
	RET()
	JMP(LabelRef("false")) // pointless jump
	ADDQ(cond, cond)       // useless instruction to prevent JMP from being removed
	Label("false")
	Store(b, ReturnIndex(0))
	RET()
	Generate()
}
