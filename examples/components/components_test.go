package components

import (
	"testing"
	"testing/quick"
)

//go:generate go run asm.go -out components.s -stubs stub.go

func TestAdd(t *testing.T) {
	expect := func(x uint64, s Struct, y uint64) uint64 { return x + y }
	if err := quick.CheckEqual(Add, expect, nil); err != nil {
		t.Fatal(err)
	}
}
