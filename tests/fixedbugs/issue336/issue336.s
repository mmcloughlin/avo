// Code generated by command: go run asm.go -out issue336.s -stubs stub.go. DO NOT EDIT.

#include "textflag.h"

// func Not8(x bool) bool
TEXT ·Not8(SB), NOSPLIT, $0-9
	MOVB x+0(FP), AL
	XORB $0x01, AL
	MOVB AL, ret+8(FP)
	RET

// func Not16(x bool) bool
TEXT ·Not16(SB), NOSPLIT, $0-9
	MOVBWZX x+0(FP), AX
	XORW    $0x01, AX
	MOVB    AL, ret+8(FP)
	RET

// func Not32(x bool) bool
TEXT ·Not32(SB), NOSPLIT, $0-9
	MOVBLZX x+0(FP), AX
	XORL    $0x01, AX
	MOVB    AL, ret+8(FP)
	RET

// func Not64(x bool) bool
TEXT ·Not64(SB), NOSPLIT, $0-9
	MOVBQZX x+0(FP), AX
	XORQ    $0x01, AX
	MOVB    AL, ret+8(FP)
	RET
