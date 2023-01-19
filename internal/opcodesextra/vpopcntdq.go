package opcodesextra

import "github.com/mmcloughlin/avo/internal/inst"

// vpopcntdq is the "Vector Population Count" instruction set.
var vpopcntdq = []*inst.Instruction{
	// Reference: https://github.com/golang/go/blob/go1.19.3/src/cmd/internal/obj/x86/avx_optabs.go#L3741-L3750
	//
	//		{as: AVPOPCNTD, ytab: _yvexpandpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW0, evexN16 | evexBcstN4 | evexZeroingEnabled, 0x55,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW0, evexN32 | evexBcstN4 | evexZeroingEnabled, 0x55,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW0, evexN64 | evexBcstN4 | evexZeroingEnabled, 0x55,
	//		}},
	//		{as: AVPOPCNTQ, ytab: _yvexpandpd, prefix: Pavx, op: opBytes{
	//			avxEscape | evex128 | evex66 | evex0F38 | evexW1, evexN16 | evexBcstN8 | evexZeroingEnabled, 0x55,
	//			avxEscape | evex256 | evex66 | evex0F38 | evexW1, evexN32 | evexBcstN8 | evexZeroingEnabled, 0x55,
	//			avxEscape | evex512 | evex66 | evex0F38 | evexW1, evexN64 | evexBcstN8 | evexZeroingEnabled, 0x55,
	//		}},
	//
	// Note the opcodes database already contains the non-AVX512VL forms. They
	// are overridden by the definitions below.
	{
		Opcode:  "VPOPCNTD",
		Summary: "Packed Population Count for Doubleword Integers",
		Forms:   vpopcntdqforms("/m32bcst"),
	},
	{
		Opcode:  "VPOPCNTQ",
		Summary: "Packed Population Count for Quadword Integers",
		Forms:   vpopcntdqforms("/m64bcst"),
	},
}

// VPOPCNTD and VPOPCNTQ forms.
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
func vpopcntdqforms(bcst string) inst.Forms { return _yvexpandpd("AVX512VPOPCNTDQ", bcst) }
