// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	TEXT("Real", "func(x complex128) float64")
	r := Load(Param("x").Real(), Xv())
	Store(r, ReturnIndex(0))
	RET()

	TEXT("Imag", "func(x complex128) float64")
	i := Load(Param("x").Imag(), Xv())
	Store(i, ReturnIndex(0))
	RET()

	TEXT("Norm", "func(x complex128) float64")
	r = Load(Param("x").Real(), Xv())
	i = Load(Param("x").Imag(), Xv())
	MULSD(r, r)
	MULSD(i, i)
	ADDSD(i, r)
	n := Xv()
	SQRTSD(r, n)
	Store(n, ReturnIndex(0))
	RET()

	Generate()
}
