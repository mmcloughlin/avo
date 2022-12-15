package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// MOVLQZX does not appear in either x86 CSV or Opcodes, but does appear in stdlib assembly.
//
// Reference: https://github.com/golang/go/blob/048c9164a0c5572df18325e377473e7893dbfb07/src/runtime/asm_amd64.s#L451-L453
//
//	TEXT ·reflectcall(SB), NOSPLIT, $0-32
//		MOVLQZX argsize+24(FP), CX
//		DISPATCH(runtime·call32, 32)
//
// Reference: https://github.com/golang/go/blob/048c9164a0c5572df18325e377473e7893dbfb07/src/cmd/internal/obj/x86/asm6.go#L1217
//
//	{AMOVLQZX, yml_rl, Px, opBytes{0x8b}},
//
// Reference: https://github.com/golang/go/blob/048c9164a0c5572df18325e377473e7893dbfb07/src/cmd/internal/obj/x86/asm6.go#L515-L517
//
//	var yml_rl = []ytab{
//		{Zm_r, 1, argList{Yml, Yrl}},
//	}
var movlqzx = []*inst.Instruction{
	{
		Opcode:  "MOVLQZX",
		Summary: "Move with Zero-Extend",
		Forms: []inst.Form{
			{
				Operands: []inst.Operand{
					{Type: "m32", Action: inst.R},
					{Type: "r64", Action: inst.W},
				},
			},
		},
	},
}
