package issue83

import (
	"testing"
)

//go:generate go run asm.go -out issue83.s -stubs stubs.go

func TestIssue83(t *testing.T) {
	expectations := []bool{
		Issue83(0, 1, 2) == 2,
		Issue83(0, 42, 13) == 13,
		Issue83(1, 1, 2) == 1,
		Issue83(3, 42, 13) == 42,
		Issue83(30000, 4200, 1300000000) == 4200,
	}
	for _, expect := range expectations {
		if !expect {
			t.FailNow()
		}
	}
}
