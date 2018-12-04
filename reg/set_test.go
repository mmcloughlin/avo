package reg

import "testing"

func TestFamilyRegisterSets(t *testing.T) {
	fs := []*Family{GeneralPurpose, SIMD}
	for _, f := range fs {
		if len(f.Set()) != len(f.Registers()) {
			t.Fatal("family set and list should have same size")
		}
	}
}
