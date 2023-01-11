package opcodesextra

import (
	"github.com/mmcloughlin/avo/internal/inst"
)

// vbmi2 is the "Vector Bit Manipulation Instructions 2" instruction set.
var vbmi2 = []*inst.Instruction{
	// Insert: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3026-L3030
	// Insert: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3041-L3045
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
	// Insert: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3200-L3204
	// Insert: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3215-L3219
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
	// Insert: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3837-L3896
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
}

// VPCOMPRESSB and VPCOMPRESSW forms.
//
// Insert: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L240-L247
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
// Insert: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L376-L383
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
// Insert: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L128-L135
func vpshld(bcst string) inst.Forms {
	return inst.Forms{
		// EVEX.128.66.0F3A.W1 70 /r /ib VPSHLDW xmm1{k1}{z}, xmm2, xmm3/m128, imm8	A	V/V	AVX512_VBMI2 AVX512VL
		// EVEX.128.66.0F3A.W0 71 /r /ib VPSHLDD xmm1{k1}{z}, xmm2, xmm3/m128/m32bcst, imm8	B	V/V	AVX512_VBMI2 AVX512VL
		// EVEX.128.66.0F3A.W1 71 /r /ib VPSHLDQ xmm1{k1}{z}, xmm2, xmm3/m128/m64bcst, imm8	B	V/V	AVX512_VBMI2 AVX512VL
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
				{Type: "m512" + bcst, Action: inst.R},
				{Type: "zmm", Action: inst.R},
				{Type: "zmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
	}
}
