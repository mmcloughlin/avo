package reg

// Register families.
const (
	Internal Kind = iota
	GP
	MMX
	SSEAVX
	Mask
)

var Families = []*Family{
	Pseudo,
	GeneralPurpose,
	SIMD,
}

var familiesByKind = map[Kind]*Family{}

func init() {
	for _, f := range Families {
		familiesByKind[f.Kind] = f
	}
}

func FamilyOfKind(k Kind) *Family {
	return familiesByKind[k]
}

// Pseudo registers.
var (
	Pseudo = &Family{Kind: Internal}

	FramePointer   = Pseudo.define(S0, 0, "FP")
	ProgramCounter = Pseudo.define(S0, 0, "PC")
	StaticBase     = Pseudo.define(S0, 0, "SB")
	StackPointer   = Pseudo.define(S0, 0, "SP")
)

// General purpose registers.
var (
	GeneralPurpose = &Family{Kind: GP}

	// Low byte
	AL = GeneralPurpose.define(S8L, 0, "AL")
	CL = GeneralPurpose.define(S8L, 1, "CL")
	DL = GeneralPurpose.define(S8L, 2, "DL")
	BL = GeneralPurpose.define(S8L, 3, "BL")

	// High byte
	AH = GeneralPurpose.define(S8H, 0, "AH")
	CH = GeneralPurpose.define(S8H, 1, "CH")
	DH = GeneralPurpose.define(S8H, 2, "DH")
	BH = GeneralPurpose.define(S8H, 3, "BH")

	// 8-bit
	SPB  = GeneralPurpose.restricted(S8, 4, "SP")
	BPB  = GeneralPurpose.define(S8, 5, "BP")
	SIB  = GeneralPurpose.define(S8, 6, "SI")
	DIB  = GeneralPurpose.define(S8, 7, "DI")
	R8B  = GeneralPurpose.define(S8, 8, "R8")
	R9B  = GeneralPurpose.define(S8, 9, "R9")
	R10B = GeneralPurpose.define(S8, 10, "R10")
	R11B = GeneralPurpose.define(S8, 11, "R11")
	R12B = GeneralPurpose.define(S8, 12, "R12")
	R13B = GeneralPurpose.define(S8, 13, "R13")
	R14B = GeneralPurpose.define(S8, 14, "R14")
	R15B = GeneralPurpose.define(S8, 15, "R15")

	// 16-bit
	AX   = GeneralPurpose.define(S16, 0, "AX")
	CX   = GeneralPurpose.define(S16, 1, "CX")
	DX   = GeneralPurpose.define(S16, 2, "DX")
	BX   = GeneralPurpose.define(S16, 3, "BX")
	SP   = GeneralPurpose.restricted(S16, 4, "SP")
	BP   = GeneralPurpose.define(S16, 5, "BP")
	SI   = GeneralPurpose.define(S16, 6, "SI")
	DI   = GeneralPurpose.define(S16, 7, "DI")
	R8W  = GeneralPurpose.define(S16, 8, "R8")
	R9W  = GeneralPurpose.define(S16, 9, "R9")
	R10W = GeneralPurpose.define(S16, 10, "R10")
	R11W = GeneralPurpose.define(S16, 11, "R11")
	R12W = GeneralPurpose.define(S16, 12, "R12")
	R13W = GeneralPurpose.define(S16, 13, "R13")
	R14W = GeneralPurpose.define(S16, 14, "R14")
	R15W = GeneralPurpose.define(S16, 15, "R15")

	// 32-bit
	EAX  = GeneralPurpose.define(S32, 0, "AX")
	ECX  = GeneralPurpose.define(S32, 1, "CX")
	EDX  = GeneralPurpose.define(S32, 2, "DX")
	EBX  = GeneralPurpose.define(S32, 3, "BX")
	ESP  = GeneralPurpose.restricted(S32, 4, "SP")
	EBP  = GeneralPurpose.define(S32, 5, "BP")
	ESI  = GeneralPurpose.define(S32, 6, "SI")
	EDI  = GeneralPurpose.define(S32, 7, "DI")
	R8L  = GeneralPurpose.define(S32, 8, "R8")
	R9L  = GeneralPurpose.define(S32, 9, "R9")
	R10L = GeneralPurpose.define(S32, 10, "R10")
	R11L = GeneralPurpose.define(S32, 11, "R11")
	R12L = GeneralPurpose.define(S32, 12, "R12")
	R13L = GeneralPurpose.define(S32, 13, "R13")
	R14L = GeneralPurpose.define(S32, 14, "R14")
	R15L = GeneralPurpose.define(S32, 15, "R15")

	// 64-bit
	RAX = GeneralPurpose.define(S64, 0, "AX")
	RCX = GeneralPurpose.define(S64, 1, "CX")
	RDX = GeneralPurpose.define(S64, 2, "DX")
	RBX = GeneralPurpose.define(S64, 3, "BX")
	RSP = GeneralPurpose.restricted(S64, 4, "SP")
	RBP = GeneralPurpose.define(S64, 5, "BP")
	RSI = GeneralPurpose.define(S64, 6, "SI")
	RDI = GeneralPurpose.define(S64, 7, "DI")
	R8  = GeneralPurpose.define(S64, 8, "R8")
	R9  = GeneralPurpose.define(S64, 9, "R9")
	R10 = GeneralPurpose.define(S64, 10, "R10")
	R11 = GeneralPurpose.define(S64, 11, "R11")
	R12 = GeneralPurpose.define(S64, 12, "R12")
	R13 = GeneralPurpose.define(S64, 13, "R13")
	R14 = GeneralPurpose.define(S64, 14, "R14")
	R15 = GeneralPurpose.define(S64, 15, "R15")
)

