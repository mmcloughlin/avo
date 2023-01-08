package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// vnni is the "Vector Neural Network Instructions" instruction set.
var vnni = []*inst.Instruction{
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3056-L3075
	//
	//		{as: AVPDPBUSD, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x50,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x50,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x50,
	//		}},
	//		{as: AVPDPBUSDS, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x51,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x51,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x51,
	//		}},
	//		{as: AVPDPWSSD, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x52,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x52,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x52,
	//		}},
	//		{as: AVPDPWSSDS, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x53,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x53,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x53,
	//		}},
	//
	// Note the Go assembler does not support the VEX encoded "AVX-VNNI" forms of these instructions.
	{
		Opcode:  "VPDPBUSD",
		Summary: "Multiply and Add Unsigned and Signed Bytes",
		Forms:   vnniforms,
	},
	{
		Opcode:  "VPDPBUSDS",
		Summary: "Multiply and Add Unsigned and Signed Bytes with Saturation",
		Forms:   vnniforms,
	},
	{
		Opcode:  "VPDPWSSD",
		Summary: "Multiply and Add Signed Word Integers",
		Forms:   vnniforms,
	},
	{
		Opcode:  "VPDPWSSDS",
		Summary: "Multiply and Add Signed Word Integers with Saturation",
		Forms:   vnniforms,
	},
}

// VPDPBUSD, VPDPBUSDS, VPDPWSSD and VPDPWSSDS forms.
//
// See: https://www.felixcloutier.com/x86/vpdpbusd
// See: https://www.felixcloutier.com/x86/vpdpbusds
// See: https://www.felixcloutier.com/x86/vpdpwssd
// See: https://www.felixcloutier.com/x86/vpdpwssds
//
// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L148-L155
//
//	var _yvblendmpd = []ytab{
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{YxmEvex, YxrEvex, YxrEvex}},
//		{zcase: Zevex_rm_v_k_r, zoffset: 3, args: argList{YxmEvex, YxrEvex, Yknot0, YxrEvex}},
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{YymEvex, YyrEvex, YyrEvex}},
//		{zcase: Zevex_rm_v_k_r, zoffset: 3, args: argList{YymEvex, YyrEvex, Yknot0, YyrEvex}},
//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{Yzm, Yzr, Yzr}},
//		{zcase: Zevex_rm_v_k_r, zoffset: 3, args: argList{Yzm, Yzr, Yknot0, Yzr}},
//	}
var vnniforms = inst.Forms{
	// EVEX.128.66.0F38.W0 50 /r VPDPBUSD xmm1{k1}{z}, xmm2, xmm3/m128/m32bcst  AVX512VNNI AVX512VL
	// EVEX.128.66.0F38.W0 51 /r VPDPBUSDS xmm1{k1}{z}, xmm2, xmm3/m128/m32bcst  AVX512VNNI AVX512VL
	// EVEX.128.66.0F38.W0 52 /r VPDPWSSD xmm1{k1}{z}, xmm2, xmm3/m128/m32bcst  AVX512VNNI AVX512VL
	// EVEX.128.66.0F38.W0 53 /r VPDPWSSDS xmm1{k1}{z}, xmm2, xmm3/m128/m32bcst  AVX512VNNI AVX512VL
	{
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "xmm{k}{z}", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "AVX512VNNI"},
	},
	{
		Operands: []inst.Operand{
			{Type: "m128/m32bcst", Action: inst.R},
			{Type: "xmm", Action: inst.R},
			{Type: "xmm{k}{z}", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "AVX512VNNI"},
	},
	// EVEX.256.66.0F38.W0 50 /r VPDPBUSD ymm1{k1}{z}, ymm2, ymm3/m256/m32bcst  AVX512VNNI AVX512VL
	// EVEX.256.66.0F38.W0 51 /r VPDPBUSDS ymm1{k1}{z}, ymm2, ymm3/m256/m32bcst  AVX512VNNI AVX512VL
	// EVEX.256.66.0F38.W0 52 /r VPDPWSSD ymm1{k1}{z}, ymm2, ymm3/m256/m32bcst  AVX512VNNI AVX512VL
	// EVEX.256.66.0F38.W0 53 /r VPDPWSSDS ymm1{k1}{z}, ymm2, ymm3/m256/m32bcst  AVX512VNNI AVX512VL
	{
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm{k}{z}", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "AVX512VNNI"},
	},
	{
		Operands: []inst.Operand{
			{Type: "m256/m32bcst", Action: inst.R},
			{Type: "ymm", Action: inst.R},
			{Type: "ymm{k}{z}", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VL", "AVX512VNNI"},
	},
	// EVEX.512.66.0F38.W0 50 /r VPDPBUSD zmm1{k1}{z}, zmm2, zmm3/m512/m32bcst  AVX512VNNI
	// EVEX.512.66.0F38.W0 51 /r VPDPBUSDS zmm1{k1}{z}, zmm2, zmm3/m512/m32bcst  AVX512VNNI
	// EVEX.512.66.0F38.W0 52 /r VPDPWSSD zmm1{k1}{z}, zmm2, zmm3/m512/m32bcst  AVX512VNNI
	// EVEX.512.66.0F38.W0 53 /r VPDPWSSDS zmm1{k1}{z}, zmm2, zmm3/m512/m32bcst  AVX512VNNI
	{
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "zmm", Action: inst.R},
			{Type: "zmm{k}{z}", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VNNI"},
	},
	{
		Operands: []inst.Operand{
			{Type: "m512/m32bcst", Action: inst.R},
			{Type: "zmm", Action: inst.R},
			{Type: "zmm{k}{z}", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
		ISA:          []string{"AVX512VNNI"},
	},
}
