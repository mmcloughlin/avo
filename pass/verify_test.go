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

func TestVerifyNoLabelOnPCALIGN(t *testing.T) {
	pcalign := func() ir.Node { return &ir.Instruction{Opcode: "PCALIGN", Operands: []operand.Op{operand.U8(64)}} }
	ret := func() ir.Node { return &ir.Instruction{Opcode: "RET"} }
	cases := []struct {
		Name    string
		Nodes   []ir.Node
		WantErr bool
	}{
		{"label on pcalign", []ir.Node{ir.Label("skip"), pcalign(), ret()}, true},
		{"comment between", []ir.Node{ir.Label("skip"), ir.NewComment("x"), pcalign(), ret()}, true},
		{"pcalign before label", []ir.Node{pcalign(), ir.Label("loop"), ret()}, false},
		{"instruction between", []ir.Node{ir.Label("skip"), ret(), pcalign(), ret()}, false},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			fn := ir.NewFunction("f")
			fn.Nodes = c.Nodes
			err := VerifyNoLabelOnPCALIGN(fn)
			if (err != nil) != c.WantErr {
				t.Fatalf("got error %v, want error %v", err, c.WantErr)
			}
		})
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
