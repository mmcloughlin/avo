// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	Package("github.com/mmcloughlin/avo/examples/components")

	// Add confirms that we correctly deduce the packing of Struct.
	TEXT("Add", "func(x uint64, s Struct, y uint64) uint64")
	x := Load(Param("x"), GP64v())
	y := Load(Param("y"), GP64v())
	ADDQ(x, y)
	Store(y, ReturnIndex(0))
	RET()

	Generate()
}
