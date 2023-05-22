package issue387

import (
	"testing"
)

//go:generate go run asm.go -out issue387.s -stubs stub.go

func TestFloat32(t *testing.T) {
	for i := 0; i < 10; i++ {
		got := Float32(i)
		expect := float32(i)
		if got != expect {
			t.Fatalf("Float32(%d) = %#v; expect %#v", i, got, expect)
		}
	}
}

func TestFloat64(t *testing.T) {
	for i := 0; i < 10; i++ {
		got := Float64(i)
		expect := float64(i)
		if got != expect {
			t.Fatalf("Float64(%d) = %#v; expect %#v", i, got, expect)
		}
	}
}
