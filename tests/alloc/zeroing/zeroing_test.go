package zeroing

import (
	"testing"

	"golang.org/x/sys/cpu"
)

//go:generate go run asm.go -out zeroing.s -stubs stub.go

func TestZeroing(t *testing.T) {
	const (
		n      = 32
		expect = n * (n + 1) / 2
	)

	if !cpu.X86.HasAVX512F {
		t.Skip("require AVX512F")
	}

	var got [8]uint64
	Zeroing(&got)

	for i := 0; i < 8; i++ {
		if got[i] != expect {
			t.Errorf("got[%d] = %d; expect %d", i, got[i], expect)
		}
	}
}
