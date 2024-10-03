package upper32

import (
	"testing"
)

//go:generate go run asm.go -out upper32.s -stubs stub.go

func TestUpper32(t *testing.T) {
	const n = 14 * 3
	const expect = n * (n + 1) / 2
	if got := Upper32(); got != expect {
		t.Fatalf("Upper32() = %v; expect %v", got, expect)
	}
}
