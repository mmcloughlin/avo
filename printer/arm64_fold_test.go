package printer_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/printer"
	"github.com/mmcloughlin/avo/reg"
)

// foldContext builds a single function from body, for the copy-fold tests.
func foldContext(body func(ctx *build.Context)) *build.Context {
	ctx := build.NewContext()
	ctx.Function("f")
	ctx.SignatureExpr("func()")
	body(ctx)
	ctx.RET()
	return ctx
}

// expectLines checks for the presence and absence of instructions in out.
// Operands are column-aligned per block, so runs of spaces are collapsed on
// both sides before comparing.
func expectLines(t *testing.T, out string, want []string, reject []string) {
	t.Helper()
	norm := regexp.MustCompile(`[ \t]+`)
	flat := norm.ReplaceAllString(out, " ")
	for _, w := range want {
		if !strings.Contains(flat, norm.ReplaceAllString(w, " ")) {
			t.Errorf("expected %q in:\n%s", w, out)
		}
	}
	for _, r := range reject {
		if strings.Contains(flat, norm.ReplaceAllString(r, " ")) {
			t.Errorf("did not expect %q in:\n%s", r, out)
		}
	}
}

// TestARM64ShiftCountFold covers the count fold (see shiftFolds): the copy
// into CL is dropped and the shift reads the original register, but only
// when RCX is provably dead after the shift and the source is unchanged.
func TestARM64ShiftCountFold(t *testing.T) {
	t.Run("Folded", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.MOVQ(reg.R9, reg.RBX)
			ctx.SHLQ(reg.CL, reg.RBX)
			ctx.MOVQ(operand.U32(0), reg.RCX) // fresh definition: the copy is dead
		}), printer.NewARM64Asm)
		// Both folds apply: count from R7 (x86 R8), source from R8 (x86 R9).
		expectLines(t, out, []string{"LSL R7, R8, R3"}, []string{"MOVD R7, R1", "MOVD R8, R3"})
	})
	t.Run("FoldedROL", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.ROLQ(reg.CL, reg.RBX)
			ctx.MOVQ(operand.U32(0), reg.RCX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"NEG R7, R16", "ROR R16, R3, R3"}, []string{"MOVD R7, R1"})
	})
	t.Run("Folded32", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.SHRL(reg.CL, reg.EBX)
			ctx.XORQ(reg.RCX, reg.RCX) // a cancelling XOR is a pure definition
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"LSRW R7, R3, R3"}, []string{"MOVD R7, R1"})
	})
	t.Run("Folded32BitRedefinition", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.SHLQ(reg.CL, reg.RBX)
			ctx.MOVL(operand.U32(1), reg.ECX) // zero-extends: all of RCX replaced
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"LSL R7, R3, R3"}, []string{"MOVD R7, R1"})
	})
	// A narrower write keeps the rest of RCX, so the copy is still observable.
	for _, c := range []struct {
		name  string
		write func(ctx *build.Context)
	}{
		{"RefusedByteRedefinition", func(ctx *build.Context) { ctx.MOVB(operand.U8(1), reg.CL) }},
		{"RefusedWordRedefinition", func(ctx *build.Context) { ctx.MOVW(operand.U16(1), reg.CX) }},
		{"RefusedSetccRedefinition", func(ctx *build.Context) {
			ctx.TESTQ(reg.RBX, reg.RBX)
			ctx.SETPL(reg.CL)
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			out := Print(t, foldContext(func(ctx *build.Context) {
				ctx.MOVQ(reg.R8, reg.RCX)
				ctx.SHLQ(reg.CL, reg.RBX)
				c.write(ctx)
			}), printer.NewARM64Asm)
			expectLines(t, out, []string{"MOVD R7, R1", "LSL R1, R3, R3"}, nil)
		})
	}
	t.Run("RefusedLabelBeforeRedefinition", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.SHLQ(reg.CL, reg.RBX)
			ctx.Label("join")
			ctx.MOVQ(operand.U32(0), reg.RCX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R7, R1", "LSL R1, R3, R3"}, nil)
	})
	t.Run("RefusedReadAfterShift", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.SHLQ(reg.CL, reg.RBX)
			ctx.MOVQ(reg.RCX, reg.RDX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R7, R1", "LSL R1, R3, R3"}, nil)
	})
	t.Run("RefusedReadModifyWriteAfterShift", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.SHLQ(reg.CL, reg.RBX)
			ctx.ADDQ(operand.U32(1), reg.RCX) // writes RCX, but reads it too
			ctx.MOVQ(reg.RCX, reg.RDX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R7, R1", "LSL R1, R3, R3"}, nil)
	})
	t.Run("RefusedSourceWrittenBeforeShift", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.ADDQ(operand.U32(1), reg.R8)
			ctx.SHLQ(reg.CL, reg.RBX)
			ctx.MOVQ(operand.U32(0), reg.RCX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R7, R1", "LSL R1, R3, R3"}, nil)
	})
	t.Run("RefusedSourceByteWrittenBeforeShift", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.MOVB(operand.U8(1), reg.R8B) // a partial write is still a write
			ctx.SHLQ(reg.CL, reg.RBX)
			ctx.MOVQ(operand.U32(0), reg.RCX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R7, R1", "LSL R1, R3, R3"}, nil)
	})
	t.Run("RefusedShiftOfRCX", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.SHLQ(reg.CL, reg.RCX)
			ctx.MOVQ(operand.U32(0), reg.RDX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R7, R1", "LSL R1, R1, R1"}, nil)
	})
	t.Run("RefusedBranchBeforeRedefinition", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.SHLQ(reg.CL, reg.RBX)
			ctx.JMP(operand.LabelRef("next"))
			ctx.Label("next")
			ctx.MOVQ(operand.U32(0), reg.RCX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R7, R1", "LSL R1, R3, R3"}, nil)
	})
	t.Run("RefusedMemoryOperandReadsRCX", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.SHLQ(reg.CL, reg.RBX)
			ctx.MOVQ(operand.Mem{Base: reg.RCX}, reg.RCX) // reads RCX as an address
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R7, R1", "LSL R1, R3, R3"}, nil)
	})
	t.Run("RefusedNoRedefinitionBeforeReturn", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R8, reg.RCX)
			ctx.SHLQ(reg.CL, reg.RBX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R7, R1", "LSL R1, R3, R3"}, nil)
	})
}

