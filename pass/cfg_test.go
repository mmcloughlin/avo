package pass

import (
	"reflect"
	"testing"

	"github.com/mmcloughlin/avo"
)

func TestLabelTarget(t *testing.T) {
	expect := map[avo.Label]*avo.Instruction{
		"lblA": &avo.Instruction{Opcode: "A"},
		"lblB": &avo.Instruction{Opcode: "B"},
	}

	f := avo.NewFunction("happypath")
	for lbl, i := range expect {
		f.AddLabel(lbl)
		f.AddInstruction(i)
		f.AddInstruction(&avo.Instruction{Opcode: "IDK"})
	}

	if err := LabelTarget(f); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expect, f.LabelTarget) {
		t.Fatalf("incorrect LabelTarget value\ngot=%#v\nexpext=%#v\n", f.LabelTarget, expect)
	}
}

func TestLabelTargetDuplicate(t *testing.T) {
	f := avo.NewFunction("dupelabel")
	f.AddLabel(avo.Label("lblA"))
	f.AddInstruction(&avo.Instruction{Opcode: "A"})
	f.AddLabel(avo.Label("lblA"))
	f.AddInstruction(&avo.Instruction{Opcode: "A"})

	err := LabelTarget(f)

	if err == nil || err.Error() != "duplicate label \"lblA\"" {
		t.Fatalf("expected error on duplcate label; got %v", err)
	}
}

func TestLabelTargetEndsWithLabel(t *testing.T) {
	f := avo.NewFunction("endswithlabel")
	f.AddInstruction(&avo.Instruction{Opcode: "A"})
	f.AddLabel(avo.Label("theend"))

	err := LabelTarget(f)

	if err == nil || err.Error() != "function ends with label" {
		t.Fatalf("expected error when function ends with label; got %v", err)
	}
}

func TestLabelTargetInstructionFollowLabel(t *testing.T) {
	f := avo.NewFunction("expectinstafterlabel")
	f.AddLabel(avo.Label("lblA"))
	f.AddLabel(avo.Label("lblB"))
	f.AddInstruction(&avo.Instruction{Opcode: "A"})

	err := LabelTarget(f)

	if err == nil || err.Error() != "instruction should follow a label" {
		t.Fatalf("expected error when label is not followed by instruction; got %v", err)
	}
}

func TestCFG(t *testing.T) {
	// TODO(mbm): jump backward
	// TODO(mbm): jump forward
	// TODO(mbm): multiple returns
	// TODO(mbm): infinite loop
	// TODO(mbm): very short infinite loop
}
