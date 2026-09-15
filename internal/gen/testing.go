package gen

import (
	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/printer"
)

// DeclareTestArguments prints a block of variables declaring a valid operand of
// each operand type.
func DeclareTestArguments(g *printer.Generator) {
	g.Printf("var (\n")
	for _, arg := range validArgs {
		g.Printf("\t%s operand.Op = %s\n", TestArgumentName(arg.Type), arg.Code)
	}
	g.Printf(")\n")
}

// TestSignature returns a function signature with arguments matching the given
// instruction form. Requires variables declared by DeclareTestArguments().
func TestSignature(f inst.Form) api.Signature {
	var names []string
	for _, op := range f.Operands {
		names = append(names, TestArgumentName(op.Type))
	}
	return api.ArgsList(names)
}

// TestArgumentName returns the name of the variable of operand type t declared
// by DeclareTestArguments().
func TestArgumentName(t string) string {
	return "op" + t
}

var validArgs = []struct {
	Type string
	Code string
}{
	// Immediates
	{"1", "operand.Imm(1)"},
	{"3", "operand.Imm(3)"},
	{"imm2u", "operand.Imm(3)"},
	{"imm8", "operand.Imm(math.MaxInt8)"},
	{"imm16", "operand.Imm(math.MaxInt16)"},
	{"imm32", "operand.Imm(math.MaxInt32)"},
	{"imm64", "operand.Imm(math.MaxInt64)"},

	// Registers
	{"al", "reg.AL"},
	{"cl", "reg.CL"},
	{"ax", "reg.AX"},
	{"eax", "reg.EAX"},
	{"rax", "reg.RAX"},
	{"r8", "reg.CH"},
	{"r16", "reg.R9W"},
	{"r32", "reg.R10L"},
	{"r64", "reg.R11"},
	{"xmm0", "reg.X0"},
	{"xmm", "reg.X7"},
	{"ymm", "reg.Y15"},
	{"zmm", "reg.Z31"},
	{"k", "reg.K7"},

	// Memory
	{"m", "operand.Mem{Base: reg.BX, Index: reg.CX, Scale: 2}"},
	{"m8", "operand.Mem{Base: reg.BL, Index: reg.CH, Scale: 1}"},
	{"m16", "operand.Mem{Base: reg.BX, Index: reg.CX, Scale: 2}"},
	{"m32", "operand.Mem{Base: reg.EBX, Index: reg.ECX, Scale: 4}"},
	{"m64", "operand.Mem{Base: reg.RBX, Index: reg.RCX, Scale: 8}"},
	{"m128", "operand.Mem{Base: reg.RBX, Index: reg.RCX, Scale: 8}"},
	{"m256", "operand.Mem{Base: reg.RBX, Index: reg.RCX, Scale: 8}"},
	{"m512", "operand.Mem{Base: reg.RBX, Index: reg.RCX, Scale: 8}"},

	// Vector memory
	{"vm32x", "operand.Mem{Base: reg.R13, Index: reg.X4, Scale: 1}"},
	{"vm64x", "operand.Mem{Base: reg.R13, Index: reg.X8, Scale: 1}"},
	{"vm32y", "operand.Mem{Base: reg.R13, Index: reg.Y4, Scale: 1}"},
	{"vm64y", "operand.Mem{Base: reg.R13, Index: reg.Y8, Scale: 1}"},
	{"vm32z", "operand.Mem{Base: reg.R13, Index: reg.Z4, Scale: 1}"},
	{"vm64z", "operand.Mem{Base: reg.R13, Index: reg.Z8, Scale: 1}"},

	// Relative
	{"rel8", "operand.Rel(math.MaxInt8)"},
	{"rel32", "operand.LabelRef(\"lbl\")"},
}
