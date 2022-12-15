// Code generated by command: go run asm.go -out issue193.s -stubs stub.go. DO NOT EDIT.

#include "textflag.h"

DATA offQ<>+0(SB)/8, $0x0000000000000000
DATA offQ<>+8(SB)/8, $0x0000000000000002
DATA offQ<>+16(SB)/8, $0x0000000000000004
DATA offQ<>+24(SB)/8, $0x0000000000000006
GLOBL offQ<>(SB), RODATA|NOPTR, $32

// func AddSubPairs(a *[8]int64)
// Requires: AVX2, AVX512DQ, AVX512F, AVX512VL
TEXT ·AddSubPairs(SB), $0-8
	VMOVDQU64   offQ<>+0(SB), Y0
	MOVQ        a+0(FP), AX
	KXNORB      K0, K0, K1
	VPGATHERQQ  (AX)(Y0*8), K1, Y1
	KXNORB      K0, K0, K1
	VPGATHERQQ  8(AX)(Y0*8), K1, Y2
	VPADDQ      Y2, Y1, Y3
	KXNORB      K0, K0, K1
	VPSCATTERQQ Y3, K1, (AX)(Y0*8)
	VPSUBQ      Y2, Y1, Y3
	KXNORB      K0, K0, K1
	VPSCATTERQQ Y3, K1, 8(AX)(Y0*8)
	RET

// func AddSubPairsNoBase(a *[8]int64)
// Requires: AVX2, AVX512DQ, AVX512F, AVX512VL
TEXT ·AddSubPairsNoBase(SB), $0-8
	VMOVDQU64   offQ<>+0(SB), Y0
	VPSLLQ      $0x03, Y0, Y0
	VPADDQ.BCST a+0(FP), Y0, Y0
	KXNORB      K0, K0, K1
	VPGATHERQQ  (Y0*1), K1, Y1
	KXNORB      K0, K0, K1
	VPGATHERQQ  8(Y0*1), K1, Y2
	VPADDQ      Y2, Y1, Y3
	KXNORB      K0, K0, K1
	VPSCATTERQQ Y3, K1, (Y0*1)
	VPSUBQ      Y2, Y1, Y3
	KXNORB      K0, K0, K1
	VPSCATTERQQ Y3, K1, 8(Y0*1)
	RET