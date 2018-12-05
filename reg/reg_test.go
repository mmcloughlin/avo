package reg

import "testing"

func TestSpecBytes(t *testing.T) {
	cases := []struct {
		Spec  Spec
		Bytes uint
	}{
		{S8L, 1},
		{S8H, 1},
		{S16, 2},
		{S32, 4},
		{S64, 8},
		{S128, 16},
		{S256, 32},
		{S512, 64},
	}
	for _, c := range cases {
		if c.Spec.Bytes() != c.Bytes {
			t.Errorf("%v.Bytes() = %d; expect = %d", c.Spec, c.Spec.Bytes(), c.Bytes)
		}
	}
}

func TestToVirtual(t *testing.T) {
	v := GeneralPurpose.Virtual(42, B32)
	if ToVirtual(v) != v {
		t.Errorf("ToVirtual(v) != v for virtual register")
	}
	if ToVirtual(ECX) != nil {
		t.Errorf("ToVirtual should be nil for physical registers")
	}
}

func TestToPhysical(t *testing.T) {
	v := GeneralPurpose.Virtual(42, B32)
	if ToPhysical(v) != nil {
		t.Errorf("ToPhysical should be nil for virtual registers")
	}
	if ToPhysical(ECX) != ECX {
		t.Errorf("ToPhysical(p) != p for physical register")
	}
}

func TestAreConflicting(t *testing.T) {
	cases := []struct {
		X, Y   Physical
		Expect bool
	}{
		{ECX, X3, false},
		{AL, AH, false},
		{AL, AX, true},
		{AL, BX, false},
		{X3, Y4, false},
		{X3, Y3, true},
		{Y3, Z4, false},
		{Y3, Z3, true},
	}
	for _, c := range cases {
		if AreConflicting(c.X, c.Y) != c.Expect {
			t.Errorf("AreConflicting(%s, %s) != %v", c.X, c.Y, c.Expect)
		}
	}
}
