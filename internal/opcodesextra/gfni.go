package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

var gfni = []*inst.Instruction{
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
}
