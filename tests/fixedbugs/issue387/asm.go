//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

// Float32 generates a function which indexes into an array of single-precision
// integer float values.
func Float32() {
	f32 := GLOBL("f32", RODATA|NOPTR)
	for i := 0; i < 10; i++ {
		DATA(4*i, F32(i))
	}

	TEXT("Float32", NOSPLIT, "func(i int) float32")
	Doc("Float32 indexes into an array of single-precision integral floats.")
	i := Load(Param("i"), GP64())
	ptr := Mem{Base: GP64()}
	LEAQ(f32, ptr.Base)
	x := XMM()
	MOVSS(ptr.Idx(i, 4), x)
	Store(x, ReturnIndex(0))
	RET()
}

// Float64 generates a function which indexes into an array of double-precision
// integer float values.
func Float64() {
	f64 := GLOBL("f64", RODATA|NOPTR)
	for i := 0; i < 10; i++ {
		DATA(8*i, F64(i))
	}

	TEXT("Float64", NOSPLIT, "func(i int) float64")
	Doc("Float64 indexes into an array of double-precision integral floats.")
	i := Load(Param("i"), GP64())
	ptr := Mem{Base: GP64()}
	LEAQ(f64, ptr.Base)
	x := XMM()
	MOVSD(ptr.Idx(i, 8), x)
	Store(x, ReturnIndex(0))
	RET()
}

func main() {
	Float32()
	Float64()
	Generate()
}
