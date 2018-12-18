
#include "textflag.h"

// func Add(x uint64, s Struct, y uint64) uint64
TEXT ·Add(SB),0,$0-200
	MOVQ	x(FP), AX
	MOVQ	y+184(FP), CX
	ADDQ	AX, CX
	MOVQ	CX, ret+192(FP)
	RET	
