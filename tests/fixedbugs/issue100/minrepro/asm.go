//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("Issue100", NOSPLIT, "func() uint64")
	x := GP64()
	XORQ(x, x)
	for i := 1; i <= 100; i++ {
		t := GP64()
		MOVQ(U32(i), t)
		ADDQ(t.As64(), x)
	}
	Store(x, ReturnIndex(0))
	RET()
	Generate()
}
