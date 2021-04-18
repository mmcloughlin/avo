package issue145

import (
	"testing"
)

//go:generate go run asm.go -out issue145.s -stubs stub.go

func TestShuffle(t *testing.T) {
	x := uint64(0xfeeddeadbeefcafe)
	got := Halves(x)
	expect := [2]uint32{0xbeefcafe, 0xfeeddead}
	if got != expect {
		t.Errorf("Halves(%x) = %x; expect %x", x, got, expect)
	}
}
