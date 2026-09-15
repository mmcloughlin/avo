//go:build ignore

// This program generates functions that exercise the EXPERIMENTAL arm64
// lowering printer. The same avo source is emitted to both amd64 and arm64; the
// test file runs each generated function against a pure-Go reference (via
// testing/quick), so on each architecture CI verifies that the lowering is
// semantics-preserving. It deliberately covers the operand-width and
// sub-register cases that produce plausible-but-wrong asm (MOVL zero-extension,
// high-byte reads) and the conditional mappings real generators may not hit.
package main

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
	"github.com/mmcloughlin/avo/tests/arm64lower/condspec"
	"github.com/mmcloughlin/avo/tests/arm64lower/propspec"
)

func main() {
	// Add: ADDQ.
	TEXT("Add", NOSPLIT, "func(x, y uint64) uint64")
	{
		x := Load(Param("x"), GP64())
		y := Load(Param("y"), GP64())
		ADDQ(x, y)
		Store(y, ReturnIndex(0))
		RET()
	}

	// Sub: SUBQ.
	TEXT("Sub", NOSPLIT, "func(x, y uint64) uint64")
	{
		x := Load(Param("x"), GP64())
		y := Load(Param("y"), GP64())
		SUBQ(y, x) // x -= y
		Store(x, ReturnIndex(0))
		RET()
	}

	// Neg: NEGQ.
	TEXT("Neg", NOSPLIT, "func(x uint64) uint64")
	{
		x := Load(Param("x"), GP64())
		NEGQ(x)
		Store(x, ReturnIndex(0))
		RET()
	}

	// AndOrXor: ANDQ, ORQ, XORQ combined.
	TEXT("AndOrXor", NOSPLIT, "func(x, y uint64) uint64")
	{
		x := Load(Param("x"), GP64())
		y := Load(Param("y"), GP64())
		a := GP64()
		MOVQ(x, a)
		ANDQ(y, a) // a = x & y
		o := GP64()
		MOVQ(x, o)
		ORQ(y, o)  // o = x | y
		XORQ(o, a) // a = (x&y) ^ (x|y)
		Store(a, ReturnIndex(0))
		RET()
	}

	// Variable shifts by CL: SHLQ, SHRQ, ROLQ.
	TEXT("ShlCL", NOSPLIT, "func(x, n uint64) uint64")
	{
		x := GP64()
		Load(Param("x"), x)
		n := GP64()
		Load(Param("n"), n)
		MOVQ(n, reg.RCX)
		SHLQ(reg.CL, x)
		Store(x, ReturnIndex(0))
		RET()
	}
	TEXT("ShrCL", NOSPLIT, "func(x, n uint64) uint64")
	{
		x := GP64()
		Load(Param("x"), x)
		n := GP64()
		Load(Param("n"), n)
		MOVQ(n, reg.RCX)
		SHRQ(reg.CL, x)
		Store(x, ReturnIndex(0))
		RET()
	}
	TEXT("RolCL", NOSPLIT, "func(x, n uint64) uint64")
	{
		x := GP64()
		Load(Param("x"), x)
		n := GP64()
		Load(Param("n"), n)
		MOVQ(n, reg.RCX)
		ROLQ(reg.CL, x)
		Store(x, ReturnIndex(0))
		RET()
	}

	// ZeroExt32: MOVL register move must zero the upper 32 bits (x86 semantics).
	// The case where Go arm64 MOVW would wrongly sign-extend.
	TEXT("ZeroExt32", NOSPLIT, "func(x uint64) uint64")
	{
		x := GP64()
		Load(Param("x"), x)
		y := GP64()
		MOVL(x.As32(), y.As32()) // y = uint64(uint32(x))
		// Keep x live so the truncating move cannot be coalesced/elided; the sum
		// observes whether the upper 32 bits of y were correctly zeroed.
		ADDQ(x, y)
		Store(y, ReturnIndex(0))
		RET()
	}

	// HighByte: read bits 8-15 via the AH-style high-byte register.
	TEXT("HighByte", NOSPLIT, "func(x uint64) uint64")
	{
		x := GP64()
		Load(Param("x"), x)
		y := GP64()
		MOVB(x.As8H(), y.As8L()) // y_low = (x >> 8) & 0xff
		MOVBQZX(y.As8(), y)      // zero-extend on both architectures
		Store(y, ReturnIndex(0))
		RET()
	}

	// ShiftExtractByte64/32: MOVQ/MOVL + SHRQ/SHRL $n + MOVBQZX/MOVBLZX, all
	// writing back into the same register -- the x86 two-address idiom for
	// extracting one byte at a fixed bit offset (see zstd/_generate/gen.go's
	// "moB" comment), which the arm64 printer's shiftExtractFold collapses to
	// one UBFX. Covers both operand widths and the shift-amount boundary
	// (0 and width-8, the extremes shiftExtractFold accepts).
	shiftExtractByte := func(name string, shift uint8) {
		TEXT(name, NOSPLIT, "func(x uint64) uint64")
		x := GP64()
		Load(Param("x"), x)
		tmp := GP64()
		MOVQ(x, tmp)
		SHRQ(operand.U8(shift), tmp)
		MOVBQZX(tmp.As8(), tmp)
		// x must stay live past the shift (its value is needed below), which is
		// what forces the register allocator to keep it in a register separate
		// from tmp instead of coalescing the MOVQ away -- reproducing the shape
		// zstd/_generate/gen.go's "moB" site actually has (see the comment
		// there: "copy ofState, its current value is needed below"). Without
		// this, x and tmp end up in the same register, the MOVQ is elided
		// before the printer ever sees it, and shiftExtractFold has no MOVQ
		// node to match -- silently testing nothing.
		ADDQ(x, tmp)
		Store(tmp, ReturnIndex(0))
		RET()
	}
	shiftExtractByte("ShiftExtractByte64Lo", 0)
	shiftExtractByte("ShiftExtractByte64Mid", 8)
	shiftExtractByte("ShiftExtractByte64Hi", 56)

	TEXT("ShiftExtractByte32", NOSPLIT, "func(x uint64) uint64")
	{
		x := GP64()
		Load(Param("x"), x)
		tmp := GP64()
		MOVL(x.As32(), tmp.As32())
		SHRL(operand.U8(16), tmp.As32())
		MOVBLZX(tmp.As8(), tmp.As32())
		ADDQ(x, tmp) // keep x live past the shift; see shiftExtractByte above
		Store(tmp, ReturnIndex(0))
		RET()
	}

	// LoadIdx: indexed memory load (base+index*scale), lowered via a scratch ADD.
	TEXT("LoadIdx", NOSPLIT, "func(p *[8]uint64, i uint64) uint64")
	{
		p := Load(Param("p"), GP64())
		i := Load(Param("i"), GP64())
		v := GP64()
		MOVQ(operand.Mem{Base: p, Index: i, Scale: 8}, v)
		Store(v, ReturnIndex(0))
		RET()
	}

	// Copy16: MOVUPS 16-byte copy through a vector register.
	TEXT("Copy16", NOSPLIT, "func(dst, src *[16]byte)")
	{
		dst := Load(Param("dst"), GP64())
		src := Load(Param("src"), GP64())
		x := XMM()
		MOVUPS(operand.Mem{Base: src}, x)
		MOVUPS(x, operand.Mem{Base: dst})
		RET()
	}

	// Prefetch: PREFETCHT0/T1/T2/NTA are hints with no architectural effect,
	// so the reference is the load that follows. The value proves the
	// lowering assembled, left the address registers alone and did not
	// disturb the flags a following branch reads. The operand shapes cover
	// every path lowerPrefetch takes: plain base, base+aligned disp (folded
	// into PRFM), base+index and misaligned/large/negative disp (each through
	// the scratch ADD).
	TEXT("Prefetch", NOSPLIT, "func(p *[64]uint64, i uint64) uint64")
	{
		p := Load(Param("p"), GP64())
		i := Load(Param("i"), GP64())
		ANDQ(operand.Imm(63), i)
		PREFETCHT0(operand.Mem{Base: p})
		PREFETCHT1(operand.Mem{Base: p, Disp: 64})
		PREFETCHT2(operand.Mem{Base: p, Disp: 32760})
		PREFETCHNTA(operand.Mem{Base: p, Disp: 4})
		PREFETCHT0(operand.Mem{Base: p, Disp: 260})
		PREFETCHT0(operand.Mem{Base: p, Disp: 32768})
		PREFETCHT0(operand.Mem{Base: p, Disp: -64})
		PREFETCHT0(operand.Mem{Base: p, Index: i, Scale: 8})
		PREFETCHT0(operand.Mem{Base: p, Index: i, Scale: 1, Disp: 8})
		// A prefetch between a flag producer and its consumer must be
		// transparent on both architectures.
		TESTQ(i, i)
		v := GP64()
		MOVQ(operand.Mem{Base: p, Index: i, Scale: 8}, v)
		PREFETCHT0(operand.Mem{Base: p, Index: i, Scale: 8, Disp: 64})
		JZ(operand.LabelRef("prefetch_done"))
		ADDQ(operand.Imm(1), v)
		Label("prefetch_done")
		Store(v, ReturnIndex(0))
		RET()
	}

	// Condition helpers: each returns 1 if the condition holds, else 0,
	// exercising a distinct branch mnemonic (including JGE, unused by zstd).
	branch := func(name string, jmp func(operand.Op)) {
		TEXT(name, NOSPLIT, "func(a, b uint64) uint64")
		a := Load(Param("a"), GP64())
		b := Load(Param("b"), GP64())
		res := GP64()
		CMPQ(a, b)
		jmp(operand.LabelRef(name + "_yes"))
		MOVQ(operand.U64(0), res)
		JMP(operand.LabelRef(name + "_end"))
		Label(name + "_yes")
		MOVQ(operand.U64(1), res)
		Label(name + "_end")
		Store(res, ReturnIndex(0))
		RET()
	}
	branch("LessS", JLT)      // signed <
	branch("GreaterEqS", JGE) // signed >= (not used by zstd)
	branch("LessU", JCS)      // unsigned <
	branch("AboveU", JHI)     // unsigned >
	branch("Equal", JEQ)

	// SelectEq: CMOVQEQ — res = (a==b) ? c : d.
	TEXT("SelectEq", NOSPLIT, "func(a, b, c, d uint64) uint64")
	{
		a := Load(Param("a"), GP64())
		b := Load(Param("b"), GP64())
		c := Load(Param("c"), GP64())
		res := Load(Param("d"), GP64())
		CMPQ(a, b)
		CMOVQEQ(c, res)
		Store(res, ReturnIndex(0))
		RET()
	}

	// selBranch emits "res = (op sets ZF) ? 1 : 0" with a flag-transparent MOVQ
	// wedged between the flag-setting op and the JEQ that consumes it. x86 leaves
	// flags untouched across a MOV, so a lowering that only inspects the single
	// instruction following the producer fails to emit a flag-setting variant and
	// the branch then reads stale NZCV.
	selBranch := func(name string, produce func(res reg.GPVirtual)) {
		TEXT(name, NOSPLIT, "func(x, y uint64) uint64")
		res := GP64()
		produce(res)
		MOVQ(operand.U64(0), res)            // flag-transparent gap instruction
		JEQ(operand.LabelRef(name + "_yes")) // consumes ZF from the producer above
		JMP(operand.LabelRef(name + "_end"))
		Label(name + "_yes")
		MOVQ(operand.U64(1), res)
		Label(name + "_end")
		Store(res, ReturnIndex(0))
		RET()
	}

	// SubGapEq: SUBQ producer separated from JEQ by a MOV. res = (x-y==0).
	selBranch("SubGapEq", func(res reg.GPVirtual) {
		x := Load(Param("x"), GP64())
		y := Load(Param("y"), GP64())
		SUBQ(y, x) // x -= y; ZF set iff x == y
	})

	// DecGapZero: DECQ producer separated from JEQ by a MOV. res = (x-1==0).
	// Covers the lowerIncDec path (the loop-counter idiom) as well as lowerArith.
	selBranch("DecGapZero", func(res reg.GPVirtual) {
		x := Load(Param("x"), GP64())
		DECQ(x) // x--; ZF set iff x == 1
	})

	// cmpLow32 emits "res = (cmp of low 32 bits) ? 1 : 0" via a 32-bit compare on
	// operands whose upper 32 bits differ, so folding CMPL to a 64-bit CMP is
	// observably wrong. jmp selects the (signed or unsigned) branch under test.
	cmpLow32 := func(name string, jmp func(operand.Op)) {
		TEXT(name, NOSPLIT, "func(a, b uint64) uint64")
		a := GP64()
		Load(Param("a"), a)
		b := GP64()
		Load(Param("b"), b)
		res := GP64()
		CMPL(a.As32(), b.As32())
		jmp(operand.LabelRef(name + "_yes"))
		MOVQ(operand.U64(0), res)
		JMP(operand.LabelRef(name + "_end"))
		Label(name + "_yes")
		MOVQ(operand.U64(1), res)
		Label(name + "_end")
		Store(res, ReturnIndex(0))
		RET()
	}
	cmpLow32("CmpL32Eq", JEQ)    // low32(a) == low32(b)
	cmpLow32("CmpL32LessS", JLT) // int32(a) < int32(b)
	cmpLow32("CmpL32LessU", JCS) // uint32(a) < uint32(b)

	// TestL32: TESTL sets ZF from the low-32-bit AND; the 64-bit fold is wrong
	// when the operands share set bits only above bit 31. res = (low32 AND == 0).
	TEXT("TestL32", NOSPLIT, "func(a, m uint64) uint64")
	{
		a := GP64()
		Load(Param("a"), a)
		m := GP64()
		Load(Param("m"), m)
		res := GP64()
		TESTL(a.As32(), m.As32())
		JEQ(operand.LabelRef("TestL32_zero"))
		MOVQ(operand.U64(0), res)
		JMP(operand.LabelRef("TestL32_end"))
		Label("TestL32_zero")
		MOVQ(operand.U64(1), res)
		Label("TestL32_end")
		Store(res, ReturnIndex(0))
		RET()
	}

	// CmpW16Eq: res = (low 16 bits of a == low 16 bits of b) via a 16-bit
	// compare on operands whose upper 48 bits differ, exercising the EQ/NE-only
	// zero-extend fallback for sub-32-bit compares: a naive 64-bit fold would
	// see the differing upper bits and wrongly report inequality.
	TEXT("CmpW16Eq", NOSPLIT, "func(a, b uint64) uint64")
	{
		a := GP64()
		Load(Param("a"), a)
		b := GP64()
		Load(Param("b"), b)
		res := GP64()
		CMPW(a.As16(), b.As16())
		JEQ(operand.LabelRef("CmpW16Eq_yes"))
		MOVQ(operand.U64(0), res)
		JMP(operand.LabelRef("CmpW16Eq_end"))
		Label("CmpW16Eq_yes")
		MOVQ(operand.U64(1), res)
		Label("CmpW16Eq_end")
		Store(res, ReturnIndex(0))
		RET()
	}

	// CmpB8Ne: same idea as CmpW16Eq but 8-bit (CMPB), consumed by JNE instead
	// of JEQ.
	TEXT("CmpB8Ne", NOSPLIT, "func(a, b uint64) uint64")
	{
		a := GP64()
		Load(Param("a"), a)
		b := GP64()
		Load(Param("b"), b)
		res := GP64()
		CMPB(a.As8(), b.As8())
		JNE(operand.LabelRef("CmpB8Ne_diff"))
		MOVQ(operand.U64(0), res)
		JMP(operand.LabelRef("CmpB8Ne_end"))
		Label("CmpB8Ne_diff")
		MOVQ(operand.U64(1), res)
		Label("CmpB8Ne_end")
		Store(res, ReturnIndex(0))
		RET()
	}

	// TestW16Eq: TESTW sets ZF from the low-16-bit AND; the 64-bit fold is wrong
	// when the operands share set bits only above bit 15. res = (low16 AND == 0).
	TEXT("TestW16Eq", NOSPLIT, "func(a, m uint64) uint64")
	{
		a := GP64()
		Load(Param("a"), a)
		m := GP64()
		Load(Param("m"), m)
		res := GP64()
		TESTW(a.As16(), m.As16())
		JEQ(operand.LabelRef("TestW16Eq_zero"))
		MOVQ(operand.U64(0), res)
		JMP(operand.LabelRef("TestW16Eq_end"))
		Label("TestW16Eq_zero")
		MOVQ(operand.U64(1), res)
		Label("TestW16Eq_end")
		Store(res, ReturnIndex(0))
		RET()
	}

	// ---- multiplication ----

	// IMul2: two-operand IMULQ (dst *= src), low 64 bits.
	TEXT("IMul2", NOSPLIT, "func(x, y uint64) uint64")
	{
		x := Load(Param("x"), GP64())
		y := Load(Param("y"), GP64())
		IMULQ(x, y) // y *= x
		Store(y, ReturnIndex(0))
		RET()
	}

	// IMul3: three-operand IMUL3Q (dst = src * imm), low 64 bits.
	TEXT("IMul3", NOSPLIT, "func(x uint64) uint64")
	{
		x := Load(Param("x"), GP64())
		dst := GP64()
		IMUL3Q(operand.Imm(0x9E3779B1), x, dst) // dst = x * prime
		Store(dst, ReturnIndex(0))
		RET()
	}

	// MulWide: one-operand MULQ (RDX:RAX = RAX * src), unsigned 128-bit product.
	TEXT("MulWide", NOSPLIT, "func(x, y uint64) (lo uint64, hi uint64)")
	{
		Load(Param("x"), reg.RAX)
		Load(Param("y"), reg.RSI)
		MULQ(reg.RSI)
		Store(reg.RAX, ReturnIndex(0)) // low
		Store(reg.RDX, ReturnIndex(1)) // high
		RET()
	}

	// IMulWide: one-operand IMULQ (RDX:RAX = RAX * src), signed 128-bit product.
	TEXT("IMulWide", NOSPLIT, "func(x, y int64) (lo uint64, hi uint64)")
	{
		Load(Param("x"), reg.RAX)
		Load(Param("y"), reg.RSI)
		IMULQ(reg.RSI)
		Store(reg.RAX, ReturnIndex(0)) // low
		Store(reg.RDX, ReturnIndex(1)) // high (signed)
		RET()
	}

	// MulX: BMI2 MULXQ (flag-free wide multiply, implicit RDX factor).
	TEXT("MulX", NOSPLIT, "func(x, y uint64) (lo uint64, hi uint64)")
	{
		Load(Param("x"), reg.RSI)
		Load(Param("y"), reg.RDX)        // implicit MULX factor
		MULXQ(reg.RSI, reg.RAX, reg.RBX) // {RBX,RAX} = RSI * RDX
		Store(reg.RAX, ReturnIndex(0))   // low
		Store(reg.RBX, ReturnIndex(1))   // high
		RET()
	}

	// ---- BMI2 flag-free shifts / rotate ----
	shiftX := func(name string, op func(count, src, dst operand.Op)) {
		TEXT(name, NOSPLIT, "func(x, n uint64) uint64")
		x := Load(Param("x"), GP64())
		n := Load(Param("n"), GP64())
		dst := GP64()
		op(n, x, dst)
		Store(dst, ReturnIndex(0))
		RET()
	}
	shiftX("ShlX", func(count, src, dst operand.Op) { SHLXQ(count, src, dst) })
	shiftX("ShrX", func(count, src, dst operand.Op) { SHRXQ(count, src, dst) })
	shiftX("SarX", func(count, src, dst operand.Op) { SARXQ(count, src, dst) })

	// RorX: BMI2 RORXQ with an immediate rotate.
	TEXT("RorX", NOSPLIT, "func(x uint64) uint64")
	{
		x := Load(Param("x"), GP64())
		dst := GP64()
		RORXQ(operand.Imm(56), x, dst) // dst = ror(x, 56)
		Store(dst, ReturnIndex(0))
		RET()
	}

	// ---- BMI2 bit-field ops ----

	// Bzhi: BZHIQ zeroes bits at index n and above (n < 64).
	TEXT("Bzhi", NOSPLIT, "func(x, n uint64) uint64")
	{
		x := Load(Param("x"), GP64())
		n := Load(Param("n"), GP64())
		dst := GP64()
		BZHIQ(n, x, dst) // dst = x & ((1<<n)-1)
		Store(dst, ReturnIndex(0))
		RET()
	}

	// bextr: BEXTRQ extracts `length` bits starting at `start` (control built the
	// way zstd does, MOVQ of start|(length<<8) into a register).
	bextr := func(name string, start, length uint32) {
		TEXT(name, NOSPLIT, "func(x uint64) uint64")
		x := Load(Param("x"), GP64())
		ctrl := GP64()
		MOVQ(operand.U32(start|(length<<8)), ctrl)
		dst := GP64()
		BEXTRQ(ctrl, x, dst)
		Store(dst, ReturnIndex(0))
		RET()
	}
	bextr("Bextr88", 8, 8)    // zstd's exact usage
	bextr("Bextr4_12", 4, 12) // asymmetric start/len
	// Constant-fold boundary cases: field reaching bit 63 (LSR form), field
	// crossing past bit 63 (effective width shrinks), empty field, and start
	// past bit 63 (both constant zero).
	bextr("Bextr56_8", 56, 8)
	bextr("Bextr8_56", 8, 56)
	bextr("Bextr8_60", 8, 60)
	bextr("Bextr0_0", 0, 0)
	bextr("Bextr70_8", 70, 8)

	// BZHI/SHLX/SHRX with the count loaded as an adjacent constant, exercising
	// the immediate fold on arm64 (real BMI2 register semantics on amd64).
	constCount := func(name string, val uint32, op func(count, src, dst operand.Op)) {
		TEXT(name, NOSPLIT, "func(x uint64) uint64")
		x := Load(Param("x"), GP64())
		cnt := GP64()
		MOVQ(operand.U32(val), cnt)
		dst := GP64()
		op(cnt, x, dst)
		Store(dst, ReturnIndex(0))
		RET()
	}
	constCount("BzhiConst13", 13, func(c, s, d operand.Op) { BZHIQ(c, s, d) })
	constCount("BzhiConst0", 0, func(c, s, d operand.Op) { BZHIQ(c, s, d) })
	constCount("BzhiConst64", 64, func(c, s, d operand.Op) { BZHIQ(c, s, d) })
	constCount("ShlXConst9", 9, func(c, s, d operand.Op) { SHLXQ(c, s, d) })
	constCount("ShrXConst9", 9, func(c, s, d operand.Op) { SHRXQ(c, s, d) })

	// ---- huff0-style idioms: SETcc, ADC carry-accumulate, BSWAPL, ADDB ----

	// SetGe: XOR-zeroed destination + SETGE, the huff0 exhausted-flag pattern.
	TEXT("SetGe", NOSPLIT, "func(a, b uint64) uint64")
	{
		a, b, dst := GP64(), GP64(), GP64()
		Load(Param("a"), a)
		Load(Param("b"), b)
		XORL(dst.As32(), dst.As32())
		CMPQ(a, b)
		SETGE(dst.As8())
		Store(dst, ReturnIndex(0))
		RET()
	}

	// AdcAccum: the huff0 fillFast32 exhausted counter: acc's low byte += (x<4),
	// on an arbitrary accumulator (upper bits preserved, byte wrap exact).
	TEXT("AdcAccum", NOSPLIT, "func(x, acc uint64) uint64")
	{
		x, acc := GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("acc"), acc)
		CMPQ(x, operand.Imm(4))
		ADCB(operand.I8(0), acc.As8())
		Store(acc, ReturnIndex(0))
		RET()
	}

	// BswapL: 32-bit byte reverse, zero-extending like x86.
	TEXT("BswapL", NOSPLIT, "func(x uint64) uint64")
	{
		x := GP64()
		Load(Param("x"), x)
		BSWAPL(x.As32())
		Store(x, ReturnIndex(0))
		RET()
	}

	// AddByte: exact x86 byte add on arbitrary values: y's low byte gets
	// (x+y) mod 256, all other bits of y preserved.
	TEXT("AddByte", NOSPLIT, "func(x, y uint64) uint64")
	{
		x, y := GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("y"), y)
		ADDB(x.As8(), y.As8())
		Store(y, ReturnIndex(0))
		RET()
	}

	// MovbHighDst: MOVB into AH: replaces bits 15:8 of y with x's low byte,
	// preserving everything else (the huff0 byte-packing idiom).
	TEXT("MovbHighDst", NOSPLIT, "func(x, y uint64) uint64")
	{
		x := GP64()
		Load(Param("x"), x)
		y := reg.RAX // fixed: high-byte access requires AX-DX
		Load(Param("y"), y)
		MOVB(x.As8(), y.As8H())
		Store(y, ReturnIndex(0))
		RET()
	}

	// MovbLowPreserve: MOVB from AH to a low byte: replaces bits 7:0 of y with
	// bits 15:8 of x, preserving y's upper bits.
	TEXT("MovbLowPreserve", NOSPLIT, "func(x, y uint64) uint64")
	{
		x := reg.RCX // fixed: high-byte access requires AX-DX
		Load(Param("x"), x)
		y := GP64()
		Load(Param("y"), y)
		MOVB(x.As8H(), y.As8())
		Store(y, ReturnIndex(0))
		RET()
	}

	// ---- CMOVcc conditions newly enabled by the completed table ----
	selectCC := func(name string, cmov func(src, dst operand.Op)) {
		TEXT(name, NOSPLIT, "func(a, b, c, d uint64) uint64")
		a := Load(Param("a"), GP64())
		b := Load(Param("b"), GP64())
		c := Load(Param("c"), GP64())
		res := Load(Param("d"), GP64())
		CMPQ(a, b)
		cmov(c, res) // res = cond(a,b) ? c : d
		Store(res, ReturnIndex(0))
		RET()
	}
	selectCC("SelLtS", func(s, d operand.Op) { CMOVQLT(s, d) }) // signed <
	selectCC("SelLeS", func(s, d operand.Op) { CMOVQLE(s, d) }) // signed <=
	selectCC("SelGtS", func(s, d operand.Op) { CMOVQGT(s, d) }) // signed >
	selectCC("SelGeS", func(s, d operand.Op) { CMOVQGE(s, d) }) // signed >=
	selectCC("SelLsU", func(s, d operand.Op) { CMOVQLS(s, d) }) // unsigned <=
	selectCC("SelMi", func(s, d operand.Op) { CMOVQMI(s, d) })  // sign set
	selectCC("SelPl", func(s, d operand.Op) { CMOVQPL(s, d) })  // sign clear

	// ---- extension, 32-bit ALU, rotates, bit ops (s2 + general coverage) ----

	un := func(name string, emit func(x, dst reg.GPVirtual)) {
		TEXT(name, NOSPLIT, "func(x uint64) uint64")
		x, dst := GP64(), GP64()
		Load(Param("x"), x)
		emit(x, dst)
		Store(dst, ReturnIndex(0))
		RET()
	}
	un("SxL", func(x, d reg.GPVirtual) { MOVLQSX(x.As32(), d) })
	un("SxB", func(x, d reg.GPVirtual) { MOVBQSX(x.As8(), d) })
	un("SxBL", func(x, d reg.GPVirtual) { MOVBLSX(x.As8(), d.As32()) })
	un("SxWL", func(x, d reg.GPVirtual) { MOVWLSX(x.As16(), d.As32()) })
	un("NegL", func(x, d reg.GPVirtual) { MOVQ(x, d); NEGL(d.As32()) })
	un("NotL", func(x, d reg.GPVirtual) { MOVQ(x, d); NOTL(d.As32()) })
	un("NotQ", func(x, d reg.GPVirtual) { MOVQ(x, d); NOTQ(d) })
	un("RolL7", func(x, d reg.GPVirtual) { MOVQ(x, d); ROLL(operand.U8(7), d.As32()) })
	un("RorQ9", func(x, d reg.GPVirtual) { MOVQ(x, d); RORQ(operand.U8(9), d) })
	un("RorL9", func(x, d reg.GPVirtual) { MOVQ(x, d); RORL(operand.U8(9), d.As32()) })
	un("BtrQ5", func(x, d reg.GPVirtual) {
		MOVQ(x, d)
		n := GP64()
		MOVQ(operand.U64(5), n)
		BTRQ(n, d)
	})
	un("BtcQ5", func(x, d reg.GPVirtual) {
		MOVQ(x, d)
		n := GP64()
		MOVQ(operand.U64(5), n)
		BTCQ(n, d)
	})
	un("PopcntQ", func(x, d reg.GPVirtual) { POPCNTQ(x, d) })
	un("SarQ3", func(x, d reg.GPVirtual) { MOVQ(x, d); SARQ(operand.U8(3), d) })
	un("SarL3", func(x, d reg.GPVirtual) { MOVQ(x, d); SARL(operand.U8(3), d.As32()) })
	un("IncL", func(x, d reg.GPVirtual) { MOVQ(x, d); INCL(d.As32()) })
	un("ShlB2", func(x, d reg.GPVirtual) { MOVQ(x, d); SHLB(operand.U8(2), d.As8()) })
	un("BsfQ", func(x, d reg.GPVirtual) { BSFQ(x, d) })
	un("TzcntQ", func(x, d reg.GPVirtual) { TZCNTQ(x, d) })

	bin := func(name string, emit func(x, y, dst reg.GPVirtual)) {
		TEXT(name, NOSPLIT, "func(x, y uint64) uint64")
		x, y, dst := GP64(), GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("y"), y)
		emit(x, y, dst)
		Store(dst, ReturnIndex(0))
		RET()
	}
	bin("AddL", func(x, y, d reg.GPVirtual) { MOVQ(x, d); ADDL(y.As32(), d.As32()) })
	bin("SubL", func(x, y, d reg.GPVirtual) { MOVQ(x, d); SUBL(y.As32(), d.As32()) })
	bin("AndL", func(x, y, d reg.GPVirtual) { MOVQ(x, d); ANDL(y.As32(), d.As32()) })
	bin("OrL", func(x, y, d reg.GPVirtual) { MOVQ(x, d); ORL(y.As32(), d.As32()) })
	bin("ImulL", func(x, y, d reg.GPVirtual) { MOVQ(x, d); IMULL(y.As32(), d.As32()) })
	bin("XchgQ", func(x, y, d reg.GPVirtual) {
		a, b := GP64(), GP64()
		MOVQ(x, a)
		MOVQ(y, b)
		XCHGQ(a, b)
		// Return a after the swap, which must be y.
		MOVQ(a, d)
	})

	// LeaL: 32-bit address arithmetic, truncated and zero-extended.
	TEXT("LeaL", NOSPLIT, "func(x, y uint64) uint64")
	{
		x, y, d := GP64(), GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("y"), y)
		LEAL(operand.Mem{Base: x, Index: y, Scale: 4, Disp: 7}, d.As32())
		Store(d, ReturnIndex(0))
		RET()
	}

	// VecXor: MOVOU load, PXOR against itself to zero, MOVOU store. Verifies the
	// 128-bit move and vector-xor lowerings end to end.
	TEXT("VecZero", NOSPLIT, "func(dst *[16]byte)")
	{
		dst := GP64()
		Load(Param("dst"), dst)
		z := XMM()
		PXOR(z, z)
		MOVOU(z, operand.Mem{Base: dst})
		RET()
	}

	// VecCopyIdx: 128-bit move through an indexed operand, the shape zstd's
	// match-copy loop emits. arm64 has no register-offset form for these, so the
	// address must be materialized.
	TEXT("VecCopyIdx", NOSPLIT, "func(dst, src *[32]byte, i uint64)")
	{
		dst, src, i := GP64(), GP64(), GP64()
		Load(Param("dst"), dst)
		Load(Param("src"), src)
		Load(Param("i"), i)
		v := XMM()
		MOVOU(operand.Mem{Base: src, Index: i, Scale: 1}, v)
		MOVOU(v, operand.Mem{Base: dst, Index: i, Scale: 1})
		RET()
	}

	// VecCopy: MOVOU load + MOVOU store with a displacement on each side.
	TEXT("VecCopy", NOSPLIT, "func(dst, src *[32]byte)")
	{
		dst, src := GP64(), GP64()
		Load(Param("dst"), dst)
		Load(Param("src"), src)
		v := XMM()
		MOVOU(operand.Mem{Base: src, Disp: 16}, v)
		MOVOU(v, operand.Mem{Base: dst, Disp: 16})
		RET()
	}

	// Randomized programs. Each chains operations from propspec, whose Go
	// references the test replays over the same sequence. This is what catches
	// the interaction bugs a hand-written case has to be thought of first: an
	// operand width that only matters after a particular predecessor, or a flag
	// producer separated from its consumer.
	for n := 0; n < propspec.NumPrograms; n++ {
		TEXT(fmt.Sprintf("Prop%d", n), NOSPLIT, "func(x, y uint64) uint64")
		acc, y := GP64(), GP64()
		Load(Param("x"), acc)
		Load(Param("y"), y)
		for _, op := range propspec.Program(n, propspec.ProgramLength) {
			propspec.Ops[op].Emit(acc, y)
		}
		Store(acc, ReturnIndex(0))
		RET()
	}

	// The memory-operand family. Each program interleaves loads, stores and
	// register work over one array, so the address rendering, the scratch
	// staging and the address cache are exercised in combination rather than one
	// access at a time.
	for n := 0; n < propspec.NumMemPrograms; n++ {
		TEXT(fmt.Sprintf("MemProp%d", n), NOSPLIT, "func(x, y uint64, p *[8]uint64) uint64")
		acc, y, p, idx := GP64(), GP64(), GP64(), GP64()
		Load(Param("x"), acc)
		Load(Param("y"), y)
		Load(Param("p"), p)
		// A scaled index needs to stay inside the array; masking it here keeps
		// every generated access in bounds for any input.
		MOVQ(y, idx)
		ANDQ(operand.U32(3), idx)
		for _, op := range propspec.MemProgram(n, propspec.MemProgramLength) {
			propspec.MemOps[op].Emit(acc, y, p, idx)
		}
		Store(acc, ReturnIndex(0))
		RET()
	}

	// BtBranch/BtBranchClear: BTL followed immediately by a carry branch,
	// which lowers to a single arm64 TBNZ or TBZ rather than to anything that
	// materializes CF. Setting CF and letting the generic branch path run would
	// invert the test -- armCond maps the carry conditions by their meaning
	// after a compare, where the two architectures use opposite borrow senses,
	// and BT's CF is a raw bit with no borrow about it. A lowering that made
	// that mistake returns exactly the wrong branch's value on every input, so
	// these two pin the polarity in both directions.
	//
	// Bit 3 rather than bit 0 so the immediate is carried through rather than
	// happening to work as a zero test.
	btBranch := func(name string, bit uint64, jmp func(operand.Op)) {
		TEXT(name, NOSPLIT, "func(x uint64) uint64")
		x := GP64()
		Load(Param("x"), x)
		res := GP64()
		BTL(operand.U8(uint8(bit)), x.As32())
		jmp(operand.LabelRef(name + "_taken"))
		MOVQ(operand.U64(20), res)
		JMP(operand.LabelRef(name + "_end"))
		Label(name + "_taken")
		MOVQ(operand.U64(10), res)
		Label(name + "_end")
		Store(res, ReturnIndex(0))
		RET()
	}
	btBranch("BtBranch", 0, JC)       // branch when the bit is set
	btBranch("BtBranchHigh", 3, JC)   // ... with a non-zero bit index
	btBranch("BtBranchClear", 0, JNC) // branch when the bit is clear
	btBranch("BtBranchClearHigh", 3, JNC)

	// ---- regression cases for lowering bugs found in review ----

	// MovwzxL: the 32-bit-destination form of a halfword zero-extend, loaded
	// through a scaled index -- the shape minlz's block encoders use to read
	// their hash tables. It differs from MOVWQZX only in the named width of the
	// destination, not in the value produced, since x86 zeroes bits 63:32 on any
	// 32-bit register write. The destination is pre-filled with ones so that a
	// lowering which failed to clear the upper half would be caught rather than
	// masked by an already-zero register.
	TEXT("MovwzxL", NOSPLIT, "func(p *[4]uint16, i uint64) uint64")
	{
		p, i, d := GP64(), GP64(), GP64()
		Load(Param("p"), p)
		Load(Param("i"), i)
		MOVQ(operand.U64(^uint64(0)), d)
		MOVWLZX(operand.Mem{Base: p, Index: i, Scale: 2}, d.As32())
		Store(d, ReturnIndex(0))
		RET()
	}

	// MovbzxHigh: a byte-extend from a high-byte source must read bits 15:8. AH
	// renames to the same arm64 register as AL, so a plain byte load silently
	// reads 7:0 instead. The 32-bit destination form is used deliberately: with a
	// 64-bit destination x86-64 needs a REX prefix, and REX redefines that
	// register slot as SPL, so AH is unreachable there.
	TEXT("MovbzxHigh", NOSPLIT, "func(x uint64) uint64")
	{
		x := reg.RAX // high-byte access requires AX-DX
		Load(Param("x"), x)
		// The destination is pinned low on purpose: an extended register would
		// force a REX prefix, and REX renames the high-byte source to SPL, so an
		// allocator-chosen register could silently change what amd64 reads.
		d := reg.RBX
		MOVBLZX(x.As8H(), d.As32())
		Store(d, ReturnIndex(0))
		RET()
	}

	// CmovL32: a 32-bit CMOV writes its destination on BOTH condition outcomes,
	// zero-extending it. Here the condition is false, so a 64-bit CSEL would
	// leave the destination's upper half intact where x86 clears it.
	TEXT("CmovL32", NOSPLIT, "func(x, y uint64) uint64")
	{
		x, y, d := GP64(), GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("y"), y)
		MOVQ(y, d)
		CMPQ(x, x)                  // equal, so NE is false
		CMOVLNE(x.As32(), d.As32()) // no move, but x86 still zero-extends d
		Store(d, ReturnIndex(0))
		RET()
	}

	// BextrMem: BEXTR with a runtime control and an indexed memory source. The
	// control fields and the source address both want a scratch register, and
	// staging them in the wrong order makes the shift count the address.
	TEXT("BextrMem", NOSPLIT, "func(p *[4]uint64, i uint64, ctrl uint64) uint64")
	{
		ptr, i, ctrl, d := GP64(), GP64(), GP64(), GP64()
		Load(Param("p"), ptr)
		Load(Param("i"), i)
		Load(Param("ctrl"), ctrl)
		BEXTRQ(ctrl, operand.Mem{Base: ptr, Index: i, Scale: 8}, d)
		Store(d, ReturnIndex(0))
		RET()
	}

	// XorSelfEq: x86's XOR-self zeroing idiom also sets ZF, and a consumer may
	// read it. The arm64 shortcut replaces the XOR with a move, which sets no
	// flags, so it has to supply them separately.
	TEXT("XorSelfEq", NOSPLIT, "func(x uint64) uint64")
	{
		x, res := GP64(), GP64()
		Load(Param("x"), x)
		XORQ(x, x)
		JEQ(operand.LabelRef("xorself_yes"))
		MOVQ(operand.U64(0), res)
		JMP(operand.LabelRef("xorself_end"))
		Label("xorself_yes")
		MOVQ(operand.U64(1), res)
		Label("xorself_end")
		Store(res, ReturnIndex(0))
		RET()
	}

	// MovLNegImm: a 32-bit move zero-extends, so a negative immediate must land
	// as its unsigned 32-bit value rather than sign-extended across all 64 bits.
	TEXT("MovLNegImm", NOSPLIT, "func() uint64")
	{
		d := GP64()
		MOVL(operand.I32(-1), d.As32())
		Store(d, ReturnIndex(0))
		RET()
	}

	// ---- regressions from the operand-form review ----
	//
	// Four silent miscompiles, each reachable from an ordinary x86 operand form
	// that the in-tree generators simply never used.

	// IncLMem/DecLMem: a read-modify-write on a 32-bit memory destination. The
	// lowering loaded and stored eight bytes around a 32-bit increment, and
	// since the W-form arithmetic zeroes bits 63:32 of the scratch register, the
	// store cleared the four bytes FOLLOWING the target word. The second array
	// element is the witness: it must come back untouched.
	TEXT("IncLMem", NOSPLIT, "func(p *[2]uint32)")
	{
		p := GP64()
		Load(Param("p"), p)
		INCL(operand.Mem{Base: p})
		RET()
	}
	TEXT("DecLMem", NOSPLIT, "func(p *[2]uint32)")
	{
		p := GP64()
		Load(Param("p"), p)
		DECL(operand.Mem{Base: p})
		RET()
	}
	TEXT("IncQMem", NOSPLIT, "func(p *[2]uint64)")
	{
		p := GP64()
		Load(Param("p"), p)
		INCQ(operand.Mem{Base: p})
		RET()
	}

	// CmpBHigh: a byte compare against a high-byte register. AH renames to the
	// same arm64 register as AL, so the operand has to be extracted from bits
	// 15:8; masking the renamed register compared the low byte instead. The
	// inputs are chosen so the two bytes disagree, and disagree in both
	// directions across the test cases.
	TEXT("CmpBHigh", NOSPLIT, "func(x uint64) uint64")
	{
		x := reg.RAX // high-byte access requires AX-DX
		Load(Param("x"), x)
		d := reg.RBX // pinned low: an extended register would force a REX prefix,
		XORQ(d, d)   // which renames the high-byte source to SPL
		CMPB(x.As8H(), operand.U8(2))
		SETEQ(d.As8())
		Store(d, ReturnIndex(0))
		RET()
	}

	// MovWLoad: a 16-bit load writes bits 15:0 and leaves the upper 48 alone,
	// exactly like the register and immediate forms. A bare zero-extending load
	// is shorter and was what the lowering emitted.
	TEXT("MovWLoad", NOSPLIT, "func(x uint64, p *uint16) uint64")
	{
		x, p, d := GP64(), GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("p"), p)
		MOVQ(x, d)
		MOVW(operand.Mem{Base: p}, d.As16())
		Store(d, ReturnIndex(0))
		RET()
	}

	// MulXAlias: x86 lets MULX name one register for both destinations and
	// defines the high half as written second, so the high half is what
	// survives. Staging the low half and copying it out unconditionally left the
	// wrong one.
	TEXT("MulXAlias", NOSPLIT, "func(x, y uint64) uint64")
	{
		Load(Param("x"), reg.RSI)
		Load(Param("y"), reg.RDX)        // implicit MULX factor
		MULXQ(reg.RSI, reg.RAX, reg.RAX) // both destinations alias
		Store(reg.RAX, ReturnIndex(0))
		RET()
	}

	// TestQImm: the canonical immediate-test form, which x86 writes with the
	// immediate first. The lowering only accepted an immediate in the second
	// position, a place x86 never puts it.
	bin("TestQImm", func(x, y, d reg.GPVirtual) {
		XORQ(d, d)
		TESTQ(operand.U32(0x10), x)
		SETEQ(d.As8())
	})

	// FlagsAcrossTransparent: instructions that preserve EFLAGS on x86 must not
	// break the link between a compare and the branch that reads it. These four
	// used to abort generation; the risk in allowing them is the opposite one,
	// so the branch outcome is checked rather than just that it compiles.
	TEXT("FlagsAcrossTransparent", NOSPLIT, "func(a, b uint64) uint64")
	{
		a, b, res, t := GP64(), GP64(), GP64(), GP64()
		Load(Param("a"), a)
		Load(Param("b"), b)
		CMPQ(a, b)
		MOVQ(a, t)
		NOTQ(t) // flag-preserving on x86
		BSWAPL(t.As32())
		XCHGQ(t, res)
		JEQ(operand.LabelRef("fat_eq"))
		MOVQ(operand.U64(0), res)
		JMP(operand.LabelRef("fat_end"))
		Label("fat_eq")
		MOVQ(operand.U64(1), res)
		Label("fat_end")
		Store(res, ReturnIndex(0))
		RET()
	}

	// ---- regressions from the flags/control-flow review ----

	// XorlSign: XORL is a 32-bit operation, so a sign test after it must read
	// bit 31. Lowering it at 64-bit width synthesized a 64-bit TST, whose N bit
	// reads bit 63 -- always zero after a zero-extending EORW -- so the branch
	// was never taken. The inputs make bit 31 of x^y the deciding bit.
	TEXT("XorlSign", NOSPLIT, "func(x, y uint64) uint64")
	{
		x, y, res := GP64(), GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("y"), y)
		XORL(y.As32(), x.As32())
		JMI(operand.LabelRef("xorlsign_neg"))
		MOVQ(operand.U64(0), res)
		JMP(operand.LabelRef("xorlsign_end"))
		Label("xorlsign_neg")
		MOVQ(operand.U64(1), res)
		Label("xorlsign_end")
		Store(res, ReturnIndex(0))
		RET()
	}

	// XorlMem: the same width bug on the memory path, where it corrupted the
	// adjacent word rather than a branch. Second element is the witness.
	TEXT("XorlMem", NOSPLIT, "func(p *[2]uint32, v uint64)")
	{
		p, v := GP64(), GP64()
		Load(Param("p"), p)
		Load(Param("v"), v)
		XORL(v.As32(), operand.Mem{Base: p})
		RET()
	}

	// BzhiConst256/BzhiConstNeg: x86 reads BZHI's bit count from the low 8 bits
	// of the control operand and ignores the rest, so the whole constant is the
	// wrong thing to classify. $256 means zero bits (clear), not "64 or more"
	// (copy); $-1 means 255 (copy), not a negative count (clear). Both get the
	// answer exactly backwards on every nonzero input.
	un("BzhiConst256", func(x, d reg.GPVirtual) {
		n := GP64()
		MOVQ(operand.U64(256), n)
		BZHIQ(n, x, d)
	})
	un("BzhiConstNeg", func(x, d reg.GPVirtual) {
		n := GP64()
		MOVQ(operand.I32(-1), n)
		BZHIQ(n, x, d)
	})

	// ---- 64-bit immediates with bit 31 set ----
	//
	// x86-64 has no 64-bit immediate for these forms: it encodes imm32 and the
	// CPU sign-extends. So $0x80000000 in a Q-width slot means
	// 0xffffffff80000000, and Go's assembler accepts the unsigned spelling
	// without complaint. Every case below is chosen so the two readings give
	// different answers.

	un("AndQBit31", func(x, d reg.GPVirtual) { MOVQ(x, d); ANDQ(operand.U32(0x80000000), d) })
	un("AddQBit31", func(x, d reg.GPVirtual) { MOVQ(x, d); ADDQ(operand.U32(0x80000000), d) })
	un("OrQBit31", func(x, d reg.GPVirtual) { MOVQ(x, d); ORQ(operand.U32(0xffffffe0), d) })

	// A compare against such an immediate, read through a branch.
	TEXT("CmpQBit31", NOSPLIT, "func(x uint64) uint64")
	{
		x, res := GP64(), GP64()
		Load(Param("x"), x)
		CMPQ(x, operand.U32(0x80000000))
		JEQ(operand.LabelRef("cqb_eq"))
		MOVQ(operand.U64(0), res)
		JMP(operand.LabelRef("cqb_end"))
		Label("cqb_eq")
		MOVQ(operand.U64(1), res)
		Label("cqb_end")
		Store(res, ReturnIndex(0))
		RET()
	}

	// TESTQ's mask likewise covers bits 63:31, not just bit 31.
	un("TestQBit31", func(x, d reg.GPVirtual) {
		XORQ(d, d)
		TESTQ(operand.U32(0x80000000), x)
		SETEQ(d.As8())
	})

	// A 64-bit store of such an immediate writes the sign-extended value.
	TEXT("MovQBit31Mem", NOSPLIT, "func(p *uint64)")
	{
		p := GP64()
		Load(Param("p"), p)
		MOVQ(operand.U32(0x80000000), operand.Mem{Base: p})
		RET()
	}

	// Control: MOVQ of the same immediate INTO A REGISTER is narrowed by the
	// assembler to a zero-extending MOVL, so here the literal reading is right.
	// This must keep agreeing -- it is what stops the fix from over-applying.
	TEXT("MovQBit31Reg", NOSPLIT, "func() uint64")
	{
		d := GP64()
		MOVQ(operand.U32(0x80000000), d)
		Store(d, ReturnIndex(0))
		RET()
	}

	// CmpLIntMin: "CMP a, -b" lowers to "CMN a, b" because subtracting a
	// negative adds its magnitude -- but at the signed minimum the negated
	// magnitude is not representable as a positive at that width, so arm64 would
	// add -2^31 where x86 subtracts it. N, Z and C survive; V is computed from a
	// different true value and inverts, flipping every signed condition. The
	// results are packed so all four signed conditions are checked at once.
	TEXT("CmpLIntMin", NOSPLIT, "func(x uint64) uint64")
	{
		x, acc := GP64(), GP64()
		Load(Param("x"), x)
		XORQ(acc, acc)
		for i, set := range []func(operand.Op){SETLT, SETGE, SETGT, SETLE} {
			t := GP64()
			XORQ(t, t)
			CMPL(x.As32(), operand.I32(-2147483648))
			set(t.As8())
			if i > 0 {
				SHLQ(operand.U8(i), t)
			}
			ORQ(t, acc)
		}
		Store(acc, ReturnIndex(0))
		RET()
	}

	// x86 masks shift counts to the low 6 bits (64-bit) or 5 (32-bit), so a
	// count at the width is a no-op and one past it shifts by the remainder.
	// arm64's immediate shift form rejects such counts outright, so the lowering
	// has to apply the same mask rather than pass the count through.
	un("ShrQ64", func(x, d reg.GPVirtual) { MOVQ(x, d); SHRQ(operand.U8(64), d) })
	un("ShrQ65", func(x, d reg.GPVirtual) { MOVQ(x, d); SHRQ(operand.U8(65), d) })
	un("ShlQ64", func(x, d reg.GPVirtual) { MOVQ(x, d); SHLQ(operand.U8(64), d) })
	un("ShrL32", func(x, d reg.GPVirtual) { MOVQ(x, d); SHRL(operand.U8(32), d.As32()) })

	// ---- filling out the dispatch table ----
	//
	// Everything below exists because TestOpcodeDispatchIsCovered found it
	// reachable in the printer's dispatch switch but absent from the generated
	// amd64 assembly -- lowered, but never executed against a reference. The
	// in-tree generators exercise about two thirds of the switch; the rest was
	// only as good as the review that wrote it.

	un("IncQ", func(x, d reg.GPVirtual) { MOVQ(x, d); INCQ(d) })
	un("DecL", func(x, d reg.GPVirtual) { MOVQ(x, d); DECL(d.As32()) })
	un("ShrL5", func(x, d reg.GPVirtual) { MOVQ(x, d); SHRL(operand.U8(5), d.As32()) })
	un("SxWQ", func(x, d reg.GPVirtual) { MOVWQSX(x.As16(), d) })
	// BSR's result is architecturally undefined for a zero input, so the test
	// feeds it non-zero values only.
	un("BsrQ", func(x, d reg.GPVirtual) { BSRQ(x, d) })
	un("BtsQ5", func(x, d reg.GPVirtual) {
		MOVQ(x, d)
		n := GP64()
		MOVQ(operand.U64(5), n)
		BTSQ(n, d)
	})

	// MOVLQZX has no register-to-register encoding on x86 -- zero-extending a
	// register is what a plain MOVL does -- so it only ever appears as a load,
	// and the lowering must use a 4-byte one.
	TEXT("ZxLQ", NOSPLIT, "func(p *uint64) uint64")
	{
		p, d := GP64(), GP64()
		Load(Param("p"), p)
		MOVLQZX(operand.Mem{Base: p}, d)
		Store(d, ReturnIndex(0))
		RET()
	}

	// MOVW writes only the low 16 bits and leaves the rest of the destination
	// alone -- the same partial-register shape that produced silent miscompiles
	// at byte width.
	bin("MovW", func(x, y, d reg.GPVirtual) { MOVQ(x, d); MOVW(y.As16(), d.As16()) })

	// LEAQ with a full base+index*scale+displacement operand. Only the 32-bit
	// form had coverage, and the two take different paths through the address
	// rendering.
	bin("LeaQ", func(x, y, d reg.GPVirtual) {
		LEAQ(operand.Mem{Base: x, Index: y, Scale: 8, Disp: -24}, d)
	})

	// TESTQ/TESTB discard the AND and keep only its flags. TESTB is a sub-word
	// test, which the lowering accepts only when every consumer is EQ/NE.
	bin("TestQZero", func(x, y, d reg.GPVirtual) {
		XORQ(d, d)
		TESTQ(x, y)
		SETEQ(d.As8())
	})
	bin("TestBZero", func(x, y, d reg.GPVirtual) {
		XORQ(d, d)
		TESTB(x.As8(), y.As8())
		SETEQ(d.As8())
	})

	// TESTB/TESTW dst, dst -- the same register on both sides, x86's "is dst
	// zero?" idiom -- is a special case inside lowerSubwordTestEqNe: it needs
	// no zero-extended scratch copy, only an immediate-masked TST against the
	// original register (see shiftExtractFold's sibling change in the same
	// commit). Covers both widths.
	un("TestBSelfZero", func(x, d reg.GPVirtual) {
		XORQ(d, d)
		TESTB(x.As8(), x.As8())
		SETEQ(d.As8())
	})
	un("TestWSelfZero", func(x, d reg.GPVirtual) {
		XORQ(d, d)
		TESTW(x.As16(), x.As16())
		SETEQ(d.As8())
	})

	// MOVOA is the aligned 128-bit move. Only the register-to-register form is
	// generated: the memory form faults on x86 unless the address is 16-byte
	// aligned, which Go does not guarantee for a heap [16]byte.
	TEXT("VecCopyReg", NOSPLIT, "func(dst, src *[16]byte)")
	{
		dst, src := GP64(), GP64()
		Load(Param("dst"), dst)
		Load(Param("src"), src)
		a, b := XMM(), XMM()
		MOVOU(operand.Mem{Base: src}, a)
		MOVOA(a, b)
		MOVOU(b, operand.Mem{Base: dst})
		RET()
	}

	// ---- condition-code coverage ----
	//
	// One function per consumer family evaluates every condition in condspec
	// after the same compare and packs the results into a bitmask, so the three
	// families cover the whole translation table with three symbols and no
	// per-condition glue. Before this, most conditions had never been executed
	// in some family, and the carry conditions -- where the two architectures
	// disagree most -- were the thinnest covered of all.

	TEXT("SetAll", NOSPLIT, "func(a, b uint64) uint64")
	{
		a, b, acc := GP64(), GP64(), GP64()
		Load(Param("a"), a)
		Load(Param("b"), b)
		XORQ(acc, acc)
		for i, c := range condspec.Conds {
			t := GP64()
			XORQ(t, t) // SETcc writes only the low byte
			CMPQ(a, b)
			setcc(c.Name, t)
			if i > 0 {
				SHLQ(operand.U8(i), t)
			}
			ORQ(t, acc)
		}
		Store(acc, ReturnIndex(0))
		RET()
	}

	TEXT("JmpAll", NOSPLIT, "func(a, b uint64) uint64")
	{
		a, b, acc := GP64(), GP64(), GP64()
		Load(Param("a"), a)
		Load(Param("b"), b)
		XORQ(acc, acc)
		for i, c := range condspec.Conds {
			taken := fmt.Sprintf("jmpall_%d_taken", i)
			done := fmt.Sprintf("jmpall_%d_done", i)
			CMPQ(a, b)
			jcc(c.Name, operand.LabelRef(taken))
			JMP(operand.LabelRef(done))
			Label(taken)
			ORQ(operand.U32(1<<uint(i)), acc)
			Label(done)
		}
		Store(acc, ReturnIndex(0))
		RET()
	}

	TEXT("CmovAll", NOSPLIT, "func(a, b uint64) uint64")
	{
		a, b, acc := GP64(), GP64(), GP64()
		Load(Param("a"), a)
		Load(Param("b"), b)
		XORQ(acc, acc)
		for i, c := range condspec.Conds {
			bit, d := GP64(), GP64()
			MOVQ(operand.U64(1<<uint(i)), bit)
			XORQ(d, d)
			CMPQ(a, b)
			cmovcc(c.Name, bit, d)
			ORQ(d, acc)
		}
		Store(acc, ReturnIndex(0))
		RET()
	}

	// The copy folds the arm64 printer applies to x86's shift idioms (see
	// shiftFolds in printer/arm64.go). x86 needs the variable count in CL and
	// shifts in place, so avo programs copy the count and the source before
	// every shift; arm64 reads both from wherever they are. Each function
	// keeps the original registers live past the shift so the allocator
	// cannot coalesce the copies away before the printer sees them (see the
	// shiftExtractByte comment above), and follows the shift with a fresh
	// definition of RCX, the shape the fold requires.
	TEXT("ShlCountFold", NOSPLIT, "func(x, n uint64) uint64")
	{
		x, n, y := GP64(), GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("n"), n)
		MOVQ(n, reg.RCX)
		MOVQ(x, y)
		SHLQ(reg.CL, y)
		MOVQ(x, reg.RCX) // RCX redefined without being read: the count copy is dead
		ADDQ(reg.RCX, y)
		ADDQ(n, y)
		Store(y, ReturnIndex(0))
		RET()
	}
	TEXT("ShrCountFold32", NOSPLIT, "func(x, n uint64) uint64")
	{
		x, n, y := GP64(), GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("n"), n)
		MOVQ(n, reg.RCX)
		MOVQ(x, y)
		SHRL(reg.CL, y.As32())
		MOVQ(x, reg.RCX)
		ADDQ(reg.RCX, y)
		ADDQ(n, y)
		Store(y, ReturnIndex(0))
		RET()
	}
	TEXT("RolCountFold", NOSPLIT, "func(x, n uint64) uint64")
	{
		x, n, y := GP64(), GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("n"), n)
		MOVQ(n, reg.RCX)
		MOVQ(x, y)
		ROLQ(reg.CL, y)
		MOVQ(x, reg.RCX)
		ADDQ(reg.RCX, y)
		ADDQ(n, y)
		Store(y, ReturnIndex(0))
		RET()
	}
	// CountFoldRefusedRead reads the count copy after the shift, so the copy
	// must survive; the lowering has to stay correct with the fold refused.
	TEXT("CountFoldRefusedRead", NOSPLIT, "func(x, n uint64) uint64")
	{
		x, n, y := GP64(), GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("n"), n)
		MOVQ(n, reg.RCX)
		MOVQ(x, y)
		SHLQ(reg.CL, y)
		ADDQ(reg.RCX, y)
		ADDQ(n, y)
		Store(y, ReturnIndex(0))
		RET()
	}
	// ShlCountSelf shifts RCX by CL: the copied value is both count and data,
	// which the fold refuses.
	TEXT("ShlCountSelf", NOSPLIT, "func(n uint64) uint64")
	{
		n := GP64()
		Load(Param("n"), n)
		MOVQ(n, reg.RCX)
		SHLQ(reg.CL, reg.RCX)
		ADDQ(n, reg.RCX)
		Store(reg.RCX, ReturnIndex(0))
		RET()
	}
	// SarCopyFold: only the source copy is redundant (immediate count).
	TEXT("SarCopyFold", NOSPLIT, "func(x uint64) uint64")
	{
		x, y := GP64(), GP64()
		Load(Param("x"), x)
		MOVQ(x, y)
		SARQ(operand.U8(3), y)
		ADDQ(x, y)
		Store(y, ReturnIndex(0))
		RET()
	}
	// ShlCopyFoldCX: the copy's destination is RCX, but with an immediate
	// count nothing aliases it, so the source fold still applies.
	TEXT("ShlCopyFoldCX", NOSPLIT, "func(x uint64) uint64")
	{
		x := GP64()
		Load(Param("x"), x)
		MOVQ(x, reg.RCX)
		SHLQ(operand.U8(2), reg.RCX)
		ADDQ(x, reg.RCX)
		Store(reg.RCX, ReturnIndex(0))
		RET()
	}

	// AdcAccumQ: the full-width form of AdcAccum (ADCQ $0), the shape the
	// huff0 4X decoders now use for their exhausted counter.
	TEXT("AdcAccumQ", NOSPLIT, "func(x, acc uint64) uint64")
	{
		x, acc := GP64(), GP64()
		Load(Param("x"), x)
		Load(Param("acc"), acc)
		CMPQ(x, operand.Imm(4))
		ADCQ(operand.I8(0), acc)
		Store(acc, ReturnIndex(0))
		RET()
	}
	// SetGeZeroMov: SetGe with the destination zeroed by a move rather than an
	// XOR, the other zeroing setccFolds recognizes.
	TEXT("SetGeZeroMov", NOSPLIT, "func(a, b uint64) uint64")
	{
		a, b, dst := GP64(), GP64(), GP64()
		Load(Param("a"), a)
		Load(Param("b"), b)
		MOVQ(operand.U32(0), dst)
		CMPQ(a, b)
		SETGE(dst.As8())
		Store(dst, ReturnIndex(0))
		RET()
	}
	// SetGeRefusedRead reads the zeroed register before the SETcc, so the
	// zeroing must survive and the byte insert must be exact.
	TEXT("SetGeRefusedRead", NOSPLIT, "func(a, b uint64) uint64")
	{
		a, b, dst := GP64(), GP64(), GP64()
		Load(Param("a"), a)
		Load(Param("b"), b)
		XORQ(dst, dst)
		ADDQ(dst, a)
		CMPQ(a, b)
		SETGE(dst.As8())
		Store(dst, ReturnIndex(0))
		RET()
	}

	Generate()
}

