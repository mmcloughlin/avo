// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	TEXT("Real", "func(z complex128) float64")
	Doc("Real returns the real part of z.")
	r := Load(Param("z").Real(), Xv())
	Store(r, ReturnIndex(0))
	RET()

	TEXT("Imag", "func(z complex128) float64")
	Doc("Imag returns the imaginary part of z.")
	i := Load(Param("z").Imag(), Xv())
	Store(i, ReturnIndex(0))
	RET()

	TEXT("Norm", "func(z complex128) float64")
	Doc("Norm returns the complex norm of z.")
	r = Load(Param("z").Real(), Xv())
	i = Load(Param("z").Imag(), Xv())
	MULSD(r, r)
	MULSD(i, i)
	ADDSD(i, r)
	n := Xv()
	SQRTSD(r, n)
	Store(n, ReturnIndex(0))
	RET()

	Generate()
}
