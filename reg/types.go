package reg

import (
	"errors"
	"fmt"
)

// Kind is a class of registers.
type Kind uint8

// Index of a register within a kind.
type Index uint16

// Family is a collection of Physical registers of a common kind.
type Family struct {
	Kind      Kind
	registers []Physical
}

// define builds a register and adds it to the Family.
func (f *Family) define(s Spec, id Index, name string, flags ...Info) Physical {
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
func (f *Family) Virtual(id Index, s Spec) Virtual {
	return NewVirtual(id, f.Kind, s)
}

// Registers returns the registers in this family.
func (f *Family) Registers() []Physical {
	return append([]Physical(nil), f.registers...)
}

// Lookup returns the register with given physical ID and spec. Returns nil if no such register exists.
func (f *Family) Lookup(id Index, s Spec) Physical {
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
func newid(v uint8, kind Kind, idx Index) ID {
	return ID(v) | (ID(kind) << 8) | (ID(idx) << 16)
}

// IsVirtual reports whether this is an ID for a virtual register.
func (id ID) IsVirtual() bool { return (id & 1) == 1 }

// IsPhysical reports whether this is an ID for a physical register.
func (id ID) IsPhysical() bool { return !id.IsVirtual() }

// Kind extracts the kind from the register ID.
func (id ID) Kind() Kind { return Kind(id >> 8) }

// Index extracts the index from the register ID.
func (id ID) Index() Index { return Index(id >> 16) }

// Register represents a virtual or physical register.
type Register interface {
	ID() ID
	Kind() Kind
	Size() uint
	Mask() uint16
	Asm() string
	as(Spec) Register
	spec() Spec
	register()
}

// Virtual is a register of a given type and size, not yet allocated to a physical register.
type Virtual interface {
	VirtualID() Index
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
	id   Index
	kind Kind
	Spec
}

// NewVirtual builds a Virtual register.
func NewVirtual(id Index, k Kind, s Spec) Virtual {
	return virtual{
		id:   id,
		kind: k,
		Spec: s,
	}
}

func (v virtual) ID() ID           { return newid(1, v.kind, v.id) }
func (v virtual) VirtualID() Index { return v.id }
func (v virtual) Kind() Kind       { return v.kind }

func (v virtual) Asm() string {
	// TODO(mbm): decide on virtual register syntax
	return fmt.Sprintf("<virtual:%v:%v:%v>", v.id, v.Kind(), v.Size())
}

func (v virtual) as(s Spec) Register {
	return virtual{
		id:   v.id,
		kind: v.kind,
		Spec: s,
	}
}

func (v virtual) spec() Spec { return v.Spec }
func (v virtual) register()  {}

// Info is a bitmask of register properties.
type Info uint8

// Defined register Info flags.
const (
	None       Info = 0
	Restricted Info = 1 << iota
)

// Physical is a concrete register.
type Physical interface {
	PhysicalID() Index
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
	id     Index
	name   string
	info   Info
	Spec
}

func newregister(f *Family, s Spec, id Index, name string, flags ...Info) register {
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

func (r register) ID() ID            { return newid(0, r.Kind(), r.id) }
func (r register) PhysicalID() Index { return r.id }
func (r register) Kind() Kind        { return r.family.Kind }
func (r register) Asm() string       { return r.name }
func (r register) Info() Info        { return r.info }

func (r register) as(s Spec) Register {
	return r.family.Lookup(r.PhysicalID(), s)
}

func (r register) spec() Spec { return r.Spec }
func (r register) register()  {}

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

func LookupPhysical(k Kind, idx Index, s Spec) Physical {
	f := FamilyOfKind(k)
	if f == nil {
		return nil
	}
	return f.Lookup(idx, s)
}

func LookupID(id ID, s Spec) Physical {
	if id.IsVirtual() {
		return nil
	}
	return LookupPhysical(id.Kind(), id.Index(), s)
}

// Allocation records a register allocation.
type Allocation map[ID]ID

// NewEmptyAllocation builds an empty register allocation.
func NewEmptyAllocation() Allocation {
	return Allocation{}
}

// Merge allocations from b into a. Errors if there is disagreement on a common
// register.
func (a Allocation) Merge(b Allocation) error {
	for id, p := range b {
		if alt, found := a[id]; found && alt != p {
			return errors.New("disagreement on overlapping register")
		}
		a[id] = p
	}
	return nil
}

func (a Allocation) LookupDefault(id ID) ID {
	if _, found := a[id]; found {
		return a[id]
	}
	return id
}

// LookupRegister the allocation for register r, or return nil if there is none.
func (a Allocation) LookupRegister(r Register) Physical {
	// Return immediately if it is already a physical register.
	if p := ToPhysical(r); p != nil {
		return p
	}

	// Lookup an allocation for this virtual ID.
	id, found := a[r.ID()]
	if !found {
		return nil
	}

	return LookupID(id, r.spec())
}

// LookupDefault returns the register assigned to r, or r itself if there is none.
func (a Allocation) LookupRegisterDefault(r Register) Register {
	if r == nil {
		return nil
	}
	if p := a.LookupRegister(r); p != nil {
		return p
	}
	return r
}
