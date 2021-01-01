package zeroing

import (
	"testing"

	"golang.org/x/sys/cpu"
)

//go:generate go run asm.go -out zeroing.s -stubs stub.go

func TestZeroing(t *testing.T) {
	if !cpu.X86.HasAVX512F {
		t.Skip("require AVX512F")
	}

	var got [8]uint64
	Zeroing(&got)
}
