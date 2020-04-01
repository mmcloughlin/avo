package max

import (
	"math/rand"
	"testing"
)

//go:generate go run asm.go -out max.s -stubs stub.go

func TestMax(t *testing.T) {
	n := 16 * 43
	x, y := RandomArray(n), RandomArray(n)

	got := Copy(x)
	Max(got, y)

	expect := Copy(x)
	Expect(expect, y)

	AssertEqualArray(t, expect, got)
}

func Expect(x, y []uint8) {
	for i := range x {
		if y[i] > x[i] {
			x[i] = y[i]
		}
	}
}

func RandomArray(n int) []uint8 {
	x := make([]uint8, n)
	for i := 0; i < n; i++ {
		x[i] = uint8(rand.Intn(256))
	}
	return x
}

func Copy(x []uint8) []uint8 {
	y := make([]uint8, len(x))
	copy(y, x)
	return y
}

func AssertEqualArray(t *testing.T, expect, got []uint8) {
	if len(expect) != len(got) {
		t.Fatalf("length mismatch")
	}

	for i := range expect {
		if expect[i] != got[i] {
			t.Fatalf("index %d: got %x expect %x", i, got[i], expect[i])
		}
	}
}
