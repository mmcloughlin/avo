package opcodesextra

import (
	"fmt"

	"github.com/mmcloughlin/avo/internal/inst"
)

// vbmi2 is the "Vector Bit Manipulation Instructions 2" instruction set.
var vbmi2 = []*inst.Instruction{
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3026-L3030
	//
	//	{as: AVPCOMPRESSB, ytab: _yvcompresspd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x63,
	//		avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x63,
	//		avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x63,
	//	}},
	//
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3041-L3045
	//
	//	{as: AVPCOMPRESSW, ytab: _yvcompresspd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x63,
	//		avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x63,
	//		avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x63,
	//	}},
	{
		Opcode:  "VPCOMPRESSB",
		Summary: "Store Sparse Packed Byte Integer Values into Dense Memory/Register",
		Forms:   vbmi2Compress,
	},
	{
		Opcode:  "VPCOMPRESSW",
		Summary: "Store Sparse Packed Word Integer Values into Dense Memory/Register",
		Forms:   vbmi2Compress,
	},
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3200-L3204
	//
	//	{as: AVPEXPANDB, ytab: _yvexpandpd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x62,
	//		avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x62,
	//		avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN1 | evexZeroingEnabled, 0x62,
	//	}},
	//
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3215-L3219
	//
	//	{as: AVPEXPANDW, ytab: _yvexpandpd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x62,
	//		avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x62,
	//		avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN2 | evexZeroingEnabled, 0x62,
	//	}},
	{
		Opcode:  "VPEXPANDB",
		Summary: "Load Sparse Packed Byte Integer Values from Dense Memory/Register",
		Forms:   vbmi2Expand,
	},
	{
		Opcode:  "VPEXPANDW",
		Summary: "Load Sparse Packed Word Integer Values from Dense Memory/Register",
		Forms:   vbmi2Expand,
	},

	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3837-L3896
	//
	//	{as: AVPSHLDD, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F3A | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//		avxEscape | evex256 | evex66 | evex0F3A | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//		avxEscape | evex512 | evex66 | evex0F3A | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//	}},
	//	{as: AVPSHLDQ, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F3A | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//		avxEscape | evex256 | evex66 | evex0F3A | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//		avxEscape | evex512 | evex66 | evex0F3A | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//	}},
	//	{as: AVPSHLDVD, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//		avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//		avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x71,
	//	}},
	//	{as: AVPSHLDVQ, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//		avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//		avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0x71,
	//	}},
	//	{as: AVPSHLDVW, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN16 | evexZeroingEnabled, 0x70,
	//		avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN32 | evexZeroingEnabled, 0x70,
	//		avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN64 | evexZeroingEnabled, 0x70,
	//	}},
	//	{as: AVPSHLDW, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F3A | evexW1, evexN16 | evexZeroingEnabled, 0x70,
	//		avxEscape | evex256 | evex66 | evex0F3A | evexW1, evexN32 | evexZeroingEnabled, 0x70,
	//		avxEscape | evex512 | evex66 | evex0F3A | evexW1, evexN64 | evexZeroingEnabled, 0x70,
	//	}},
	//	{as: AVPSHRDD, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F3A | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//		avxEscape | evex256 | evex66 | evex0F3A | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//		avxEscape | evex512 | evex66 | evex0F3A | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//	}},
	//	{as: AVPSHRDQ, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F3A | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//		avxEscape | evex256 | evex66 | evex0F3A | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//		avxEscape | evex512 | evex66 | evex0F3A | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//	}},
	//	{as: AVPSHRDVD, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//		avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//		avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x73,
	//	}},
	//	{as: AVPSHRDVQ, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//		avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//		avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0x73,
	//	}},
	//	{as: AVPSHRDVW, ytab: _yvblendmpd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN16 | evexZeroingEnabled, 0x72,
	//		avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN32 | evexZeroingEnabled, 0x72,
	//		avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN64 | evexZeroingEnabled, 0x72,
	//	}},
	//	{as: AVPSHRDW, ytab: _yvalignd, prefix: Pavx, op: opBytes{
	//		avxEscape | evex128 | evex66 | evex0F3A | evexW1, evexN16 | evexZeroingEnabled, 0x72,
	//		avxEscape | evex256 | evex66 | evex0F3A | evexW1, evexN32 | evexZeroingEnabled, 0x72,
	//		avxEscape | evex512 | evex66 | evex0F3A | evexW1, evexN64 | evexZeroingEnabled, 0x72,
	//	}},
	{
		Opcode:  "VPSHLDD",
		Summary: "Concatenate Dwords and Shift Packed Data Left Logical",
		Forms:   immShift(32),
	},
	{
		Opcode:  "VPSHLDQ",
		Summary: "Concatenate Quadwords and Shift Packed Data Left Logical",
		Forms:   immShift(64),
	},
	{
		Opcode:  "VPSHLDVD",
		Summary: "Concatenate Dwords and Variable Shift Packed Data Left Logical",
		Forms:   varShift(32),
	},
	{
		Opcode:  "VPSHLDVQ",
		Summary: "Concatenate Quadwords and Variable Shift Packed Data Left Logical",
		Forms:   varShift(64),
	},
	{
		Opcode:  "VPSHLDVW",
		Summary: "Concatenate Words and Variable Shift Packed Data Left Logical",
		Forms:   varShift(16),
	},
	{
		Opcode:  "VPSHLDW",
		Summary: "Concatenate Words and Shift Packed Data Left Logical",
		Forms:   immShift(16),
	},
	{
		Opcode:  "VPSHRDD",
		Summary: "Concatenate Dwords and Shift Packed Data Right Logical",
		Forms:   immShift(32),
	},
	{
		Opcode:  "VPSHRDQ",
		Summary: "Concatenate Quadwords and Shift Packed Data Right Logical",
		Forms:   immShift(64),
	},
	{
		Opcode:  "VPSHRDVD",
		Summary: "Concatenate Dwords and Variable Shift Packed Data Right Logical",
		Forms:   varShift(32),
	},
	{
		Opcode:  "VPSHRDVQ",
		Summary: "Concatenate Quadwords and Variable Shift Packed Data Right Logical",
		Forms:   varShift(64),
	},
	{
		Opcode:  "VPSHRDVW",
		Summary: "Concatenate Words and Variable Shift Packed Data Right Logical",
		Forms:   varShift(16),
	},
	{
		Opcode:  "VPSHRDW",
		Summary: "Concatenate Words and Shift Packed Data Right Logical",
		Forms:   immShift(16),
	},
}

