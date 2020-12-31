package api

import (
	"sort"

	"github.com/mmcloughlin/avo/internal/inst"
)

// Function represents a function that constructs some collection of
// instruction forms.
type Function struct {
	Instruction inst.Instruction
	Suffixes    []string
	Forms       []inst.Form
}

// Name returns the function name.
func (f *Function) Name() string {
	return f.opcodesuffix("_")
}

// Opcode returns the full Go opcode of the instruction built by this function. Includes any suffixes.
func (f *Function) Opcode() string {
	return f.opcodesuffix(".")
}

func (f *Function) opcodesuffix(sep string) string {
	n := f.Instruction.Opcode
	for _, suffix := range f.Suffixes {
		n += sep
		n += suffix
	}
	return n
}

// InstructionFunctions builds the list of all functions for a given
// instruction.
func InstructionFunctions(i inst.Instruction) []*Function {
	// // One constructor for each possible suffix combination.
	// bysuffix := map[string]*Function{}
	// for _, f := range i.Forms {
	// 	for _, suffixes := range f.SupportedSuffixes() {
	// 		k := strings.Join(suffixes, ".")
	// 		if _, ok := bysuffix[k]; !ok {
	// 			bysuffix[k] = &Function{
	// 				Instruction: i,
	// 				Suffixes:    suffixes,
	// 			}
	// 		}
	// 		bysuffix[k].Forms = append(bysuffix[k].Forms, f)
	// 	}
	// }

	// // Convert to a sorted slice.
	// var ctors []*Function
	// for _, ctor := range bysuffix {
	// 	ctors = append(ctors, ctor)
	// }

	// SortFunctions(ctors)

	fns := []*Function{
		{
			Instruction: i,
			Suffixes:    []string{},
			Forms:       i.Forms,
		},
	}

	return fns
}

// InstructionsFunctions builds all functions for a list of instructions.
func InstructionsFunctions(is []inst.Instruction) []*Function {
	var all []*Function
	for _, i := range is {
		ctors := InstructionFunctions(i)
		all = append(all, ctors...)
	}

	SortFunctions(all)

	return all
}

// SortFunctions sorts a list of functions by name.
func SortFunctions(ctors []*Function) {
	sort.Slice(ctors, func(i, j int) bool {
		return ctors[i].Name() < ctors[j].Name()
	})
}