// SIMD registers.
var (
	SIMD = &Family{Kind: SSEAVX}

	// 128-bit
	X0  = SIMD.define(S128, 0, "X0")
	X1  = SIMD.define(S128, 1, "X1")
	X2  = SIMD.define(S128, 2, "X2")
	X3  = SIMD.define(S128, 3, "X3")
	X4  = SIMD.define(S128, 4, "X4")
	X5  = SIMD.define(S128, 5, "X5")
	X6  = SIMD.define(S128, 6, "X6")
	X7  = SIMD.define(S128, 7, "X7")
	X8  = SIMD.define(S128, 8, "X8")
	X9  = SIMD.define(S128, 9, "X9")
	X10 = SIMD.define(S128, 10, "X10")
	X11 = SIMD.define(S128, 11, "X11")
	X12 = SIMD.define(S128, 12, "X12")
	X13 = SIMD.define(S128, 13, "X13")
	X14 = SIMD.define(S128, 14, "X14")
	X15 = SIMD.define(S128, 15, "X15")
	X16 = SIMD.define(S128, 16, "X16")
	X17 = SIMD.define(S128, 17, "X17")
	X18 = SIMD.define(S128, 18, "X18")
	X19 = SIMD.define(S128, 19, "X19")
	X20 = SIMD.define(S128, 20, "X20")
	X21 = SIMD.define(S128, 21, "X21")
	X22 = SIMD.define(S128, 22, "X22")
	X23 = SIMD.define(S128, 23, "X23")
	X24 = SIMD.define(S128, 24, "X24")
	X25 = SIMD.define(S128, 25, "X25")
	X26 = SIMD.define(S128, 26, "X26")
	X27 = SIMD.define(S128, 27, "X27")
	X28 = SIMD.define(S128, 28, "X28")
	X29 = SIMD.define(S128, 29, "X29")
	X30 = SIMD.define(S128, 30, "X30")
	X31 = SIMD.define(S128, 31, "X31")

	// 256-bit
	Y0  = SIMD.define(S256, 0, "Y0")
	Y1  = SIMD.define(S256, 1, "Y1")
	Y2  = SIMD.define(S256, 2, "Y2")
	Y3  = SIMD.define(S256, 3, "Y3")
	Y4  = SIMD.define(S256, 4, "Y4")
	Y5  = SIMD.define(S256, 5, "Y5")
	Y6  = SIMD.define(S256, 6, "Y6")
	Y7  = SIMD.define(S256, 7, "Y7")
	Y8  = SIMD.define(S256, 8, "Y8")
	Y9  = SIMD.define(S256, 9, "Y9")
	Y10 = SIMD.define(S256, 10, "Y10")
	Y11 = SIMD.define(S256, 11, "Y11")
	Y12 = SIMD.define(S256, 12, "Y12")
	Y13 = SIMD.define(S256, 13, "Y13")
	Y14 = SIMD.define(S256, 14, "Y14")
	Y15 = SIMD.define(S256, 15, "Y15")
	Y16 = SIMD.define(S256, 16, "Y16")
	Y17 = SIMD.define(S256, 17, "Y17")
	Y18 = SIMD.define(S256, 18, "Y18")
	Y19 = SIMD.define(S256, 19, "Y19")
	Y20 = SIMD.define(S256, 20, "Y20")
	Y21 = SIMD.define(S256, 21, "Y21")
	Y22 = SIMD.define(S256, 22, "Y22")
	Y23 = SIMD.define(S256, 23, "Y23")
	Y24 = SIMD.define(S256, 24, "Y24")
	Y25 = SIMD.define(S256, 25, "Y25")
	Y26 = SIMD.define(S256, 26, "Y26")
	Y27 = SIMD.define(S256, 27, "Y27")
	Y28 = SIMD.define(S256, 28, "Y28")
	Y29 = SIMD.define(S256, 29, "Y29")
	Y30 = SIMD.define(S256, 30, "Y30")
	Y31 = SIMD.define(S256, 31, "Y31")

	// 512-bit
	Z0  = SIMD.define(S512, 0, "Z0")
	Z1  = SIMD.define(S512, 1, "Z1")
	Z2  = SIMD.define(S512, 2, "Z2")
	Z3  = SIMD.define(S512, 3, "Z3")
	Z4  = SIMD.define(S512, 4, "Z4")
	Z5  = SIMD.define(S512, 5, "Z5")
	Z6  = SIMD.define(S512, 6, "Z6")
	Z7  = SIMD.define(S512, 7, "Z7")
	Z8  = SIMD.define(S512, 8, "Z8")
	Z9  = SIMD.define(S512, 9, "Z9")
	Z10 = SIMD.define(S512, 10, "Z10")
	Z11 = SIMD.define(S512, 11, "Z11")
	Z12 = SIMD.define(S512, 12, "Z12")
	Z13 = SIMD.define(S512, 13, "Z13")
	Z14 = SIMD.define(S512, 14, "Z14")
	Z15 = SIMD.define(S512, 15, "Z15")
	Z16 = SIMD.define(S512, 16, "Z16")
	Z17 = SIMD.define(S512, 17, "Z17")
	Z18 = SIMD.define(S512, 18, "Z18")
	Z19 = SIMD.define(S512, 19, "Z19")
	Z20 = SIMD.define(S512, 20, "Z20")
	Z21 = SIMD.define(S512, 21, "Z21")
	Z22 = SIMD.define(S512, 22, "Z22")
	Z23 = SIMD.define(S512, 23, "Z23")
	Z24 = SIMD.define(S512, 24, "Z24")
	Z25 = SIMD.define(S512, 25, "Z25")
	Z26 = SIMD.define(S512, 26, "Z26")
	Z27 = SIMD.define(S512, 27, "Z27")
	Z28 = SIMD.define(S512, 28, "Z28")
	Z29 = SIMD.define(S512, 29, "Z29")
	Z30 = SIMD.define(S512, 30, "Z30")
	Z31 = SIMD.define(S512, 31, "Z31")
)
