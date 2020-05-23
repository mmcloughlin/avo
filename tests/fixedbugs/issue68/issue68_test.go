package custom

import (
	"testing"
)

//go:generate go run asm.go -pkg custom -out issue69.s -stubs stub.go

func TestIssue68(t *testing.T) {
	if got := Issue68(); got != 68 {
		t.Fail()
	}
}
