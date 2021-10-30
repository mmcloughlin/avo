package issue195

import "testing"

//go:generate go run asm.go -out issue195.s -stubs stub.go

func TestIssue195(t *testing.T) {
	x := uint64(42)
	Issue195(&x, 42)
}
