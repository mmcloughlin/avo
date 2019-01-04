package build

import (
	"github.com/mmcloughlin/avo/internal/stack"
	"github.com/mmcloughlin/avo/src"
)

// Error represents an error during building, optionally tagged with the position at which it happened.
type Error struct {
	Position src.Position
	Err      error
}

// exterr constructs an Error with position derived from the first frame in the
// call stack outside this package.
func exterr(err error) Error {
	e := Error{Err: err}
	if f := stack.ExternalCaller(); f != nil {
		e.Position = src.FramePosition(*f).Relwd()
	}
	return e
}

func (e Error) Error() string {
	msg := e.Err.Error()
	if e.Position.IsValid() {
		return e.Position.String() + ": " + msg
	}
	return msg
}
