// Package condspec enumerates the x86 condition codes the arm64 lowering
// translates, paired with a pure-Go model of what each one means after a
// 64-bit compare.
//
// The lowering maps x86 conditions onto arm64 ones through a table that is
// correct only for the flags a compare or subtract produces -- the two
// architectures use opposite borrow conventions, so the map is written in terms
// of post-compare meaning. Every entry of that table therefore needs to be
// executed on both architectures, not just the handful the in-tree generators
// happen to emit.
//
// The generator emits one function per consumer family (SETcc, Jcc, CMOVcc)
// that evaluates every condition and packs the results into a bitmask, so the
// whole table is covered by three symbols and needs no per-condition glue. Both
// the generator and the test range over Conds, so the bit positions agree
// without either side recording anything.
package condspec

// Cond is one condition code and the meaning it carries after CMPQ(a, b),
// which sets flags from a-b.
type Cond struct {
	// Name is the mnemonic suffix avo uses: SET<Name>, J<Name>, CMOVQ<Name>.
	Name string
	// Ref reports whether the condition holds, modelling x86 semantics.
	Ref func(a, b uint64) bool
}

// subOverflow reports signed overflow of a-b, which is what x86's OF and
// arm64's V record.
func subOverflow(a, b uint64) bool {
	sa, sb := int64(a), int64(b)
	r := int64(a - b)
	return (sa < 0) != (sb < 0) && (r < 0) != (sa < 0)
}

// Conds is every condition armCond translates. Order fixes the bit positions
// used by the generated masks, so entries may be appended but not reordered.
var Conds = []Cond{
	{"EQ", func(a, b uint64) bool { return a == b }},
	{"NE", func(a, b uint64) bool { return a != b }},
	// Signed comparisons: N, V and Z together.
	{"LT", func(a, b uint64) bool { return int64(a) < int64(b) }},
	{"LE", func(a, b uint64) bool { return int64(a) <= int64(b) }},
	{"GT", func(a, b uint64) bool { return int64(a) > int64(b) }},
	{"GE", func(a, b uint64) bool { return int64(a) >= int64(b) }},
	// Unsigned comparisons: the carry conditions, where x86 sets CF on borrow
	// and arm64 clears C on borrow. These are the ones the translation table
	// deliberately inverts, and the ones a wrong producer silently breaks.
	{"CS", func(a, b uint64) bool { return a < b }},
	{"CC", func(a, b uint64) bool { return a >= b }},
	{"HI", func(a, b uint64) bool { return a > b }},
	{"LS", func(a, b uint64) bool { return a <= b }},
	// The raw sign and overflow bits, which are not the same as LT/GE: MI reads
	// N alone, so it differs from LT exactly when the subtraction overflowed.
	{"MI", func(a, b uint64) bool { return int64(a-b) < 0 }},
	{"PL", func(a, b uint64) bool { return int64(a-b) >= 0 }},
	{"OS", subOverflow},
	{"OC", func(a, b uint64) bool { return !subOverflow(a, b) }},
}

// Mask returns the reference bitmask for a compare of a against b, with bit i
// set when Conds[i] holds. The generated SetAll/JmpAll/CmovAll must all return
// this value.
func Mask(a, b uint64) uint64 {
	var m uint64
	for i, c := range Conds {
		if c.Ref(a, b) {
			m |= 1 << uint(i)
		}
	}
	return m
}