// VPCOMPRESSB and VPCOMPRESSW forms
//
// See: https://www.felixcloutier.com/x86/vpcompressb:vcompressw
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
var vbmi2Compress = inst.Forms{
	// EVEX.128.66.0F38.W0 63 /r VPCOMPRESSB m128{k1}, xmm1  AVX512VBMI2 AVX512VL
	// EVEX.128.66.0F38.W1 63 /r VPCOMPRESSW m128{k1}, xmm1  AVX512VBMI2 AVX512VL
	// EVEX.128.66.0F38.W0 63 /r VPCOMPRESSB xmm1{k1}{z}, xmm2  AVX512VBMI2 AVX512VL
	// EVEX.128.66.0F38.W1 63 /r VPCOMPRESSW xmm1{k1}{z}, xmm2  AVX512VBMI2 AVX512VL
	// EVEX.256.66.0F38.W0 63 /r VPCOMPRESSB m256{k1}, ymm1  AVX512VBMI2 AVX512VL
	// EVEX.256.66.0F38.W1 63 /r VPCOMPRESSW m256{k1}, ymm1  AVX512VBMI2 AVX512VL
	// EVEX.256.66.0F38.W0 63 /r VPCOMPRESSB ymm1{k1}{z}, ymm2  AVX512VBMI2 AVX512VL
	// EVEX.256.66.0F38.W1 63 /r VPCOMPRESSW ymm1{k1}{z}, ymm2  AVX512VBMI2 AVX512VL
	// EVEX.512.66.0F38.W0 63 /r VPCOMPRESSB m512{k1}, zmm1  AVX512VBMI2
	// EVEX.512.66.0F38.W1 63 /r VPCOMPRESSW m512{k1}, zmm1  AVX512VBMI2
	// EVEX.512.66.0F38.W0 63 /r VPCOMPRESSB zmm1{k1}{z}, zmm2  AVX512VBMI2
	// EVEX.512.66.0F38.W1 63 /r VPCOMPRESSW zmm1{k1}{z}, zmm2  AVX512VBMI2
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "m128", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "m128", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "xmm", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "xmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "m128", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "m256", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "m256", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "ymm", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "ymm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "m256", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "m512", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "m512", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "zmm", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "zmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "m512", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "zmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
}

