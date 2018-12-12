
#include "textflag.h"

// func Sum(xs []uint64) uint64
TEXT ·Sum(SB),0,$0-32
	MOVQ	xs_base(FP), DX
	MOVQ	xs_len+8(FP), CX
	XORQ	AX, AX
loop:
	CMPQ	CX, $0x0
	JE	done
	ADDQ	(DX), AX
	ADDQ	$0x8, DX
	DECQ	CX
	JMP	loop
done:
	MOVQ	AX, ret+24(FP)
	RET	