// TestARM64ShiftSourceFold covers the source fold (see shiftFolds): a copy
// whose only reader is a shift that overwrites it becomes the shift's source.
func TestARM64ShiftSourceFold(t *testing.T) {
	t.Run("Folded", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R9, reg.RBX)
			ctx.SHRQ(operand.U8(3), reg.RBX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"LSR $0x03, R8, R3"}, []string{"MOVD R8, R3"})
	})
	t.Run("FoldedROLImmediate", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R9, reg.RBX)
			ctx.ROLQ(operand.U8(5), reg.RBX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"ROR $59, R8, R3"}, []string{"MOVD R8, R3"})
	})
	t.Run("FoldedAcrossNeutralInstruction", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R9, reg.RBX)
			ctx.ADDQ(operand.U32(1), reg.RDX)
			ctx.SHRQ(operand.U8(3), reg.RBX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"LSR $0x03, R8, R3"}, []string{"MOVD R8, R3"})
	})
	t.Run("FoldedIntoRCXWithImmediateCount", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R9, reg.RCX)
			ctx.SHLQ(operand.U8(2), reg.RCX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"LSL $0x02, R8, R1"}, []string{"MOVD R8, R1"})
	})
	t.Run("RefusedDestinationReadBetween", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R9, reg.RBX)
			ctx.ADDQ(reg.RBX, reg.RDX)
			ctx.SHRQ(operand.U8(3), reg.RBX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R8, R3", "LSR $0x03, R3, R3"}, nil)
	})
	t.Run("RefusedSourceWrittenBetween", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R9, reg.RBX)
			ctx.ADDQ(operand.U32(1), reg.R9)
			ctx.SHRQ(operand.U8(3), reg.RBX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R8, R3", "LSR $0x03, R3, R3"}, nil)
	})
	t.Run("RefusedLabelBetween", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R9, reg.RBX)
			ctx.Label("join")
			ctx.SHRQ(operand.U8(3), reg.RBX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R8, R3", "LSR $0x03, R3, R3"}, nil)
	})
	t.Run("RefusedCountAliasesCopy", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R9, reg.RCX)
			ctx.SHLQ(reg.CL, reg.RCX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD R8, R1", "LSL R1, R1, R1"}, nil)
	})
	t.Run("RefusedDoubleShift", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected the three-operand shift to be rejected")
			}
		}()
		Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(reg.R9, reg.RBX)
			ctx.SHLQ(operand.U8(8), reg.RDX, reg.RBX)
		}), printer.NewARM64Asm)
	})
}

