package printer_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/printer"
	"github.com/mmcloughlin/avo/reg"
)

// twinTestContext builds a file with a generic/BMI2 twin pair and a
// single-variant function, shared by both twin-preference tests below.
func twinTestContext() *build.Context {
	ctx := build.NewContext()
	for _, name := range []string{"twin_amd64", "twin_bmi2", "solo_amd64"} {
		ctx.Function(name)
		ctx.SignatureExpr("func()")
		ctx.RET()
	}
	return ctx
}

// printARM64 prints ctx with the arm64 lowering printer under the given config.
func printARM64(t *testing.T, ctx *build.Context, cfg printer.Config) string {
	t.Helper()
	f, errs := ctx.Result()
	if errs != nil {
		t.Fatal(errs)
	}
	b, err := printer.NewARM64Asm(cfg).Print(f)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestARM64PrefersGenericTwinByDefault checks that with the default config the
// arm64 lowering picks a function's generic twin over its BMI2 one: BMI2 x86
// code is tuned for x86 and is not reliably faster once mechanically lowered
// (e.g. BEXTR packs into one instruction what arm64 needs two UBFX to unpack),
// so this is the safe default absent a measured reason to prefer BMI2. A
// function with only one variant is lowered regardless.
func TestARM64PrefersGenericTwinByDefault(t *testing.T) {
	out := printARM64(t, twinTestContext(), printer.NewDefaultConfig())

	if n := strings.Count(out, "TEXT ·twin_arm64(SB)"); n != 1 {
		t.Errorf("expected exactly one twin_arm64 definition, got %d:\n%s", n, out)
	}
	if !strings.Contains(out, "skipped twin_bmi2") {
		t.Errorf("expected BMI2 twin_bmi2 to be skipped by default:\n%s", out)
	}
	if strings.Contains(out, "TEXT ·twin_amd64(SB)") {
		t.Errorf("generic variant should be renamed, not emitted as twin_amd64:\n%s", out)
	}
	if !strings.Contains(out, "TEXT ·solo_arm64(SB)") {
		t.Errorf("expected single-variant solo_amd64 lowered to solo_arm64:\n%s", out)
	}
}

// TestARM64PreferBMI2TwinOptIn checks the ARM64PreferBMI2 config opt-in flips
// the choice to the BMI2 twin, for callers who have actually measured it to be
// faster for their case.
func TestARM64PreferBMI2TwinOptIn(t *testing.T) {
	cfg := printer.NewDefaultConfig()
	cfg.ARM64PreferBMI2 = true
	out := printARM64(t, twinTestContext(), cfg)

	if n := strings.Count(out, "TEXT ·twin_arm64(SB)"); n != 1 {
		t.Errorf("expected exactly one twin_arm64 definition, got %d:\n%s", n, out)
	}
	if !strings.Contains(out, "skipped twin_amd64") {
		t.Errorf("expected generic twin_amd64 to be skipped when BMI2 is preferred:\n%s", out)
	}
	if strings.Contains(out, "TEXT ·twin_bmi2(SB)") {
		t.Errorf("BMI2 variant should be renamed, not emitted as twin_bmi2:\n%s", out)
	}
	if !strings.Contains(out, "TEXT ·solo_arm64(SB)") {
		t.Errorf("expected single-variant solo_amd64 lowered to solo_arm64:\n%s", out)
	}
}

// TestARM64SubwordCompareGuard checks the sub-32-bit CMPW guard: lowering
// succeeds when the only consumer is EQ/NE (lowerSubwordCompareEqNe applies),
// and still panics for an ordering consumer (JLT), which would need
// sign-aware operand extension this printer does not model.
func TestARM64SubwordCompareGuard(t *testing.T) {
	t.Run("EqNeAllowed", func(t *testing.T) {
		ctx := build.NewContext()
		ctx.Function("f")
		ctx.SignatureExpr("func()")
		ctx.CMPW(reg.RAX.As16(), reg.RCX.As16())
		ctx.JEQ(operand.LabelRef("yes"))
		ctx.Label("yes")
		ctx.RET()

		out := Print(t, ctx, printer.NewARM64Asm)
		if !strings.Contains(out, "CMP") {
			t.Errorf("expected a lowered CMP, got:\n%s", out)
		}
	})

	t.Run("OrderingPanics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected lowering CMPW followed by JLT to panic")
			}
		}()

		ctx := build.NewContext()
		ctx.Function("f")
		ctx.SignatureExpr("func()")
		ctx.CMPW(reg.RAX.As16(), reg.RCX.As16())
		ctx.JLT(operand.LabelRef("yes"))
		ctx.Label("yes")
		ctx.RET()

		Print(t, ctx, printer.NewARM64Asm)
	})
}

