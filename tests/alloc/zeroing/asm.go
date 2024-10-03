//go:build ignore

package main

import (
	"strconv"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

// The goal of this test is to confirm correct liveness analysis of zeroing mode
// when masking in AVX-512. In merge masking, some of the bits of the output
// register will be preserved, so the register is live coming into the
// instruction. Zeroing mode removes any input dependency.
//
// This synthetic test sets up a situation where we allocate multiple temporary
// registers. Allocation is only feasible if the liveness pass correctly
// identifies that they are not all live at once.

func main() {
	const n = 32

	TEXT("Zeroing", NOSPLIT, "func(out *[8]uint64)")
	Doc("Zeroing computes the sum 1+2+...+" + strconv.Itoa(n) + " in 8 lanes of 512-bit register.")

	out := Load(Param("out"), GP64())

	Comment("Initialize sum.")
	s := ZMM()
	VPXORD(s, s, s)

	// Allocate registers for the terms of the sum. Write garbage to them.
	//
	// The point here is that under merge-masking, or an incorrect handling of
	// zeroing-masking, these registers would be live from this point. And there
	// would be too many of them so register allocation would fail.
	Comment("Initialize summand registers.")
	filler := GP64()
	MOVQ(U64(0x9e77d78aacb8cbcc), filler)

	z := make([]VecVirtual, n)
	for i := 0; i < n; i++ {
		z[i] = ZMM()
		VPBROADCASTQ(filler, z[i])
	}

	// Prepare a mask register set to all ones.
	Comment("Prepare mask register.")
	k := K()
	KXNORW(k, k, k)

	// Prepare an increment register set to 1 in each lane.
	Comment("Prepare constant registers.")
	one := GP64()
	MOVQ(U64(1), one)
	ones := ZMM()
	VPBROADCASTQ(one, ones)

	zero := ZMM()
	VPXORD(zero, zero, zero)

	last := zero
	for i := 0; i < n; i++ {
		Commentf("Summand %d.", i+1)
		VPADDD_Z(last, ones, k, z[i])
		VPADDD(s, z[i], s)
		last = z[i]
	}

	Comment("Write result to output pointer.")
	VMOVDQU64(s, Mem{Base: out})

	RET()

	Generate()
}
