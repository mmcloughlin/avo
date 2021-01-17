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
	// Collect set.
	set := map[string]bool{}
	for _, i := range is {
		for _, f := range i.Forms {
			for _, op := range f.Operands {
				set[op.Type] = true
			}
		}
	}

	// Convert to sorted slice.
	types := make([]string, 0, len(set))
	for t := range set {
		types = append(types, t)
	}

	sort.Strings(types)

	return types
}