// TestARM64CrossLabelFlagsRejected checks that a conditional branch whose flags
// are produced before a label fails generation instead of silently emitting a
// non-flag-setting op. The lowering recovers the producer/consumer link by
// scanning backwards through a straight-line run; when a label intervenes the
// flags arrive along a control-flow edge it does not model, and quietly leaving
// the producer unmarked would let the branch read whatever NZCV survived.
func TestARM64CrossLabelFlagsRejected(t *testing.T) {
	ctx := build.NewContext()
	ctx.Function("crosslabel")
	ctx.SignatureExpr("func(x, y uint64) uint64")
	x, y := reg.RAX, reg.RCX
	ctx.SUBQ(y, x) // producer
	ctx.Label("join")
	ctx.JEQ(operand.LabelRef("yes")) // consumer, separated by the label
	ctx.Label("yes")
	ctx.RET()

	f, errs := ctx.Result()
	if errs != nil {
		t.Fatal(errs)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic for flags read across a label, got none")
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "across a label") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	_, _ = printer.NewARM64Asm(printer.NewGoRunConfig()).Print(f)
}

// TestARM64CallRejected checks that a CALL fails generation. The lowering emits
// NOFRAME leaf functions and uses caller-saved registers (including R16/R17 and
// the two scratch registers) without preserving them, all of which is only
// sound while the function makes no calls.
func TestARM64CallRejected(t *testing.T) {
	ctx := build.NewContext()
	ctx.Function("calls")
	ctx.SignatureExpr("func()")
	ctx.CALL(operand.LabelRef("somewhere"))
	ctx.RET()

	f, errs := ctx.Result()
	if errs != nil {
		t.Fatal(errs)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic for CALL, got none")
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "CALL") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	_, _ = printer.NewARM64Asm(printer.NewGoRunConfig()).Print(f)
}

// TestARM64FlagSemanticGuards covers sequences the lowering must refuse rather
// than miscompile. Each pairs a flag producer with a consumer whose condition
// the arm64 translation cannot faithfully reproduce, because the two
// architectures disagree about what the flag means after that producer.
func TestARM64FlagSemanticGuards(t *testing.T) {
	cases := []struct {
		name  string
		build func(ctx *build.Context)
		want  string
	}{
		{
			// x86 TEST forces CF=0 and so does arm64 TST, meaning the raw bits
			// agree; the meaning-inverting condition map (correct after a
			// compare) then flips the answer. JA is never taken after this TEST
			// on x86, but BHI would always be taken.
			name: "carry condition after TEST",
			build: func(ctx *build.Context) {
				ctx.TESTQ(reg.RAX, reg.RCX)
				ctx.JHI(operand.LabelRef("l"))
			},
			want: "only the ZF/SF conditions",
		},
		{
			// x86 INC leaves CF alone, so this reads the compare's borrow. The
			// arm64 ADDS it lowers to overwrites C with the increment's carry.
			name: "carry condition across INC",
			build: func(ctx *build.Context) {
				ctx.CMPQ(reg.RAX, reg.RCX)
				ctx.INCQ(reg.RDX)
				ctx.JCS(operand.LabelRef("l"))
			},
			want: "preserves CF on x86",
		},
		{
			// ADC's lowering hardcodes the borrow convention, which only holds
			// after a compare or subtract. After an addition both architectures
			// use the same carry-out sense, so it would increment inversely.
			name: "ADC after an addition",
			build: func(ctx *build.Context) {
				ctx.ADDQ(reg.RAX, reg.RCX)
				ctx.ADCB(operand.I8(0), reg.DL)
			},
			want: "borrow-producing compare or subtract",
		},
		{
			// A signed sub-word compare whose consumer sits past a
			// flag-transparent MOV. The zero-extending translation is only valid
			// for EQ/NE, so this must not be silently blessed.
			name: "signed sub-word compare past a MOV",
			build: func(ctx *build.Context) {
				ctx.CMPB(reg.AL, reg.CL)
				ctx.MOVQ(operand.U64(1), reg.RAX)
				ctx.JLT(operand.LabelRef("l"))
			},
			want: "sub-32-bit compare",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := build.NewContext()
			ctx.Function("guard")
			ctx.SignatureExpr("func()")
			c.build(ctx)
			ctx.Label("l")
			ctx.RET()
			f, errs := ctx.Result()
			if errs != nil {
				t.Fatal(errs)
			}
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("expected a panic, got none")
				}
				msg, ok := r.(string)
				if !ok || !strings.Contains(msg, c.want) {
					t.Fatalf("panic %q does not mention %q", r, c.want)
				}
			}()
			_, _ = printer.NewARM64Asm(printer.NewGoRunConfig()).Print(f)
		})
	}
}

