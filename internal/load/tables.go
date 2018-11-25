package load

// Go contains a list of self-proclaimed "Annoying aliases", as follows. We use
// a script to automatically extract this list from the source code (see the
// following go:generate line).
//
// Reference: https://github.com/golang/go/blob/048c9164a0c5572df18325e377473e7893dbfb07/src/cmd/asm/internal/arch/arch.go#L126-L182
//
//		}
//		// Annoying aliases.
//		instructions["JA"] = x86.AJHI   /* alternate */
//		instructions["JAE"] = x86.AJCC  /* alternate */
//		instructions["JB"] = x86.AJCS   /* alternate */
//		instructions["JBE"] = x86.AJLS  /* alternate */
//		instructions["JC"] = x86.AJCS   /* alternate */
//		instructions["JCC"] = x86.AJCC  /* carry clear (CF = 0) */
//		instructions["JCS"] = x86.AJCS  /* carry set (CF = 1) */
//		instructions["JE"] = x86.AJEQ   /* alternate */
//		instructions["JEQ"] = x86.AJEQ  /* equal (ZF = 1) */
//		instructions["JG"] = x86.AJGT   /* alternate */
//		instructions["JGE"] = x86.AJGE  /* greater than or equal (signed) (SF = OF) */
//		instructions["JGT"] = x86.AJGT  /* greater than (signed) (ZF = 0 && SF = OF) */
//		instructions["JHI"] = x86.AJHI  /* higher (unsigned) (CF = 0 && ZF = 0) */
//		instructions["JHS"] = x86.AJCC  /* alternate */
//		instructions["JL"] = x86.AJLT   /* alternate */
//		instructions["JLE"] = x86.AJLE  /* less than or equal (signed) (ZF = 1 || SF != OF) */
//		instructions["JLO"] = x86.AJCS  /* alternate */
//		instructions["JLS"] = x86.AJLS  /* lower or same (unsigned) (CF = 1 || ZF = 1) */
//		instructions["JLT"] = x86.AJLT  /* less than (signed) (SF != OF) */
//		instructions["JMI"] = x86.AJMI  /* negative (minus) (SF = 1) */
//		instructions["JNA"] = x86.AJLS  /* alternate */
//		instructions["JNAE"] = x86.AJCS /* alternate */
//		instructions["JNB"] = x86.AJCC  /* alternate */
//		instructions["JNBE"] = x86.AJHI /* alternate */
//		instructions["JNC"] = x86.AJCC  /* alternate */
//		instructions["JNE"] = x86.AJNE  /* not equal (ZF = 0) */
//		instructions["JNG"] = x86.AJLE  /* alternate */
//		instructions["JNGE"] = x86.AJLT /* alternate */
//		instructions["JNL"] = x86.AJGE  /* alternate */
//		instructions["JNLE"] = x86.AJGT /* alternate */
//		instructions["JNO"] = x86.AJOC  /* alternate */
//		instructions["JNP"] = x86.AJPC  /* alternate */
//		instructions["JNS"] = x86.AJPL  /* alternate */
//		instructions["JNZ"] = x86.AJNE  /* alternate */
//		instructions["JO"] = x86.AJOS   /* alternate */
//		instructions["JOC"] = x86.AJOC  /* overflow clear (OF = 0) */
//		instructions["JOS"] = x86.AJOS  /* overflow set (OF = 1) */
//		instructions["JP"] = x86.AJPS   /* alternate */
//		instructions["JPC"] = x86.AJPC  /* parity clear (PF = 0) */
//		instructions["JPE"] = x86.AJPS  /* alternate */
//		instructions["JPL"] = x86.AJPL  /* non-negative (plus) (SF = 0) */
//		instructions["JPO"] = x86.AJPC  /* alternate */
//		instructions["JPS"] = x86.AJPS  /* parity set (PF = 1) */
//		instructions["JS"] = x86.AJMI   /* alternate */
//		instructions["JZ"] = x86.AJEQ   /* alternate */
//		instructions["MASKMOVDQU"] = x86.AMASKMOVOU
//		instructions["MOVD"] = x86.AMOVQ
//		instructions["MOVDQ2Q"] = x86.AMOVQ
//		instructions["MOVNTDQ"] = x86.AMOVNTO
//		instructions["MOVOA"] = x86.AMOVO
//		instructions["PSLLDQ"] = x86.APSLLO
//		instructions["PSRLDQ"] = x86.APSRLO
//		instructions["PADDD"] = x86.APADDL
//
//		return &Arch{
//

//go:generate ./annoyingaliases.sh zannoyingaliases.go
