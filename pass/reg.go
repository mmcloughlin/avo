package pass

import (
	"github.com/mmcloughlin/avo"
	"github.com/mmcloughlin/avo/reg"
)

// Liveness computes register liveness.
func Liveness(fn *avo.Function) error {
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
	return nil
}
