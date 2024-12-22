package pass

import (
	"testing"

	"github.com/mmcloughlin/avo/reg"
)

func TestAllocatorSimple(t *testing.T) {
	c := reg.NewCollection()
	x, y := c.XMM(), c.YMM()

	a, err := NewAllocatorForKind(reg.KindVector)
	if err != nil {
		t.Fatal(err)
	}

	a.Add(x.ID())
	a.Add(y.ID())
	a.AddInterference(x.ID(), y.ID())

	alloc, err := a.Allocate()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(alloc)

	if alloc.LookupRegister(x) != reg.X0 || alloc.LookupRegister(y) != reg.Y1 {
		t.Fatalf("unexpected allocation")
	}
}

func TestAllocatorImpossible(t *testing.T) {
	a, err := NewAllocatorForKind(reg.KindVector)
	if err != nil {
		t.Fatal(err)
	}

	a.AddInterference(reg.X7.ID(), reg.Z7.ID())

	_, err = a.Allocate()
	if err == nil {
		t.Fatal("expected allocation error")
	}
}

func TestAllocatorPriority(t *testing.T) {
	const n = 4

	// Create an allocator with custom priorities.
	a, err := NewAllocatorForKind(reg.KindVector)
	if err != nil {
		t.Fatal(err)
	}

	a.SetPriority(reg.X0.ID(), -1)
	a.SetPriority(reg.X7.ID(), 1)
	a.SetPriority(reg.X13.ID(), 1)
	a.SetPriority(reg.X3.ID(), 2)

	// The expected n highest priority registers.
	expect := [n]reg.Physical{
		reg.X3,  // priority 2, id 3
		reg.X7,  // priority 1, id 7
		reg.X13, // priority 1, id 13
		reg.X1,  // priority 0, id 1 (X0 has priority -1)
	}

	// Setup allocation problem with n conflicting registers.
	c := reg.NewCollection()
	x := make([]reg.Virtual, n)
	for i := range x {
		x[i] = c.XMM()
	}

	for i := range x {
		a.Add(x[i].ID())
	}

	for i := range n {
		for j := i + 1; j < n; j++ {
			a.AddInterference(x[i].ID(), x[j].ID())
		}
	}

	// Allocate and confirm expectation.
	alloc, err := a.Allocate()
	if err != nil {
		t.Fatal(err)
	}

	for i := range x {
		if got := alloc.LookupRegister(x[i]); got != expect[i] {
			t.Errorf("x[%d] allocated %s; expected %s", i, got.Asm(), expect[i].Asm())
		}
	}
}
