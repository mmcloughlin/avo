package issue89

import (
	"testing"
)

//go:generate go run asm.go -out issue89.s -stubs stub.go

func TestIssue89(t *testing.T) {
	if Issue89() != 42 {
		t.FailNow()
	}
}
