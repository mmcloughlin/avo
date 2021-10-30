package issue191

import "testing"

//go:generate go run asm.go -out issue191.s -stubs stub.go

func TestUint16(t *testing.T) {
	Uint16(42)
}
