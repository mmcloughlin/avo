//go:build ignore
// +build ignore

package main

import (
	"strconv"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func main() {
	// n is the number of 8-bit registers to use.
	// 15 low-byte registers (excluding SP)
	// 4 high-byte registers AH,BH,CH,DH
	const n = 19

	TEXT("GP8", NOSPLIT, "func() uint8")
	Doc("GP8 returns the sum 1+2+...+" + strconv.Itoa(n) + " using " + strconv.Itoa(n) + " distinct 8-bit registers.")

	// Allocate registers and initialize.
	x := make([]Register, n)
	i := 0

	// Low byte registers.
	for ; i < 15; i++ {
		x[i] = GP8L()
		MOVB(U8(i+1), x[i])
	}

	// High byte registers.
	for ; i < n; i++ {
		x[i] = GP8H()
		MOVB(U8(i+1), x[i])
	}

	// Sum them up.
	for i := 1; i < n; i++ {
		ADDB(x[i], x[0])
	}

	// Return.
	Store(x[0], ReturnIndex(0))
	RET()

	Generate()
}
