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
