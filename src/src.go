package src

import "strconv"

// Position represents a position in a source file.
type Position struct {
	Filename string
	Line     int // 1-up
}

// IsValid reports whether the position is valid: Line must be positive, but
// Filename may be empty.
func (p Position) IsValid() bool {
	return p.Line > 0
}

// String represents Position as a string.
func (p Position) String() string {
	if !p.IsValid() {
		return "-"
	}
	var s string
	if p.Filename != "" {
		s += p.Filename + ":"
	}
	s += strconv.Itoa(p.Line)
	return s
}
