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

func TestVirtualPhysicalHaveDifferentIDs(t *testing.T) {
	// Confirm that ID() returns different results even when virtual and physical IDs are the same.
	var v Virtual = virtual{id: 42}
	var p Physical = register{id: 42}
	if uint16(v.VirtualID()) != uint16(p.PhysicalID()) {
		t.Fatal("test assumption violated: VirtualID and PhysicalID should agree")
	}
	if v.ID() == p.ID() {
		t.Errorf("virtual and physical IDs should be different")
	}
}