// TestARM64AddCarryGuard checks that a carry-condition consumer after an
// addition fails generation. x86 and arm64 both set carry-out on an add, so the
// condition map -- which translates by post-compare meaning, where the two use
// opposite borrow conventions -- would invert the test.
func TestARM64AddCarryGuard(t *testing.T) {
	ctx := build.NewContext()
	ctx.Function("addcarry")
	ctx.SignatureExpr("func()")
	ctx.ADDQ(reg.RAX, reg.RCX)
	ctx.JCS(operand.LabelRef("wrap"))
	ctx.Label("wrap")
	ctx.RET()

	f, errs := ctx.Result()
	if errs != nil {
		t.Fatal(errs)
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic for a carry condition after ADD")
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "only a compare or subtract") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	_, _ = printer.NewARM64Asm(printer.NewGoRunConfig()).Print(f)
}

// TestARM64HighByteEncodability checks that byte extends from a high-byte
// register are refused where x86-64 cannot encode them. Any REX-carrying form
// renames that operand to SPL, so emitting a faithful arm64 extract would make
// the two architectures compute different values from one program.
func TestARM64HighByteEncodability(t *testing.T) {
	cases := []struct {
		name  string
		build func(ctx *build.Context)
		want  string
	}{
		{
			name:  "64-bit destination",
			build: func(ctx *build.Context) { ctx.MOVBQZX(reg.AH, reg.RBX) },
			want:  "forces a REX prefix",
		},
		{
			name:  "extended destination",
			build: func(ctx *build.Context) { ctx.MOVBLZX(reg.AH, reg.R9L) },
			want:  "forces a REX prefix",
		},
		// The pairings below all encode on amd64 rather than being rejected --
		// the assembler silently substitutes SPL -- so each is a case where the
		// two architectures would otherwise compute different things.
		{
			name:  "byte move to extended register",
			build: func(ctx *build.Context) { ctx.MOVB(reg.AH, reg.R8B) },
			want:  "forces a REX prefix",
		},
		{
			name:  "byte move from extended register",
			build: func(ctx *build.Context) { ctx.MOVB(reg.R8B, reg.AH) },
			want:  "forces a REX prefix",
		},
		{
			name:  "byte move through extended base",
			build: func(ctx *build.Context) { ctx.MOVB(reg.AH, operand.Mem{Base: reg.R8}) },
			want:  "forces a REX prefix",
		},
		{
			// SIL is not an extended register, but it exists only under REX:
			// without the prefix that encoding names AH.
			name:  "byte add against SIL",
			build: func(ctx *build.Context) { ctx.ADDB(reg.SIB, reg.AH) },
			want:  "forces a REX prefix",
		},
		{
			name: "sub-word compare against extended register",
			build: func(ctx *build.Context) {
				ctx.CMPB(reg.AH, reg.R8B)
				ctx.JEQ(operand.LabelRef("hb_yes"))
				ctx.Label("hb_yes")
			},
			want: "forces a REX prefix",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := build.NewContext()
			ctx.Function("hb")
			ctx.SignatureExpr("func()")
			c.build(ctx)
			ctx.RET()
			f, errs := ctx.Result()
			if errs != nil {
				t.Fatal(errs)
			}
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("expected a panic")
				}
				if msg, ok := r.(string); !ok || !strings.Contains(msg, c.want) {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()
			_, _ = printer.NewARM64Asm(printer.NewGoRunConfig()).Print(f)
		})
	}
}

// dispatch loop, and tests here that fail when that barrier is removed.

