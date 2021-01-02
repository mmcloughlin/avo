package x86

import (
	"reflect"
	"testing"

	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

func TestCases(t *testing.T) {
	must := MustInstruction(t)

	cases := []struct {
		Name        string
		Instruction *ir.Instruction
		Expect      *ir.Instruction
	}{
		// In the merge-masking case, the output register should also be an
		// input. This test confirms that Z3 appears in the input operands list.
		{
			Name:        "avx512_masking_merging_input_registers",
			Instruction: must(VPADDD(reg.Z1, reg.Z2, reg.K1, reg.Z3)),
			Expect: &ir.Instruction{
				Opcode:   "VPADDD",
				Operands: []operand.Op{reg.Z1, reg.Z2, reg.K1, reg.Z3},
				Inputs:   []operand.Op{reg.Z1, reg.Z2, reg.K1, reg.Z3},
				Outputs:  []operand.Op{reg.Z3},
				ISA:      []string{"AVX512F"},
			},
		},
		// In the zeroing-masking case, the output register is not an input.
		// This test case is the same as above, but with the zeroing suffix. In
		// this case Z3 should not be an input.
		{
			Name:        "avx512_masking_zeroing_input_registers",
			Instruction: must(VPADDD_Z(reg.Z1, reg.Z2, reg.K1, reg.Z3)),
			Expect: &ir.Instruction{
				Opcode:   "VPADDD",
				Suffixes: []string{"Z"},
				Operands: []operand.Op{reg.Z1, reg.Z2, reg.K1}, // not Z3
				Inputs:   []operand.Op{reg.Z1, reg.Z2, reg.K1, reg.Z3},
				Outputs:  []operand.Op{reg.Z3},
				ISA:      []string{"AVX512F"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if !reflect.DeepEqual(c.Instruction, c.Expect) {
				t.Logf("   got = %#v", c.Instruction)
				t.Logf("expect = %#v", c.Expect)
				t.FailNow()
			}
		})
	}
}

func MustInstruction(t *testing.T) func(*ir.Instruction, error) *ir.Instruction {
	return func(i *ir.Instruction, err error) *ir.Instruction {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return i
	}
}
