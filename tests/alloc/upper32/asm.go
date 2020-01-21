// +build ignore

package main

import (
	"strconv"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

// The goal is to test for correct handling of 32-bit operands in 64-bit mode,
// specifically that writes are zero-extended to 64 bits. This test is
// constructed such that the register allocator would fail if this feature is
// not accounted for. It consists of multiple copies of a 32-bit write followed
// by a 64-bit read of the same register. Without special treatment liveness
// analysis would consider the upper 32 bits to still be live prior to the
// write. Therefore if we stack up enough copies of this, we could cause the
// register allocator to fail.

func main() {
	const (
		r = 14 // number of registers
		m = 3  // number of iterations
		n = r * m
	)

	TEXT("Upper32", NOSPLIT, "func() uint64")
	Doc("Upper32 computes the sum 1+2+...+" + strconv.Itoa(n) + ".")

	Comment("Initialize sum.")
	s := GP64()
	XORQ(s, s)

	// Allocate n 64-bit registers and populate them.
	Comment("Initialize registers.")
	x := make([]GPVirtual, n)
	for i := 0; i < n; i++ {
		x[i] = GP64()
		MOVQ(U64(0x9e77d78aacb8cbcc), x[i])
	}

	k := 0
	for i := 0; i < m; i++ {
		Commentf("Iteration %d.", i+1)

		// Write to the 32-bit aliases of r registers.
		for j := 0; j < r; j++ {
			MOVL(U32(k+j+1), x[k+j].As32())
		}

		// Sum them up.
		for j := 0; j < r; j++ {
			ADDQ(x[k+j], s)
		}

		k += r
	}

	Comment("Store result and return.")
	Store(s, ReturnIndex(0))
	RET()

	Generate()
}
