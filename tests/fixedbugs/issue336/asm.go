//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	TEXT("Not", NOSPLIT, "func(x bool) bool")
	Doc("Not returns the boolean negation of x.")
	// x := Load(Param("x"), GP8())
	// XORB(U8(1), x)
	// Store(x, ReturnIndex(0))
	RET()
	Generate()
}
