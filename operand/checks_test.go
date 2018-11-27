package operand

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/mmcloughlin/avo"
	"github.com/mmcloughlin/avo/reg"
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

		// SIMD registers
		{IsXmm0, reg.X0, true},
		{IsXmm0, reg.X13, false},
		{IsXmm0, reg.Y3, false},

		{IsXmm, reg.X0, true},
		{IsXmm, reg.X13, true},
		{IsXmm, reg.Y3, false},
		{IsXmm, reg.Z23, false},

		{IsYmm, reg.Y0, true},
		{IsYmm, reg.Y13, true},
		{IsYmm, reg.Y31, true},
		{IsYmm, reg.X3, false},
		{IsYmm, reg.Z3, false},

		// Memory operands
		{IsM, Mem{Base: reg.CX}, true},
		{IsM, Mem{Base: reg.ECX}, true},
		{IsM, Mem{Base: reg.RCX}, true},
		{IsM, Mem{Base: reg.CL}, false},

		{IsM8, Mem{Disp: 8, Base: reg.CL}, true},
		{IsM8, Mem{Disp: 8, Base: reg.CL, Index: reg.AH, Scale: 2}, true},
		{IsM8, Mem{Disp: 8, Base: reg.AX, Index: reg.AH, Scale: 2}, false},
		{IsM8, Mem{Disp: 8, Base: reg.CL, Index: reg.R10, Scale: 2}, false},

		{IsM16, Mem{Disp: 4, Base: reg.DX}, true},
		{IsM16, Mem{Disp: 4, Base: reg.R13W, Index: reg.R8W, Scale: 2}, true},
		{IsM16, Mem{Disp: 4, Base: reg.ESI, Index: reg.R8W, Scale: 2}, false},
		{IsM16, Mem{Disp: 4, Base: reg.R13W, Index: reg.R9, Scale: 2}, false},

		{IsM32, Mem{Base: reg.R13L, Index: reg.EBX, Scale: 2}, true},
		{IsM32, Mem{Base: reg.R13W}, false},

		{IsM64, Mem{Base: reg.RBX, Index: reg.R12, Scale: 2}, true},
		{IsM64, Mem{Base: reg.R13L}, false},

		{IsM128, Mem{Base: reg.RBX, Index: reg.R12, Scale: 2}, true},
		{IsM128, Mem{Base: reg.R13L}, false},

		{IsM256, Mem{Base: reg.RBX, Index: reg.R12, Scale: 2}, true},
		{IsM256, Mem{Base: reg.R13L}, false},
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
