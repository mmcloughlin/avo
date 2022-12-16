package issue336

import "testing"

//go:generate go run asm.go -out issue336.s -stubs stub.go

func TestNot(t *testing.T) {
	nots := []func(bool) bool{Not8, Not16, Not32, Not64}
	for _, not := range nots {
		for _, x := range []bool{true, false} {
			if not(x) != !x {
				t.Fail()
			}
		}
	}
}
