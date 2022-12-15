package issue193

import (
	"testing"

	"golang.org/x/sys/cpu"
)

//go:generate go run asm.go -out issue193.s -stubs stub.go

func TestAddSubs(t *testing.T) {
	if !(cpu.X86.HasAVX2 && cpu.X86.HasAVX512F && cpu.X86.HasAVX512DQ && cpu.X86.HasAVX512VL) {
		t.Skipf("Skipping issue193 test because runtime lacks needed CPUID features")
	}

	in := [8]int64{7, 6, 5, 4, 3, 2, 1, 0}
	out := [8]int64{13, 1, 9, 1, 5, 1, 1, 1}

	val := in
	AddSubPairs(&val)
	if val != out {
		t.Errorf("Bad result %v", val)
	}

	val = in
	AddSubPairsNoBase(&val)
	if val != out {
		t.Errorf("Bad result %v", val)
	}
}
