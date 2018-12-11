
#include "textflag.h"

// func Add(x uint64, y uint64) uint64
TEXT ·Add(SB),0,$0-24
	MOVQ	x(FP), CX
	MOVQ	y+8(FP), AX
	ADDQ	CX, AX
	MOVQ	CX, ret+16(FP)
	RET	
