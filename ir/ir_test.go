package ir

import (
	"reflect"
	"testing"

	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

func TestFunctionLabels(t *testing.T) {
	f := NewFunction("labels")
	f.AddInstruction(&Instruction{})
	f.AddLabel("a")
	f.AddInstruction(&Instruction{})
	f.AddLabel("b")
	f.AddInstruction(&Instruction{})
	f.AddLabel("c")
	f.AddInstruction(&Instruction{})

	expect := []Label{"a", "b", "c"}
	if got := f.Labels(); !reflect.DeepEqual(expect, got) {
		t.Fatalf("f.Labels() = %v; expect %v", got, expect)
	}
}

func TestIsNotInternal(t *testing.T) {
	f := NewFunction("isNotInternal")
	if f.IsInternal {
		t.Fatalf("expected f.IsInternal to be false, got %t", f.IsInternal)
	}
}

func TestIsInternal(t *testing.T) {
	f := NewInternalFunction("isInternal")
	if !f.IsInternal {
		t.Fatalf("expected f.IsInternal to be true, got %t", f.IsInternal)
	}
}

func TestInputRegisters(t *testing.T) {
	cases := []struct {
		Name   string
		Inst   *Instruction
		Expect []reg.Register
	}{
		{
			Name: "reg",
			Inst: &Instruction{
				Inputs: []operand.Op{
					reg.RAX,
					reg.R13,
				},
				Outputs: []operand.Op{
					reg.RBX,
				},
			},
			Expect: []reg.Register{reg.RAX, reg.R13},
		},
		{
			Name: "mem",
			Inst: &Instruction{
				Inputs: []operand.Op{
					operand.Mem{
						Base:  reg.RSI,
						Index: reg.RDI,
					},
					reg.R13,
				},
				Outputs: []operand.Op{
					operand.Mem{
						Base:  reg.R9,
						Index: reg.R11,
					},
				},
			},
			Expect: []reg.Register{
				reg.RSI,
				reg.RDI,
				reg.R13,
				reg.R9,
				reg.R11,
			},
		},
		{
			Name: "cancelling_inputs",
			Inst: &Instruction{
				CancellingInputs: true,
				Inputs: []operand.Op{
					reg.R13,
					reg.R13,
				},
				Outputs: []operand.Op{
					operand.Mem{
						Base:  reg.R9,
						Index: reg.R11,
					},
				},
			},
			Expect: []reg.Register{reg.R9, reg.R11},
		},
	}
	for _, c := range cases {
		if got := c.Inst.InputRegisters(); !reflect.DeepEqual(got, c.Expect) {
			t.Errorf("%s: got %v; expect %v", c.Inst.Opcode, got, c.Expect)
		}
	}
}
