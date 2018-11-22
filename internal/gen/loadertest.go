package gen

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/mmcloughlin/avo/internal/inst"
)

type LoaderTest struct{}

func (l LoaderTest) Generate(w io.Writer, is []*inst.Instruction) error {
	p := &printer{w: w}

	p.printf("TEXT loadertest(SB), 0, $0\n")

	counts := map[string]int{}

	for _, i := range is {
		p.printf("\t// %s %s\n", i.Opcode, i.Summary)
		if skip, msg := l.skip(i.Opcode); skip {
			p.printf("\t// SKIP: %s\n", msg)
			counts["skip"]++
			continue
		}

		for _, f := range i.Forms {
			as := args(f.Operands)
			p.printf("\t// %#v\n", f.Operands)
			if as == nil {
				p.printf("\t// TODO\n")
				counts["todo"]++
				continue
			}
			p.printf("\t%s\t%s\n", i.Opcode, strings.Join(as, ", "))
			counts["total"]++
		}
		p.printf("\n")
	}

	p.printf("\tRET\n")

	for m, c := range counts {
		p.printf("// %s: %d\n", m, c)
	}

	return p.Err()
}

func (l LoaderTest) skip(opcode string) (bool, string) {
	prefixes := map[string]string{
		"PUSH": "PUSH can produce 'unbalanced PUSH/POP' assembler error",
		"POP":  "POP can produce 'unbalanced PUSH/POP' assembler error",
	}
	for p, m := range prefixes {
		if strings.HasPrefix(opcode, p) {
			return true, m
		}
	}
	return false, ""
}

func args(ops []inst.Operand) []string {
	as := make([]string, len(ops))
	for i, op := range ops {
		a := arg(op.Type)
		if a == "" {
			return nil
		}
		as[i] = a
	}
	return as
}

// arg generates an argument for an operand of the given type.
func arg(t string) string {
	m := map[string]string{
		"1":     "$1", // <xs:enumeration value="1" />
		"3":     "$3", // <xs:enumeration value="3" />
		"imm2u": "$3",
		// <xs:enumeration value="imm4" />
		"imm8":  fmt.Sprintf("$%d", math.MaxInt8),  // <xs:enumeration value="imm8" />
		"imm16": fmt.Sprintf("$%d", math.MaxInt16), // <xs:enumeration value="imm16" />
		"imm32": fmt.Sprintf("$%d", math.MaxInt32), // <xs:enumeration value="imm32" />
		"imm64": fmt.Sprintf("$%d", math.MaxInt64), // <xs:enumeration value="imm64" />

		"al":   "AL",  // <xs:enumeration value="al" />
		"cl":   "CL",  // <xs:enumeration value="cl" />
		"r8":   "CH",  // <xs:enumeration value="r8" />
		"ax":   "AX",  // <xs:enumeration value="ax" />
		"r16":  "SI",  // <xs:enumeration value="r16" />
		"eax":  "AX",  // <xs:enumeration value="eax" />
		"r32":  "DX",  // <xs:enumeration value="r32" />
		"rax":  "AX",  // <xs:enumeration value="rax" />
		"r64":  "R15", // <xs:enumeration value="r64" />
		"mm":   "M5",  // <xs:enumeration value="mm" />
		"xmm0": "X0",  // <xs:enumeration value="xmm0" />
		"xmm":  "X7",  // <xs:enumeration value="xmm" />
		// <xs:enumeration value="xmm{k}" />
		// <xs:enumeration value="xmm{k}{z}" />
		"ymm": "Y13", // <xs:enumeration value="ymm" />
		// <xs:enumeration value="ymm{k}" />
		// <xs:enumeration value="ymm{k}{z}" />
		// <xs:enumeration value="zmm" />
		// <xs:enumeration value="zmm{k}" />
		// <xs:enumeration value="zmm{k}{z}" />
		// <xs:enumeration value="k" />
		// <xs:enumeration value="k{k}" />
		// <xs:enumeration value="moffs32" />
		// <xs:enumeration value="moffs64" />
		// <xs:enumeration value="m" />
		"m8":  "8(AX)(CX*2)",  // <xs:enumeration value="m8" />
		"m16": "16(AX)(CX*2)", // <xs:enumeration value="m16" />
		// <xs:enumeration value="m16{k}{z}" />
		"m32": "32(AX)(CX*2)", // <xs:enumeration value="m32" />
		// <xs:enumeration value="m32{k}" />
		// <xs:enumeration value="m32{k}{z}" />
		"m64": "64(AX)(CX*2)", // <xs:enumeration value="m64" />
		// <xs:enumeration value="m64{k}" />
		// <xs:enumeration value="m64{k}{z}" />
		"m128": "128(AX)(CX*2)", // <xs:enumeration value="m128" />
		// <xs:enumeration value="m128{k}{z}" />
		"m256": "256(AX)(CX*2)", // <xs:enumeration value="m256" />
		// <xs:enumeration value="m256{k}{z}" />
		// <xs:enumeration value="m512" />
		// <xs:enumeration value="m512{k}{z}" />
		// <xs:enumeration value="m64/m32bcst" />
		// <xs:enumeration value="m128/m32bcst" />
		// <xs:enumeration value="m256/m32bcst" />
		// <xs:enumeration value="m512/m32bcst" />
		// <xs:enumeration value="m128/m64bcst" />
		// <xs:enumeration value="m256/m64bcst" />
		// <xs:enumeration value="m512/m64bcst" />
		// <xs:enumeration value="vm32x" />
		// <xs:enumeration value="vm32x{k}" />
		// <xs:enumeration value="vm64x" />
		// <xs:enumeration value="vm64x{k}" />
		// <xs:enumeration value="vm32y" />
		// <xs:enumeration value="vm32y{k}" />
		// <xs:enumeration value="vm64y" />
		// <xs:enumeration value="vm64y{k}" />
		// <xs:enumeration value="vm32z" />
		// <xs:enumeration value="vm32z{k}" />
		// <xs:enumeration value="vm64z" />
		// <xs:enumeration value="vm64z{k}" />
		// <xs:enumeration value="rel8" />
		// <xs:enumeration value="rel32" />
		// <xs:enumeration value="{er}" />
		// <xs:enumeration value="{sae}" />

		// Appear unused:
		"r8l":  "????", // <xs:enumeration value="r8l" />
		"r16l": "????", // <xs:enumeration value="r16l" />
		"r32l": "????", // <xs:enumeration value="r32l" />
	}
	return m[t]
}
