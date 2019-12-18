package issue100

import (
	"testing"
)

//go:generate go run asm.go -out issue100.s -stubs stub.go

func TestIssue100(t *testing.T) {
	n := uint64(100)
	expect := n * (n + 1) / 2
	if got := Issue100(); got != expect {
		t.Fatalf("Issue100() = %v; expect %v", got, expect)
	}
}
