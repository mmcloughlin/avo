//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	TEXT("Uint16", 0, "func(n uint16)")
	Doc("Uint16 tests argument size without return types.")
	RET()
	Generate()
}
