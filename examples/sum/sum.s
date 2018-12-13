
#include "textflag.h"

// func Sum(xs []uint64) uint64
TEXT ·Sum(SB),0,$0-32
	MOVQ	xs_base(FP), AX
	MOVQ	xs_len+8(FP), CX
	XORQ	DX, DX
loop:
	CMPQ	CX, $0x0
	JE	done
	ADDQ	(AX), DX
	ADDQ	$0x8, AX
	DECQ	CX
	JMP	loop
done:
	MOVQ	DX, ret+24(FP)
	RET	
