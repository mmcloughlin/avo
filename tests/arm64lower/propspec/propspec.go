// Package propspec defines the operation vocabulary used to build randomized
// programs for the arm64 lowering's property test.
//
// Each Op pairs an avo emitter with a pure-Go reference that models the exact
// x86 semantics of what the emitter produces, including partial-register and
// flag behaviour. Keeping the two in one struct is deliberate: the emitter and
// its reference are the two halves of a differential test, and separating them
// invites drift.
//
// The generator (asm.go) and the test both derive the same programs from the
// same seeds via Program, so neither needs to record what the other did.
package propspec

import (
	"math/bits"
	"math/rand"

	"github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

// Op is one step of a random program: acc = f(acc, y).
type Op struct {
	Name string
	// Emit appends the operation to the avo program being built, updating acc
	// in place. y is a second live value the operation may read.
	Emit func(acc, y reg.GPVirtual)
	// Ref is the pure-Go model of Emit's x86 semantics.
	Ref func(acc, y uint64) uint64
}

// Ops is the vocabulary. It deliberately over-samples the classes that have
// produced silent miscompiles in this lowering: partial-register writes
// (MOVB/ADDB/SHLB), operand-width truncation (32-bit ALU, MOVL), and
// flag-producer/consumer pairs (CMOVcc, SETcc).
var Ops = []Op{
	{
		Name: "AddQ",
		Emit: func(acc, y reg.GPVirtual) { build.ADDQ(y, acc) },
		Ref:  func(acc, y uint64) uint64 { return acc + y },
	},
	{
		Name: "SubQ",
		Emit: func(acc, y reg.GPVirtual) { build.SUBQ(y, acc) },
		Ref:  func(acc, y uint64) uint64 { return acc - y },
	},
	{
		Name: "AndQ",
		Emit: func(acc, y reg.GPVirtual) { build.ANDQ(y, acc) },
		Ref:  func(acc, y uint64) uint64 { return acc & y },
	},
	{
		Name: "OrQ",
		Emit: func(acc, y reg.GPVirtual) { build.ORQ(y, acc) },
		Ref:  func(acc, y uint64) uint64 { return acc | y },
	},
	{
		Name: "XorQ",
		Emit: func(acc, y reg.GPVirtual) { build.XORQ(y, acc) },
		Ref:  func(acc, y uint64) uint64 { return acc ^ y },
	},
	{
		Name: "ImulQ",
		Emit: func(acc, y reg.GPVirtual) { build.IMULQ(y, acc) },
		Ref:  func(acc, y uint64) uint64 { return acc * y },
	},
	// 32-bit ALU: the result zero-extends into the 64-bit register.
	{
		Name: "AddL",
		Emit: func(acc, y reg.GPVirtual) { build.ADDL(y.As32(), acc.As32()) },
		Ref:  func(acc, y uint64) uint64 { return uint64(uint32(acc) + uint32(y)) },
	},
	{
		Name: "SubL",
		Emit: func(acc, y reg.GPVirtual) { build.SUBL(y.As32(), acc.As32()) },
		Ref:  func(acc, y uint64) uint64 { return uint64(uint32(acc) - uint32(y)) },
	},
	{
		Name: "AndL",
		Emit: func(acc, y reg.GPVirtual) { build.ANDL(y.As32(), acc.As32()) },
		Ref:  func(acc, y uint64) uint64 { return uint64(uint32(acc) & uint32(y)) },
	},
	{
		Name: "ImulL",
		Emit: func(acc, y reg.GPVirtual) { build.IMULL(y.As32(), acc.As32()) },
		Ref:  func(acc, y uint64) uint64 { return uint64(uint32(acc) * uint32(y)) },
	},
	// Shifts and rotates, immediate forms.
	{
		Name: "ShlQ3",
		Emit: func(acc, y reg.GPVirtual) { build.SHLQ(operand.U8(3), acc) },
		Ref:  func(acc, y uint64) uint64 { return acc << 3 },
	},
	{
		Name: "ShrQ7",
		Emit: func(acc, y reg.GPVirtual) { build.SHRQ(operand.U8(7), acc) },
		Ref:  func(acc, y uint64) uint64 { return acc >> 7 },
	},
	{
		Name: "SarQ5",
		Emit: func(acc, y reg.GPVirtual) { build.SARQ(operand.U8(5), acc) },
		Ref:  func(acc, y uint64) uint64 { return uint64(int64(acc) >> 5) },
	},
	{
		Name: "ShlL9",
		Emit: func(acc, y reg.GPVirtual) { build.SHLL(operand.U8(9), acc.As32()) },
		Ref:  func(acc, y uint64) uint64 { return uint64(uint32(acc) << 9) },
	},
	{
		Name: "RolQ13",
		Emit: func(acc, y reg.GPVirtual) { build.ROLQ(operand.U8(13), acc) },
		Ref:  func(acc, y uint64) uint64 { return bits.RotateLeft64(acc, 13) },
	},
	{
		Name: "RolL5",
		Emit: func(acc, y reg.GPVirtual) { build.ROLL(operand.U8(5), acc.As32()) },
		Ref:  func(acc, y uint64) uint64 { return uint64(bits.RotateLeft32(uint32(acc), 5)) },
	},
	// Unary.
	{
		Name: "NotQ",
		Emit: func(acc, y reg.GPVirtual) { build.NOTQ(acc) },
		Ref:  func(acc, y uint64) uint64 { return ^acc },
	},
	{
		Name: "NegQ",
		Emit: func(acc, y reg.GPVirtual) { build.NEGQ(acc) },
		Ref:  func(acc, y uint64) uint64 { return -acc },
	},
	{
		Name: "PopcntQ",
		Emit: func(acc, y reg.GPVirtual) { build.POPCNTQ(acc, acc) },
		Ref:  func(acc, y uint64) uint64 { return uint64(bits.OnesCount64(acc)) },
	},
	{
		Name: "TzcntQ",
		Emit: func(acc, y reg.GPVirtual) { build.TZCNTQ(acc, acc) },
		Ref:  func(acc, y uint64) uint64 { return uint64(bits.TrailingZeros64(acc)) },
	},
	{
		Name: "BswapL",
		Emit: func(acc, y reg.GPVirtual) { build.BSWAPL(acc.As32()) },
		Ref:  func(acc, y uint64) uint64 { return uint64(bits.ReverseBytes32(uint32(acc))) },
	},
	// Width truncation and extension.
	{
		// Copies y's low half into acc, zero-extending. Deliberately not a
		// self-move: avo elides "MOVL AX, AX" as redundant, even though on x86
		// it clears the upper 32 bits.
		Name: "MovLFromY",
		Emit: func(acc, y reg.GPVirtual) { build.MOVL(y.As32(), acc.As32()) },
		Ref:  func(acc, y uint64) uint64 { return uint64(uint32(y)) },
	},
	{
		Name: "SxL",
		Emit: func(acc, y reg.GPVirtual) { build.MOVLQSX(acc.As32(), acc) },
		Ref:  func(acc, y uint64) uint64 { return uint64(int64(int32(acc))) },
	},
	{
		Name: "ZxW",
		Emit: func(acc, y reg.GPVirtual) { build.MOVWQZX(acc.As16(), acc) },
		Ref:  func(acc, y uint64) uint64 { return uint64(uint16(acc)) },
	},
	// Partial-register writes: only the addressed byte changes.
	{
		Name: "MovBLow",
		Emit: func(acc, y reg.GPVirtual) { build.MOVB(y.As8(), acc.As8()) },
		Ref:  func(acc, y uint64) uint64 { return acc&^0xff | y&0xff },
	},
	{
		Name: "AddB",
		Emit: func(acc, y reg.GPVirtual) { build.ADDB(y.As8(), acc.As8()) },
		Ref:  func(acc, y uint64) uint64 { return acc&^0xff | (acc+y)&0xff },
	},
	{
		Name: "ShlB2",
		Emit: func(acc, y reg.GPVirtual) { build.SHLB(operand.U8(2), acc.As8()) },
		Ref:  func(acc, y uint64) uint64 { return acc&^0xff | (acc&0xff)<<2&0xff },
	},
	// Flag producer/consumer pairs. Two of the three miscompiles found in this
	// lowering lived here, so both CMOVcc and SETcc are represented.
	{
		Name: "CmovBelow",
		Emit: func(acc, y reg.GPVirtual) {
			build.CMPQ(acc, y)
			build.CMOVQCS(y, acc) // acc = acc < y ? y : acc
		},
		Ref: func(acc, y uint64) uint64 {
			if acc < y {
				return y
			}
			return acc
		},
	},
	{
		Name: "CmovEq",
		Emit: func(acc, y reg.GPVirtual) {
			build.CMPQ(acc, y)
			build.CMOVQEQ(y, acc)
		},
		Ref: func(acc, y uint64) uint64 { return acc },
	},
	{
		Name: "SetLess",
		Emit: func(acc, y reg.GPVirtual) {
			build.CMPQ(acc, y)
			build.SETLT(acc.As8()) // signed <, written into acc's low byte
		},
		Ref: func(acc, y uint64) uint64 {
			v := uint64(0)
			if int64(acc) < int64(y) {
				v = 1
			}
			return acc&^0xff | v
		},
	},
	{
		Name: "XorSelfZero",
		Emit: func(acc, y reg.GPVirtual) { build.XORQ(acc, acc) },
		Ref:  func(acc, y uint64) uint64 { return 0 },
	},
}

// MemOp is one step of a random program that touches memory: acc = f(acc, y,
// mem), possibly writing to mem.
//
// Register-only programs leave a whole axis of the lowering untested. A memory
// operand is not just another source: it is staged through a scratch register,
// its address may be folded into the instruction or materialized with an ADD,
// the result of that materialization is cached across instructions, and the
// load must be issued at the operand's width rather than the register's. Each
// of those has its own failure mode, and three of the bugs found by review so
// far lived in exactly that machinery.
type MemOp struct {
	Name string
	// Emit appends the operation. p holds the base address of an 8-element
	// uint64 array and idx holds an index already masked to 0..3, so any
	// scaled-index operand stays in bounds.
	Emit func(acc, y, p, idx reg.GPVirtual)
	// Ref models Emit's x86 semantics, including any write to mem or to idx.
	// idx is a pointer because an operation may advance it: an index register
	// that never changes would let a stale cached address go unnoticed.
	Ref func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64
}

// MemOps mixes register and memory operands. Displacements, scaled indices and
// sub-word widths are all represented, since the address rendering treats them
// differently: a folded base+index form is only encodable on arm64 under
// constraints a displacement breaks, and a load wider than its operand can
// fault at a page boundary even when the value it produces is right.
var MemOps = []MemOp{
	{
		Name: "AddQMemIdx",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.ADDQ(operand.Mem{Base: p, Index: idx, Scale: 8}, acc)
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 { return acc + mem[*idx] },
	},
	{
		Name: "XorQMemIdx",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.XORQ(operand.Mem{Base: p, Index: idx, Scale: 8}, acc)
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 { return acc ^ mem[*idx] },
	},
	{
		Name: "ImulQMemDisp",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.IMULQ(operand.Mem{Base: p, Disp: 16}, acc)
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 { return acc * mem[2] },
	},
	{
		// A 32-bit operand four bytes into the array: the load must take four
		// bytes, and the result must zero-extend.
		Name: "MovLMemDisp",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.MOVL(operand.Mem{Base: p, Disp: 4}, acc.As32())
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 { return uint64(uint32(mem[0] >> 32)) },
	},
	{
		Name: "AddLMemIdx",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.ADDL(operand.Mem{Base: p, Index: idx, Scale: 8}, acc.As32())
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 {
			return uint64(uint32(acc) + uint32(mem[*idx]))
		},
	},
	{
		// A byte load into the low byte of acc, leaving the rest alone.
		Name: "MovBMemDisp",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.MOVB(operand.Mem{Base: p, Disp: 1}, acc.As8())
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 {
			return acc&^0xff | mem[0]>>8&0xff
		},
	},
	{
		Name: "MovWZXMemDisp",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.MOVWQZX(operand.Mem{Base: p, Disp: 2}, acc)
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 { return mem[0] >> 16 & 0xffff },
	},
	{
		Name: "MovLSXMemDisp",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.MOVLQSX(operand.Mem{Base: p, Disp: 8}, acc)
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 {
			return uint64(int64(int32(uint32(mem[1]))))
		},
	},
	{
		// A compare against a memory operand feeding a conditional select: the
		// flag producer and its consumer are separated by the load the lowering
		// has to insert.
		Name: "CmpMemSelect",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.CMPQ(acc, operand.Mem{Base: p, Index: idx, Scale: 8})
			build.CMOVQCS(y, acc)
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 {
			if acc < mem[*idx] {
				return y
			}
			return acc
		},
	},
	{
		// Address arithmetic whose result is compared against the base, so the
		// answer does not depend on where the array happens to live.
		Name: "LeaIdxOffset",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			t := build.GP64()
			build.LEAQ(operand.Mem{Base: p, Index: idx, Scale: 8, Disp: 8}, t)
			build.SUBQ(p, t)
			build.MOVQ(t, acc)
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 { return *idx*8 + 8 },
	},
	{
		// Stores: the destination-is-memory path, at two widths. A 32-bit store
		// must leave the upper half of the destination word untouched.
		Name: "StoreQIdx",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.MOVQ(acc, operand.Mem{Base: p, Index: idx, Scale: 8})
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 {
			mem[*idx] = acc
			return acc
		},
	},
	{
		Name: "StoreLDisp",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.MOVL(acc.As32(), operand.Mem{Base: p, Disp: 24})
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 {
			mem[3] = mem[3]&0xffffffff00000000 | uint64(uint32(acc))
			return acc
		},
	},
	{
		// Two accesses to the same base and index in a row, which is what the
		// address cache is for; the store in between must not let a stale cached
		// address survive a write to the registers it was built from.
		Name: "ReadModifyWriteIdx",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			m := operand.Mem{Base: p, Index: idx, Scale: 8}
			build.ADDQ(m, acc)
			build.MOVQ(acc, m)
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 {
			acc += mem[*idx]
			mem[*idx] = acc
			return acc
		},
	},
	{
		// Two accesses through the same base and index registers with a write to
		// the index between them.
		//
		// arm64 can fold a base+index operand into the instruction only when the
		// scale matches the access width and there is no displacement; anything
		// else has to be materialized into a scratch register, and that computed
		// address is cached so a second access can reuse it. This operation
		// deliberately takes the materializing path twice -- a 4-byte access at
		// scale 8 -- around a write to the index. If the write does not
		// invalidate the cache, the second access silently reads the first
		// address. Nothing else in this vocabulary reaches that path: with a
		// matching scale everything folds, and the cache is never consulted.
		Name: "StaleIdxAccess",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			m := operand.Mem{Base: p, Index: idx, Scale: 8}
			build.ADDL(m, acc.As32())
			build.INCQ(idx)
			build.ANDQ(operand.U32(3), idx)
			build.ADDL(m, acc.As32())
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 {
			a := uint64(uint32(acc) + uint32(mem[*idx]))
			*idx = (*idx + 1) & 3
			return uint64(uint32(a) + uint32(mem[*idx]))
		},
	},
	{
		// A displacement alongside a scaled index, which also cannot be folded.
		Name: "AddQMemIdxDisp",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.ADDQ(operand.Mem{Base: p, Index: idx, Scale: 8, Disp: 8}, acc)
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 {
			return acc + mem[*idx+1]
		},
	},
	{
		// Advances the index register, so the next access to the same base uses
		// a different address. The lowering caches a materialized base+index
		// address to avoid recomputing it; if a write to either register does
		// not invalidate that cache, the following access reads the old address
		// and this is what notices.
		Name: "AdvanceIdx",
		Emit: func(acc, y, p, idx reg.GPVirtual) {
			build.INCQ(idx)
			build.ANDQ(operand.U32(3), idx)
		},
		Ref: func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 {
			*idx = (*idx + 1) & 3
			return acc
		},
	},
	// Register-only operations are kept in the mix so memory accesses are
	// interleaved with the register pressure and flag traffic that surrounds
	// them in real code.
	{
		Name: "AddQReg",
		Emit: func(acc, y, p, idx reg.GPVirtual) { build.ADDQ(y, acc) },
		Ref:  func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 { return acc + y },
	},
	{
		Name: "RolQ11",
		Emit: func(acc, y, p, idx reg.GPVirtual) { build.ROLQ(operand.U8(11), acc) },
		Ref:  func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 { return bits.RotateLeft64(acc, 11) },
	},
	{
		Name: "SubQReg",
		Emit: func(acc, y, p, idx reg.GPVirtual) { build.SUBQ(y, acc) },
		Ref:  func(acc, y uint64, mem *[8]uint64, idx *uint64) uint64 { return acc - y },
	},
}

// Program returns the op indices making up program n. Both the generator and
// the test call this, so the two always agree without recording anything.
func Program(n, length int) []int {
	r := rand.New(rand.NewSource(int64(n) * 7919))
	out := make([]int, length)
	for i := range out {
		out[i] = r.Intn(len(Ops))
	}
	return out
}

// MemProgram is Program for the memory-operand vocabulary. The seed is offset
// so the two families do not generate the same sequences.
func MemProgram(n, length int) []int {
	r := rand.New(rand.NewSource(int64(n)*7919 + 104729))
	out := make([]int, length)
	for i := range out {
		out[i] = r.Intn(len(MemOps))
	}
	return out
}

// NumMemPrograms and MemProgramLength size the memory-operand family.
const (
	NumMemPrograms   = 8
	MemProgramLength = 10
)

// NumPrograms is how many random programs the suite builds.
const NumPrograms = 12

// ProgramLength is how many operations each one chains together. Long enough to
// interleave widths and flag pairs, short enough to point at a culprit when a
// program fails.
const ProgramLength = 12

// SlotOp is one step of a frame-slot program. Four adjacent 8-byte frame
// slots s[0..3] (one AllocLocal) hold state alongside acc; k picks the slot an
// op works on and lbl is a label prefix unique to this step. The mix decides
// which slots stack-slot promotion may keep in registers: whole-slot MOVQs and
// 64-bit ALU operations keep a slot eligible, and a narrower write disqualifies
// it, so different programs promote different subsets.
type SlotOp struct {
	Name string
	Emit func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string)
	Ref  func(acc, y uint64, s *[4]uint64, k int) uint64
}

