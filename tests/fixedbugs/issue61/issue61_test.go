package issue61

import (
	"testing"
)

//go:generate go run asm.go -out issue61.s

func TestPrivate(t *testing.T) {
	private()
}
