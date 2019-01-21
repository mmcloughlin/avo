// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	Package("github.com/mmcloughlin/avo/examples/ext")
	Implement("ExtStructFieldB")
	e := Load(Param("e").Field("B"), GP8())
	Store(e, ReturnIndex(0))
	RET()
	Generate()
}
