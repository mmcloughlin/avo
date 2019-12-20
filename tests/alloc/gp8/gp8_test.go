package gp8

import (
	"testing"
)

//go:generate go run asm.go -out gp8.s -stubs stub.go

func TestGP8(t *testing.T) {
	const n = 19
	expect := uint8(n * (n + 1) / 2)
	if got := GP8(); got != expect {
		t.Fatalf("GP8() = %d; expect %d", got, expect)
	}
}
