package complex

import (
	"math"
	"testing"
	"testing/quick"
)

//go:generate go run asm.go -out complex.s -stubs stub.go

func TestReal(t *testing.T) {
	expect := func(x complex128) float64 {
		return real(x)
	}
	if err := quick.CheckEqual(Real, expect, nil); err != nil {
		t.Fatal(err)
	}
}

func TestImag(t *testing.T) {
	expect := func(x complex128) float64 {
		return imag(x)
	}
	if err := quick.CheckEqual(Imag, expect, nil); err != nil {
		t.Fatal(err)
	}
}

func TestNorm(t *testing.T) {
	expect := func(x complex128) float64 {
		return math.Sqrt(real(x)*real(x) + imag(x)*imag(x))
	}
	if err := quick.CheckEqual(Norm, expect, nil); err != nil {
		t.Fatal(err)
	}
}
