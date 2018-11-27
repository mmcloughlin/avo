package operand

import (
	"testing"

	"github.com/mmcloughlin/avo/reg"
)

func TestMemAsm(t *testing.T) {
	cases := []struct {
		Mem    Mem
		Expect string
	}{
		{Mem{Base: reg.EAX}, "(AX)"},
		{Mem{Disp: 16, Base: reg.RAX}, "16(AX)"},
		{Mem{Base: reg.R11, Index: reg.RAX, Scale: 4}, "(R11)(AX*4)"},
		{Mem{Base: reg.R11, Index: reg.RAX, Scale: 1}, "(R11)(AX*1)"},
		{Mem{Base: reg.R11, Index: reg.RAX}, "(R11)"},
		{Mem{Base: reg.R11, Scale: 8}, "(R11)"},
		{Mem{Disp: 2048, Base: reg.R11, Index: reg.RAX, Scale: 8}, "2048(R11)(AX*8)"},
	}
	for _, c := range cases {
		got := c.Mem.Asm()
		if got != c.Expect {
			t.Errorf("%#v.Asm() = %s expected %s", c.Mem, got, c.Expect)
		}
	}
}
