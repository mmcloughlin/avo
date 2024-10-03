package issue122

import (
	"testing"
)

//go:generate go run asm.go -out issue122.s -stubs stub.go

func TestTriangle(t *testing.T) {
	expect := func(n uint64) uint64 { return n * (n + 1) / 2 }
	for n := uint64(1); n < 42; n++ {
		if got := Triangle(n); expect(n) != got {
			t.Fatalf("Triangle(%v) = %v; expect %v", n, got, expect(n))
		}
	}
}
