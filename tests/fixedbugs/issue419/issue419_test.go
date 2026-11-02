package issue419

import (
	"testing"
)

//go:generate go run asm.go -out issue419.s -stubs stub.go

func TestSum(t *testing.T) {
	cases := []struct {
		n      uint64
		expect uint64
	}{
		{n: 0, expect: 0},
		{n: 1, expect: 0},
		{n: 2, expect: 1},
		{n: 10, expect: 45},
		{n: 1000, expect: 499500},
	}
	for _, c := range cases {
		if got := Sum(c.n); got != c.expect {
			t.Errorf("Sum(%d) = %d; expect %d", c.n, got, c.expect)
		}
	}
}

func TestSumAlignment(t *testing.T) {
	// PCALIGN $1024 raises the alignment of the enclosing function symbol. The
	// default function alignment on amd64 is 32, so this would fail if the
	// directive were dropped.
	if addr := SumAddr(); addr%1024 != 0 {
		t.Fatalf("Sum is at %#x; expect 1024-byte alignment", addr)
	}
}
