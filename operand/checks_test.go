package operand

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/mmcloughlin/avo/reg"

	"github.com/mmcloughlin/avo"
)

func TestChecks(t *testing.T) {
	cases := []struct {
		Predicate func(avo.Operand) bool
		Operand   avo.Operand
		Expect    bool
	}{
		// Immediates
		{Is1, Imm(1), true},
		{Is1, Imm(23), false},
		{Is3, Imm(3), true},
		{Is3, Imm(23), false},
		{IsImm2u, Imm(3), true},
		{IsImm2u, Imm(4), false},
		{IsImm8, Imm(255), true},
		{IsImm8, Imm(256), false},
		{IsImm16, Imm((1 << 16) - 1), true},
		{IsImm16, Imm(1 << 16), false},
		{IsImm32, Imm((1 << 32) - 1), true},
		{IsImm32, Imm(1 << 32), false},
		{IsImm64, Imm((1 << 64) - 1), true},

		// Specific registers
		{IsAl, reg.AL, true},
		{IsAl, reg.CL, false},
		{IsCl, reg.CL, true},
		{IsCl, reg.DH, false},
		{IsAx, reg.AX, true},
		{IsAx, reg.DX, false},
		{IsEax, reg.EAX, true},
		{IsEax, reg.ECX, false},
		{IsRax, reg.RAX, true},
		{IsRax, reg.R13, false},

		// General-purpose registers
		{IsR8, reg.AL, true},
		{IsR8, reg.CH, true},
		{IsR8, reg.EAX, false},
		{IsR16, reg.DX, true},
		{IsR16, reg.R10W, true},
		{IsR16, reg.R10B, false},
		{IsR32, reg.EBP, true},
		{IsR32, reg.R14L, true},
		{IsR32, reg.R8, false},
		{IsR64, reg.RDX, true},
		{IsR64, reg.R10, true},
		{IsR64, reg.EBX, false},
	}
	for _, c := range cases {
		if c.Predicate(c.Operand) != c.Expect {
			t.Errorf("%s( %#v ) != %v", funcname(c.Predicate), c.Operand, c.Expect)
		}
	}
}

func funcname(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
