package reg

import (
	"errors"
	"fmt"
)

type Size uint

const (
	B8 Size = 1 << iota
	B16
	B32
	B64
	B128
	B256
	B512
)

func (s Size) Bytes() uint { return uint(s) }

type Kind uint8

type Family struct {
	Kind      Kind
	registers []Physical
}

func (f *Family) add(s Spec, id PID, name string, info Info) Physical {
	r := register{
		id:   id,
		kind: f.Kind,
		name: name,
		info: info,
		Spec: s,
	}
	f.registers = append(f.registers, r)
	return r
}

func (f *Family) define(s Spec, id PID, name string) Physical {
	return f.add(s, id, name, None)
}

func (f *Family) restricted(s Spec, id PID, name string) Physical {
	return f.add(s, id, name, Restricted)
}

func (f *Family) Virtual(id VID, s Size) Virtual {
	return NewVirtual(id, f.Kind, s)
}

// Registers returns the registers in this family.
func (f *Family) Registers() []Physical {
	return append([]Physical(nil), f.registers...)
}

// Set returns the set of registers in the family.
func (f *Family) Set() Set {
	s := NewEmptySet()
	for _, r := range f.registers {
		s.Add(r)
	}
	return s
}

type (
	ID  uint64
	VID uint16
	PID uint16
)

type Register interface {
	ID() ID
	Kind() Kind
	Bytes() uint
	Asm() string
	register()
}

type Virtual interface {
	VirtualID() VID
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
	Size
}

func NewVirtual(id VID, k Kind, s Size) Virtual {
	return virtual{
		id:   id,
		kind: k,
		Size: s,
	}
}

func (v virtual) VirtualID() VID { return v.id }
func (v virtual) Kind() Kind     { return v.kind }

func (v virtual) ID() ID {
	return (ID(1) << 63) | (ID(v.Size) << 24) | (ID(v.kind) << 16) | ID(v.VirtualID())
}

func (v virtual) Asm() string {
	// TODO(mbm): decide on virtual register syntax
	return fmt.Sprintf("<virtual:%v:%v:%v>", v.id, v.Kind(), v.Bytes())
}

func (v virtual) register() {}

type Info uint8

const (
	None       Info = 0
	Restricted Info = 1 << iota
)

type Physical interface {
	PhysicalID() PID
	Mask() uint16
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

type register struct {
	id   PID
	kind Kind
	name string
	info Info
	Spec
}

func (r register) PhysicalID() PID { return r.id }
func (r register) ID() ID          { return (ID(r.Mask()) << 16) | ID(r.id) }
func (r register) Kind() Kind      { return r.kind }
func (r register) Asm() string     { return r.name }
func (r register) Info() Info      { return r.info }
func (r register) register()       {}

type Spec uint16

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

// Bytes returns the register size in bytes.
func (s Spec) Bytes() uint {
	x := uint(s)
	return (x >> 1) + (x & 1)
}

// AreConflicting returns whether registers conflict with each other.
func AreConflicting(x, y Physical) bool {
	return x.Kind() == y.Kind() && x.PhysicalID() == y.PhysicalID() && (x.Mask()&y.Mask()) != 0
}

// Allocation records a register allocation.
type Allocation map[Register]Physical

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

func (a Allocation) LookupDefault(r Register) Register {
	if p, found := a[r]; found {
		return p
	}
	return r
}
