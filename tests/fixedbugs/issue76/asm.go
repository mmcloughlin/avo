// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/reg"
)

func main() {
	TEXT("Issue76", NOSPLIT, "func(x, y uint64) uint64")
	x := Load(Param("x"), GP64())
	y := Load(Param("y"), GP64())
	s := add(x, y)
	Store(s, ReturnIndex(0))
	RET()
	Generate()
}

func add(x, y Register) Register {
	s := GP64()
	MOVQ(x, s)
	ADDQ(y, s)
	return s
}