// TestARM64NoProducerRejected checks that a flag consumer with no producer
// anywhere in the function fails generation rather than branching on whatever
// NZCV the caller happened to leave.
func TestARM64NoProducerRejected(t *testing.T) {
	ctx := build.NewContext()
	ctx.Function("noprod")
	ctx.SignatureExpr("func()")
	ctx.JEQ(operand.LabelRef("np_yes"))
	ctx.Label("np_yes")
	ctx.RET()

	f, errs := ctx.Result()
	if errs != nil {
		t.Fatal(errs)
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic for a consumer with no producer")
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "no producer") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	_, _ = printer.NewARM64Asm(printer.NewGoRunConfig()).Print(f)
}

// TestARM64DoubleShiftRejected checks the three-operand SHL/SHR forms, which
// Go's assembler encodes as the double-precision shifts SHLD/SHRD -- a
// different instruction, whose destination is the third operand.
func TestARM64DoubleShiftRejected(t *testing.T) {
	for _, op := range []string{"SHLQ", "SHRQ", "SHLL", "SHRL"} {
		t.Run(op, func(t *testing.T) {
			ctx := build.NewContext()
			ctx.Function("dbl")
			ctx.SignatureExpr("func()")
			switch op {
			case "SHLQ":
				ctx.SHLQ(operand.U8(8), reg.RDX, reg.RAX)
			case "SHRQ":
				ctx.SHRQ(operand.U8(8), reg.RDX, reg.RAX)
			case "SHLL":
				ctx.SHLL(operand.U8(8), reg.EDX, reg.EAX)
			case "SHRL":
				ctx.SHRL(operand.U8(8), reg.EDX, reg.EAX)
			}
			ctx.RET()

			f, errs := ctx.Result()
			if errs != nil {
				t.Fatal(errs)
			}
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("expected a panic for three-operand %s", op)
				}
				if msg, ok := r.(string); !ok || !strings.Contains(msg, "double-precision shift") {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()
			_, _ = printer.NewARM64Asm(printer.NewGoRunConfig()).Print(f)
		})
	}
}

// TestARM64BitTestGuards checks that BTL is accepted only in the one shape that
// fuses into an arm64 test-and-branch, and refused everywhere else.
//
// The refusals are the substance of the feature, not trimming around it. x86 BT
// defines only CF; OF/SF/AF/PF are architecturally undefined, and AMD has
// printed ZF that way too. So a BTL whose flags reach anything other than the
// carry branch fused with it would be a branch on flags no manual promises.
// Keeping BTL out of both flagSetter and isFlagTransparent is what makes those
// cases land on the existing unknown-producer panic rather than lower quietly.
func TestARM64BitTestGuards(t *testing.T) {
	for _, tc := range []struct {
		name  string
		build func(ctx *build.Context)
		want  string
	}{
		{
			// The one accepted shape, as a control on the refusals below.
			name: "carry branch fused",
			build: func(ctx *build.Context) {
				ctx.BTL(operand.U8(0), reg.ECX)
				ctx.JC(operand.LabelRef("target"))
			},
			want: "",
		},
		{
			name: "non-carry consumer",
			build: func(ctx *build.Context) {
				ctx.BTL(operand.U8(0), reg.ECX)
				ctx.JEQ(operand.LabelRef("target"))
			},
			want: "carry branch",
		},
		{
			// x86 leaves CF live after BT, so a second consumer is legal there;
			// the fused TBNZ writes no flags, so it must not be lowered.
			name: "second carry consumer",
			build: func(ctx *build.Context) {
				ctx.BTL(operand.U8(0), reg.ECX)
				ctx.JC(operand.LabelRef("target"))
				ctx.JC(operand.LabelRef("other"))
			},
			want: "cannot emit as a flag-setter",
		},
		{
			name: "no consumer",
			build: func(ctx *build.Context) {
				ctx.BTL(operand.U8(0), reg.ECX)
				ctx.MOVQ(reg.RAX, reg.RBX)
			},
			want: "carry branch",
		},
		{
			name: "register bit index",
			build: func(ctx *build.Context) {
				ctx.BTL(reg.EAX, reg.ECX)
				ctx.JC(operand.LabelRef("target"))
			},
			want: "register bit index",
		},
		{
			name: "bit index at operand width",
			build: func(ctx *build.Context) {
				ctx.BTL(operand.U8(32), reg.ECX)
				ctx.JC(operand.LabelRef("target"))
			},
			want: "not below the 32-bit operand width",
		},
		{
			name: "memory operand",
			build: func(ctx *build.Context) {
				ctx.BTL(operand.U8(0), operand.Mem{Base: reg.RAX})
				ctx.JC(operand.LabelRef("target"))
			},
			want: "memory operand",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := build.NewContext()
			ctx.Function("bt")
			ctx.SignatureExpr("func()")
			tc.build(ctx)
			ctx.Label("target")
			ctx.Label("other")
			ctx.RET()

			var recovered interface{}
			out := func() string {
				defer func() { recovered = recover() }()
				return printARM64(t, ctx, printer.NewGoRunConfig())
			}()

			if tc.want == "" {
				if recovered != nil {
					t.Fatalf("expected this shape to lower, got panic: %v", recovered)
				}
				if !strings.Contains(out, "TBNZ $0x00, R1, target") {
					t.Errorf("expected a fused TBNZ:\n%s", out)
				}
				return
			}
			if recovered == nil {
				t.Fatalf("expected a panic containing %q, got output:\n%s", tc.want, out)
			}
			if msg := fmt.Sprint(recovered); !strings.Contains(msg, tc.want) {
				t.Errorf("panic %q does not mention %q", msg, tc.want)
			}
		})
	}
}

