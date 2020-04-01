package max

import (
	"math/rand"
	"strconv"
	"testing"
)

//go:generate go run asm.go -out max.s -stubs stub.go

func TestMax(t *testing.T) {
	n := 64 * 43
	x, y := RandomArray(n), RandomArray(n)

	got := Copy(x)
	Max(got, y)

	expect := Copy(x)
	Expect(expect, y)

	AssertEqualArray(t, expect, got)
}

func BenchmarkAsm600K(b *testing.B) {
	benchmarkSize(b, 600000, Max)
}

func BenchmarkNaive600K(b *testing.B) {
	benchmarkSize(b, 600000, Expect)
}

func BenchmarkAsmSizes(b *testing.B) {
	benchmarkSizes(b, Max)
}

func BenchmarkNaiveSizes(b *testing.B) {
	benchmarkSizes(b, Expect)
}

func benchmarkSizes(b *testing.B, mx func(x, y []uint8)) {
	for n := 6; n <= 26; n += 4 {
		size := 1 << uint(n)
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			benchmarkSize(b, size, mx)
		})
	}
}

func benchmarkSize(b *testing.B, size int, mx func(x, y []uint8)) {
	x, y := RandomArray(size), RandomArray(size)
	for i := 0; i < b.N; i++ {
		mx(x, y)
	}
	b.SetBytes(2 * int64(size))
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
