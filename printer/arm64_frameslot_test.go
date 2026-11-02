package printer

import (
	"testing"

	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

// TestFrameSlotBase checks that only AllocLocal's pseudo SP makes a frame slot.
// The pseudo registers share one ID, so a check by ID would take a
// symbol-less FP, SB or PC operand for a slot too.
func TestFrameSlotBase(t *testing.T) {
	cases := []struct {
		base reg.Register
		want bool
	}{
		{reg.StackPointer, true},
		{reg.FramePointer, false},
		{reg.StaticBase, false},
		{reg.ProgramCounter, false},
		{reg.RSP, false},
		{reg.RAX, false},
	}
	for _, c := range cases {
		d, ok := frameSlot(operand.Mem{Base: c.base, Disp: 8})
		if ok != c.want || (ok && d != 8) {
			t.Errorf("frameSlot(8(%s)) = %d, %v; want slot=%v", c.base.Asm(), d, ok, c.want)
		}
	}
}
