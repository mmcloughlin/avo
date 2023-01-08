package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// vaes is the "Vector Advanced Encryption Standard" instruction set.
var vaes = []*inst.Instruction{
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L1217-L1244
	//
	//		{as: AVAESDEC, ytab: _yvaesdec, prefix: Pavx, op: opBytes{
	//			avxEscape | vex128 | vex66 | vex0F38 | vexW0, 0xDE,
	//			avxEscape | vex256 | vex66 | vex0F38 | vexW0, 0xDE,
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16, 0xDE,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32, 0xDE,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64, 0xDE,
	//		}},
	//		{as: AVAESDECLAST, ytab: _yvaesdec, prefix: Pavx, op: opBytes{
	//			avxEscape | vex128 | vex66 | vex0F38 | vexW0, 0xDF,
	//			avxEscape | vex256 | vex66 | vex0F38 | vexW0, 0xDF,
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16, 0xDF,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32, 0xDF,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64, 0xDF,
	//		}},
	//		{as: AVAESENC, ytab: _yvaesdec, prefix: Pavx, op: opBytes{
	//			avxEscape | vex128 | vex66 | vex0F38 | vexW0, 0xDC,
	//			avxEscape | vex256 | vex66 | vex0F38 | vexW0, 0xDC,
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16, 0xDC,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32, 0xDC,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64, 0xDC,
	//		}},
	//		{as: AVAESENCLAST, ytab: _yvaesdec, prefix: Pavx, op: opBytes{
	//			avxEscape | vex128 | vex66 | vex0F38 | vexW0, 0xDD,
	//			avxEscape | vex256 | vex66 | vex0F38 | vexW0, 0xDD,
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16, 0xDD,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32, 0xDD,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64, 0xDD,
	//		}},
	//
	{
		Opcode:  "VAESDEC",
		Summary: "Perform One Round of an AES Decryption Flow",
		Forms:   vaesforms,
	},
	{
		Opcode:  "VAESDECLAST",
		Summary: "Perform Last Round of an AES Decryption Flow",
		Forms:   vaesforms,
	},
	{
		Opcode:  "VAESENC",
		Summary: "Perform One Round of an AES Encryption Flow",
		Forms:   vaesforms,
	},
	{
		Opcode:  "VAESENCLAST",
		Summary: "Perform Last Round of an AES Encryption Flow",
		Forms:   vaesforms,
	},
}

// VAESDEC, VAESDECLAST, VAESENC and VAESENCLAST forms.
//
// See: https://software.intel.com/en-us/download/intel-64-and-ia-32-architectures-sdm-combined-volumes-1-2a-2b-2c-2d-3a-3b-3c-3d-and-4
//
// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L111-L117
//
//	var _yvaesdec = []ytab{
//		{zcase: Zvex_rm_v_r, zoffset: 2, args: argList{Yxm, Yxr, Yxr}},
//		{zcase: Zvex_rm_v_r, zoffset: 2, args: argList{Yym, Yyr, Yyr}},
//		{zcase: Zevex_rm_v_r, zoffset: 3, args: argList{YxmEvex, YxrEvex, YxrEvex}},
//		{zcase: Zevex_rm_v_r, zoffset: 3, args: argList{YymEvex, YyrEvex, YyrEvex}},
//		{zcase: Zevex_rm_v_r, zoffset: 3, args: argList{Yzm, Yzr, Yzr}},
//	}
var vaesforms = inst.Forms{
	// VEX.128.66.0F38.WIG DE /r VAESDEC xmm1, xmm2, xmm3/m128  AVX + AES
	// VEX.128.66.0F38.WIG DF /r VAESDECLAST xmm1, xmm2, xmm3/m128  AVX + AES
	// VEX.128.66.0F38.WIG DC /r VAESENC xmm1, xmm2, xmm3/m128  AVX + AES
	// VEX.128.66.0F38.WIG DD /r VAESENCLAST xmm1, xmm2, xmm3/m128  AVX + AES
	{
		ISA: []string{"AES", "AVX"},
		Operands: []inst.Operand{
			{Type: "m128", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeVEX,
	},
	{
		ISA: []string{"AES", "AVX"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeVEX,
	},
	// VEX.256.66.0F38.WIG DE /r VAESDEC ymm1, ymm2, ymm3/m256  VAES
	// VEX.256.66.0F38.WIG DF /r VAESDECLAST ymm1, ymm2, ymm3/m256  VAES
	// VEX.256.66.0F38.WIG DC /r VAESENC ymm1, ymm2, ymm3/m256  VAES
	// VEX.256.66.0F38.WIG DD /r VAESENCLAST ymm1, ymm2, ymm3/m256  VAES
	{
		ISA: []string{"VAES"},
		Operands: []inst.Operand{
			{Type: "m256", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeVEX,
	},
	{
		ISA: []string{"VAES"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeVEX,
	},
	// EVEX.128.66.0F38.WIG DE /r VAESDEC xmm1, xmm2, xmm3/m128  AVX512VL + VAES
	// EVEX.128.66.0F38.WIG DF /r VAESDECLAST xmm1, xmm2, xmm3/m128  AVX512VL + VAES
	// EVEX.128.66.0F38.WIG DC /r VAESENC xmm1, xmm2, xmm3/m128  AVX512VL + VAES
	// EVEX.128.66.0F38.WIG DD /r VAESENCLAST xmm1, xmm2, xmm3/m128  AVX512VL + VAES
	{
		ISA: []string{"AVX512VL", "VAES"},
		Operands: []inst.Operand{
			{Type: "m128", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VL", "VAES"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.256.66.0F38.WIG DE /r VAESDEC ymm1, ymm2, ymm3/m256  AVX512VL + VAES
	// EVEX.256.66.0F38.WIG DF /r VAESDECLAST ymm1, ymm2, ymm3/m256  AVX512VL + VAES
	// EVEX.256.66.0F38.WIG DC /r VAESENC ymm1, ymm2, ymm3/m256  AVX512VL + VAES
	// EVEX.256.66.0F38.WIG DD /r VAESENCLAST ymm1, ymm2, ymm3/m256  AVX512VL + VAES
	{
		ISA: []string{"AVX512VL", "VAES"},
		Operands: []inst.Operand{
			{Type: "m256", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VL", "VAES"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.512.66.0F38.WIG DE /r VAESDEC zmm1, zmm2, zmm3/m512  AVX512F + VAES
	// EVEX.512.66.0F38.WIG DF /r VAESDECLAST zmm1, zmm2, zmm3/m512  AVX512F + VAES
	// EVEX.512.66.0F38.WIG DC /r VAESENC zmm1, zmm2, zmm3/m512  AVX512F + VAES
	// EVEX.512.66.0F38.WIG DD /r VAESENCLAST zmm1, zmm2, zmm3/m512  AVX512F + VAES
	{
		ISA: []string{"AVX512F", "VAES"},
		Operands: []inst.Operand{
			{Type: "m512", Action: inst.R},
			{Type: "zmm", Action: inst.R},
			{Type: "zmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512F", "VAES"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "zmm", Action: inst.R},
			{Type: "zmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
}
