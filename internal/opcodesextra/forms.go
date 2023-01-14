package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

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
func _yvexpandpd(isa, bcst string) inst.Forms {
	return inst.Forms{
		// EVEX.128.66.0F38.W0 62 /r VPEXPANDB xmm1{k1}{z}, m128	A	V/V	AVX512_VBMI2 AVX512VL
		{
			ISA: []string{isa, "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "m128" + bcst, Action: inst.R},
				{Type: "xmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		// EVEX.128.66.0F38.W0 62 /r VPEXPANDB xmm1{k1}{z}, xmm2	B	V/V	AVX512_VBMI2 AVX512VL
		{
			ISA: []string{isa, "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "xmm", Action: inst.R},
				{Type: "xmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		// EVEX.256.66.0F38.W0 62 /r VPEXPANDB ymm1{k1}{z}, m256	A	V/V	AVX512_VBMI2 AVX512VL
		{
			ISA: []string{isa, "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "m256" + bcst, Action: inst.R},
				{Type: "ymm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		// EVEX.256.66.0F38.W0 62 /r VPEXPANDB ymm1{k1}{z}, ymm2	B	V/V	AVX512_VBMI2 AVX512VL
		{
			ISA: []string{isa, "AVX512VL"},
			Operands: []inst.Operand{
				{Type: "ymm", Action: inst.R},
				{Type: "ymm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		// EVEX.512.66.0F38.W0 62 /r VPEXPANDB zmm1{k1}{z}, m512	A	V/V	AVX512_VBMI2
		{
			ISA: []string{isa},
			Operands: []inst.Operand{
				{Type: "m512" + bcst, Action: inst.R},
				{Type: "zmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
		// EVEX.512.66.0F38.W0 62 /r VPEXPANDB zmm1{k1}{z}, zmm2	B	V/V	AVX512_VBMI2
		{
			ISA: []string{isa},
			Operands: []inst.Operand{
				{Type: "zmm", Action: inst.R},
				{Type: "zmm{k}{z}", Action: inst.W},
			},
			EncodingType: inst.EncodingTypeEVEX,
		},
	}
}
