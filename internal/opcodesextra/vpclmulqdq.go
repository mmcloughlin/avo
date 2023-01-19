package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// vpclmulqdq is the "Vector Carry-less Multiplication" instruction set.
var vpclmulqdq = []*inst.Instruction{
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L2911-L2917
	//
	//		{as: AVPCLMULQDQ, ytab: _yvpclmulqdq, prefix: Pavx, op: opBytes{
	//			avxEscape | vex128 | vex66 | vex0F3A | vexW0, 0x44,
	//			avxEscape | vex256 | vex66 | vex0F3A | vexW0, 0x44,
	//			avxEscape | evex128 | evex66 | evex0F3A | evexW0, evexN16, 0x44,
	//			avxEscape | evex256 | evex66 | evex0F3A | evexW0, evexN32, 0x44,
	//			avxEscape | evex512 | evex66 | evex0F3A | evexW0, evexN64, 0x44,
	//		}},
	//
	{
		Opcode:  "VPCLMULQDQ",
		Summary: "Carry-Less Quadword Multiplication",
		// See: https://www.felixcloutier.com/x86/pclmulqdq
		//
		// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L676-L682
		//
		//	var _yvpclmulqdq = []ytab{
		//		{zcase: Zvex_i_rm_v_r, zoffset: 2, args: argList{Yu8, Yxm, Yxr, Yxr}},
		//		{zcase: Zvex_i_rm_v_r, zoffset: 2, args: argList{Yu8, Yym, Yyr, Yyr}},
		//		{zcase: Zevex_i_rm_v_r, zoffset: 3, args: argList{Yu8, YxmEvex, YxrEvex, YxrEvex}},
		//		{zcase: Zevex_i_rm_v_r, zoffset: 3, args: argList{Yu8, YymEvex, YyrEvex, YyrEvex}},
		//		{zcase: Zevex_i_rm_v_r, zoffset: 3, args: argList{Yu8, Yzm, Yzr, Yzr}},
		//	}
		//
		Forms: inst.Forms{
			// VEX.128.66.0F3A.WIG 44 /r ib VPCLMULQDQ xmm1, xmm2, xmm3/m128, imm8
			{
				ISA: []string{"AVX", "PCLMULQDQ"},
				Operands: []inst.Operand{
					{Type: "imm8"},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
			},
			{
				ISA: []string{"AVX", "PCLMULQDQ"},
				Operands: []inst.Operand{
					{Type: "imm8"},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
			},
			// VEX.256.66.0F3A.WIG 44 /r /ib VPCLMULQDQ ymm1, ymm2, ymm3/m256, imm8
			{
				ISA: []string{"VPCLMULQDQ"},
				Operands: []inst.Operand{
					{Type: "imm8"},
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
			},
			{
				ISA: []string{"VPCLMULQDQ"},
				Operands: []inst.Operand{
					{Type: "imm8"},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeVEX,
			},
			// EVEX.128.66.0F3A.WIG 44 /r /ib VPCLMULQDQ xmm1, xmm2, xmm3/m128, imm8
			{
				ISA: []string{"AVX512VL", "VPCLMULQDQ"},
				Operands: []inst.Operand{
					{Type: "imm8"},
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "VPCLMULQDQ"},
				Operands: []inst.Operand{
					{Type: "imm8"},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			// EVEX.256.66.0F3A.WIG 44 /r /ib VPCLMULQDQ ymm1, ymm2, ymm3/m256, imm8
			{
				ISA: []string{"AVX512VL", "VPCLMULQDQ"},
				Operands: []inst.Operand{
					{Type: "imm8"},
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "VPCLMULQDQ"},
				Operands: []inst.Operand{
					{Type: "imm8"},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			// EVEX.512.66.0F3A.WIG 44 /r /ib VPCLMULQDQ zmm1, zmm2, zmm3/m512, imm8
			{
				ISA: []string{"AVX512F", "VPCLMULQDQ"},
				Operands: []inst.Operand{
					{Type: "imm8"},
					{Type: "m512", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512F", "VPCLMULQDQ"},
				Operands: []inst.Operand{
					{Type: "imm8"},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
		},
	},
}