// TestARM64ZeroExtendWidthPairs checks that the L- and Q-destination forms of
// each zero-extending load lower identically.
//
// They differ only in the named width of the destination, not in the value
// produced: x86 zeroes bits 63:32 on any 32-bit register write, so MOVWLZX
// leaves the same fully zero-extended register MOVWQZX does, and the arm64
// MOVBU/MOVHU zero-extend across the whole register either way. Emitting a
// narrowing step for the L forms would be wrong, and refusing them outright
// would reject programs the Q forms already accept.
func TestARM64ZeroExtendWidthPairs(t *testing.T) {
	for _, tc := range []struct {
		name  string
		build func(ctx *build.Context)
		want  string
	}{
		{
			name:  "MOVWQZX",
			build: func(ctx *build.Context) { ctx.MOVWQZX(operand.Mem{Base: reg.RAX}, reg.RCX) },
			want:  "MOVHU (R0), R1",
		},
		{
			// The L forms take a 32-bit destination; the point of the pairing
			// is that this still yields the same fully zero-extended register.
			name:  "MOVWLZX",
			build: func(ctx *build.Context) { ctx.MOVWLZX(operand.Mem{Base: reg.RAX}, reg.ECX) },
			want:  "MOVHU (R0), R1",
		},
		{
			name:  "MOVBQZX",
			build: func(ctx *build.Context) { ctx.MOVBQZX(operand.Mem{Base: reg.RAX}, reg.RCX) },
			want:  "MOVBU (R0), R1",
		},
		{
			name:  "MOVBLZX",
			build: func(ctx *build.Context) { ctx.MOVBLZX(operand.Mem{Base: reg.RAX}, reg.ECX) },
			want:  "MOVBU (R0), R1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := build.NewContext()
			ctx.Function("zeroextend")
			ctx.SignatureExpr("func()")
			tc.build(ctx)
			ctx.RET()

			out := printARM64(t, ctx, printer.NewGoRunConfig())
			if !strings.Contains(out, tc.want) {
				t.Errorf("expected %q in output:\n%s", tc.want, out)
			}
			// A narrowing step after the load would mean the lowering thought
			// the upper half needed clearing, which it never does here.
			if strings.Contains(out, "MOVWU R1, R1") {
				t.Errorf("unexpected narrowing of an already zero-extended value:\n%s", out)
			}
		})
	}
}

// TestARM64PseudoSPLocalsOffset checks that avo's frame-relative locals land
// above the saved link register. x86 puts local 0 at SP+0 because CALL has
// already pushed the return address above the frame; arm64 spills the link
// register to the bottom of the frame, so locals start at 8(RSP).
func TestARM64PseudoSPLocalsOffset(t *testing.T) {
	ctx := build.NewContext()
	ctx.Function("locals")
	ctx.SignatureExpr("func()")
	ctx.MOVQ(reg.RAX, operand.Mem{Base: reg.StackPointer})          // avo local 0
	ctx.MOVQ(operand.Mem{Base: reg.StackPointer, Disp: 8}, reg.RCX) // avo local 8
	ctx.RET()

	out := printARM64(t, ctx, printer.NewGoRunConfig())
	if strings.Contains(out, ", (RSP)") || strings.Contains(out, "(RSP), ") && strings.Contains(out, " (RSP)") {
		t.Errorf("a local was emitted at 0(RSP), which is the saved link register:\n%s", out)
	}
	if !strings.Contains(out, "8(RSP)") || !strings.Contains(out, "16(RSP)") {
		t.Errorf("expected locals shifted to 8(RSP) and 16(RSP):\n%s", out)
	}
}

