package issue336

import "testing"

//go:generate go run asm.go -out issue336.s -stubs stub.go

func TestNot(t *testing.T) {
	for _, x := range []bool{true, false} {
		expect := !x
		if got := Not(x); got != expect {
			t.Errorf("Not(%v) = %v, expect %v", x, got, expect)
		}
	}
}
