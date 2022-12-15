package issue193

import "testing"

//go:generate go run asm.go -out issue193.s -stubs stub.go

func TestAddSubs(t *testing.T) {
	in := [8]int64{7, 6, 5, 4, 3, 2, 1, 0}
	out := [8]int64{13, 1, 9, 1, 5, 1, 1, 1}

	val := in
	AddSubPairs(&val)
	if val != out {
		t.Errorf("Bad result %v", val)
	}

	val = in
	AddSubPairsNoBase(&val)
	if val != out {
		t.Errorf("Bad result %v", val)
	}
}
