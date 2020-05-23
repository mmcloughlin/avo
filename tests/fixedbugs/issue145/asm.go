// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	TEXT("Halves", NOSPLIT, "func(x uint64) [2]uint32")
	Doc("Halves returns the two 32-bit halves of a 64-bit word.")
	x := Load(Param("x"), GP64())
	MOVQ(x, ReturnIndex(0).MustAddr())
	RET()
	Generate()
}
