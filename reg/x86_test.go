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
