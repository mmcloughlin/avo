package pass

import (
	"strings"
	"testing"

	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

func TestVerifyMemOperands(t *testing.T) {
	i := &ir.Instruction{
		Operands: []operand.Op{
			reg.RAX,
			operand.Mem{
				Base: reg.R10,
				Disp: 42,
			},
		},
	}
	if err := VerifyMemOperands(i); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyMemOperandsErrors(t *testing.T) {
	cases := []struct {
		Operands       []operand.Op
		ErrorSubstring string
	}{
		{
			Operands: []operand.Op{
				reg.RAX,
				operand.Mem{
					Disp: 42,
				},
			},
			ErrorSubstring: "missing base",
		},
		{
			Operands: []operand.Op{
				operand.Mem{
					Base:  reg.EBX,
					Index: reg.R9L,
				},
				reg.ECX,
			},
			ErrorSubstring: "index register with scale 0",
		},
	}
	for _, c := range cases {
		i := &ir.Instruction{Operands: c.Operands}
		if err := VerifyMemOperands(i); err == nil || !strings.Contains(err.Error(), c.ErrorSubstring) {
			t.Errorf("got error %v; expected error to contain %q", err, c.ErrorSubstring)
		}
	}
}
