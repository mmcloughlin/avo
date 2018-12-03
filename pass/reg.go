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
		i.LiveIn = map[reg.ID]bool{}
		i.LiveOut = map[reg.ID]bool{}
	}

	// Iterative dataflow analysis.
	for {
		changes := false

		for _, i := range is {
			// in[n] = use[n] UNION (out[n] - def[n])
			nin := len(i.LiveIn)
			for _, r := range i.InputRegisters() {
				i.LiveIn[r.ID()] = true
			}
			def := map[reg.ID]bool{}
			for _, r := range i.OutputRegisters() {
				def[r.ID()] = true
			}
			for id := range i.LiveOut {
				if !def[id] {
					i.LiveIn[id] = true
				}
			}
			if len(i.LiveIn) != nin {
				changes = true
			}

			// out[n] = UNION[s IN succ[n]] in[s]
			nout := len(i.LiveOut)
			for _, s := range i.Succ {
				if s == nil {
					continue
				}
				for id := range s.LiveIn {
					i.LiveOut[id] = true
				}
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