// VPEXPANDB and VPEXPANDW forms
//
// See: https://www.felixcloutier.com/x86/vpexpandb:vpexpandw
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
var vbmi2Expand = inst.Forms{
	// EVEX.128.66.0F38.W0 62 /r VPEXPANDB xmm1{k1}{z}, m128  AVX512VBMI2 AVX512VL
	// EVEX.128.66.0F38.W1 62 /r VPEXPANDW xmm1{k1}{z}, m128  AVX512VBMI2 AVX512VL
	// EVEX.128.66.0F38.W0 62 /r VPEXPANDB xmm1{k1}{z}, xmm2  AVX512VBMI2 AVX512VL
	// EVEX.128.66.0F38.W1 62 /r VPEXPANDW xmm1{k1}{z}, xmm2  AVX512VBMI2 AVX512VL
	// EVEX.256.66.0F38.W0 62 /r VPEXPANDB ymm1{k1}{z}, m256  AVX512VBMI2 AVX512VL
	// EVEX.256.66.0F38.W1 62 /r VPEXPANDW ymm1{k1}{z}, m256  AVX512VBMI2 AVX512VL
	// EVEX.256.66.0F38.W0 62 /r VPEXPANDB ymm1{k1}{z}, ymm2  AVX512VBMI2 AVX512VL
	// EVEX.256.66.0F38.W1 62 /r VPEXPANDW ymm1{k1}{z}, ymm2  AVX512VBMI2 AVX512VL
	// EVEX.512.66.0F38.W0 62 /r VPEXPANDB zmm1{k1}{z}, m512  AVX512VBMI2
	// EVEX.512.66.0F38.W1 62 /r VPEXPANDW zmm1{k1}{z}, m512  AVX512VBMI2
	// EVEX.512.66.0F38.W0 62 /r VPEXPANDB zmm1{k1}{z}, zmm2  AVX512VBMI2
	// EVEX.512.66.0F38.W1 62 /r VPEXPANDW zmm1{k1}{z}, zmm2  AVX512VBMI2
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "m128", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "xmm", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "m128", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "xmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "m128", Action: inst.R},
			{Type: "xmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "m256", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "ymm", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "m256", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "ymm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "m256", Action: inst.R},
			{Type: "ymm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "xmm", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "xmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "xmm", Action: inst.R},
			{Type: "xmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "ymm", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "ymm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2", "AVX512VL"},
		Operands: []inst.Operand{
			{Type: "ymm", Action: inst.R},
			{Type: "ymm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "m512", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "zmm", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "m512", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "zmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "m512", Action: inst.R},
			{Type: "zmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "zmm", Action: inst.RW},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "k", Action: inst.R},
			{Type: "zmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
		Zeroing:      true,
	},
	{
		ISA: []string{"AVX512VBMI2"},
		Operands: []inst.Operand{
			{Type: "zmm", Action: inst.R},
			{Type: "zmm", Action: inst.W},
		},
		EncodingType: inst.EncodingTypeEVEX,
	},
}

// VPSHLDD, VPSHLDQ, VPSHLDW, VPSHRDD, VPSHRDQ and VPSHRDW forms
//
// See:	https://www.felixcloutier.com/x86/vpshld
// https://www.felixcloutier.com/x86/vpshrd
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
func immShift(bcastWidth int) inst.Forms {
	// EVEX.128.66.0F3A.W1 70 /r /ib VPSHLDW xmm1{k1}{z}, xmm2, xmm3/m128, imm8  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F3A.W1 70 /r /ib VPSHLDW ymm1{k1}{z}, ymm2, ymm3/m256, imm8  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F3A.W1 70 /r /ib VPSHLDW zmm1{k1}{z}, zmm2, zmm3/m512, imm8  AVX512VMBI2
	// EVEX.128.66.0F3A.W0 71 /r /ib VPSHLDD xmm1{k1}{z}, xmm2, xmm3/m128/m32bcst, imm8  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F3A.W0 71 /r /ib VPSHLDD ymm1{k1}{z}, ymm2, ymm3/m256/m32bcst, imm8  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F3A.W0 71 /r /ib VPSHLDD zmm1{k1}{z}, zmm2, zmm3/m512/m32bcst, imm8  AVX512VMBI2
	// EVEX.128.66.0F3A.W1 71 /r /ib VPSHLDQ xmm1{k1}{z}, xmm2, xmm3/m128/m64bcst, imm8  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F3A.W1 71 /r /ib VPSHLDQ ymm1{k1}{z}, ymm2, ymm3/m256/m64bcst, imm8  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F3A.W1 71 /r /ib VPSHLDQ zmm1{k1}{z}, zmm2, zmm3/m512/m64bcst, imm8  AVX512VMBI2
	// EVEX.128.66.0F3A.W1 72 /r /ib VPSHRDW xmm1{k1}{z}, xmm2, xmm3/m128, imm8  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F3A.W1 72 /r /ib VPSHRDW ymm1{k1}{z}, ymm2, ymm3/m256, imm8  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F3A.W1 72 /r /ib VPSHRDW zmm1{k1}{z}, zmm2, zmm3/m512, imm8  AVX512VMBI2
	// EVEX.128.66.0F3A.W0 73 /r /ib VPSHRDD xmm1{k1}{z}, xmm2, xmm3/m128/m32bcst, imm8  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F3A.W0 73 /r /ib VPSHRDD ymm1{k1}{z}, ymm2, ymm3/m256/m32bcst, imm8  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F3A.W0 73 /r /ib VPSHRDD zmm1{k1}{z}, zmm2, zmm3/m512/m32bcst, imm8  AVX512VMBI2
	// EVEX.128.66.0F3A.W1 73 /r /ib VPSHRDQ xmm1{k1}{z}, xmm2, xmm3/m128/m64bcst, imm8  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F3A.W1 73 /r /ib VPSHRDQ ymm1{k1}{z}, ymm2, ymm3/m256/m64bcst, imm8  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F3A.W1 73 /r /ib VPSHRDQ zmm1{k1}{z}, zmm2, zmm3/m512/m64bcst, imm8  AVX512VMBI2

	ret := inst.Forms{
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "m128", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "xmm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "m128", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "xmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "m128", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "m256", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "ymm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "m256", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "ymm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "m256", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "xmm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "xmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "ymm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "ymm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "m512", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "zmm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "m512", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "zmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "m512", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "zmm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "zmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "imm8", Action: inst.I},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
	}

	if bcastWidth > 16 {
		bcastMem := fmt.Sprintf("m%d", bcastWidth)
		ret = append(ret, inst.Forms{
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.I},
					{Type: bcastMem, Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.I},
					{Type: bcastMem, Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.I},
					{Type: bcastMem, Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.I},
					{Type: bcastMem, Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.I},
					{Type: bcastMem, Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.I},
					{Type: bcastMem, Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2"},
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.I},
					{Type: bcastMem, Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2"},
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.I},
					{Type: bcastMem, Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2"},
				Operands: []inst.Operand{
					{Type: "imm8", Action: inst.I},
					{Type: bcastMem, Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
		}...)
	}
	return ret
}

// VPSHLDVD, VPSHLDVQ, VPSHLDVW, VPSHRDVD, VPSHRDVQ and VPSHRDVW forms
//
// See:	https://www.felixcloutier.com/x86/vpshldv
// https://www.felixcloutier.com/x86/vpshrdv
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
func varShift(bcastWidth int) inst.Forms {
	// EVEX.128.66.0F38.W1 70 /r VPSHLDVW xmm1{k1}{z}, xmm2, xmm3/m128  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F38.W1 70 /r VPSHLDVW ymm1{k1}{z}, ymm2, ymm3/m256  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F38.W1 70 /r VPSHLDVW zmm1{k1}{z}, zmm2, zmm3/m512  AVX512VMBI2
	// EVEX.128.66.0F38.W0 71 /r VPSHLDVD xmm1{k1}{z}, xmm2, xmm3/m128/m32bcst  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F38.W0 71 /r VPSHLDVD ymm1{k1}{z}, ymm2, ymm3/m256/m32bcst  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F38.W0 71 /r VPSHLDVD zmm1{k1}{z}, zmm2, zmm3/m512/m32bcst  AVX512VMBI2
	// EVEX.128.66.0F38.W1 71 /r VPSHLDVQ xmm1{k1}{z}, xmm2, xmm3/m128/m64bcst  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F38.W1 71 /r VPSHLDVQ ymm1{k1}{z}, ymm2, ymm3/m256/m64bcst  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F38.W1 71 /r VPSHLDVQ zmm1{k1}{z}, zmm2, zmm3/m512/m64bcst  AVX512VMBI2
	// EVEX.128.66.0F38.W1 72 /r VPSHRDVW xmm1{k1}{z}, xmm2, xmm3/m128  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F38.W1 72 /r VPSHRDVW ymm1{k1}{z}, ymm2, ymm3/m256  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F38.W1 72 /r VPSHRDVW zmm1{k1}{z}, zmm2, zmm3/m512  AVX512VMBI2
	// EVEX.128.66.0F38.W0 73 /r VPSHRDVD xmm1{k1}{z}, xmm2, xmm3/m128/m32bcst  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F38.W0 73 /r VPSHRDVD ymm1{k1}{z}, ymm2, ymm3/m256/m32bcst  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F38.W0 73 /r VPSHRDVD zmm1{k1}{z}, zmm2, zmm3/m512/m32bcst  AVX512VMBI2
	// EVEX.128.66.0F38.W1 73 /r VPSHRDVQ xmm1{k1}{z}, xmm2, xmm3/m128/m64bcst  AVX512VMBI2 AVX512VL
	// EVEX.256.66.0F38.W1 73 /r VPSHRDVQ ymm1{k1}{z}, ymm2, ymm3/m256/m64bcst  AVX512VMBI2 AVX512VL
	// EVEX.512.66.0F38.W1 73 /r VPSHRDVQ zmm1{k1}{z}, zmm2, zmm3/m512/m64bcst  AVX512VMBI2
	ret := inst.Forms{
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "m128", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "xmm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "m128", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "xmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "m128", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "m256", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "ymm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "m256", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "ymm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "m256", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "xmm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "xmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.R},
				{Type: "xmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "ymm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "ymm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2", "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.R},
				{Type: "ymm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "m512", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "zmm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "m512", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "zmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "m512", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "zmm", Action: inst.RW},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "k", Action: inst.R},
				{Type: "zmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
			Zeroing:      true,
		},
		{
			ISA: []string{"AVX512VBMI2"},
			Operands: []inst.Operand{
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
	}

	if bcastWidth > 16 {
		bcastMem := fmt.Sprintf("m%d", bcastWidth)
		ret = append(ret, inst.Forms{
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: bcastMem, Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: bcastMem, Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: bcastMem, Action: inst.R},
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: bcastMem, Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: bcastMem, Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2", "AVX512VL"},
				Operands: []inst.Operand{
					{Type: bcastMem, Action: inst.R},
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2"},
				Operands: []inst.Operand{
					{Type: bcastMem, Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2"},
				Operands: []inst.Operand{
					{Type: bcastMem, Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VBMI2"},
				Operands: []inst.Operand{
					{Type: bcastMem, Action: inst.R},
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
		}...)
	}
	return ret
}