// TestARM64SetccZeroFold covers setccFolds: a SETcc into a register that was
// zeroed and not touched since writes the whole register with CSET.
func TestARM64SetccZeroFold(t *testing.T) {
	t.Run("FoldedXORL", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.XORL(reg.EAX, reg.EAX)
			ctx.CMPQ(reg.RBX, reg.RDX)
			ctx.SETGE(reg.AL)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"CSET GE, R0"}, []string{"MOVD $0, R0", "BFI"})
	})
	t.Run("FoldedMOVQZero", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.MOVQ(operand.U32(0), reg.RAX)
			ctx.CMPQ(reg.RBX, reg.RDX)
			ctx.SETEQ(reg.AL)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"CSET EQ, R0"}, []string{"MOVD $0x00000000, R0", "MOVD $0, R0", "BFI"})
	})
	t.Run("RefusedRegisterReadBetween", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.XORQ(reg.RAX, reg.RAX)
			ctx.ADDQ(reg.RAX, reg.RBX)
			ctx.CMPQ(reg.RBX, reg.RDX)
			ctx.SETGE(reg.AL)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD $0, R0", "CSET GE, R16", "BFI $0, R16, $8, R0"}, nil)
	})
	t.Run("RefusedZeroingFlagsConsumed", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.XORQ(reg.RAX, reg.RAX)
			ctx.SETEQ(reg.BL) // reads the XOR's ZF
			ctx.CMPQ(reg.RBX, reg.RDX)
			ctx.SETGE(reg.AL)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD $0, R0", "TST R0, R0", "BFI $0, R16, $8, R0"}, nil)
	})
	t.Run("RefusedLabelBetween", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.XORQ(reg.RAX, reg.RAX)
			ctx.Label("join")
			ctx.CMPQ(reg.RBX, reg.RDX)
			ctx.SETGE(reg.AL)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"MOVD $0, R0", "BFI $0, R16, $8, R0"}, nil)
	})
	t.Run("RefusedHighByte", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected SETcc into a high byte to be rejected")
			}
		}()
		Print(t, foldContext(func(ctx *build.Context) {
			ctx.XORQ(reg.RAX, reg.RAX)
			ctx.CMPQ(reg.RBX, reg.RDX)
			ctx.SETGE(reg.AH)
		}), printer.NewARM64Asm)
	})
}

// TestARM64ADCQ covers the full-width carry accumulate.
func TestARM64ADCQ(t *testing.T) {
	t.Run("Lowered", func(t *testing.T) {
		out := Print(t, foldContext(func(ctx *build.Context) {
			ctx.CMPQ(reg.RAX, operand.U8(4))
			ctx.ADCQ(operand.I8(0), reg.RBX)
		}), printer.NewARM64Asm)
		expectLines(t, out, []string{"CSINC HS, R3, R3, R3"}, []string{"BFI"})
	})
	t.Run("NonZeroImmediateRejected", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected ADCQ with a non-zero immediate to be rejected")
			}
		}()
		Print(t, foldContext(func(ctx *build.Context) {
			ctx.CMPQ(reg.RAX, operand.U8(4))
			ctx.ADCQ(operand.I8(1), reg.RBX)
		}), printer.NewARM64Asm)
	})
}

