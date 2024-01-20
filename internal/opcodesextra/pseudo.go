package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// pseudo is pseudo-ops supported by the Go assembler.
var pseudo = []*inst.Instruction{
	// Reference: https://github.com/golang/go/blob/go1.22rc1/doc/asm.html#L469-L482
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
	{
		Opcode:  "PCALIGN",
		Summary: "Align the next instruction to the specified boundary",
		Forms: []inst.Form{
			{
				Operands: []inst.Operand{
					{Type: "imm8"},
				},
			},
		},
	},
}
