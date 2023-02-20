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
var vnniforms = _yvblendmpd("AVX512VNNI", "/m32bcst", inst.W)
