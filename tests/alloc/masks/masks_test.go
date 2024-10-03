package masks

import (
	"testing"
)

//go:generate go run asm.go -out masks.s -stubs stub.go

func TestMasks(t *testing.T) {
	const n = 15
	const expect = n * (n + 1) / 2
	if got16, got64 := Masks(); got16 != expect || got64 != expect {
		t.Fatalf("Masks() = %v, %v; expect %v, %v", got16, got64, expect, expect)
	}
}
