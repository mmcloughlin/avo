package reg

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
	registers []Register
}

func (f *Family) define(s Spec, id uint16, name string) Register {
	r := register{
		id:   id,
		kind: f.Kind,
		name: name,
		Spec: s,
	}
	f.registers = append(f.registers, r)
	return r
}

func (f *Family) Virtual(id uint16, s Size) Virtual {
	return virtual{
		id:   id,
		kind: f.Kind,
		Size: s,
	}
}

type private interface {
	private()
}

type Virtual interface {
	VirtualID() uint16
	Kind() Kind
	Bytes() uint
}

type virtual struct {
	id   uint16
	kind Kind
	Size
}

func (v virtual) VirtualID() uint16 { return v.id }
func (v virtual) Kind() Kind        { return v.kind }

type Register interface {
	PhysicalID() uint16
	Kind() Kind
	Mask() uint16
	Bytes() uint
	Asm() string
	private
}

type register struct {
	id   uint16
	kind Kind
	name string
	Spec
}

func (r register) PhysicalID() uint16 { return r.id }
func (r register) Kind() Kind         { return r.kind }
func (r register) Asm() string        { return r.name }
func (r register) private()           {}

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
