package sum

import (
	"testing"
	"testing/quick"
)

//go:generate go run asm.go -out sum.s -stubs stub.go

func expect(xs []uint64) uint64 {
	var s uint64
	for _, x := range xs {
		s += x
	}
	return s
}

func TestSum(t *testing.T) {
	quick.CheckEqual(Sum, expect, nil)
}
