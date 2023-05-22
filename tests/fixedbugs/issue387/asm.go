//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	// Initialize array of floating point values.
	f32 := GLOBL("f32", RODATA|NOPTR)
	for i := 0; i < 10; i++ {
		DATA(4*i, F32(i))
	}

	TEXT("Float32At", NOSPLIT, "func(i int) float32")
	i := Load(Param("i"), GP64())
	ptr := Mem{Base: GP64()}
	LEAQ(f32, ptr.Base)
	x := XMM()
	MOVSS(ptr.Idx(i, 4), x)
	Store(x, ReturnIndex(0))
	RET()

	Generate()
}
