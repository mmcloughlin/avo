package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// ifma is the "Integer Fused Multiply and Accumulate" instruction set.
var ifma = []*inst.Instruction{
	// VPMADD52LUQ 	IFMA 	Packed multiply of unsigned 52-bit integers and add the low 52-bit products to qword accumulators
	// VPMADD52HUQ 	IFMA 	Packed multiply of unsigned 52-bit integers and add the high 52-bit products to 64-bit accumulators
}
