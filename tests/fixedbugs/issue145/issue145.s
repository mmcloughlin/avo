// Code generated by command: go run asm.go -out issue145.s -stubs stub.go. DO NOT EDIT.

#include "textflag.h"

// func Shuffle(mask []byte, data []byte) [4]uint32
// Requires: SSE2, SSSE3
TEXT ·Shuffle(SB), NOSPLIT, $0-64
	MOVQ   mask_base+0(FP), AX
	MOVQ   data_base+24(FP), CX
	MOVOU  (CX), X0
	PSHUFB (AX), X0
	MOVOU  X0, ret+48(FP)
	RET
