package operand

import (
	"math"

	"github.com/mmcloughlin/avo/reg"

	"github.com/mmcloughlin/avo"
)

// Is1 returns true if op is the immediate constant 1.
func Is1(op avo.Operand) bool {
	i, ok := op.(Imm)
	return ok && i == 1
}

// Is3 returns true if op is the immediate constant 3.
func Is3(op avo.Operand) bool {
	i, ok := op.(Imm)
	return ok && i == 3
}

// IsImm2u returns true if op is a 2-bit unsigned immediate (less than 4).
func IsImm2u(op avo.Operand) bool {
	i, ok := op.(Imm)
	return ok && i < 4
}

// IsImm8 returns true is op is an 8-bit immediate.
func IsImm8(op avo.Operand) bool {
	i, ok := op.(Imm)
	return ok && i <= math.MaxUint8
}

// IsImm16 returns true is op is a 16-bit immediate.
func IsImm16(op avo.Operand) bool {
	i, ok := op.(Imm)
	return ok && i <= math.MaxUint16
}

// IsImm32 returns true is op is a 32-bit immediate.
func IsImm32(op avo.Operand) bool {
	i, ok := op.(Imm)
	return ok && i <= math.MaxUint32
}

// IsImm64 returns true is op is a 64-bit immediate.
func IsImm64(op avo.Operand) bool {
	_, ok := op.(Imm)
	return ok
}

// IsAl returns true if op is the AL register.
func IsAl(op avo.Operand) bool {
	return op == reg.AL
}

// IsCl returns true if op is the CL register.
func IsCl(op avo.Operand) bool {
	return op == reg.CL
}

// IsAx returns true if op is the 16-bit AX register.
func IsAx(op avo.Operand) bool {
	return op == reg.AX
}

// IsEax returns true if op is the 32-bit EAX register.
func IsEax(op avo.Operand) bool {
	return op == reg.EAX
}

// IsRax returns true if op is the 64-bit RAX register.
func IsRax(op avo.Operand) bool {
	return op == reg.RAX
}

// IsR8 returns true if op is an 8-bit general-purpose register.
func IsR8(op avo.Operand) bool {
	return IsGP(op, 1)
}

// IsR16 returns true if op is a 16-bit general-purpose register.
func IsR16(op avo.Operand) bool {
	return IsGP(op, 2)
}

// IsR32 returns true if op is a 32-bit general-purpose register.
func IsR32(op avo.Operand) bool {
	return IsGP(op, 4)
}

// IsR64 returns true if op is a 64-bit general-purpose register.
func IsR64(op avo.Operand) bool {
	return IsGP(op, 8)
}

// IsGP returns true if op is a general-purpose register of size n bytes.
func IsGP(op avo.Operand, n uint) bool {
	r, ok := op.(reg.Register)
	return ok && r.Kind() == reg.GeneralPurpose.Kind && r.Bytes() == n
}

func IsXmm0(op avo.Operand) bool {
	return false
}

func IsXmm(op avo.Operand) bool {
	return false
}

func IsYmm(op avo.Operand) bool {
	return false
}

func IsM(op avo.Operand) bool {
	return false
}

func IsM8(op avo.Operand) bool {
	return false
}

func IsM16(op avo.Operand) bool {
	return false
}

func IsM32(op avo.Operand) bool {
	return false
}

func IsM64(op avo.Operand) bool {
	return false
}

func IsM128(op avo.Operand) bool {
	return false
}

func IsM256(op avo.Operand) bool {
	return false
}

func IsVm32x(op avo.Operand) bool {
	return false
}

func IsVm64x(op avo.Operand) bool {
	return false
}

func IsVm32y(op avo.Operand) bool {
	return false
}

func IsVm64y(op avo.Operand) bool {
	return false
}

func IsRel8(op avo.Operand) bool {
	return false
}

func IsRel32(op avo.Operand) bool {
	return false
}
