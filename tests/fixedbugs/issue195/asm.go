//go:build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	TEXT("Issue195", NOSPLIT|NOFRAME, "func(x *uint64, y uint32)")
	Doc("Issue195 tests for correct argument size.")
	RET()
	Generate()
}
