package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	TEXT("add", "func(x, y uint64) uint64")
	x := Load(Param("x"), GP64v())
	y := Load(Param("y"), GP64v())
	ADDQ(x, y)
	Store(x, ReturnIndex(0))
	RET()
	EOF()
}