// setcc, jcc and cmovcc emit one condspec condition in each consumer family.
// They are exhaustive switches rather than table lookups so that adding a
// condition to condspec without teaching all three families about it fails at
// generation time, instead of silently contributing a zero bit that the
// reference would then have to match.

func setcc(name string, dst reg.GPVirtual) {
	d := dst.As8()
	switch name {
	case "EQ":
		SETEQ(d)
	case "NE":
		SETNE(d)
	case "LT":
		SETLT(d)
	case "LE":
		SETLE(d)
	case "GT":
		SETGT(d)
	case "GE":
		SETGE(d)
	case "CS":
		SETCS(d)
	case "CC":
		SETCC(d)
	case "HI":
		SETHI(d)
	case "LS":
		SETLS(d)
	case "MI":
		SETMI(d)
	case "PL":
		SETPL(d)
	case "OS":
		SETOS(d)
	case "OC":
		SETOC(d)
	default:
		panic("no SETcc emitter for condition " + name)
	}
}

func jcc(name string, target operand.Op) {
	switch name {
	case "EQ":
		JEQ(target)
	case "NE":
		JNE(target)
	case "LT":
		JLT(target)
	case "LE":
		JLE(target)
	case "GT":
		JGT(target)
	case "GE":
		JGE(target)
	case "CS":
		JCS(target)
	case "CC":
		JCC(target)
	case "HI":
		JHI(target)
	case "LS":
		JLS(target)
	case "MI":
		JMI(target)
	case "PL":
		JPL(target)
	case "OS":
		JOS(target)
	case "OC":
		JOC(target)
	default:
		panic("no Jcc emitter for condition " + name)
	}
}

func cmovcc(name string, src, dst reg.GPVirtual) {
	switch name {
	case "EQ":
		CMOVQEQ(src, dst)
	case "NE":
		CMOVQNE(src, dst)
	case "LT":
		CMOVQLT(src, dst)
	case "LE":
		CMOVQLE(src, dst)
	case "GT":
		CMOVQGT(src, dst)
	case "GE":
		CMOVQGE(src, dst)
	case "CS":
		CMOVQCS(src, dst)
	case "CC":
		CMOVQCC(src, dst)
	case "HI":
		CMOVQHI(src, dst)
	case "LS":
		CMOVQLS(src, dst)
	case "MI":
		CMOVQMI(src, dst)
	case "PL":
		CMOVQPL(src, dst)
	case "OS":
		CMOVQOS(src, dst)
	case "OC":
		CMOVQOC(src, dst)
	default:
		panic("no CMOVcc emitter for condition " + name)
	}
}
