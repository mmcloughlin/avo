package inst

import (
	"sort"
	"strings"
)

//go:generate avogen -bootstrap -data ../data -output ztable.go godata

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

// SuffixesClasses returns all possible classes of suffix combinations.
func SuffixesClasses(is []Instruction) map[string][]Suffixes {
	classes := map[string][]Suffixes{}
	for _, i := range is {
		for _, f := range i.Forms {
			class := f.SuffixesClass()
			if _, ok := classes[class]; ok {
				continue
			}
			classes[class] = f.SupportedSuffixes()
		}
	}
	return classes
}

// ISAs returns all the unique ISAs seen in the given instructions.
func ISAs(is []Instruction) []string {
	set := map[string]bool{}
	for _, i := range is {
		for _, f := range i.Forms {
			for _, isa := range f.ISA {
				set[isa] = true
			}
		}
	}
	return sortedslice(set)
}

// ISACombinations returns all the unique combinations of ISAs seen in the given
// instructions.
func ISACombinations(is []Instruction) [][]string {
	var combinations [][]string
	seen := map[string]bool{}
	for _, i := range is {
		for _, f := range i.Forms {
			isas := append([]string(nil), f.ISA...)
			sort.Strings(isas)
			key := strings.Join(isas, ",")

			if !seen[key] {
				combinations = append(combinations, isas)
				seen[key] = true
			}
		}
	}
	return combinations
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
