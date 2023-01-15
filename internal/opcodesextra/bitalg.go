package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// bitalg is the "Bit Algorithms" instruction set.
var bitalg = []*inst.Instruction{
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3904-L3908
	//
	//		{as: AVPSHUFBITQMB, ytab: _yvpshufbitqmb, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16, 0x8F,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32, 0x8F,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64, 0x8F,
	//		}},
	//
	{
		Opcode:  "VPSHUFBITQMB",
		Summary: "Shuffle Bits from Quadword Elements Using Byte Indexes into Mask",
		Forms:   vpshufbitqmb,
	},
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3736-L3740
	//
	//		{as: AVPOPCNTB, ytab: _yvexpandpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexZeroingEnabled, 0x54,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexZeroingEnabled, 0x54,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexZeroingEnabled, 0x54,
	//		}},
	//
	{
		Opcode:  "VPOPCNTB",
		Summary: "Packed Population Count for Byte Integers",
		Forms:   vpopcntb,
	},
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3751-L3755
	//
	//		{as: AVPOPCNTW, ytab: _yvexpandpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN16 | evexZeroingEnabled, 0x54,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN32 | evexZeroingEnabled, 0x54,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN64 | evexZeroingEnabled, 0x54,
	//		}},
	//
	{
		Opcode:  "VPOPCNTW",
		Summary: "Packed Population Count for Word Integers",
		Forms:   vpopcntb,
	},
}

// VPSHUFBITQMB forms.
//
// See: https://www.felixcloutier.com/x86/vpshufbitqmb
//
// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L827-L834
//
//	var _yvpshufbitqmb = []ytab{
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{YxmEvex, YxrEvex, Yk}},
//		{zcase: Zevex_rm_v_k_r, zoffset: 3, args: argList{YxmEvex, YxrEvex, Yknot0, Yk}},
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{YymEvex, YyrEvex, Yk}},
//		{zcase: Zevex_rm_v_k_r, zoffset: 3, args: argList{YymEvex, YyrEvex, Yknot0, Yk}},
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{Yzm, Yzr, Yk}},
//		{zcase: Zevex_rm_v_k_r, zoffset: 3, args: argList{Yzm, Yzr, Yknot0, Yk}},
//	}
var vpshufbitqmb = inst.Forms{
	// EVEX.128.66.0F38.W0 8F /r VPSHUFBITQMB k1{k2}, xmm2, xmm3/m128
	{
		ISA: []string{"AVX512BITALG", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "m128", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "k{k}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512BITALG", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "k{k}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.256.66.0F38.W0 8F /r VPSHUFBITQMB k1{k2}, ymm2, ymm3/m256
	{
		ISA: []string{"AVX512BITALG", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "m256", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "k{k}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512BITALG", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "k{k}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.512.66.0F38.W0 8F /r VPSHUFBITQMB k1{k2}, zmm2, zmm3/m512
	{
		ISA: []string{"AVX512BITALG"},
		Operands: []inst.Operand{
			{Type: "m512", Action: inst.R},
			{Type: "zmm", Action: inst.R},
			{Type: "k{k}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512BITALG", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "zmm", Action: inst.R},
			{Type: "k{k}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
}

// VPOPCNTB and VPOPCNTW forms.
//
// See: https://www.felixcloutier.com/x86/vpopcnt
//
// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L376-L383
//
//	var _yvexpandpd = []ytab{
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{YxmEvex, YxrEvex}},
//		{zcase: Zevex_rm_k_r, zoffset: 3, args: argList{YxmEvex, Yknot0, YxrEvex}},
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{YymEvex, YyrEvex}},
//		{zcase: Zevex_rm_k_r, zoffset: 3, args: argList{YymEvex, Yknot0, YyrEvex}},
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{Yzm, Yzr}},
//		{zcase: Zevex_rm_k_r, zoffset: 3, args: argList{Yzm, Yknot0, Yzr}},
//	}
var vpopcntb = _yvexpandpd("AVX512BITALG", "")
