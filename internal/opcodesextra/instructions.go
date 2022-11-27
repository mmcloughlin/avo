// Package opcodesextra provides curated extensions to the instruction database.
package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// sets of extra instructions.
var sets = [][]*inst.Instruction{
	movlqzx,
	gfni,
}

// Instructions returns a list of extras to add to the instructions database.
func Instructions() []*inst.Instruction {
	// Concatenate and clone the instruction lists.  It can be convenient for
	// forms lists and other data structures to be shared in the curated lists,
	// but we want to return distinct copies here to avoid subtle bugs in
	// consumers.
	var is []*inst.Instruction
	for _, set := range sets {
		for _, i := range set {
			c := i.Clone()
			is = append(is, &c)
		}
	}
	return is
}
