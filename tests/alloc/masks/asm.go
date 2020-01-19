// +build ignore

package main

import (
	"strconv"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

// The goal of this test is to create a synthetic scenario in which register
// allocation would fail if register liveness and allocation passes didn't take
// masks into account.
//
// The idea is to create a set of 15 64-bit virtual registers (15 being total
// number of allocatable 64-bit general purpose registers). For each one: write
// to the whole register and then later write to only the low 16 bits, and
// finally consume the whole 64-bit register. This means there is an interval in
// which only the high 48-bits are live. During this interval we should be able
// to allocate and use a set of 15 16-bit virtual registers.

func main() {
	const n = 15

	TEXT("Masks", NOSPLIT, "func() (uint16, uint64)")
	Doc("Masks computes the sum 1+2+...+" + strconv.Itoa(n) + " in two ways.")

	// Step 1: Allocate n 64-bit registers A that we will arrange to live in their top 48 bits.
	A := make([]GPVirtual, n)
	for i := 0; i < n; i++ {
		A[i] = GP64()
		c := ((i + 1) << 16) | 42 // 42 in low bits will be cleared later
		MOVQ(U32(c), A[i])
	}

	// Step 3: Allocate n 16-bit registers B.
	B := make([]Register, n)
	for i := 0; i < n; i++ {
		B[i] = GP16()
		MOVW(U16(i+1), B[i])
	}

	// Step 3: Sum up all the B registers and return.
	for i := 1; i < n; i++ {
		ADDW(B[i], B[0])
	}
	Store(B[0], ReturnIndex(0))

	// Step 4: Clear the low 16-bits of the A registers.
	for i := 0; i < n; i++ {
		MOVW(U16(0), A[i].As16())
	}

	// Step 5: Sum up all the A registers and return.
	for i := 1; i < n; i++ {
		ADDQ(A[i], A[0])
	}
	SHRQ(U8(16), A[0])
	Store(A[0], ReturnIndex(1))

	RET()

	Generate()
}
