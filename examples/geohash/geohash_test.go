package geohash

import (
	"testing"

	"golang.org/x/sys/cpu"
)

//go:generate go run asm.go -out geohash.s -stubs stub.go

func TestEncodeIntMountEverest(t *testing.T) {
	if !(cpu.X86.HasSSE2 && cpu.X86.HasBMI2) {
		t.Skip("requires SSE2 and BMI2 instruction sets")
	}
	if EncodeInt(27.988056, 86.925278) != 0xceb7f254240fd612 {
		t.Fail()
	}
}
