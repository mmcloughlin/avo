package reg

// MaskSet maps register IDs to masks.
type MaskSet map[ID]uint16

// NewEmptyMaskSet builds an empty register mask set.
func NewEmptyMaskSet() MaskSet {
	return MaskSet{}
}

// NewMaskSetFromRegisters forms a mask set from the given register list.
func NewMaskSetFromRegisters(rs []Register) MaskSet {
	s := NewEmptyMaskSet()
	for _, r := range rs {
		s.AddRegister(r)
	}
	return s
}

// Clone returns a copy of s.
func (s MaskSet) Clone() MaskSet {
	c := NewEmptyMaskSet()
	for id, mask := range s {
		c.Add(id, mask)
	}
	return c
}

// Add mask to the given register ID.
func (s MaskSet) Add(id ID, mask uint16) {
	s[id] |= mask
}

// AddRegister is a convenience for adding the register's (ID, mask) to the set.
func (s MaskSet) AddRegister(r Register) {
	s.Add(r.ID(), r.Mask())
}

// Discard clears masked bits from register ID.
func (s MaskSet) Discard(id ID, mask uint16) {
	s[id] &^= mask
	if s[id] == 0 {
		delete(s, id)
	}
}

// DiscardRegister is a convenience for discarding the register's (ID, mask) from the set.
func (s MaskSet) DiscardRegister(r Register) {
	s.Discard(r.ID(), r.Mask())
}

// Update adds masks in t to s.
func (s MaskSet) Update(t MaskSet) {
	for id, mask := range t {
		s.Add(id, mask)
	}
}

// Difference returns the set of registers in s but not t.
func (s MaskSet) Difference(t MaskSet) MaskSet {
	d := s.Clone()
	d.DifferenceUpdate(t)
	return d
}

// DifferenceUpdate removes every element of t from s.
func (s MaskSet) DifferenceUpdate(t MaskSet) {
	for id, mask := range t {
		s.Discard(id, mask)
	}
}

// Equals returns true if s and t contain the same masks.
func (s MaskSet) Equals(t MaskSet) bool {
	if len(s) != len(t) {
		return false
	}
	for r := range s {
		if _, found := t[r]; !found {
			return false
		}
	}
	return true
}

// OfKind returns the set of elements of s with kind k.
func (s MaskSet) OfKind(k Kind) MaskSet {
	t := NewEmptyMaskSet()
	for id, mask := range s {
		if id.Kind() == k {
			t.Add(id, mask)
		}
	}
	return t
}
