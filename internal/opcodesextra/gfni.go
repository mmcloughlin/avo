package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// gfni is the "Galois Field New Instructions" instruction set.
var gfni = []*inst.Instruction{
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L2250-L2269
	//
	//		{as: AVGF2P8AFFINEINVQB, ytab: _yvgf2p8affineinvqb, prefix: Pavx, op: opBytes{
	//			avxEscape | vex128 | vex66 | vex0F3A | vexW1, 0xCF,
	//			avxEscape | vex256 | vex66 | vex0F3A | vexW1, 0xCF,
	//			avxEscape | evex128 | evex66 | evex0F3A | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0xCF,
	//			avxEscape | evex256 | evex66 | evex0F3A | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0xCF,
	//			avxEscape | evex512 | evex66 | evex0F3A | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0xCF,
	//		}},
	//		{as: AVGF2P8AFFINEQB, ytab: _yvgf2p8affineinvqb, prefix: Pavx, op: opBytes{
	//			avxEscape | vex128 | vex66 | vex0F3A | vexW1, 0xCE,
	//			avxEscape | vex256 | vex66 | vex0F3A | vexW1, 0xCE,
	//			avxEscape | evex128 | evex66 | evex0F3A | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0xCE,
	//			avxEscape | evex256 | evex66 | evex0F3A | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0xCE,
	//			avxEscape | evex512 | evex66 | evex0F3A | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0xCE,
	//		}},
	//		{as: AVGF2P8MULB, ytab: _yvandnpd, prefix: Pavx, op: opBytes{
	//			avxEscape | vex128 | vex66 | vex0F38 | vexW0, 0xCF,
	//			avxEscape | vex256 | vex66 | vex0F38 | vexW0, 0xCF,
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexZeroingEnabled, 0xCF,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexZeroingEnabled, 0xCF,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexZeroingEnabled, 0xCF,
	//
	{
		Opcode:  "VGF2P8AFFINEQB",
		Summary: "Galois Field Affine Transformation",
		Forms:   vgf2p8affineqb,
	},
	{
		Opcode:  "VGF2P8AFFINEINVQB",
		Summary: "Galois Field Affine Transformation Inverse",
		Forms:   vgf2p8affineqb,
	},
	{
		Opcode:  "VGF2P8MULB",
		Summary: "Galois Field Multiply Bytes",
		Forms:   vgf2p8mulb,
	},
}

// VGF2P8AFFINEQB and VGF2P8AFFINEINVQB forms.
//
// See: https://www.felixcloutier.com/x86/gf2p8affineqb
//
// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L483-L492
//
//	var _yvgf2p8affineinvqb = []ytab{
//		{zcase: Zvex_i_rm_v_r, zoffset: 2, args: argList{Yu8, Yxm, Yxr, Yxr}},
//		{zcase: Zvex_i_rm_v_r, zoffset: 2, args: argList{Yu8, Yym, Yyr, Yyr}},
//		{zcase: Zevex_i_rm_v_r, zoffset: 0, args: argList{Yu8, YxmEvex, YxrEvex, YxrEvex}},
//		{zcase: Zevex_i_rm_v_k_r, zoffset: 3, args: argList{Yu8, YxmEvex, YxrEvex, Yknot0, YxrEvex}},
//		{zcase: Zevex_i_rm_v_r, zoffset: 0, args: argList{Yu8, YymEvex, YyrEvex, YyrEvex}},
//		{zcase: Zevex_i_rm_v_k_r, zoffset: 3, args: argList{Yu8, YymEvex, YyrEvex, Yknot0, YyrEvex}},
//		{zcase: Zevex_i_rm_v_r, zoffset: 0, args: argList{Yu8, Yzm, Yzr, Yzr}},
//		{zcase: Zevex_i_rm_v_k_r, zoffset: 3, args: argList{Yu8, Yzm, Yzr, Yknot0, Yzr}},
//	}
var vgf2p8affineqb = inst.Forms{
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
			{Type: "xmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "GFNI"},
	},
	{
		Operands: []inst.Operand{
			{Type: "imm8", Action: inst.R},
			{Type: "m128/m64bcst", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "xmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "GFNI"},
	},
	// EVEX.256.66.0F3A.W1 CE /r /ib VGF2P8AFFINEQB ymm1{k1}{z}, ymm2, ymm3/m256/m64bcst, imm8
	{
		Operands: []inst.Operand{
			{Type: "imm8", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "GFNI"},
	},
	{
		Operands: []inst.Operand{
			{Type: "imm8", Action: inst.R},
			{Type: "m256/m64bcst", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "GFNI"},
	},
	// EVEX.512.66.0F3A.W1 CE /r /ib VGF2P8AFFINEQB zmm1{k1}{z}, zmm2, zmm3/m512/m64bcst, imm8
	{
		Operands: []inst.Operand{
			{Type: "imm8", Action: inst.R},
			{Type: "zmm", Action: inst.R},
			{Type: "zmm", Action: inst.R},
			{Type: "zmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512F", "GFNI"},
	},
	{
		Operands: []inst.Operand{
			{Type: "imm8", Action: inst.R},
			{Type: "m512/m64bcst", Action: inst.R},
			{Type: "zmm", Action: inst.R},
			{Type: "zmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512F", "GFNI"},
	},
}

// VGF2P8MULB forms.
//
// See: https://www.felixcloutier.com/x86/gf2p8mulb
//
// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L137-L146
//
//	var _yvandnpd = []ytab{
//		{zcase: Zvex_rm_v_r, zoffset: 2, args: argList{Yxm, Yxr, Yxr}},
//		{zcase: Zvex_rm_v_r, zoffset: 2, args: argList{Yym, Yyr, Yyr}},
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{YxmEvex, YxrEvex, YxrEvex}},
//		{zcase: Zevex_rm_v_k_r, zoffset: 3, args: argList{YxmEvex, YxrEvex, Yknot0, YxrEvex}},
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{YymEvex, YyrEvex, YyrEvex}},
//		{zcase: Zevex_rm_v_k_r, zoffset: 3, args: argList{YymEvex, YyrEvex, Yknot0, YyrEvex}},
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{Yzm, Yzr, Yzr}},
//		{zcase: Zevex_rm_v_k_r, zoffset: 3, args: argList{Yzm, Yzr, Yknot0, Yzr}},
//	}
var vgf2p8mulb = inst.Forms{
	// VEX.128.66.0F38.W0 CF /r VGF2P8MULB xmm1, xmm2, xmm3/m128
	{
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeVEX,
		ISA:          []string{"AVX", "GFNI"},
	},
	{
		Operands: []inst.Operand{
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
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeVEX,
		ISA:          []string{"AVX", "GFNI"},
	},
	{
		Operands: []inst.Operand{
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
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "xmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "GFNI"},
	},
	{
		Operands: []inst.Operand{
			{Type: "m128", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "xmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "GFNI"},
	},
	// EVEX.256.66.0F38.W0 CF /r VGF2P8MULB ymm1{k1}{z}, ymm2, ymm3/m256
	{
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "GFNI"},
	},
	{
		Operands: []inst.Operand{
			{Type: "m256", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "GFNI"},
	},
	// EVEX.512.66.0F38.W0 CF /r VGF2P8MULB zmm1{k1}{z}, zmm2, zmm3/m512
	{
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "zmm", Action: inst.R},
			{Type: "zmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512F", "GFNI"},
	},
	{
		Operands: []inst.Operand{
			{Type: "m512", Action: inst.R},
			{Type: "zmm", Action: inst.R},
			{Type: "zmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512F", "GFNI"},
	},
}
