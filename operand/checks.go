package operand

import (
	"math"

	"github.com/mmcloughlin/avo/reg"
)

// Pure type assertion checks:

// IsRegister returns whether op has type reg.Register.
func IsRegister(op Op) bool { _, ok := op.(reg.Register); return ok }

// IsMem returns whether op has type Mem.
func IsMem(op Op) bool { _, ok := op.(Mem); return ok }

// IsImm returns whether op has type Imm.
func IsImm(op Op) bool { _, ok := op.(Imm); return ok }

// IsRel returns whether op has type Rel.
func IsRel(op Op) bool { _, ok := op.(Rel); return ok }

// Checks corresponding to specific operand types in the Intel Manual:

// Is1 returns true if op is the immediate constant 1.
func Is1(op Op) bool {
	i, ok := op.(Imm)
	return ok && i == 1
}

// Is3 returns true if op is the immediate constant 3.
func Is3(op Op) bool {
	i, ok := op.(Imm)
	return ok && i == 3
}

// IsImm2u returns true if op is a 2-bit unsigned immediate (less than 4).
func IsImm2u(op Op) bool {
	i, ok := op.(Imm)
	return ok && i < 4
}

// IsImm8 returns true is op is an 8-bit immediate.
func IsImm8(op Op) bool {
	i, ok := op.(Imm)
	return ok && i <= math.MaxUint8
}

// IsImm16 returns true is op is a 16-bit immediate.
func IsImm16(op Op) bool {
	i, ok := op.(Imm)
	return ok && i <= math.MaxUint16
}

// IsImm32 returns true is op is a 32-bit immediate.
func IsImm32(op Op) bool {
	i, ok := op.(Imm)
	return ok && i <= math.MaxUint32
}

// IsImm64 returns true is op is a 64-bit immediate.
func IsImm64(op Op) bool {
	_, ok := op.(Imm)
	return ok
}

// IsAl returns true if op is the AL register.
func IsAl(op Op) bool {
	return op == reg.AL
}

// IsCl returns true if op is the CL register.
func IsCl(op Op) bool {
	return op == reg.CL
}

// IsAx returns true if op is the 16-bit AX register.
func IsAx(op Op) bool {
	return op == reg.AX
}

// IsEax returns true if op is the 32-bit EAX register.
func IsEax(op Op) bool {
	return op == reg.EAX
}

// IsRax returns true if op is the 64-bit RAX register.
func IsRax(op Op) bool {
	return op == reg.RAX
}

// IsR8 returns true if op is an 8-bit general-purpose register.
func IsR8(op Op) bool {
	return IsGP(op, 1)
}

// IsR16 returns true if op is a 16-bit general-purpose register.
func IsR16(op Op) bool {
	return IsGP(op, 2)
}

// IsR32 returns true if op is a 32-bit general-purpose register.
func IsR32(op Op) bool {
	return IsGP(op, 4)
}

// IsR64 returns true if op is a 64-bit general-purpose register.
func IsR64(op Op) bool {
	return IsGP(op, 8)
}

// IsPseudo returns true if op is a pseudo register.
func IsPseudo(op Op) bool {
	return IsRegisterKind(op, reg.Internal)
}

// IsGP returns true if op is a general-purpose register of size n bytes.
func IsGP(op Op, n uint) bool {
	return IsRegisterKindSize(op, reg.GP, n)
}

// IsXmm0 returns true if op is the X0 register.
func IsXmm0(op Op) bool {
	return op == reg.X0
}

// IsXmm returns true if op is a 128-bit XMM register.
func IsXmm(op Op) bool {
	return IsRegisterKindSize(op, reg.SSEAVX, 16)
}

// IsYmm returns true if op is a 256-bit YMM register.
func IsYmm(op Op) bool {
	return IsRegisterKindSize(op, reg.SSEAVX, 32)
}

