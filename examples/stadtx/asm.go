// +build ignore

package main

import (
	"strconv"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

const (
	k0U64 = 0xb89b0f8e1655514f
	k1U64 = 0x8c6f736011bd5127
	k2U64 = 0x8f29bd94edce7b39
	k3U64 = 0x9c1b8e1e9628323f

	k2U32 = 0x802910e3
	k3U32 = 0x819b13af
	k4U32 = 0x91cb27e5
	k5U32 = 0xc1a269c1
)

func imul(k uint64, r Register) {
	t := GP64()
	MOVQ(U64(k), t)
	IMULQ(t, r)
}

func makelabels(name string, n int) []string {
	l := make([]string, n)
	for i := 0; i < n; i++ {
		l[i] = name + strconv.Itoa(i)
	}
	return l
}

func main() {
	Package("github.com/mmcloughlin/avo/examples/stadtx")
	TEXT("Hash", "func(state *State, key []byte) uint64")
	Doc("Hash computes the Stadtx hash.")

	statePtr := Load(Param("state"), GP64())
	ptr := Load(Param("key").Base(), GP64())
	n := Load(Param("key").Len(), GP64())

	v0 := GP64()                           // reg_v0 = GeneralPurposeRegister64()
	v1 := GP64()                           // reg_v1 = GeneralPurposeRegister64()
	MOVQ(Mem{Base: statePtr}, v0)          // MOV(reg_v0, [reg_state_ptr])
	MOVQ(Mem{Base: statePtr, Disp: 8}, v1) // MOV(reg_v1, [reg_state_ptr+8])

	t := GP64()     // t = GeneralPurposeRegister64()
	MOVQ(n, t)      // MOV(t, reg_ptr_len)
	ADDQ(U32(1), t) // ADD(t, 1)
	imul(k0U64, t)  // imul(t, k0U64)
	XORQ(t, v0)     // XOR(reg_v0, t)

	MOVQ(n, t)      // MOV(t, reg_ptr_len)
	ADDQ(U32(2), t) // ADD(t, 2)
	imul(k1U64, t)  // imul(t, k1U64)
	XORQ(t, v1)     // XOR(reg_v1, t)

	long := "coreLong"  //		    coreLong = Label("coreLong")
	CMPQ(n, U32(32))    //		    CMP(reg_ptr_len, 32)
	JGE(LabelRef(long)) //		    JGE(coreLong)
	//
	u64s := GP64()    //		    reg_u64s = GeneralPurposeRegister64()
	MOVQ(n, u64s)     //		    MOV(reg_u64s, reg_ptr_len)
	SHRQ(U8(3), u64s) //		    SHR(reg_u64s, 3)
	//
	labels := makelabels("shortCore", 4) //		    labels = [Label("shortCore%d" % i) for i in range(4)]
	//
	for i := 0; i < 4; i++ { //		    for i in range(4):
		CMPQ(u64s, U32(i))      //		        CMP(reg_u64s, i)
		JE(LabelRef(labels[i])) //		        JE(labels[i])
	} //
	for i := 3; i > 0; i-- { //		    for i in range(3, 0, -1):
		Label(labels[i])        //		        Label(labels[i])
		r := GP64()             //		        r = GeneralPurposeRegister64()
		MOVQ(Mem{Base: ptr}, r) //		        MOV(r, [reg_ptr])
		imul(k3U64, r)          //		        imul(r, k3U64)
		ADDQ(r, v0)             //		        ADD(reg_v0, r)
		RORQ(U8(17), v0)        //		        ROR(reg_v0, 17)
		XORQ(v1, v0)            //		        XOR(reg_v0, reg_v1)
		RORQ(U8(53), v1)        //		        ROR(reg_v1, 53)
		ADDQ(v0, v1)            //		        ADD(reg_v1, reg_v0)
		ADDQ(U32(8), ptr)       //		        ADD(reg_ptr,8)
		SUBQ(U32(8), n)         //		        SUB(reg_ptr_len,8)
	} //
	Label(labels[0]) //		    Label(labels[0])
	//
	labels = makelabels("shortTail", 8) //		    labels = [Label("shortTail%d" % i) for i in range(8)]
	//
	//		    split(labels, reg_ptr_len,0,7)
	for i := 0; i < 8; i++ {
		CMPQ(n, U32(i))
		JE(LabelRef(labels[i]))
	}
	//
	after := "shortAfter" //		    after = Label("shortAfter")
	//
	ch := GP64() //		    reg_ch = GeneralPurposeRegister64()
	//
	Label(labels[7])                     //		    Label(labels[7])
	MOVBQZX(Mem{Base: ptr, Disp: 6}, ch) //		    MOVZX(reg_ch, byte[reg_ptr+6])
	SHLQ(U8(32), ch)                     //		    SHL(reg_ch, 32)
	ADDQ(ch, v0)                         //		    ADD(reg_v0, reg_ch)
	//
	Label(labels[6])                     //		    Label(labels[6])
	MOVBQZX(Mem{Base: ptr, Disp: 5}, ch) //		    MOVZX(reg_ch, byte[reg_ptr+5])
	SHLQ(U8(48), ch)                     //		    SHL(reg_ch, 48)
	ADDQ(ch, v1)                         //		    ADD(reg_v1, reg_ch)
	//
	Label(labels[5])                     //		    Label(labels[5])
	MOVBQZX(Mem{Base: ptr, Disp: 4}, ch) //		    MOVZX(reg_ch, byte[reg_ptr+4])
	SHLQ(U8(16), ch)                     //		    SHL(reg_ch, 16)
	ADDQ(ch, v0)                         //		    ADD(reg_v0, reg_ch)
	//
	Label(labels[4]) //		    Label(labels[4])
	//		    XOR(reg_ch, reg_ch)
	MOVLQZX(Mem{Base: ptr}, ch) //		    MOV(reg_ch.as_dword, dword[reg_ptr])
	ADDQ(ch, v1)                //		    ADD(reg_v1, reg_ch)
	//
	JMP(LabelRef(after)) //		    JMP(after)
	//
	Label(labels[3])                     //		    Label(labels[3])
	MOVBQZX(Mem{Base: ptr, Disp: 2}, ch) //		    MOVZX(reg_ch, byte[reg_ptr+2])
	SHLQ(U8(48), ch)                     //		    SHL(reg_ch, 48)
	ADDQ(ch, v0)                         //		    ADD(reg_v0, reg_ch)
	//
	Label(labels[2]) //		    Label(labels[2])
	//		    XOR(reg_ch, reg_ch)
	MOVWQZX(Mem{Base: ptr}, ch) //		    MOV(reg_ch.as_word, word[reg_ptr])
	ADDQ(ch, v1)                //		    ADD(reg_v1, reg_ch)
	//
	JMP(LabelRef(after)) //		    JMP(after)
	//
	Label(labels[1])            //		    Label(labels[1])
	MOVBQZX(Mem{Base: ptr}, ch) //		    MOVZX(reg_ch, byte[reg_ptr])
	ADDQ(ch, v0)                //		    ADD(reg_v0, reg_ch)
	//
	Label(labels[0])    //		    Label(labels[0])
	RORQ(U8(32), v1)    //		    ROR(reg_v1, 32)
	XORQ(U32(0xff), v1) //		    XOR(reg_v1, 0xFF)
	//
	Label(after) //		    Label(after)
	//
	XORQ(v0, v1) //		    XOR(reg_v1, reg_v0)
	//
	RORQ(U8(33), v0) //		    ROR(reg_v0, 33)
	ADDQ(v1, v0)     //		    ADD(reg_v0, reg_v1)
	//
	ROLQ(U8(17), v1) //		    ROL(reg_v1, 17)
	XORQ(v0, v1)     //		    XOR(reg_v1, reg_v0)
	//
	ROLQ(U8(43), v0) //		    ROL(reg_v0, 43)
	ADDQ(v1, v0)     //		    ADD(reg_v0, reg_v1)
	//
	ROLQ(U8(31), v1) //		    ROL(reg_v1, 31)
	SUBQ(v0, v1)     //		    SUB(reg_v1, reg_v0)
	//
	ROLQ(U8(13), v0) //		    ROL(reg_v0, 13)
	XORQ(v1, v0)     //		    XOR(reg_v0, reg_v1)
	//
	SUBQ(v0, v1) //		    SUB(reg_v1, reg_v0)
	//
	ROLQ(U8(41), v0) //		    ROL(reg_v0, 41)
	ADDQ(v1, v0)     //		    ADD(reg_v0, reg_v1)
	//
	ROLQ(U8(37), v1) //		    ROL(reg_v1, 37)
	XORQ(v0, v1)     //		    XOR(reg_v1, reg_v0)
	//
	RORQ(U8(39), v0) //		    ROR(reg_v0, 39)
	ADDQ(v1, v0)     //		    ADD(reg_v0, reg_v1)
	//
	RORQ(U8(15), v1) //		    ROR(reg_v1, 15)
	ADDQ(v0, v1)     //		    ADD(reg_v1, reg_v0)
	//
	ROLQ(U8(15), v0) //		    ROL(reg_v0, 15)
	XORQ(v1, v0)     //		    XOR(reg_v0, reg_v1)
	//
	RORQ(U8(5), v1) //		    ROR(reg_v1, 5)
	//
	XORQ(v1, v0) //		    XOR(reg_v0, reg_v1)
	//
	Store(v0, ReturnIndex(0))
	RET() //		    RETURN(reg_v0)
	//
	Label(long) //		    Label(coreLong)
	//
	v2 := GP64() //		    reg_v2 = GeneralPurposeRegister64()
	v3 := GP64() //		    reg_v3 = GeneralPurposeRegister64()
	//
	MOVQ(Mem{Base: statePtr, Disp: 16}, v2) //		    MOV(reg_v2, [reg_state_ptr+16])
	MOVQ(Mem{Base: statePtr, Disp: 24}, v3) //		    MOV(reg_v3, [reg_state_ptr+24])
	//
	//		    t = GeneralPurposeRegister64()
	MOVQ(n, t)      //		    MOV(t, reg_ptr_len)
	ADDQ(U32(3), t) //		    ADD(t, 3)
	imul(k2U64, t)  //		    imul(t, k2U64)
	XORQ(t, v2)     //		    XOR(reg_v2, t)
	//
	MOVQ(n, t)      //		    MOV(t, reg_ptr_len)
	ADDQ(U32(4), t) //		    ADD(t, 4)
	imul(k3U64, t)  //		    imul(t, k3U64)
	XORQ(t, v3)     //		    XOR(reg_v3, t)
	//
	r := GP64()     //		    r = GeneralPurposeRegister64()
	loop := "block" //		    with Loop() as loop:
	Label(loop)
	MOVQ(Mem{Base: ptr}, r) //		        MOV(r, [reg_ptr])
	imul(k2U32, r)          //		        imul(r, k2U32)
	ADDQ(r, v0)             //		        ADD(reg_v0, r)
	ROLQ(U8(57), v0)        //		        ROL(reg_v0, 57)
	XORQ(v3, v0)            //		        XOR(reg_v0, reg_v3)
	//
	MOVQ(Mem{Base: ptr, Disp: 8}, r) //		        MOV(r, [reg_ptr + 8])
	imul(k3U32, r)                   //		        imul(r, k3U32)
	ADDQ(r, v1)                      //		        ADD(reg_v1, r)
	ROLQ(U8(63), v1)                 //		        ROL(reg_v1, 63)
	XORQ(v2, v1)                     //		        XOR(reg_v1, reg_v2)
	//
	MOVQ(Mem{Base: ptr, Disp: 16}, r) //		        MOV(r, [reg_ptr + 16])
	imul(k4U32, r)                    //		        imul(r, k4U32)
	ADDQ(r, v2)                       //		        ADD(reg_v2, r)
	RORQ(U8(47), v2)                  //		        ROR(reg_v2, 47)
	ADDQ(v0, v2)                      //		        ADD(reg_v2, reg_v0)
	//
	MOVQ(Mem{Base: ptr, Disp: 24}, r) //		        MOV(r, [reg_ptr + 24])
	imul(k5U32, r)                    //		        imul(r, k5U32)
	ADDQ(r, v3)                       //		        ADD(reg_v3, r)
	RORQ(U8(11), v3)                  //		        ROR(reg_v3, 11)
	SUBQ(v1, v3)                      //		        SUB(reg_v3, reg_v1)
	//
	ADDQ(U32(32), ptr) //		        ADD(reg_ptr, 32)
	SUBQ(U32(32), n)   //		        SUB(reg_ptr_len, 32)
	//
	CMPQ(n, U32(32))    //		        CMP(reg_ptr_len, 32)
	JGE(LabelRef(loop)) //		        JGE(loop.begin)
	//
	//
	nsave := GP64() //		    reg_ptr_len_saved = GeneralPurposeRegister64()
	MOVQ(n, nsave)  //		    MOV(reg_ptr_len_saved, reg_ptr_len)
	//
	//		    reg_u64s = GeneralPurposeRegister64()
	MOVQ(n, u64s)     //		    MOV(reg_u64s, reg_ptr_len)
	SHRQ(U8(3), u64s) //		    SHR(reg_u64s, 3)
	//
	labels = makelabels("longCore", 4) //		    labels = [Label("longCore%d" % i) for i in range(4)]
	//
	for i := 0; i < 4; i++ { //		    for i in range(4):
		CMPQ(u64s, U32(i))      //		        CMP(reg_u64s, i)
		JE(LabelRef(labels[i])) //		        JE(labels[i])
	} //
	Label(labels[3]) //		    Label(labels[3])
	//
	MOVQ(Mem{Base: ptr}, r) //		    MOV(r, [reg_ptr])
	imul(k2U32, r)          //		    imul(r, k2U32)
	ADDQ(r, v0)             //		    ADD(reg_v0, r)
	ROLQ(U8(57), v0)        //		    ROL(reg_v0, 57)
	XORQ(v3, v0)            //		    XOR(reg_v0, reg_v3)
	ADDQ(U32(8), ptr)       //		    ADD(reg_ptr, 8)
	SUBQ(U32(8), n)         //		    SUB(reg_ptr_len, 8)
	//
	Label(labels[2]) //		    Label(labels[2])
	//
	MOVQ(Mem{Base: ptr}, r) //		    MOV(r, [reg_ptr])
	imul(k3U32, r)          //		    imul(r, k3U32)
	ADDQ(r, v1)             //		    ADD(reg_v1, r)
	ROLQ(U8(63), v1)        //		    ROL(reg_v1, 63)
	XORQ(v2, v1)            //		    XOR(reg_v1, reg_v2)
	ADDQ(U32(8), ptr)       //		    ADD(reg_ptr, 8)
	SUBQ(U32(8), n)         //		    SUB(reg_ptr_len, 8)
	//
	Label(labels[1]) //		    Label(labels[1])
	//
	MOVQ(Mem{Base: ptr}, r) //		    MOV(r, [reg_ptr])
	imul(k4U32, r)          //		    imul(r, k4U32)
	ADDQ(r, v2)             //		    ADD(reg_v2, r)
	RORQ(U8(47), v2)        //		    ROR(reg_v2, 47)
	ADDQ(v0, v2)            //		    ADD(reg_v2, reg_v0)
	ADDQ(U32(8), ptr)       //		    ADD(reg_ptr, 8)
	SUBQ(U32(8), n)         //		    SUB(reg_ptr_len, 8)
	//
	Label(labels[0]) //		    Label(labels[0])
	//
	RORQ(U8(11), v3) //		    ROR(reg_v3, 11)
	SUBQ(v1, v3)     //		    SUB(reg_v3, reg_v1)
	//
	ADDQ(U32(1), nsave) //		    ADD(reg_ptr_len_saved, 1)
	imul(k3U64, nsave)  //		    imul(reg_ptr_len_saved, k3U64)
	XORQ(nsave, v0)     //		    XOR(reg_v0, reg_ptr_len_saved)
	//
	labels = makelabels("longTail", 8) //		    labels = [Label("longTail%d" % i) for i in range(8)]
	//
	//		    split(labels, reg_ptr_len, 0, 7)
	for i := 0; i < 8; i++ {
		CMPQ(n, U32(i))
		JE(LabelRef(labels[i]))
	}
	//
	after = "longAfter" //		    after = Label("longAfter")
	//
	//		    reg_ch = GeneralPurposeRegister64()
	//
	Label(labels[7])                     //		    Label(labels[7])
	MOVBQZX(Mem{Base: ptr, Disp: 6}, ch) //		    MOVZX(reg_ch, byte[reg_ptr+6])
	ADDQ(ch, v1)                         //		    ADD(reg_v1, reg_ch)
	//
	Label(labels[6]) //		    Label(labels[6])
	//		    XOR(reg_ch, reg_ch)
	MOVWQZX(Mem{Base: ptr, Disp: 4}, ch) //		    MOV(reg_ch.as_word, word[reg_ptr + 4])
	ADDQ(ch, v2)                         //		    ADD(reg_v2, reg_ch)
	MOVLQZX(Mem{Base: ptr}, ch)          //		    MOV(reg_ch.as_dword, dword[reg_ptr])
	ADDQ(ch, v3)                         //		    ADD(reg_v3, reg_ch)
	JMP(LabelRef(after))                 //		    JMP(after)
	//
	Label(labels[5])                     //		    Label(labels[5])
	MOVBQZX(Mem{Base: ptr, Disp: 4}, ch) //		    MOVZX(reg_ch, byte[reg_ptr+4])
	ADDQ(ch, v1)                         //		    ADD(reg_v1, reg_ch)
	//
	Label(labels[4]) //		    Label(labels[4])
	//		    XOR(reg_ch, reg_ch)
	MOVLQZX(Mem{Base: ptr}, ch) //		    MOV(reg_ch.as_dword, dword[reg_ptr])
	ADDQ(ch, v2)                //		    ADD(reg_v2, reg_ch)
	//
	JMP(LabelRef(after)) //		    JMP(after)
	//
	Label(labels[3])                     //		    Label(labels[3])
	MOVBQZX(Mem{Base: ptr, Disp: 2}, ch) //		    MOVZX(reg_ch, byte[reg_ptr+2])
	ADDQ(ch, v3)                         //		    ADD(reg_v3, reg_ch)
	//
	Label(labels[2]) //		    Label(labels[2])
	//		    XOR(reg_ch, reg_ch)
	MOVWQZX(Mem{Base: ptr}, ch) //		    MOV(reg_ch.as_word, word[reg_ptr])
	ADDQ(ch, v1)                //		    ADD(reg_v1, reg_ch)
	//
	JMP(LabelRef(after)) //		    JMP(after)
	//
	Label(labels[1])            //		    Label(labels[1])
	MOVBQZX(Mem{Base: ptr}, ch) //		    MOVZX(reg_ch, byte[reg_ptr])
	ADDQ(ch, v2)                //		    ADD(reg_v2, reg_ch)
	//
	Label(labels[0])    //		    Label(labels[0])
	ROLQ(U8(32), v3)    //		    ROL(reg_v3, 32)
	XORQ(U32(0xff), v3) //		    XOR(reg_v3, 0xFF)
	//
	Label(after) //		    Label(after)
	//
	//		    ## finalize
	//
	SUBQ(v2, v1)     //		    SUB(reg_v1, reg_v2)
	RORQ(U8(19), v0) //		    ROR(reg_v0, 19)
	SUBQ(v0, v1)     //		    SUB(reg_v1, reg_v0)
	RORQ(U8(53), v1) //		    ROR(reg_v1, 53)
	XORQ(v1, v3)     //		    XOR(reg_v3, reg_v1)
	SUBQ(v3, v0)     //		    SUB(reg_v0, reg_v3)
	ROLQ(U8(43), v3) //		    ROL(reg_v3, 43)
	ADDQ(v3, v0)     //		    ADD(reg_v0, reg_v3)
	RORQ(U8(3), v0)  //		    ROR(reg_v0, 3)
	SUBQ(v0, v3)     //		    SUB(reg_v3, reg_v0)
	RORQ(U8(43), v2) //		    ROR(reg_v2, 43)
	SUBQ(v3, v2)     //		    SUB(reg_v2, reg_v3)
	ROLQ(U8(55), v2) //		    ROL(reg_v2, 55)
	XORQ(v0, v2)     //		    XOR(reg_v2, reg_v0)
	SUBQ(v2, v1)     //		    SUB(reg_v1, reg_v2)
	RORQ(U8(7), v3)  //		    ROR(reg_v3, 7)
	SUBQ(v2, v3)     //		    SUB(reg_v3, reg_v2)
	RORQ(U8(31), v2) //		    ROR(reg_v2, 31)
	ADDQ(v2, v3)     //		    ADD(reg_v3, reg_v2)
	SUBQ(v1, v2)     //		    SUB(reg_v2, reg_v1)
	RORQ(U8(39), v3) //		    ROR(reg_v3, 39)
	XORQ(v3, v2)     //		    XOR(reg_v2, reg_v3)
	RORQ(U8(17), v3) //		    ROR(reg_v3, 17)
	XORQ(v2, v3)     //		    XOR(reg_v3, reg_v2)
	ADDQ(v3, v1)     //		    ADD(reg_v1, reg_v3)
	RORQ(U8(9), v1)  //		    ROR(reg_v1, 9)
	XORQ(v1, v2)     //		    XOR(reg_v2, reg_v1)
	ROLQ(U8(24), v2) //		    ROL(reg_v2, 24)
	XORQ(v2, v3)     //		    XOR(reg_v3, reg_v2)
	RORQ(U8(59), v3) //		    ROR(reg_v3, 59)
	RORQ(U8(1), v0)  //		    ROR(reg_v0, 1)
	SUBQ(v1, v0)     //		    SUB(reg_v0, reg_v1)
	//
	XORQ(v1, v0) //		    XOR(reg_v0, reg_v1)
	XORQ(v3, v2) //		    XOR(reg_v2, reg_v3)
	//
	XORQ(v2, v0) //		    XOR(reg_v0, reg_v2)
	//
	Store(v0, ReturnIndex(0))
	RET() //			RETURN(reg_v0)

	Generate()
}
