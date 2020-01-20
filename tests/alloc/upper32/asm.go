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
		n = 14 // number of registers
		m = 3  // number of iterations
	)

	TEXT("Upper32", NOSPLIT, "func() uint64")
	Doc("Upper32 computes the sum 1+2+...+" + strconv.Itoa(n*m) + ".")

	Comment("Initialize sum.")
	s := GP64()
	XORQ(s, s)

	k := 1
	for i := 0; i < m; i++ {
		Commentf("Iteration %d.", i+1)

		// Allocate n 64-bit registers and write to their 32-bit aliases.
		x := make([]GPVirtual, n)
		for i := 0; i < n; i++ {
			x[i] = GP64()
			MOVL(U32(k), x[i].As32())
			k++
		}

		// Sum them up.
		for i := 0; i < n; i++ {
			ADDQ(x[i], s)
		}
	}

	Store(s, ReturnIndex(0))
	RET()

	Generate()
}
