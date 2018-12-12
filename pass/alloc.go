package pass

import (
	"errors"
	"math"

	"github.com/mmcloughlin/avo/reg"
)

// edge is an edge of the interference graph, indicating that registers X and Y
// must be in non-conflicting registers.
type edge struct {
	X, Y reg.Register
}

type Allocator struct {
	registers  []reg.Physical
	allocation reg.Allocation
	edges      []*edge
	possible   map[reg.Virtual][]reg.Physical
}

func NewAllocator(rs []reg.Physical) (*Allocator, error) {
	if len(rs) == 0 {
		return nil, errors.New("no registers")
	}
	return &Allocator{
		registers:  rs,
		allocation: reg.NewEmptyAllocation(),
		possible:   map[reg.Virtual][]reg.Physical{},
	}, nil
}

func NewAllocatorForKind(k reg.Kind) (*Allocator, error) {
	f := reg.FamilyOfKind(k)
	if f == nil {
		return nil, errors.New("unknown register family")
	}
	return NewAllocator(f.Registers())
}

func (a *Allocator) AddInterferenceSet(r reg.Register, s reg.Set) {
	for y := range s {
		a.AddInterference(r, y)
	}
}

func (a *Allocator) AddInterference(x, y reg.Register) {
	a.Add(x)
	a.Add(y)
	a.edges = append(a.edges, &edge{X: x, Y: y})
}

// Add adds a register to be allocated. Does nothing if the register has already been added.
func (a *Allocator) Add(r reg.Register) {
	v, ok := r.(reg.Virtual)
	if !ok {
		return
	}
	if _, found := a.possible[v]; found {
		return
	}
	a.possible[v] = a.registersofsize(v.Bytes())
}

func (a *Allocator) Allocate() (reg.Allocation, error) {
	for a.remaining() > 0 {
		if err := a.update(); err != nil {
			return nil, err
		}

		v := a.mostrestricted()
		if err := a.alloc(v); err != nil {
			return nil, err
		}
	}
	return a.allocation, nil
}

// update possible allocations based on edges.
func (a *Allocator) update() error {
	var rem []*edge
	for _, e := range a.edges {
		e.X, e.Y = a.allocation.LookupDefault(e.X), a.allocation.LookupDefault(e.Y)

		px, py := reg.ToPhysical(e.X), reg.ToPhysical(e.Y)
		vx, vy := reg.ToVirtual(e.X), reg.ToVirtual(e.Y)

		switch {
		case vx != nil && vy != nil:
			rem = append(rem, e)
			continue
		case px != nil && py != nil:
			if reg.AreConflicting(px, py) {
				return errors.New("impossible register allocation")
			}
		case px != nil && vy != nil:
			a.discardconflicting(vy, px)
		case vx != nil && py != nil:
			a.discardconflicting(vx, py)
		default:
			panic("unreachable")
		}
	}
	a.edges = rem
	return nil
}

// mostrestricted returns the virtual register with the least possibilities.
func (a *Allocator) mostrestricted() reg.Virtual {
	n := int(math.MaxInt32)
	var v reg.Virtual
	for r, p := range a.possible {
		if len(p) < n {
			n = len(p)
			v = r
		}
	}
	return v
}

// discardconflicting removes registers from vs possible list that conflict with p.
func (a *Allocator) discardconflicting(v reg.Virtual, p reg.Physical) {
	var rs []reg.Physical
	for _, r := range a.possible[v] {
		if !reg.AreConflicting(r, p) {
			rs = append(rs, r)
		}
	}
	a.possible[v] = rs
}

// alloc attempts to allocate a register to v.
func (a *Allocator) alloc(v reg.Virtual) error {
	ps := a.possible[v]
	if len(ps) == 0 {
		return errors.New("failed to allocate registers")
	}
	a.allocation[v] = ps[0]
	delete(a.possible, v)
	return nil
}

// remaining returns the number of unallocated registers.
func (a *Allocator) remaining() int {
	return len(a.possible)
}

// registersofsize returns all registers of the given size.
func (a *Allocator) registersofsize(n uint) []reg.Physical {
	var rs []reg.Physical
	for _, r := range a.registers {
		if r.Bytes() == n {
			rs = append(rs, r)
		}
	}
	return rs
}
