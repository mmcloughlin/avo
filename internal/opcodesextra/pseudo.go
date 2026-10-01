package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// pseudo is pseudo-ops supported by the Go assembler.
var pseudo = []*inst.Instruction{
	// PCALIGN is handled by the architecture-independent part of the Go
	// assembler. Backend support for amd64 was added in Go 1.22.
	//
	// Reference: https://github.com/golang/go/blob/go1.27.1/doc/asm.html#L468-L481
	//
	//	<p>
	//	The <code>PCALIGN</code> pseudo-instruction is used to indicate that the next instruction should be aligned
	//	to a specified boundary by padding with no-op instructions.
	//	</p>
	//
	//	<p>
	//	It is currently supported on arm64, amd64, ppc64, loong64 and riscv64.
	//
	//	For example, the start of the <code>MOVD</code> instruction below is aligned to 32 bytes:
	//	<pre>
	//	PCALIGN $32
	//	MOVD $2, R0
	//	</pre>
	//	</p>
	//
	// Reference: https://github.com/golang/go/blob/go1.27.1/src/cmd/internal/obj/util.go#L760-L761
	//
	//	// AlignmentPaddingLength is the number of bytes to add to align code as requested.
	//	// Alignment is restricted to powers of 2 between 8 and 2048 inclusive.
	//
	// Alignments of 256 and above do not fit in an 8-bit immediate, hence the
	// additional imm16 form.
	{
		Opcode:  "PCALIGN",
		Summary: "Align the next instruction to the specified boundary",
		Forms: []inst.Form{
			{
				Operands: []inst.Operand{
					{Type: "imm8"},
				},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm16"},
				},
			},
		},
	},
}
