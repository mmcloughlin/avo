package inst

import "sort"

//go:generate avogen -bootstrap -data ../data -output ztable.go godata
//go:generate avogen -bootstrap -data ../data -output ztable_test.go godatatest

// Lookup returns the instruction with the given opcode. Boolean return value
// indicates whether the instruction was found.
func Lookup(opcode string) (Instruction, bool) {
	for _, i := range Instructions {
		if i.Opcode == opcode {
			return i, true
		}
	}
	return Instruction{}, false
}

// OperandTypes returns all the operand types that appear in the provided
// instructions.
func OperandTypes(is []Instruction) []string {
	set := map[string]bool{}
	for _, i := range is {
		for _, f := range i.Forms {
			for _, op := range f.Operands {
				set[op.Type] = true
			}
		}
	}
	return sortedslice(set)
}

// ImplicitRegisters returns all the registers that appear as implicit operands
// in the provided instructions.
func ImplicitRegisters(is []Instruction) []string {
	set := map[string]bool{}
	for _, i := range is {
		for _, f := range i.Forms {
			for _, op := range f.ImplicitOperands {
				set[op.Register] = true
			}
		}
	}
	return sortedslice(set)
}

// UniqueSuffixes returns all the non-empty suffixes that appear in the provided
// instructions.
func UniqueSuffixes(is []Instruction) []Suffix {
	// Collect set.
	set := map[Suffix]bool{}
	for _, i := range is {
		for _, f := range i.Forms {
			for _, suffixes := range f.SupportedSuffixes() {
				for _, suffix := range suffixes {
					set[suffix] = true
				}
			}
		}
	}

	// Convert to sorted slice.
	suffixes := make([]Suffix, 0, len(set))
	for suffix := range set {
		suffixes = append(suffixes, suffix)
	}

	sort.Slice(suffixes, func(i, j int) bool {
		return suffixes[i] < suffixes[j]
	})

	return suffixes
}

// sortedslice builds a sorted slice of strings from a set.
func sortedslice(set map[string]bool) []string {
	ss := make([]string, 0, len(set))
	for s := range set {
		ss = append(ss, s)
	}

	sort.Strings(ss)

	return ss
}
