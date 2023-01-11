package opcodesextra

import (
	"github.com/mmcloughlin/avo/internal/inst"
)

// vbmi2 is the "Vector Bit Manipulation Instructions 2" instruction set.
var vbmi2 = []*inst.Instruction{
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3026-L3030
	//
	//		{as: AVPCOMPRESSB, ytab: _yvcompresspd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x63,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x63,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x63,
	//		}},
	//
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3041-L3045
	//
	//		{as: AVPCOMPRESSW, ytab: _yvcompresspd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x63,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x63,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x63,
	//		}},
	//
	{
		Opcode:  "VPCOMPRESSB",
		Summary: "Store Sparse Packed Byte Integer Values into Dense Memory/Register",
		Forms:   vpcompressb,
	},
	{
		Opcode:  "VPCOMPRESSW",
		Summary: "Store Sparse Packed Word Integer Values into Dense Memory/Register",
		Forms:   vpcompressb,
	},
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3200-L3204
	//
	//		{as: AVPEXPANDB, ytab: _yvexpandpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x62,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x62,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x62,
	//		}},
	//
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3215-L3219
	//
	//		{as: AVPEXPANDW, ytab: _yvexpandpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x62,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x62,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x62,
	//		}},
	//
	{
		Opcode:  "VPEXPANDB",
		Summary: "Load Sparse Packed Byte Integer Values from Dense Memory/Register",
		Forms:   vpexpandb,
	},
	{
		Opcode:  "VPEXPANDW",
		Summary: "Load Sparse Packed Word Integer Values from Dense Memory/Register",
		Forms:   vpexpandb,
	},
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3837-L3896
	//
	//		{as: AVPSHLDD, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F3A | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//			avxEscape | evex256 | evex66 | evex0F3A | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//			avxEscape | evex512 | evex66 | evex0F3A | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//		}},
	//		{as: AVPSHLDQ, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F3A | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//			avxEscape | evex256 | evex66 | evex0F3A | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//			avxEscape | evex512 | evex66 | evex0F3A | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//		}},
	//		{as: AVPSHLDVD, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//		}},
	//		{as: AVPSHLDVQ, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//		}},
	//		{as: AVPSHLDVW, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN16 | evexZeroingEnabled, 0x70,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN32 | evexZeroingEnabled, 0x70,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN64 | evexZeroingEnabled, 0x70,
	//		}},
	//		{as: AVPSHLDW, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F3A | evexW1, evexN16 | evexZeroingEnabled, 0x70,
	//			avxEscape | evex256 | evex66 | evex0F3A | evexW1, evexN32 | evexZeroingEnabled, 0x70,
	//			avxEscape | evex512 | evex66 | evex0F3A | evexW1, evexN64 | evexZeroingEnabled, 0x70,
	//		}},
	//		{as: AVPSHRDD, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F3A | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//			avxEscape | evex256 | evex66 | evex0F3A | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//			avxEscape | evex512 | evex66 | evex0F3A | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//		}},
	//		{as: AVPSHRDQ, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F3A | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//			avxEscape | evex256 | evex66 | evex0F3A | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//			avxEscape | evex512 | evex66 | evex0F3A | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//		}},
	//		{as: AVPSHRDVD, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//		}},
	//		{as: AVPSHRDVQ, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//		}},
	//		{as: AVPSHRDVW, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN16 | evexZeroingEnabled, 0x72,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN32 | evexZeroingEnabled, 0x72,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN64 | evexZeroingEnabled, 0x72,
	//		}},
	//		{as: AVPSHRDW, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F3A | evexW1, evexN16 | evexZeroingEnabled, 0x72,
	//			avxEscape | evex256 | evex66 | evex0F3A | evexW1, evexN32 | evexZeroingEnabled, 0x72,
	//			avxEscape | evex512 | evex66 | evex0F3A | evexW1, evexN64 | evexZeroingEnabled, 0x72,
	//		}},
	//
	{
		Opcode:  "VPSHLDW",
		Summary: "Concatenate Words and Shift Packed Data Left Logical",
		Forms:   vpshld(""),
	},
	{
		Opcode:  "VPSHLDD",
		Summary: "Concatenate Dwords and Shift Packed Data Left Logical",
		Forms:   vpshld("/m32bcst"),
	},
	{
		Opcode:  "VPSHLDQ",
		Summary: "Concatenate Quadwords and Shift Packed Data Left Logical",
		Forms:   vpshld("/m64bcst"),
	},
	{
		Opcode:  "VPSHRDW",
		Summary: "Concatenate Words and Shift Packed Data Right Logical",
		Forms:   vpshld(""),
	},
	{
		Opcode:  "VPSHRDD",
		Summary: "Concatenate Dwords and Shift Packed Data Right Logical",
		Forms:   vpshld("/m32bcst"),
	},
	{
		Opcode:  "VPSHRDQ",
		Summary: "Concatenate Quadwords and Shift Packed Data Right Logical",
		Forms:   vpshld("/m64bcst"),
	},
	{
		Opcode:  "VPSHLDVW",
		Summary: "Concatenate Words and Variable Shift Packed Data Left Logical",
		Forms:   vpshldv(""),
	},
	{
		Opcode:  "VPSHLDVD",
		Summary: "Concatenate Dwords and Variable Shift Packed Data Left Logical",
		Forms:   vpshldv("/m32bcst"),
	},
	{
		Opcode:  "VPSHLDVQ",
		Summary: "Concatenate Quadwords and Variable Shift Packed Data Left Logical",
		Forms:   vpshldv("/m64bcst"),
	},
	{
		Opcode:  "VPSHRDVW",
		Summary: "Concatenate Words and Variable Shift Packed Data Right Logical",
		Forms:   vpshldv(""),
	},
	{
		Opcode:  "VPSHRDVD",
		Summary: "Concatenate Dwords and Variable Shift Packed Data Right Logical",
		Forms:   vpshldv("/m32bcst"),
	},
	{
		Opcode:  "VPSHRDVQ",
		Summary: "Concatenate Quadwords and Variable Shift Packed Data Right Logical",
		Forms:   vpshldv("/m64bcst"),
	},
}

