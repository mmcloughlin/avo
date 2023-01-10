package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// vpopcntdq adds the VPOPCNTDQ instructions including the AVX512_VL "variable
// length" (128/256-bit wide) forms of VPOPCNTD and VPOPCNTQ to the instruction set.
var vpopcntdq = []*inst.Instruction{
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3741-L3750
	//
	//		{as: AVPOPCNTD, ytab: _yvexpandpd, prefix: Pavx, op: opBytes{
	// 			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x55,
	// 			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x55,
	// 			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x55,
	// 		}},
	// 		{as: AVPOPCNTQ, ytab: _yvexpandpd, prefix: Pavx, op: opBytes{
	// 			avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0x55,
	// 			avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0x55,
	// 			avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0x55,
	// 		}},
	//
	// *** NOTE! The evex512 forms above already exist in x86_64.xml, so they are overridden by the definitions below. ***
	{
		Opcode:  "VPOPCNTD",
		Summary: "Packed Population Count for Doubleword Integers",
		// VPOPCNTD including AVX512_VL forms.
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
		//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{Yzm, Yzr}}, 					/* OVERRIDES x86_64.xml */
		//		{zcase: Zevex_rm_k_r, zoffset: 3, args: argList{Yzm, Yknot0, Yzr}},				/* OVERRIDES x86_64.xml */
		//	}
		Forms: inst.Forms{
			// EVEX.128.66.0F38.W0 55 /r VPOPCNTD xmm1{k1}{z}, xmm2/m128/m32bcst
			// EVEX.256.66.0F38.W0 55 /r VPOPCNTD ymm1{k1}{z}, ymm2/m256/m32bcst
			// EVEX.512.66.0F38.W0 55 /r VPOPCNTD zmm1{k1}{z}, zmm2/m512/m32bcst  /* OVERRIDES x86_64.xml */
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m128", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m128", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m256", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m256", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m32", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m32", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m32", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m32", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m32", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m32", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m32", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m32", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m32", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m512", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m512", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m512", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
		},
	},
	{
		Opcode:  "VPOPCNTQ",
		Summary: "Packed Population Count for Quadword Integers",
		// VPOPCNTQ including AVX512_VL forms.
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
		//		{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{Yzm, Yzr}}, 					/* OVERRIDES x86_64.xml */
		//		{zcase: Zevex_rm_k_r, zoffset: 3, args: argList{Yzm, Yknot0, Yzr}},				/* OVERRIDES x86_64.xml */
		//	}
		Forms: inst.Forms{
			// EVEX.128.66.0F38.W1 55 /r VPOPCNTQ xmm1{k1}{z}, xmm2/m128/m64bcst
			// EVEX.256.66.0F38.W1 55 /r VPOPCNTQ ymm1{k1}{z}, ymm2/m256/m64bcst
			// EVEX.512.66.0F38.W1 55 /r VPOPCNTQ zmm1{k1}{z}, zmm2/m512/m64bcst  /* OVERRIDES x86_64.xml */
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m128", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m128", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m128", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m256", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m256", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m256", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m64", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m64", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m64", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m64", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m64", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m64", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "xmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "xmm", Action: inst.R},
					{Type: "xmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "ymm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VL", "AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "ymm", Action: inst.R},
					{Type: "ymm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m512", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m512", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m512", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m64", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m64", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "m64", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Broadcast:    true,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.RW},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "zmm", Action: inst.R},
					{Type: "k", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
				Zeroing:      true,
			},
			{
				ISA: []string{"AVX512VPOPCNTDQ"},
				Operands: []inst.Operand{
					{Type: "zmm", Action: inst.R},
					{Type: "zmm", Action: inst.W},
				},
				EncodingType: inst.EncodingTypeEVEX,
			},
		},
	},
}