// SlotOps is the frame-slot vocabulary.
var SlotOps = []SlotOp{
	{
		Name: "Store",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) { build.MOVQ(acc, s[k]) },
		Ref:  func(acc, y uint64, s *[4]uint64, k int) uint64 { s[k] = acc; return acc },
	},
	{
		Name: "LoadAdd",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t := build.GP64()
			build.MOVQ(s[k], t)
			build.ADDQ(t, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 { return acc + s[k] },
	},
	{
		Name: "Swap",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t := build.GP64()
			build.MOVQ(s[k], t)
			build.MOVQ(acc, s[k])
			build.MOVQ(t, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 { acc, s[k] = s[k], acc; return acc },
	},
	{
		Name: "StoreImm",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			build.MOVQ(operand.U32(0x80000000|uint64(k)), s[k])
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			s[k] = uint64(int64(int32(uint32(0x80000000 | k))))
			return acc
		},
	},
	{
		Name: "RmwAdd",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) { build.ADDQ(acc, s[k]) },
		Ref:  func(acc, y uint64, s *[4]uint64, k int) uint64 { s[k] += acc; return acc },
	},
	{
		Name: "StoreHighL",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			build.MOVL(y.As32(), s[k].Offset(4))
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			s[k] = s[k]&0xffffffff | uint64(uint32(y))<<32
			return acc
		},
	},
	{
		Name: "CondStore",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			build.CMPQ(acc, y)
			build.JCS(operand.LabelRef(lbl + "_skip"))
			build.MOVQ(y, s[k])
			build.Label(lbl + "_skip")
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			if !(acc < y) {
				s[k] = y
			}
			return acc
		},
	},
	{
		Name: "Mix",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			build.ROLQ(operand.U8(7), acc)
			build.XORQ(y, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 { return bits.RotateLeft64(acc, 7) ^ y },
	},
}

