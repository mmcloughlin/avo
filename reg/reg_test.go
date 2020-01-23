package reg

import (
	"testing"
	"testing/quick"
)

func TestIDFields(t *testing.T) {
	f := func(v uint8, kind Kind, idx Index) bool {
		id := newid(v, kind, idx)
		return id.Kind() == kind && id.Index() == idx
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestIDIsVirtual(t *testing.T) {
	cases := []Virtual{
		GeneralPurpose.Virtual(42, S64),
		Vector.Virtual(42, S128),
	}
	for _, r := range cases {
		if !r.ID().IsVirtual() {
			t.FailNow()
		}
	}
}

func TestIDIsPhysical(t *testing.T) {
	cases := []Physical{AL, AH, AX, EAX, RAX, X1, Y2, Z31}
	for _, r := range cases {
		if !r.ID().IsPhysical() {
			t.FailNow()
		}
	}
}

func TestSpecSize(t *testing.T) {
	cases := []struct {
		Spec Spec
		Size uint
	}{
		{S0, 0},
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
		if c.Spec.Size() != c.Size {
			t.Errorf("%v.Size() = %d; expect = %d", c.Spec, c.Spec.Size(), c.Size)
		}
	}
}

func TestToVirtual(t *testing.T) {
	v := GeneralPurpose.Virtual(42, S32)
	if ToVirtual(v) != v {
		t.Errorf("ToVirtual(v) != v for virtual register")
	}
	if ToVirtual(ECX) != nil {
		t.Errorf("ToVirtual should be nil for physical registers")
	}
}

func TestToPhysical(t *testing.T) {
	v := GeneralPurpose.Virtual(42, S32)
	if ToPhysical(v) != nil {
		t.Errorf("ToPhysical should be nil for virtual registers")
	}
	if ToPhysical(ECX) != ECX {
		t.Errorf("ToPhysical(p) != p for physical register")
	}
}

func TestFamilyLookup(t *testing.T) {
	cases := []struct {
		Family *Family
		ID     Index
		Spec   Spec
		Expect Physical
	}{
		{GeneralPurpose, 0, S8, AL},
		{GeneralPurpose, 1, S8L, CL},
		{GeneralPurpose, 2, S8H, DH},
		{GeneralPurpose, 3, S16, BX},
		{GeneralPurpose, 9, S32, R9L},
		{GeneralPurpose, 13, S64, R13},
		{GeneralPurpose, 13, S512, nil},
		{GeneralPurpose, 133, S64, nil},
		{Vector, 1, S128, X1},
		{Vector, 13, S256, Y13},
		{Vector, 27, S512, Z27},
		{Vector, 1, S16, nil},
		{Vector, 299, S256, nil},
	}
	for _, c := range cases {
		got := c.Family.Lookup(c.ID, c.Spec)
		if got != c.Expect {
			t.Errorf("idx=%v spec=%v: lookup got %v expect %v", c.ID, c.Spec, got, c.Expect)
		}
	}
}

func TestPhysicalAs(t *testing.T) {
	cases := []struct {
		Register Physical
		Spec     Spec
		Expect   Physical
	}{
		{DX, S8L, DL},
		{DX, S8H, DH},
		{DX, S8, DL},
		{DX, S16, DX},
		{DX, S32, EDX},
		{DX, S64, RDX},
		{DX, S256, nil},
	}
	for _, c := range cases {
		got := c.Register.as(c.Spec)
		if got != c.Expect {
			t.Errorf("%s.as(%v) = %v; expect %v", c.Register.Asm(), c.Spec, got, c.Expect)
		}
	}
}

func TestVirtualAs(t *testing.T) {
	v := GeneralPurpose.Virtual(0, S64)
	specs := []Spec{S8, S8L, S8H, S16, S32, S64}
	for _, s := range specs {
		if v.as(s).Mask() != s.Mask() {
			t.FailNow()
		}
	}
}

func TestLookupPhysical(t *testing.T) {
	cases := []struct {
		Kind   Kind
		Index  Index
		Spec   Spec
		Expect Physical
	}{
		{KindGP, 0, S8L, AL},
		{KindGP, 1, S8H, CH},
		{KindGP, 7, S8, DIB},
		{KindGP, 8, S16, R8W},
		{KindGP, 9, S32, R9L},
		{KindGP, 10, S64, R10},

		{KindVector, 7, S128, X7},
		{KindVector, 17, S256, Y17},
		{KindVector, 27, S512, Z27},
	}
	for _, c := range cases {
		if got := LookupPhysical(c.Kind, c.Index, c.Spec); !Equal(got, c.Expect) {
			t.FailNow()
		}
	}
}

func TestLookupIDSelf(t *testing.T) {
	cases := []Physical{AL, AH, AX, EAX, RAX, X1, Y2, Z31}
	for _, r := range cases {
		if got := LookupID(r.ID(), r.spec()); !Equal(got, r) {
			t.FailNow()
		}
	}
}
