package reg

import "testing"

func TestAsMethods(t *testing.T) {
	cases := [][2]Register{
		{RAX.As8(), AL},
		{ECX.As8L(), CL},
		{EBX.As8H(), BH},
		{R9B.As16(), R9W},
		{DH.As32(), EDX},
		{R14L.As64(), R14},
		{X2.AsX(), X2},
		{X4.AsY(), Y4},
		{X9.AsZ(), Z9},
		{Y2.AsX(), X2},
		{Y4.AsY(), Y4},
		{Y9.AsZ(), Z9},
		{Z2.AsX(), X2},
		{Z4.AsY(), Y4},
		{Z9.AsZ(), Z9},
	}
	for _, c := range cases {
		if !Equal(c[0], c[1]) {
			t.FailNow()
		}
	}
}

func TestAsPreservesGPPhysical(t *testing.T) {
	cases := []Register{
		RAX.As8(),
		R13.As8L(),
		AL.As8H(),
		EAX.As16(),
		CH.As32(),
		EBX.As64(),
	}
	for _, r := range cases {
		if _, ok := r.(GPPhysical); !ok {
			t.FailNow()
		}
	}
}

func TestAsPreservesGPVirtual(t *testing.T) {
	collection := NewCollection()
	cases := []Register{
		collection.GP16().As8(),
		collection.GP32().As8L(),
		collection.GP64().As8H(),
		collection.GP8().As16(),
		collection.GP8L().As32(),
		collection.GP8H().As64(),
	}
	for _, r := range cases {
		if _, ok := r.(GPVirtual); !ok {
			t.FailNow()
		}
	}
}

func TestAsPreservesVecPhysical(t *testing.T) {
	cases := []Register{
		Y13.AsX(),
		X3.AsY(),
		Y10.AsZ(),
	}
	for _, r := range cases {
		if _, ok := r.(VecPhysical); !ok {
			t.FailNow()
		}
	}
}

func TestAsPreservesVecVirtual(t *testing.T) {
	collection := NewCollection()
	cases := []Register{
		collection.ZMM().AsX(),
		collection.XMM().AsY(),
		collection.YMM().AsZ(),
	}
	for _, r := range cases {
		if _, ok := r.(VecVirtual); !ok {
			t.FailNow()
		}
	}
}

func TestVecRegisters(t *testing.T) {
	cases := []struct {
		Spec   Spec
		N      int
		Index  Index
		Expect VecPhysical
	}{
		{S128, 16, 0, X0},
		{S128, 32, 31, X31},
		{S256, 16, 13, Y13},
		{S256, 32, 20, Y20},
		{S512, 32, 27, Z27},
		{S512, 8, 7, Z7},
	}
	for _, c := range cases {
		vs := VecRegisters(c.Spec, c.N)
		if len(vs) != c.N {
			t.Fatalf("spec=%v n=%d: got %d registers", c.Spec, c.N, len(vs))
		}
		if got := vs[c.Index]; got != c.Expect {
			t.Errorf("spec=%v idx=%v: got %v expect %v", c.Spec, c.Index, got, c.Expect)
		}
	}
}

func TestVecRegistersIndexes(t *testing.T) {
	for _, s := range []Spec{S128, S256, S512} {
		for i, r := range VecRegisters(s, 32) {
			if int(r.PhysicalIndex()) != i {
				t.Errorf("spec=%v: register %v at index %d", s, r, i)
			}
			if r.Mask() != s.Mask() {
				t.Errorf("spec=%v: register %v has mask %#x", s, r, r.Mask())
			}
		}
	}
}

func TestVecRegistersTooMany(t *testing.T) {
	for _, s := range []Spec{S128, S256, S512} {
		if vs := VecRegisters(s, 33); vs != nil {
			t.Errorf("spec=%v n=33: got %v expect nil", s, vs)
		}
	}
}

func TestVecRegistersNonVectorSpec(t *testing.T) {
	for _, s := range []Spec{S0, S8L, S8H, S16, S32, S64} {
		if vs := VecRegisters(s, 1); vs != nil {
			t.Errorf("spec=%v: got %v expect nil", s, vs)
		}
	}
}