// TestARM64MADDFold covers maddFolds: an adjacent "IMULQ a, t; ADDQ x, y" pair
// becomes one MADD when the add accumulates into the product, or adds the
// product into another register that t is provably dead after -- including
// across a loop's back edge -- and is left alone otherwise.
func TestARM64MADDFold(t *testing.T) {
	cases := []struct {
		name   string
		body   func(ctx *build.Context)
		want   []string
		reject []string
	}{
		{
			name: "AccumulateIntoProduct",
			body: func(ctx *build.Context) {
				ctx.IMULQ(reg.RBX, reg.RAX)
				ctx.ADDQ(reg.RCX, reg.RAX) // RAX = RAX*RBX + RCX
				ctx.ADDQ(reg.RAX, reg.RDX) // RAX stays live: this form needs no proof
			},
			want:   []string{"MADD R3, R1, R0, R0"},
			reject: []string{"MUL R3, R0, R0"},
		},
		{
			name: "AddProductDeadAfter",
			body: func(ctx *build.Context) {
				ctx.IMULQ(reg.RBX, reg.RAX)
				ctx.ADDQ(reg.RAX, reg.RCX) // RCX += RAX*RBX
				ctx.MOVQ(operand.U32(0), reg.RAX)
			},
			want:   []string{"MADD R3, R1, R0, R1"},
			reject: []string{"MUL R3, R0, R0"},
		},
		{
			name: "AddProductDeadAtReturn",
			body: func(ctx *build.Context) {
				ctx.IMULQ(reg.RBX, reg.RAX)
				ctx.ADDQ(reg.RAX, reg.RCX)
			},
			want:   []string{"MADD R3, R1, R0, R1"},
			reject: []string{"MUL R3, R0, R0"},
		},
		{
			name: "AddProductLiveAfter",
			body: func(ctx *build.Context) {
				ctx.IMULQ(reg.RBX, reg.RAX)
				ctx.ADDQ(reg.RAX, reg.RCX)
				ctx.ADDQ(reg.RAX, reg.RDX) // reads the product MADD would never write
			},
			want:   []string{"MUL R3, R0, R0", "ADD R0, R1, R1"},
			reject: []string{"MADD"},
		},
		{
			name: "AddProductLiveThroughPartialRedefinition",
			body: func(ctx *build.Context) {
				ctx.IMULQ(reg.RBX, reg.RAX)
				ctx.ADDQ(reg.RAX, reg.RCX)
				ctx.MOVW(operand.U16(1), reg.AX) // keeps bits 16-63 of the product
				ctx.ADDQ(reg.RAX, reg.RDX)
			},
			want:   []string{"MUL R3, R0, R0"},
			reject: []string{"MADD"},
		},
		{
			name: "AddProductLiveAcrossBackEdge",
			body: func(ctx *build.Context) {
				ctx.Label("loop")
				ctx.ADDQ(reg.RAX, reg.RDX) // the previous iteration's product
				ctx.IMULQ(reg.RBX, reg.RAX)
				ctx.ADDQ(reg.RAX, reg.RCX)
				ctx.DECQ(reg.RSI)
				ctx.JNZ(operand.LabelRef("loop"))
			},
			want:   []string{"MUL R3, R0, R0"},
			reject: []string{"MADD"},
		},
		{
			name: "AddProductDeadAcrossBackEdge",
			body: func(ctx *build.Context) {
				ctx.Label("loop")
				ctx.MOVQ(reg.RDI, reg.RAX) // redefined before any read
				ctx.IMULQ(reg.RBX, reg.RAX)
				ctx.ADDQ(reg.RAX, reg.RCX)
				ctx.DECQ(reg.RSI)
				ctx.JNZ(operand.LabelRef("loop"))
			},
			want:   []string{"MADD R3, R1, R0, R1"},
			reject: []string{"MUL R3, R0, R0"},
		},
		{
			name: "FlagsConsumed",
			body: func(ctx *build.Context) {
				ctx.IMULQ(reg.RBX, reg.RAX)
				ctx.ADDQ(reg.RCX, reg.RAX)
				ctx.JZ(operand.LabelRef("done")) // reads the ADDQ's flags
				ctx.MOVQ(operand.U32(1), reg.RDX)
				ctx.Label("done")
			},
			want:   []string{"MUL R3, R0, R0"},
			reject: []string{"MADD"},
		},
		{
			name: "NotAdjacent",
			body: func(ctx *build.Context) {
				ctx.IMULQ(reg.RBX, reg.RAX)
				ctx.MOVQ(reg.RDX, reg.RSI)
				ctx.ADDQ(reg.RCX, reg.RAX)
			},
			want:   []string{"MUL R3, R0, R0"},
			reject: []string{"MADD"},
		},
		{
			name: "Doubling",
			body: func(ctx *build.Context) {
				ctx.IMULQ(reg.RBX, reg.RAX)
				ctx.ADDQ(reg.RAX, reg.RAX)
			},
			want:   []string{"MUL R3, R0, R0", "ADD R0, R0, R0"},
			reject: []string{"MADD"},
		},
		{
			name: "ThirtyTwoBit",
			body: func(ctx *build.Context) {
				ctx.IMULL(reg.EBX, reg.EAX)
				ctx.ADDL(reg.ECX, reg.EAX)
			},
			reject: []string{"MADD"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := Print(t, foldContext(c.body), printer.NewARM64Asm)
			expectLines(t, out, c.want, c.reject)
		})
	}
}