// SlotStep is one op of a frame-slot program and the slot it works on.
type SlotStep struct{ Op, Slot int }

// SlotProgram is Program for the frame-slot vocabulary.
func SlotProgram(n, length int) []SlotStep {
	r := rand.New(rand.NewSource(int64(n)*7919 + 15485863))
	out := make([]SlotStep, length)
	for i := range out {
		out[i] = SlotStep{r.Intn(len(SlotOps)), r.Intn(4)}
	}
	return out
}

// NumSlotPrograms and SlotProgramLength size the frame-slot family.
const (
	NumSlotPrograms   = 48
	SlotProgramLength = 14
)

// FoldOps is a frame-slot vocabulary aimed at stack-slot coalescing: every
// slot access is a whole-slot MOVQ or 64-bit ALU operation, so all four slots
// stay promotable, and the
// ops put temporaries in the shapes coalescing must fold (a read-modify-write
// through a temporary, a copy chain, a slot-held loop counter across a back
// edge) and the ones it must refuse (a temporary live across a store to its
// slot, inside a loop too; a value loaded on one path only; two live copies of
// one slot).
var FoldOps = []SlotOp{
	{
		Name: "IncVia",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t := build.GP64()
			build.MOVQ(s[k], t)
			build.ADDQ(y, t)
			build.MOVQ(t, s[k])
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 { s[k] += y; return acc },
	},
	{
		Name: "LoopDec",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			c := build.GP64()
			build.MOVQ(s[k], c)
			build.ANDQ(operand.U32(3), c)
			build.MOVQ(c, s[k])
			build.Label(lbl + "_loop")
			t := build.GP64()
			build.MOVQ(s[k], t)
			build.CMPQ(t, operand.U32(0))
			build.JEQ(operand.LabelRef(lbl + "_done"))
			build.ADDQ(t, acc)
			build.ROLQ(operand.U8(3), acc)
			u := build.GP64()
			build.MOVQ(s[k], u)
			build.SUBQ(operand.U32(1), u)
			build.MOVQ(u, s[k])
			build.JMP(operand.LabelRef(lbl + "_loop"))
			build.Label(lbl + "_done")
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			for s[k] &= 3; s[k] != 0; s[k]-- {
				acc = bits.RotateLeft64(acc+s[k], 3)
			}
			return acc
		},
	},
	{
		Name: "ReadOldAfterStore",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t := build.GP64()
			build.MOVQ(s[k], t)
			build.MOVQ(acc, s[k])
			build.ADDQ(t, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			old := s[k]
			s[k] = acc
			return acc + old
		},
	},
	{
		Name: "LoadTwice",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t, u := build.GP64(), build.GP64()
			build.MOVQ(s[k], t)
			build.MOVQ(s[k], u)
			build.ROLQ(operand.U8(9), u)
			build.ADDQ(t, u)
			build.XORQ(u, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			return acc ^ (bits.RotateLeft64(s[k], 9) + s[k])
		},
	},
	{
		Name: "CopyChain",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t, u := build.GP64(), build.GP64()
			build.MOVQ(s[k], t)
			build.MOVQ(t, u)
			build.ADDQ(y, u)
			build.MOVQ(u, s[k])
			build.ADDQ(t, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			old := s[k]
			s[k] = old + y
			return acc + old
		},
	},
	{
		Name: "CondRmw",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t := build.GP64()
			build.MOVQ(s[k], t)
			build.CMPQ(acc, y)
			build.JCS(operand.LabelRef(lbl + "_skip"))
			build.ADDQ(y, t)
			build.Label(lbl + "_skip")
			build.MOVQ(t, s[k])
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			if !(acc < y) {
				s[k] += y
			}
			return acc
		},
	},
	{
		Name: "CrossCopy",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t := build.GP64()
			build.MOVQ(s[k], t)
			build.MOVQ(t, s[(k+1)%4])
			build.ADDQ(t, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			s[(k+1)%4] = s[k]
			return acc + s[k]
		},
	},
	{
		Name: "CondLoadUse",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t := build.GP64()
			build.MOVQ(y, t)
			build.CMPQ(acc, y)
			build.JCS(operand.LabelRef(lbl + "_skip"))
			build.MOVQ(s[k], t)
			build.Label(lbl + "_skip")
			build.ADDQ(t, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			v := y
			if !(acc < y) {
				v = s[k]
			}
			return acc + v
		},
	},
	{
		Name: "StoreAcc",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) { build.MOVQ(acc, s[k]) },
		Ref:  func(acc, y uint64, s *[4]uint64, k int) uint64 { s[k] = acc; return acc },
	},
	{
		Name: "SwapSlots",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t, u := build.GP64(), build.GP64()
			j := (k + 2) % 4
			build.MOVQ(s[k], t)
			build.MOVQ(s[j], u)
			build.MOVQ(u, s[k])
			build.MOVQ(t, s[j])
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			j := (k + 2) % 4
			s[k], s[j] = s[j], s[k]
			return acc
		},
	},
	{
		Name: "Mix",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			build.ROLQ(operand.U8(7), acc)
			build.XORQ(y, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 { return bits.RotateLeft64(acc, 7) ^ y },
	},
	{
		Name: "LoopCarry",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t, c := build.GP64(), build.GP64()
			build.MOVQ(s[k], t)
			build.MOVQ(operand.U32(2), c)
			build.Label(lbl + "_loop")
			build.MOVQ(acc, s[k])
			build.ADDQ(c, acc)
			build.SUBQ(operand.U32(1), c)
			build.JNE(operand.LabelRef(lbl + "_loop"))
			build.ADDQ(t, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			old := s[k]
			for c := uint64(2); c > 0; c-- {
				s[k] = acc
				acc += c
			}
			return acc + old
		},
	},
	{
		Name: "ImmThenLoad",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			build.MOVQ(operand.U32(0x80000000|uint64(k)), s[k])
			t := build.GP64()
			build.MOVQ(s[k], t)
			build.ADDQ(t, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			s[k] = uint64(int64(int32(uint32(0x80000000 | k))))
			return acc + s[k]
		},
	},
	{
		// The loaded value is live only around the back edge when the slot
		// is stored: forward-only liveness would miss the interference.
		Name: "LoopBackEdgeUse",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t, c := build.GP64(), build.GP64()
			build.MOVQ(s[k], t)
			build.MOVQ(operand.U32(2), c)
			build.Label(lbl + "_loop")
			build.ADDQ(t, acc)
			build.MOVQ(acc, s[k])
			build.SUBQ(operand.U32(1), c)
			build.JNE(operand.LabelRef(lbl + "_loop"))
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			t := s[k]
			for c := 2; c > 0; c-- {
				acc += t
				s[k] = acc
			}
			return acc
		},
	},
	{
		// The slot is live only around the back edge where the value later
		// stored to it is computed.
		Name: "LoopSlotBackEdge",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			c, t := build.GP64(), build.GP64()
			build.MOVQ(operand.U32(2), c)
			build.Label(lbl + "_loop")
			u := build.GP64()
			build.MOVQ(s[k], u)
			build.ADDQ(u, acc)
			build.MOVQ(y, t)
			build.ADDQ(acc, t)
			build.SUBQ(operand.U32(1), c)
			build.JNE(operand.LabelRef(lbl + "_loop"))
			build.MOVQ(t, s[k])
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			var t uint64
			for c := 2; c > 0; c-- {
				acc += s[k]
				t = y + acc
			}
			s[k] = t
			return acc
		},
	},
	{
		// A copy shiftFolds absorbs into a three-operand shift.
		Name: "ShiftFromSlot",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t, u := build.GP64(), build.GP64()
			build.MOVQ(s[k], t)
			build.MOVQ(t, u)
			build.SHLQ(operand.U8(5), u)
			build.XORQ(u, acc)
			build.ADDQ(t, acc) // keeps t live, so the copy stays a copy
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 { return acc ^ s[k]<<5 + s[k] },
	},
	{
		// A triple shiftExtractFold turns into one UBFX.
		Name: "ExtractFromSlot",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t, u := build.GP64(), build.GP64()
			build.MOVQ(s[k], t)
			build.MOVQ(t, u)
			build.SHRQ(operand.U8(8), u)
			build.MOVBQZX(u.As8(), u)
			build.ADDQ(u, acc)
			build.XORQ(t, acc) // keeps t live, so the copy stays a copy
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 { return (acc + s[k]>>8&0xff) ^ s[k] },
	},
	{
		Name: "AddToSlot",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) { build.ADDQ(acc, s[k]) },
		Ref:  func(acc, y uint64, s *[4]uint64, k int) uint64 { s[k] += acc; return acc },
	},
	{
		// A loop counter kept in the slot and updated in place.
		Name: "DecLoop",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			build.ANDQ(operand.U32(3), s[k])
			build.Label(lbl + "_loop")
			build.CMPQ(s[k], operand.U32(0))
			build.JEQ(operand.LabelRef(lbl + "_done"))
			build.ADDQ(s[k], acc)
			build.ROLQ(operand.U8(3), acc)
			build.DECQ(s[k])
			build.JMP(operand.LabelRef(lbl + "_loop"))
			build.Label(lbl + "_done")
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			for s[k] &= 3; s[k] != 0; s[k]-- {
				acc = bits.RotateLeft64(acc+s[k], 3)
			}
			return acc
		},
	},
	{
		Name: "CmpSlotBranch",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			build.CMPQ(y, s[k])
			build.JCS(operand.LabelRef(lbl + "_skip"))
			build.XORQ(y, acc)
			build.Label(lbl + "_skip")
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			if !(y < s[k]) {
				acc ^= y
			}
			return acc
		},
	},
	{
		// A copy of the slot added back into it: coalescing the copy into
		// the slot's register would name that register twice.
		Name: "SlotSelfAdd",
		Emit: func(acc, y reg.GPVirtual, s [4]operand.Mem, k int, lbl string) {
			t := build.GP64()
			build.MOVQ(s[k], t)
			build.ADDQ(t, s[k])
			build.ADDQ(t, acc)
		},
		Ref: func(acc, y uint64, s *[4]uint64, k int) uint64 {
			old := s[k]
			s[k] += old
			return acc + old
		},
	},
}

// FoldProgram is SlotProgram for the FoldOps vocabulary.
func FoldProgram(n, length int) []SlotStep {
	r := rand.New(rand.NewSource(int64(n)*104729 + 1299709))
	out := make([]SlotStep, length)
	for i := range out {
		out[i] = SlotStep{r.Intn(len(FoldOps)), r.Intn(4)}
	}
	return out
}

// NumFoldPrograms and FoldProgramLength size the coalescing family.
const (
	NumFoldPrograms   = 64
	FoldProgramLength = 12
)
