package pass

import (
	"errors"

	"github.com/mmcloughlin/avo"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

// Liveness computes register liveness.
func Liveness(fn *avo.Function) error {
	// Note this implementation is initially naive so as to be "obviously correct".
	// There are a well-known optimizations we can apply if necessary.

	is := fn.Instructions()

	// Initialize to empty sets.
	for _, i := range is {
		i.LiveIn = reg.NewEmptySet()
		i.LiveOut = reg.NewEmptySet()
	}

	// Iterative dataflow analysis.
	for {
		changes := false

		for _, i := range is {
			// in[n] = use[n] UNION (out[n] - def[n])
			nin := len(i.LiveIn)
			i.LiveIn.Update(reg.NewSetFromSlice(i.InputRegisters()))
			def := reg.NewSetFromSlice(i.OutputRegisters())
			i.LiveIn.Update(i.LiveOut.Difference(def))
			if len(i.LiveIn) != nin {
				changes = true
			}

			// out[n] = UNION[s IN succ[n]] in[s]
			nout := len(i.LiveOut)
			for _, s := range i.Succ {
				if s == nil {
					continue
				}
				i.LiveOut.Update(s.LiveIn)
			}
			if len(i.LiveOut) != nout {
				changes = true
			}
		}

		if !changes {
			break
		}
	}

	return nil
}

func AllocateRegisters(fn *avo.Function) error {
	// Populate allocators (one per kind).
	as := map[reg.Kind]*Allocator{}
	for _, i := range fn.Instructions() {
		for _, r := range i.Registers() {
			k := r.Kind()
			if _, found := as[k]; !found {
				a, err := NewAllocatorForKind(k)
				if err != nil {
					return err
				}
				as[k] = a
			}
			as[k].Add(r)
		}
	}

	// Record register interferences.
	for _, i := range fn.Instructions() {
		for _, d := range i.OutputRegisters() {
			k := d.Kind()
			out := i.LiveOut.OfKind(k)
			out.Discard(d)
			as[k].AddInterferenceSet(d, out)
		}
	}

	// Execute register allocation.
	fn.Allocation = reg.NewEmptyAllocation()
	for _, a := range as {
		al, err := a.Allocate()
		if err != nil {
			return err
		}
		if err := fn.Allocation.Merge(al); err != nil {
			return err
		}
	}

	return nil
}

func BindRegisters(fn *avo.Function) error {
	for _, i := range fn.Instructions() {
		for idx := range i.Operands {
			i.Operands[idx] = operand.ApplyAllocation(i.Operands[idx], fn.Allocation)
		}
	}
	return nil
}

func VerifyAllocation(fn *avo.Function) error {
	// All registers should be physical.
	for _, i := range fn.Instructions() {
		for _, r := range i.Registers() {
			if reg.ToPhysical(r) == nil {
				return errors.New("non physical register found")
			}
		}
	}

	return nil
}
