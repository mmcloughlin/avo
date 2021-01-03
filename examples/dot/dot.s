// Code generated by command: go run asm.go -out dot.s -stubs stub.go. DO NOT EDIT.

#include "textflag.h"

// func Dot(x []float32, y []float32) float32
// Requires: AVX, AVX512F, AVX512VL, FMA3, SSE
TEXT ·Dot(SB), NOSPLIT, $0-52
	MOVQ   x_base+0(FP), AX
	MOVQ   y_base+24(FP), CX
	MOVQ   x_len+8(FP), DX
	VXORPS Y0, Y0, Y0
	VXORPS Y1, Y1, Y1
	VXORPS Y2, Y2, Y2
	VXORPS Y3, Y3, Y3
	VXORPS Y4, Y4, Y4
	VXORPS Y5, Y5, Y5

blockloop:
	CMPQ        DX, $0x00000030
	JL          tail
	VMOVUPS     (AX), Y6
	VMOVUPS     32(AX), Y7
	VMOVUPS     64(AX), Y8
	VMOVUPS     96(AX), Y9
	VMOVUPS     128(AX), Y10
	VMOVUPS     160(AX), Y11
	VFMADD231PS (CX), Y6, Y0
	VFMADD231PS 32(CX), Y7, Y1
	VFMADD231PS 64(CX), Y8, Y2
	VFMADD231PS 96(CX), Y9, Y3
	VFMADD231PS 128(CX), Y10, Y4
	VFMADD231PS 160(CX), Y11, Y5
	ADDQ        $0x000000c0, AX
	ADDQ        $0x000000c0, CX
	SUBQ        $0x00000030, DX
	JMP         blockloop

tail:
	VXORPS X6, X6, X6

tailloop:
	CMPQ        DX, $0x00000000
	JE          reduce
	VMOVSS      (AX), X7
	VFMADD231SS (CX), X7, X6
	ADDQ        $0x00000004, AX
	ADDQ        $0x00000004, CX
	DECQ        DX
	JMP         tailloop

reduce:
	VADDPS       Y0, Y1, Y0
	VADDPS       Y0, Y2, Y0
	VADDPS       Y0, Y3, Y0
	VADDPS       Y0, Y4, Y0
	VADDPS       Y0, Y5, Y0
	VEXTRACTF128 $0x01, Y0, X1
	VADDPS       X0, X1, X0
	VADDPS       X0, X6, X0
	VHADDPS      X0, X0, X0
	VHADDPS      X0, X0, X0
	MOVSS        X0, ret+48(FP)
	RET
