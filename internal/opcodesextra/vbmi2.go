package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// vbmi2 is the "Vector Bit Manipulation Instructions 2" instruction set.
var vbmi2 = []*inst.Instruction{
	// VPCOMPRESSB, VPCOMPRESSW 	Store sparse packed byte/word integer values into dense memory/register
	// VPEXPANDB, VPEXPANDW 	Load sparse packed byte/word integer values from dense memory/register
	// VPSHLD 	Concatenate and shift packed data left logical
	// VPSHLDV 	Concatenate and variable shift packed data left logical
	// VPSHRD 	Concatenate and shift packed data right logical
	// VPSHRDV 	Concatenate and variable shift packed data right logical
}