// TestARM64DirectivesRejected checks that preprocessor directives are refused
// rather than interpreted.
//
// This printer used to evaluate GOAMD64 conditionals and treat other directives
// as analysis boundaries. That machinery produced roughly a third of the
// miscompiles review found -- all of them external-coupling bugs, where its
// model of when a directive is live disagreed with the assembler, with the
// "\t// #" rewrite, or with a header it could not read -- while serving no
// generator that uses this lowering. The cases below are the ones that were
// silently wrong at various points; all of them are now simply refused.
func TestARM64DirectivesRejected(t *testing.T) {
	for _, lines := range [][]string{
		{"#ifdef GOAMD64_v3"},
		{"# ifdef GOAMD64_v3"},  // the assembler accepts a space after '#'
		{" #ifdef GOAMD64_v3"},  // inert after the rewrite, live to this printer
		{"#define GOAMD64_v3"},  // would define the symbol for the assembler
		{"#include \"defs.h\""}, // could define or open anything
		{"#endif"},
		{"#if 1"}, // not a directive Go's assembler has at all
	} {
		t.Run(strings.TrimSpace(lines[0]), func(t *testing.T) {
			ctx := build.NewContext()
			ctx.Function("d")
			ctx.SignatureExpr("func()")
			ctx.Comment(lines...)
			ctx.RET()

			f, errs := ctx.Result()
			if errs != nil {
				t.Fatal(errs)
			}
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("expected a panic for %q", lines)
				}
				if msg, ok := r.(string); !ok || !strings.Contains(msg, "twin functions") {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()
			_, _ = printer.NewARM64Asm(printer.NewGoRunConfig()).Print(f)
		})
	}
}

// TestARM64PlainCommentsPass checks that ordinary comments are untouched --
// only directives are refused.
func TestARM64PlainCommentsPass(t *testing.T) {
	ctx := build.NewContext()
	ctx.Function("c")
	ctx.SignatureExpr("func()")
	ctx.Comment("this is a normal comment", "spanning two lines")
	ctx.RET()

	out := printARM64(t, ctx, printer.NewGoRunConfig())
	if !strings.Contains(out, "// this is a normal comment") {
		t.Errorf("plain comment did not survive:\n%s", out)
	}
}

// TestARM64DirectiveInSkippedTwin checks the guard runs even for a function
// that is never lowered.
//
// Where a generator produced both a generic and a BMI2 variant, arm64 emits one
// and skips the other. The skipped one still ships on the amd64 side, so a
// directive hidden in it would go live there while arm64 ran the clean twin --
// and the arm64 output would look perfectly correct, which is exactly why
// differential testing cannot see this one.
func TestARM64DirectiveInSkippedTwin(t *testing.T) {
	for _, preferBMI2 := range []bool{false, true} {
		t.Run(fmt.Sprintf("PreferBMI2=%v", preferBMI2), func(t *testing.T) {
			ctx := build.NewContext()
			// The directive sits in the BMI2 twin; with the default config that
			// is the one skipped.
			ctx.Function("twin_amd64")
			ctx.SignatureExpr("func()")
			ctx.RET()
			ctx.Function("twin_bmi2")
			ctx.SignatureExpr("func()")
			ctx.Comment("#ifdef GOAMD64_v3")
			ctx.RET()

			f, errs := ctx.Result()
			if errs != nil {
				t.Fatal(errs)
			}
			cfg := printer.NewGoRunConfig()
			cfg.ARM64PreferBMI2 = preferBMI2
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("expected a panic for a directive in a twin, selected or not")
				}
				if msg, ok := r.(string); !ok || !strings.Contains(msg, "twin functions") {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()
			_, _ = printer.NewARM64Asm(cfg).Print(f)
		})
	}
}

