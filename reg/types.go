package reg

import (
	"fmt"
	"math"
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

func (f *Family) define(s Spec, id PID, name string) Physical {
	r := register{
		id:   id,
		kind: f.Kind,
		name: name,
		Spec: s,
	}
	f.registers = append(f.registers, r)
	return r
}

func (f *Family) Virtual(id VID, s Size) Virtual {
	return NewVirtual(id, f.Kind, s)
}

type private interface {
	private()
}

type (
	ID  uint32
	VID uint16
	PID uint16
)

type Register interface {
	ID() ID
	Kind() Kind
	Bytes() uint
	Asm() string
	private
}

type Virtual interface {
	VirtualID() VID
	Register
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
	return (ID(math.MaxUint16) << 16) | ID(v.VirtualID())
}

func (v virtual) Asm() string {
	// TODO(mbm): decide on virtual register syntax
	return fmt.Sprintf("<virtual:%v:%v:%v>", v.id, v.Kind(), v.Bytes())
}

func (v virtual) private() {}

type Physical interface {
	PhysicalID() PID
	Mask() uint16
	Register
}

type register struct {
	id   PID
	kind Kind
	name string
	Spec
}

func (r register) PhysicalID() PID { return r.id }
func (r register) ID() ID          { return ID(r.id) }
func (r register) Kind() Kind      { return r.kind }
func (r register) Asm() string     { return r.name }
func (r register) private()        {}

type Spec uint16

const (
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
