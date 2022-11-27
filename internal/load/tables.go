package load

import "github.com/mmcloughlin/avo/internal/inst"

// alias defines an opcode alias.
type alias struct {
	From string
	To   string
}

// aliases defines a list of opcode aliases. Where possible these are extracted
// from the code (see note below).
var aliases = []alias{
	// The PSHUFD/PSHUFL alias is not recorded in the list of "Annoying aliases" below. However the instructions are identical.
	//
	// Reference: https://github.com/golang/go/blob/048c9164a0c5572df18325e377473e7893dbfb07/src/cmd/internal/obj/x86/asm6.go#L1365
	//
	//		{APSHUFL, yxshuf, Pq, opBytes{0x70, 00}},
	//
	// Reference: https://github.com/golang/go/blob/048c9164a0c5572df18325e377473e7893dbfb07/src/cmd/internal/obj/x86/asm6.go#L1688
	//
	//		{APSHUFD, yxshuf, Pq, opBytes{0x70, 0}},
	//
	{"PSHUFD", "PSHUFL"},
}

// Go contains a list of self-proclaimed "Annoying aliases", as follows. We use
// a script to automatically extract this list from the source code (see the
// following go:generate line). Then we merge this with the manual list above.
//
// Reference: https://github.com/golang/go/blob/048c9164a0c5572df18325e377473e7893dbfb07/src/cmd/asm/internal/arch/arch.go#L126-L182
//
//		}
//		// Annoying aliases.
//		instructions["JA"] = x86.AJHI   /* alternate */
//		instructions["JAE"] = x86.AJCC  /* alternate */
//		instructions["JB"] = x86.AJCS   /* alternate */
//		instructions["JBE"] = x86.AJLS  /* alternate */
//		instructions["JC"] = x86.AJCS   /* alternate */
//		instructions["JCC"] = x86.AJCC  /* carry clear (CF = 0) */
//		instructions["JCS"] = x86.AJCS  /* carry set (CF = 1) */
//		instructions["JE"] = x86.AJEQ   /* alternate */
//		instructions["JEQ"] = x86.AJEQ  /* equal (ZF = 1) */
//		instructions["JG"] = x86.AJGT   /* alternate */
//		instructions["JGE"] = x86.AJGE  /* greater than or equal (signed) (SF = OF) */
//		instructions["JGT"] = x86.AJGT  /* greater than (signed) (ZF = 0 && SF = OF) */
//		instructions["JHI"] = x86.AJHI  /* higher (unsigned) (CF = 0 && ZF = 0) */
//		instructions["JHS"] = x86.AJCC  /* alternate */
//		instructions["JL"] = x86.AJLT   /* alternate */
//		instructions["JLE"] = x86.AJLE  /* less than or equal (signed) (ZF = 1 || SF != OF) */
//		instructions["JLO"] = x86.AJCS  /* alternate */
//		instructions["JLS"] = x86.AJLS  /* lower or same (unsigned) (CF = 1 || ZF = 1) */
//		instructions["JLT"] = x86.AJLT  /* less than (signed) (SF != OF) */
//		instructions["JMI"] = x86.AJMI  /* negative (minus) (SF = 1) */
//		instructions["JNA"] = x86.AJLS  /* alternate */
//		instructions["JNAE"] = x86.AJCS /* alternate */
//		instructions["JNB"] = x86.AJCC  /* alternate */
//		instructions["JNBE"] = x86.AJHI /* alternate */
//		instructions["JNC"] = x86.AJCC  /* alternate */
//		instructions["JNE"] = x86.AJNE  /* not equal (ZF = 0) */
//		instructions["JNG"] = x86.AJLE  /* alternate */
//		instructions["JNGE"] = x86.AJLT /* alternate */
//		instructions["JNL"] = x86.AJGE  /* alternate */
//		instructions["JNLE"] = x86.AJGT /* alternate */
//		instructions["JNO"] = x86.AJOC  /* alternate */
//		instructions["JNP"] = x86.AJPC  /* alternate */
//		instructions["JNS"] = x86.AJPL  /* alternate */
//		instructions["JNZ"] = x86.AJNE  /* alternate */
//		instructions["JO"] = x86.AJOS   /* alternate */
//		instructions["JOC"] = x86.AJOC  /* overflow clear (OF = 0) */
//		instructions["JOS"] = x86.AJOS  /* overflow set (OF = 1) */
//		instructions["JP"] = x86.AJPS   /* alternate */
//		instructions["JPC"] = x86.AJPC  /* parity clear (PF = 0) */
//		instructions["JPE"] = x86.AJPS  /* alternate */
//		instructions["JPL"] = x86.AJPL  /* non-negative (plus) (SF = 0) */
//		instructions["JPO"] = x86.AJPC  /* alternate */
//		instructions["JPS"] = x86.AJPS  /* parity set (PF = 1) */
//		instructions["JS"] = x86.AJMI   /* alternate */
//		instructions["JZ"] = x86.AJEQ   /* alternate */
//		instructions["MASKMOVDQU"] = x86.AMASKMOVOU
//		instructions["MOVD"] = x86.AMOVQ
//		instructions["MOVDQ2Q"] = x86.AMOVQ
//		instructions["MOVNTDQ"] = x86.AMOVNTO
//		instructions["MOVOA"] = x86.AMOVO
//		instructions["PSLLDQ"] = x86.APSLLO
//		instructions["PSRLDQ"] = x86.APSRLO
//		instructions["PADDD"] = x86.APADDL
//
//		return &Arch{
//

