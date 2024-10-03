//go:build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("Add", NOSPLIT, "func(z, x, y *uint64)")
	Pragma("noescape")
	Doc("Add adds the values at x and y and writes the result to z.")
	zptr := Mem{Base: Load(Param("z"), GP64())}
	xptr := Mem{Base: Load(Param("x"), GP64())}
	yptr := Mem{Base: Load(Param("y"), GP64())}
	x, y := GP64(), GP64()
	MOVQ(xptr, x)
	MOVQ(yptr, y)
	ADDQ(x, y)
	MOVQ(y, zptr)
	RET()
	Generate()
}
