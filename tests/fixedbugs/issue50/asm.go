// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	TEXT("Issue50", NOSPLIT, "func(x uint32) uint32")
	Doc(
		"Issue50 reported that MOVD/MOVQ was missing the r32, xmm form.",
		"This function deliberately exercises this instruction form.",
	)
	x := Load(Param("x"), GP32())
	xmm := XMM()
	MOVQ(x, xmm)
	Store(xmm, ReturnIndex(0))
	RET()
	Generate()
}