// IsRegisterKindSize returns true if op is a register of the given kind and size in bytes.
func IsRegisterKindSize(op Op, k reg.Kind, n uint) bool {
	r, ok := op.(reg.Register)
	return ok && r.Kind() == k && r.Bytes() == n
}

// IsRegisterKind returns true if op is a register of the given kind.
func IsRegisterKind(op Op, k reg.Kind) bool {
	r, ok := op.(reg.Register)
	return ok && r.Kind() == k
}

// IsM returns true if op is a 16-, 32- or 64-bit memory operand.
func IsM(op Op) bool {
	// TODO(mbm): confirm "m" check is defined correctly
	// Intel manual: "A 16-, 32- or 64-bit operand in memory."
	return IsM16(op) || IsM32(op) || IsM64(op)
}

// IsM8 returns true if op is an 8-bit memory operand.
func IsM8(op Op) bool {
	// TODO(mbm): confirm "m8" check is defined correctly
	// Intel manual: "A byte operand in memory, usually expressed as a variable or
	// array name, but pointed to by the DS:(E)SI or ES:(E)DI registers. In 64-bit
	// mode, it is pointed to by the RSI or RDI registers."
	return IsMSize(op, 1)
}

// IsM16 returns true if op is a 16-bit memory operand.
func IsM16(op Op) bool {
	return IsMSize(op, 2)
}

// IsM32 returns true if op is a 16-bit memory operand.
func IsM32(op Op) bool {
	return IsMSize(op, 4)
}

// IsM64 returns true if op is a 64-bit memory operand.
func IsM64(op Op) bool {
	return IsMSize(op, 8)
}

// IsMSize returns true if op is a memory operand using general-purpose address
// registers of the given size in bytes.
func IsMSize(op Op, n uint) bool {
	// TODO(mbm): should memory operands have a size attribute as well?
	m, ok := op.(Mem)
	return ok && IsMReg(m.Base) && (m.Index == nil || IsMReg(m.Index))
}

// IsMReg returns true if op is a register that can be used in a memory operand.
func IsMReg(op Op) bool {
	return IsPseudo(op) || IsRegisterKind(op, reg.GP)
}

// IsM128 returns true if op is a 128-bit memory operand.
func IsM128(op Op) bool {
	// TODO(mbm): should "m128" be the same as "m64"?
	return IsM64(op)
}

// IsM256 returns true if op is a 256-bit memory operand.
func IsM256(op Op) bool {
	// TODO(mbm): should "m256" be the same as "m64"?
	return IsM64(op)
}

// IsVm32x returns true if op is a vector memory operand with 32-bit XMM index.
func IsVm32x(op Op) bool {
	return IsVmx(op)
}

// IsVm64x returns true if op is a vector memory operand with 64-bit XMM index.
func IsVm64x(op Op) bool {
	return IsVmx(op)
}

// IsVmx returns true if op is a vector memory operand with XMM index.
func IsVmx(op Op) bool {
	return isvm(op, IsXmm)
}

// IsVm32y returns true if op is a vector memory operand with 32-bit YMM index.
func IsVm32y(op Op) bool {
	return IsVmy(op)
}

// IsVm64y returns true if op is a vector memory operand with 64-bit YMM index.
func IsVm64y(op Op) bool {
	return IsVmy(op)
}

// IsVmy returns true if op is a vector memory operand with YMM index.
func IsVmy(op Op) bool {
	return isvm(op, IsYmm)
}

func isvm(op Op, idx func(Op) bool) bool {
	m, ok := op.(Mem)
	return ok && IsR64(m.Base) && idx(m.Index)
}

// IsRel8 returns true if op is an 8-bit offset relative to instruction pointer.
func IsRel8(op Op) bool {
	r, ok := op.(Rel)
	return ok && r == Rel(int8(r))
}

// IsRel32 returns true if op is an offset relative to instruction pointer, or a
// label reference.
func IsRel32(op Op) bool {
	// TODO(mbm): should labels be considered separately?
	_, rel := op.(Rel)
	_, label := op.(LabelRef)
	return rel || label
}
