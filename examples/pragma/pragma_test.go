package pragma

import (
	"testing"
	"testing/quick"
)

//go:generate go run asm.go -out pragma.s -stubs stub.go

func TestAdd(t *testing.T) {
	got := func(x, y uint64) (z uint64) { Add(&z, &x, &y); return }
	expect := func(x, y uint64) uint64 { return x + y }
	if err := quick.CheckEqual(got, expect, nil); err != nil {
		t.Fatal(err)
	}
}