// VPCOMPRESSB and VPCOMPRESSW forms.
//
// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L240-L247
//
//	var _yvcompresspd = []ytab{
//		{zcase: Zevex_r_v_rm, zoffset: 0, args: argList{YxrEvex, YxmEvex}},
//		{zcase: Zevex_r_k_rm, zoffset: 3, args: argList{YxrEvex, Yknot0, YxmEvex}},
//		{zcase: Zevex_r_v_rm, zoffset: 0, args: argList{YyrEvex, YymEvex}},
//		{zcase: Zevex_r_k_rm, zoffset: 3, args: argList{YyrEvex, Yknot0, YymEvex}},
//		{zcase: Zevex_r_v_rm, zoffset: 0, args: argList{Yzr, Yzm}},
//		{zcase: Zevex_r_k_rm, zoffset: 3, args: argList{Yzr, Yknot0, Yzm}},
//	}
var vpcompressb = inst.Forms{
	// EVEX.128.66.0F38.W0 63 /r VPCOMPRESSB m128{k1}, xmm1	A	V/V	AVX512_VBMI2 AVX512VL
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "m128{k}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.128.66.0F38.W0 63 /r VPCOMPRESSB xmm1{k1}{z}, xmm2	B	V/V	AVX512_VBMI2 AVX512VL
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "xmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.256.66.0F38.W0 63 /r VPCOMPRESSB m256{k1}, ymm1	A	V/V	AVX512_VBMI2 AVX512VL
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "m256{k}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.256.66.0F38.W0 63 /r VPCOMPRESSB ymm1{k1}{z}, ymm2	B	V/V	AVX512_VBMI2 AVX512VL
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "ymm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.512.66.0F38.W0 63 /r VPCOMPRESSB m512{k1}, zmm1	A	V/V	AVX512_VBMI2
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "m512{k}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.512.66.0F38.W0 63 /r VPCOMPRESSB zmm1{k1}{z}, zmm2	B	V/V	AVX512_VBMI2
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "zmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
}

// VPEXPANDB and VPEXPANDW forms.
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
var vpexpandb = inst.Forms{
	// EVEX.128.66.0F38.W0 62 /r VPEXPANDB xmm1{k1}{z}, m128	A	V/V	AVX512_VBMI2 AVX512VL
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "m128", Action: inst.R},
			{Type: "xmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.128.66.0F38.W0 62 /r VPEXPANDB xmm1{k1}{z}, xmm2	B	V/V	AVX512_VBMI2 AVX512VL
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "xmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.256.66.0F38.W0 62 /r VPEXPANDB ymm1{k1}{z}, m256	A	V/V	AVX512_VBMI2 AVX512VL
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "m256", Action: inst.R},
			{Type: "ymm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.256.66.0F38.W0 62 /r VPEXPANDB ymm1{k1}{z}, ymm2	B	V/V	AVX512_VBMI2 AVX512VL
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "ymm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.512.66.0F38.W0 62 /r VPEXPANDB zmm1{k1}{z}, m512	A	V/V	AVX512_VBMI2
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "m512", Action: inst.R},
			{Type: "zmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	// EVEX.512.66.0F38.W0 62 /r VPEXPANDB zmm1{k1}{z}, zmm2	B	V/V	AVX512_VBMI2
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "zmm{k}{z}", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
}

