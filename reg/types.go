package reg

import (
	"errors"
	"fmt"
)

// Kind is a class of registers.
type Kind uint8

// Family is a collection of Physical registers of a common kind.
type Family struct {
	Kind      Kind
	registers []Physical
}

// define builds a register and adds it to the Family.
func (f *Family) define(s Spec, id PID, name string, flags ...Info) Physical {
	r := newregister(f, s, id, name, flags...)
	f.add(r)
	return r
}

// add r to the family.
func (f *Family) add(r Physical) {
	if r.Kind() != f.Kind {
		panic("bad kind")
	}
	f.registers = append(f.registers, r)
}

// Virtual returns a virtual register from this family's kind.
func (f *Family) Virtual(id VID, s Spec) Virtual {
	return NewVirtual(id, f.Kind, s)
}

// Registers returns the registers in this family.
func (f *Family) Registers() []Physical {
	return append([]Physical(nil), f.registers...)
}

// Lookup returns the register with given physical ID and spec. Returns nil if no such register exists.
func (f *Family) Lookup(id PID, s Spec) Physical {
	for _, r := range f.registers {
		if r.PhysicalID() == id && r.Mask() == s.Mask() {
			return r
		}
	}
	return nil
}

// ID is a register identifier.
type ID uint32

// newid builds a new register ID from the virtual flag v, kind and index.
func newid(v uint8, kind Kind, idx uint16) ID {
	return ID(v) | (ID(kind) << 8) | (ID(idx) << 16)
}

// IsVirtual reports whether this is an ID for a virtual register.
func (id ID) IsVirtual() bool { return (id & 1) == 1 }

// IsPhysical reports whether this is an ID for a physical register.
func (id ID) IsPhysical() bool { return !id.IsVirtual() }

// Kind extracts the kind from the register ID.
func (id ID) Kind() Kind { return Kind(id >> 8) }

// Register represents a virtual or physical register.
type Register interface {
	ID() ID
	Kind() Kind
	Size() uint
	Mask() uint16
	Asm() string
	as(Spec) Register
	register()
}

// VID is a virtual register ID.
type VID uint16

// Virtual is a register of a given type and size, not yet allocated to a physical register.
type Virtual interface {
	VirtualID() VID
	SatisfiedBy(Physical) bool
	Register
}

// ToVirtual converts r to Virtual if possible, otherwise returns nil.
func ToVirtual(r Register) Virtual {
	if v, ok := r.(Virtual); ok {
		return v
	}
	return nil
}

type virtual struct {
	id   VID
	kind Kind
	Spec
}

// NewVirtual builds a Virtual register.
func NewVirtual(id VID, k Kind, s Spec) Virtual {
	return virtual{
		id:   id,
		kind: k,
		Spec: s,
	}
}

func (v virtual) ID() ID         { return newid(1, v.kind, uint16(v.id)) }
func (v virtual) VirtualID() VID { return v.id }
func (v virtual) Kind() Kind     { return v.kind }

func (v virtual) Asm() string {
	// TODO(mbm): decide on virtual register syntax
	return fmt.Sprintf("<virtual:%v:%v:%v>", v.id, v.Kind(), v.Size())
}

func (v virtual) SatisfiedBy(p Physical) bool {
	return v.Kind() == p.Kind() && v.Mask() == p.Mask()
}

func (v virtual) as(s Spec) Register {
	return virtual{
		id:   v.id,
		kind: v.kind,
		Spec: s,
	}
}

func (v virtual) register() {}

// Info is a bitmask of register properties.
type Info uint8

// Defined register Info flags.
const (
	None       Info = 0
	Restricted Info = 1 << iota
)

// PID is a physical register ID.
type PID uint16

// Physical is a concrete register.
type Physical interface {
	PhysicalID() PID
	Info() Info
	Register
}

// ToPhysical converts r to Physical if possible, otherwise returns nil.
func ToPhysical(r Register) Physical {
	if p, ok := r.(Physical); ok {
		return p
	}
	return nil
}

// register implements Physical.
type register struct {
	family *Family
	id     PID
	name   string
	info   Info
	Spec
}

func newregister(f *Family, s Spec, id PID, name string, flags ...Info) register {
	r := register{
		family: f,
		id:     id,
		name:   name,
		info:   None,
		Spec:   s,
	}
	for _, flag := range flags {
		r.info |= flag
	}
	return r
}

func (r register) ID() ID          { return newid(0, r.Kind(), uint16(r.id)) }
func (r register) PhysicalID() PID { return r.id }
func (r register) Kind() Kind      { return r.family.Kind }
func (r register) Asm() string     { return r.name }
func (r register) Info() Info      { return r.info }

func (r register) as(s Spec) Register {
	return r.family.Lookup(r.PhysicalID(), s)
}

func (r register) register() {}

// Spec defines the size of a register as well as the bit ranges it occupies in
// an underlying physical register.
type Spec uint16

// Spec values required for x86-64.
const (
	S0   Spec = 0x0 // zero value reserved for pseudo registers
	S8L  Spec = 0x1
	S8H  Spec = 0x2
	S8        = S8L
	S16  Spec = 0x3
	S32  Spec = 0x7
	S64  Spec = 0xf
	S128 Spec = 0x1f
	S256 Spec = 0x3f
	S512 Spec = 0x7f
)

// Mask returns a mask representing which bytes of an underlying register are
// used by this register. This is almost always the low bytes, except for the
// case of the high-byte registers. If bit n of the mask is set, this means
// bytes 2^(n-1) to 2^n-1 are used.
func (s Spec) Mask() uint16 {
	return uint16(s)
}

// Size returns the register width in bytes.
func (s Spec) Size() uint {
	x := uint(s)
	return (x >> 1) + (x & 1)
}

// AreConflicting returns whether registers conflict with each other.
func AreConflicting(x, y Physical) bool {
	return x.Kind() == y.Kind() && x.PhysicalID() == y.PhysicalID() && (x.Mask()&y.Mask()) != 0
}

// Allocation records a register allocation.
type Allocation map[Register]Physical

// NewEmptyAllocation builds an empty register allocation.
func NewEmptyAllocation() Allocation {
	return Allocation{}
}

// Merge allocations from b into a. Errors if there is disagreement on a common
// register.
func (a Allocation) Merge(b Allocation) error {
	for r, p := range b {
		if alt, found := a[r]; found && alt != p {
			return errors.New("disagreement on overlapping register")
		}
		a[r] = p
	}
	return nil
}

// LookupDefault returns the register assigned to r, or r itself if there is none.
func (a Allocation) LookupDefault(r Register) Register {
	if p, found := a[r]; found {
		return p
	}
	return r
}
