
#include "textflag.h"

// func Hash64(data []byte) uint64
TEXT ·Hash64(SB),0,$0-32
	MOVQ	data_base(FP), CX
	MOVQ	data_len+8(FP), BX
	MOVQ	$0xcbf29ce484222325, AX
	MOVQ	$0x100000001b3, BP
loop:
	CMPQ	BX, $0x0
	JE	done
	MOVBQZX	(CX), DX
	XORQ	DX, AX
	MULQ	BP
	INCQ	CX
	DECQ	BX
	JMP	loop
done:
	MOVQ	AX, ret+24(FP)
	RET