// VPSH{L,R}D{W,D,Q} forms.
//
// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L128-L135
//
//	var _yvalignd = []ytab{
//		{zcase: Zevex_i_rm_v_r, zoffset: 0, args: argList{Yu8, YxmEvex, YxrEvex, YxrEvex}},
//		{zcase: Zevex_i_rm_v_k_r, zoffset: 3, args: argList{Yu8, YxmEvex, YxrEvex, Yknot0, YxrEvex}},
//		{zcase: Zevex_i_rm_v_r, zoffset: 0, args: argList{Yu8, YymEvex, YyrEvex, YyrEvex}},
//		{zcase: Zevex_i_rm_v_k_r, zoffset: 3, args: argList{Yu8, YymEvex, YyrEvex, Yknot0, YyrEvex}},
//		{zcase: Zevex_i_rm_v_r, zoffset: 0, args: argList{Yu8, Yzm, Yzr, Yzr}},
//		{zcase: Zevex_i_rm_v_k_r, zoffset: 3, args: argList{Yu8, Yzm, Yzr, Yknot0, Yzr}},
//	}
func vpshld(bcst string) inst.Forms {
	return inst.Forms{
		// EVEX.128.66.0F3A.W1 70 /r /ib VPSHLDW xmm1{k1}{z}, xmm2, xmm3/m128, imm8	A	V/V	AVX512_VBMI2 AVX512VL
		// EVEX.128.66.0F3A.W0 71 /r /ib VPSHLDD xmm1{k1}{z}, xmm2, xmm3/m128/m32bcst, imm8	B	V/V	AVX512_VBMI2 AVX512VL
		// EVEX.128.66.0F3A.W1 71 /r /ib VPSHLDQ xmm1{k1}{z}, xmm2, xmm3/m128/m64bcst, imm8	B	V/V	AVX512_VBMI2 AVX512VL
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8"},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8"},
				{Type: "m128" + bcst, Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		// EVEX.256.66.0F3A.W1 70 /r /ib VPSHLDW ymm1{k1}{z}, ymm2, ymm3/m256, imm8	A	V/V	AVX512_VBMI2 AVX512VL
		// EVEX.256.66.0F3A.W0 71 /r /ib VPSHLDD ymm1{k1}{z}, ymm2, ymm3/m256/m32bcst, imm8	B	V/V	AVX512_VBMI2 AVX512VL
		// EVEX.256.66.0F3A.W1 71 /r /ib VPSHLDQ ymm1{k1}{z}, ymm2, ymm3/m256/m64bcst, imm8	B	V/V	AVX512_VBMI2 AVX512VL
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8"},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8"},
				{Type: "m256" + bcst, Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		// EVEX.512.66.0F3A.W1 70 /r /ib VPSHLDW zmm1{k1}{z}, zmm2, zmm3/m512, imm8	A	V/V	AVX512_VBMI2
		// EVEX.512.66.0F3A.W0 71 /r /ib VPSHLDD zmm1{k1}{z}, zmm2, zmm3/m512/m32bcst, imm8	B	V/V	AVX512_VBMI2
		// EVEX.512.66.0F3A.W1 71 /r /ib VPSHLDQ zmm1{k1}{z}, zmm2, zmm3/m512/m64bcst, imm8	B	V/V	AVX512_VBMI2
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "imm8"},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "imm8"},
				{Type: "m512" + bcst, Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
	}
}

// VPSH{L,R}DV{W,D,Q} forms.
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
func vpshldv(bcst string) inst.Forms {
	return inst.Forms{
		// EVEX.128.66.0F38.W1 70 /r VPSHLDVW xmm1{k1}{z}, xmm2, xmm3/m128	A	V/V	AVX512_VBMI2 AVX512VL
		// EVEX.128.66.0F38.W0 71 /r VPSHLDVD xmm1{k1}{z}, xmm2, xmm3/m128/m32bcst	B	V/V	AVX512_VBMI2 AVX512VL
		// EVEX.128.66.0F38.W1 71 /r VPSHLDVQ xmm1{k1}{z}, xmm2, xmm3/m128/m64bcst	B	V/V	AVX512_VBMI2 AVX512VL
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "m128" + bcst, Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		// EVEX.256.66.0F38.W1 70 /r VPSHLDVW ymm1{k1}{z}, ymm2, ymm3/m256	A	V/V	AVX512_VBMI2 AVX512VL
		// EVEX.256.66.0F38.W0 71 /r VPSHLDVD ymm1{k1}{z}, ymm2, ymm3/m256/m32bcst	B	V/V	AVX512_VBMI2 AVX512VL
		// EVEX.256.66.0F38.W1 71 /r VPSHLDVQ ymm1{k1}{z}, ymm2, ymm3/m256/m64bcst	B	V/V	AVX512_VBMI2 AVX512VL
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "m256" + bcst, Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		// EVEX.512.66.0F38.W1 70 /r VPSHLDVW zmm1{k1}{z}, zmm2, zmm3/m512	A	V/V	AVX512_VBMI2
		// EVEX.512.66.0F38.W0 71 /r VPSHLDVD zmm1{k1}{z}, zmm2, zmm3/m512/m32bcst	B	V/V	AVX512_VBMI2
		// EVEX.512.66.0F38.W1 71 /r VPSHLDVQ zmm1{k1}{z}, zmm2, zmm3/m512/m64bcst	B	V/V	AVX512_VBMI2
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "m512" + bcst, Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
	}
}
