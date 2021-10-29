//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

var unroll = 6

func main() {
	TEXT("Dot", NOSPLIT, "func(x, y []float32) float32")
	x := Mem{Base: Load(Param("x").Base(), GP64())}
	y := Mem{Base: Load(Param("y").Base(), GP64())}
	n := Load(Param("x").Len(), GP64())

	// Allocate accumulation registers.
	acc := make([]VecVirtual, unroll)
	for i := 0; i < unroll; i++ {
		acc[i] = YMM()
	}

	// Zero initialization.
	for i := 0; i < unroll; i++ {
		VXORPS(acc[i], acc[i], acc[i])
	}

	// Loop over blocks and process them with vector instructions.
	blockitems := 8 * unroll
	blocksize := 4 * blockitems
	Label("blockloop")
	CMPQ(n, U32(blockitems))
	JL(LabelRef("tail"))

	// Load x.
	xs := make([]VecVirtual, unroll)
	for i := 0; i < unroll; i++ {
		xs[i] = YMM()
	}

	for i := 0; i < unroll; i++ {
		VMOVUPS(x.Offset(32*i), xs[i])
	}

	// The actual FMA.
	for i := 0; i < unroll; i++ {
		VFMADD231PS(y.Offset(32*i), xs[i], acc[i])
	}

	ADDQ(U32(blocksize), x.Base)
	ADDQ(U32(blocksize), y.Base)
	SUBQ(U32(blockitems), n)
	JMP(LabelRef("blockloop"))

	// Process any trailing entries.
	Label("tail")
	tail := XMM()
	VXORPS(tail, tail, tail)

	Label("tailloop")
	CMPQ(n, U32(0))
	JE(LabelRef("reduce"))

	xt := XMM()
	VMOVSS(x, xt)
	VFMADD231SS(y, xt, tail)

	ADDQ(U32(4), x.Base)
	ADDQ(U32(4), y.Base)
	DECQ(n)
	JMP(LabelRef("tailloop"))

	// Reduce the lanes to one.
	Label("reduce")
	for i := 1; i < unroll; i++ {
		VADDPS(acc[0], acc[i], acc[0])
	}

	result := acc[0].AsX()
	top := XMM()
	VEXTRACTF128(U8(1), acc[0], top)
	VADDPS(result, top, result)
	VADDPS(result, tail, result)
	VHADDPS(result, result, result)
	VHADDPS(result, result, result)
	Store(result, ReturnIndex(0))

	RET()

	Generate()
}