// TestARM64CommentNewlineRejected covers comment text containing a newline.
//
// The emitter writes "\t// %s\n", so a comment line with an embedded newline
// produces a SECOND physical line carrying no comment prefix — live text at
// column 0, in both outputs. It does not have to look like a directive to do
// harm: an instruction there is assembled by both sides, and because the two
// architectures rename registers differently, the identical text reads and
// writes different values on each. It is also invisible to every analysis here,
// all of which assume a comment carries no code.
func TestARM64CommentNewlineRejected(t *testing.T) {
	for _, line := range []string{
		"benign note\nMOVB R11, (R12)", // injects a live instruction
		"benign text\n#ifdef GOAMD64_v1",
		"trailing\r\nMOVD $1, R0",
	} {
		t.Run(strings.SplitN(line, "\n", 2)[0], func(t *testing.T) {
			ctx := build.NewContext()
			ctx.Function("nl")
			ctx.SignatureExpr("func()")
			ctx.Comment(line)
			ctx.RET()

			f, errs := ctx.Result()
			if errs != nil {
				t.Fatal(errs)
			}
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("expected a panic for %q", line)
				}
				if msg, ok := r.(string); !ok || !strings.Contains(msg, "newline") {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()
			_, _ = printer.NewARM64Asm(printer.NewGoRunConfig()).Print(f)
		})
	}
}

// TestARM64BuildConstraintParenthesized checks the arm64 term is combined with
// the original constraint by AND over the WHOLE expression. Without brackets,
// "!noasm || purego" would become "arm64 && !noasm || purego", and since &&
// binds tighter that parses as "(arm64 && !noasm) || purego" — pulling the
// arm64 assembly into non-arm64 builds.
func TestARM64BuildConstraintParenthesized(t *testing.T) {
	ctx := build.NewContext()
	ctx.ConstraintExpr("!noasm purego")
	ctx.Function("c")
	ctx.SignatureExpr("func()")
	ctx.RET()

	out := printARM64(t, ctx, printer.NewGoRunConfig())
	if !strings.Contains(out, "//go:build arm64 && (") {
		t.Errorf("build constraint is not parenthesized:\n%s", out)
	}
}

// TestARM64TransparencyIsEnforced checks that a lowering claimed to leave the
// condition flags alone actually does.
//
// isFlagTransparent classifies by x86 opcode, but the property the flag scans
// rely on is that the EMITTED arm64 sequence is NZCV-neutral — including
// instructions a lowering emits incidentally, like the scratch-address ADD
// behind an indexed operand. Those two were kept in sync by hand, and the cost
// of that was a 64-bit TST synthesized where a 32-bit one was needed.
//
// The check lives in emit(); this test drives a producer/consumer pair through
// it so the enforcement itself is covered.
func TestARM64TransparencyIsEnforced(t *testing.T) {
	ctx := build.NewContext()
	ctx.Function("tr")
	ctx.SignatureExpr("func()")
	ctx.CMPQ(reg.RAX, reg.RCX)
	ctx.MOVQ(reg.RAX, reg.RDX)                                              // transparent, sits between producer and consumer
	ctx.MOVQ(operand.Mem{Base: reg.RSI, Index: reg.RDI, Scale: 4}, reg.RBX) // emits a scratch ADD
	ctx.JEQ(operand.LabelRef("tr_yes"))
	ctx.Label("tr_yes")
	ctx.RET()

	out := printARM64(t, ctx, printer.NewGoRunConfig())
	if !strings.Contains(out, "BEQ") {
		t.Errorf("expected the branch to survive:\n%s", out)
	}
}

// TestARM64UnknownMnemonicRejected checks the classification table cannot rot:
// emitting a mnemonic it does not list is a generation-time failure, so adding
// a lowering forces a conscious decision about whether it writes flags.
func TestARM64UnknownMnemonicRejected(t *testing.T) {
	if _, ok := printer.ARM64WritesNZCVForTest("MOVD"); !ok {
		t.Fatal("expected MOVD to be classified")
	}
	if _, ok := printer.ARM64WritesNZCVForTest("NOTAREALMNEMONIC"); ok {
		t.Error("an unlisted mnemonic must not be reported as classified")
	}
	// Every mnemonic the printer can emit should be listed; spot-check the
	// flag setters, which are the ones that matter for transparency.
	for _, m := range []string{"ADDS", "SUBSW", "ANDS", "CMP", "CMNW", "TSTW"} {
		writes, ok := printer.ARM64WritesNZCVForTest(m)
		if !ok || !writes {
			t.Errorf("%s should be classified as writing NZCV (listed=%v, writes=%v)", m, ok, writes)
		}
	}
}