//go:generate ./annoyingaliases.sh zannoyingaliases.go

func init() {
	aliases = append(aliases, annoyingaliases...)
}

// maskrequired is a set of AVX-512 opcodes where the mask register is required.
// Usually the mask register can be omitted, in which case K0 is implied.
var maskrequired = map[string]bool{
	// Reference: https://github.com/golang/go/blob/4fd94558820100129b98f284e21b19fc27a99926/src/cmd/internal/obj/x86/asm6.go#L4219-L4240
	//
	//		// Checks to warn about instruction/arguments combinations that
	//		// will unconditionally trigger illegal instruction trap (#UD).
	//		switch p.As {
	//		case AVGATHERDPD,
	//			AVGATHERQPD,
	//			AVGATHERDPS,
	//			AVGATHERQPS,
	//			AVPGATHERDD,
	//			AVPGATHERQD,
	//			AVPGATHERDQ,
	//			AVPGATHERQQ:
	//			// AVX512 gather requires explicit K mask.
	//			if p.GetFrom3().Reg >= REG_K0 && p.GetFrom3().Reg <= REG_K7 {
	//				if !avx512gatherValid(ctxt, p) {
	//					return
	//				}
	//			} else {
	//				if !avx2gatherValid(ctxt, p) {
	//					return
	//				}
	//			}
	//		}
	//
	"VGATHERDPD": true,
	"VGATHERQPD": true,
	"VGATHERDPS": true,
	"VGATHERQPS": true,
	"VPGATHERDD": true,
	"VPGATHERQD": true,
	"VPGATHERDQ": true,
	"VPGATHERQQ": true,

	// Restriction applies to SCATTER instructions too.
	"VPSCATTERDD": true,
	"VPSCATTERDQ": true,
	"VPSCATTERQD": true,
	"VPSCATTERQQ": true,
	"VSCATTERDPD": true,
	"VSCATTERDPS": true,
	"VSCATTERQPD": true,
	"VSCATTERQPS": true,
}

