package printer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
	"github.com/mmcloughlin/avo/x86"
)

// TestARM64PromotionKeepsSlotALUAcceptance checks that each x86 form of the
// slotALUOps on a frame slot lowers with the slot promoted exactly when it
// lowers with the slot in memory, so that whether a program generates never
// depends on which slots win a register.
func TestARM64PromotionKeepsSlotALUAcceptance(t *testing.T) {
	binary := map[string]func(a, b operand.Op) (*ir.Instruction, error){
		"ADDQ": x86.ADDQ, "SUBQ": x86.SUBQ, "ANDQ": x86.ANDQ, "ORQ": x86.ORQ,
		"XORQ": x86.XORQ, "CMPQ": x86.CMPQ, "TESTQ": x86.TESTQ,
	}
	unary := map[string]func(a operand.Op) (*ir.Instruction, error){
		"INCQ": x86.INCQ, "DECQ": x86.DECQ, "NEGQ": x86.NEGQ, "NOTQ": x86.NOTQ,
	}
	type form struct {
		name string
		mk   func(s operand.Mem) (*ir.Instruction, error)
	}
	var forms []form
	for op := range slotALUOps {
		if f, ok := binary[op]; ok {
			forms = append(forms,
				form{op + " reg, slot", func(s operand.Mem) (*ir.Instruction, error) { return f(reg.RAX, s) }},
				form{op + " imm, slot", func(s operand.Mem) (*ir.Instruction, error) { return f(operand.U32(3), s) }},
				form{op + " slot, reg", func(s operand.Mem) (*ir.Instruction, error) { return f(s, reg.RAX) }},
				form{op + " slot, imm", func(s operand.Mem) (*ir.Instruction, error) { return f(s, operand.U32(3)) }})
			continue
		}
		f, ok := unary[op]
		if !ok {
			t.Fatalf("slotALUOps has %s, which this test does not cover", op)
		}
		forms = append(forms, form{op + " slot", func(s operand.Mem) (*ir.Instruction, error) { return f(s) }})
	}
	lowers := func(fm form, promote bool) (ok, valid bool, why string) {
		fn := ir.NewFunction("accept")
		s := fn.AllocLocal(8)
		ins, err := fm.mk(s)
		if err != nil {
			return false, false, "" // not an x86 form
		}
		for _, i := range []*ir.Instruction{mustIns(x86.MOVQ(reg.RAX, s)), ins, mustIns(x86.MOVQ(s, reg.RCX)), mustIns(x86.RET())} {
			fn.AddInstruction(i)
		}
		f := &ir.File{}
		f.AddSection(fn)
		defer func() {
			if r := recover(); r != nil {
				ok, why = false, fmt.Sprint(r)
			}
		}()
		out, err := NewARM64Asm(Config{ARM64PromoteStackSlots: promote}).Print(f)
		if err != nil {
			return false, true, err.Error()
		}
		if promote && !strings.Contains(string(out), promoRegs[0]) {
			panic("slot not promoted")
		}
		return true, true, ""
	}
	for _, fm := range forms {
		on, valid, whyOn := lowers(fm, true)
		if !valid {
			continue
		}
		off, _, whyOff := lowers(fm, false)
		if on != off {
			t.Errorf("%s: lowers with the slot promoted = %v (%s), in memory = %v (%s)",
				fm.name, on, whyOn, off, whyOff)
		}
	}
}

func mustIns(i *ir.Instruction, err error) *ir.Instruction {
	if err != nil {
		panic(err)
	}
	return i
}
