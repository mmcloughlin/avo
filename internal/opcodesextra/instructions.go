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
	var is []*inst.Instruction
	for _, set := range sets {
		is = append(is, set...)
	}
	return is
}
