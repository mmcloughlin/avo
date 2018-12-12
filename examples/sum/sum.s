
#include "textflag.h"

// func Sum(xs []uint64) uint64
TEXT ·Sum(SB),0,$0-32
	MOVQ	xs_base(FP), DX
	MOVQ	xs_len+8(FP), AX
	XORQ	CX, CX
loop:
	CMPQ	AX, $0x0
	JE	done
	ADDQ	(DX), CX
	ADDQ	$0x8, DX
	DECQ	AX
	JMP	loop
done:
	MOVQ	CX, ret+24(FP)
	RET	