// extras is simply a list of extra instructions to add to the database.
var extras = []*inst.Instruction{
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
	//		{AMOVLQZX, yml_rl, Px, opBytes{0x8b}},
	//
	// Reference: https://github.com/golang/go/blob/048c9164a0c5572df18325e377473e7893dbfb07/src/cmd/internal/obj/x86/asm6.go#L515-L517
	//
	//	var yml_rl = []ytab{
	//		{Zm_r, 1, argList{Yml, Yrl}},
	//	}
	//
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
	// GFNI START
	// From https://www.felixcloutier.com/x86/gf2p8affineqb
	{
		Opcode:  "GF2P8AFFINEQB",
		Summary: "Galois Field Affine Transformation",
		Forms: []inst.Form{
			// 66 0F3A CE /r /ib GF2P8AFFINEQB xmm1, xmm2/m128, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeREX,
				ISA:          []string{"SSE2", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeREX,
				ISA:          []string{"SSE2", "GFNI"},
			},
		},
	},
	{
		Opcode:  "VGF2P8AFFINEQB",
		Summary: "Galois Field Affine Transformation",
		Forms: []inst.Form{
			// VEX.128.66.0F3A.W1 CE /r /ib VGF2P8AFFINEQB xmm1, xmm2, xmm3/m128, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			// VEX.256.66.0F3A.W1 CE /r /ib VGF2P8AFFINEQB ymm1, ymm2, ymm3/m256, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			// EVEX.128.66.0F3A.W1 CE /r /ib VGF2P8AFFINEQB xmm1{k1}{z}, xmm2, xmm3/m128/m64bcst, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			// EVEX.256.66.0F3A.W1 CE /r /ib VGF2P8AFFINEQB ymm1{k1}{z}, ymm2, ymm3/m256/m64bcst, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			// EVEX.512.66.0F3A.W1 CE /r /ib VGF2P8AFFINEQB zmm1{k1}{z}, zmm2, zmm3/m512/m64bcst, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m512", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m512", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
				Broadcast:    true,
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
				Broadcast:    true,
			},
		},
	},
	// From https://www.felixcloutier.com/x86/gf2p8affineinvqb
	{
		Opcode:  "GF2P8AFFINEINVQB",
		Summary: "Galois Field Affine Transformation Inverse",
		Forms: []inst.Form{
			// 66 0F3A CF /r /ib GF2P8AFFINEINVQB xmm1, xmm2/m128, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeREX,
				ISA:          []string{"SSE2", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeREX,
				ISA:          []string{"SSE2", "GFNI"},
			},
		},
	},
	{
		Opcode:  "VGF2P8AFFINEINVQB",
		Summary: "Galois Field Affine Transformation Inverse",
		Forms: []inst.Form{
			// VEX.128.66.0F3A.W1 CF /r /ib VGF2P8AFFINEINVQB xmm1, xmm2, xmm3/m128, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			// VEX.256.66.0F3A.W1 CF /r /ib VGF2P8AFFINEINVQB ymm1, ymm2, ymm3/m256, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			// EVEX.128.66.0F3A.W1 CF /r /ib VGF2P8AFFINEINVQB xmm1{k1}{z}, xmm2, xmm3/m128/m64bcst, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			// EVEX.256.66.0F3A.W1 CF /r /ib VGF2P8AFFINEINVQB ymm1{k1}{z}, ymm2, ymm3/m256/m64bcst, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			// EVEX.512.66.0F3A.W1 CF /r /ib VGF2P8AFFINEINVQB zmm1{k1}{z}, zmm2, zmm3/m512/m64bcst, imm8
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m512", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m512", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
				Broadcast:    true,
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
				Broadcast:    true,
			},
		},
	},
	// From https://www.felixcloutier.com/x86/gf2p8mulb
	{
		Opcode:  "GF2P8MULB",
		Summary: "Galois Field Multiply Bytes",
		Forms: []inst.Form{
			// 66 0F38 CF /r GF2P8MULB xmm1, xmm2/m128
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeREX,
				ISA:          []string{"SSE2", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeREX,
				ISA:          []string{"SSE2", "GFNI"},
			},
		},
	},
	{
		Opcode:  "VGF2P8MULB",
		Summary: "Galois Field Multiply Bytes",
		Forms: []inst.Form{
			// VEX.128.66.0F38.W0 CF /r VGF2P8MULB xmm1, xmm2, xmm3/m128
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			// VEX.256.66.0F38.W0 CF /r VGF2P8MULB ymm1, ymm2, ymm3/m256
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
				ISA:          []string{"AVX", "GFNI"},
			},
			// EVEX.128.66.0F38.W0 CF /r VGF2P8MULB xmm1{k1}{z}, xmm2, xmm3/m128
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			// EVEX.256.66.0F38.W0 CF /r VGF2P8MULB ymm1{k1}{z}, ymm2, ymm3/m256
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512VL", "GFNI"},
				Broadcast:    true,
			},
			// EVEX.512.66.0F38.W0 CF /r VGF2P8MULB zmm1{k1}{z}, zmm2, zmm3/m512
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m512", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m512", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
				Broadcast:    true,
			},
			{
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.R},
					{Type: "m64", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				ISA:          []string{"AVX512F", "GFNI"},
				Broadcast:    true,
			},
		},
	},
	// GFNI END
}
