// Code generated by command: go run asm.go -out ext.s. DO NOT EDIT.

// func ExtStructFieldB(e github.com/mmcloughlin/avo/examples/ext/ext.Struct) byte
TEXT ·ExtStructFieldB(SB), $0-25
	MOVB e_B+6(FP), AL
	MOVB AL, ret+24(FP)
	RET
