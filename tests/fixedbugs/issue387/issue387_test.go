package issue387

import (
	"testing"
)

//go:generate go run asm.go -out issue387.s -stubs stub.go

func TestFloat32At(t *testing.T) {
	for i := 0; i < 10; i++ {
		got := Float32At(i)
		expect := float32(i)
		if got != expect {
			t.Fatalf("Float32At(%d) = %#v; expect %#v", i, got, expect)
		}
	}
}
