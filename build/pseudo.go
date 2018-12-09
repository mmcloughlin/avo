package build

import (
	"github.com/mmcloughlin/avo/reg"

	"github.com/mmcloughlin/avo/gotypes"
)

//go:generate avogen -output zmov.go mov

func (c *Context) Load(src gotypes.Component, dst reg.Register) {
	b, err := src.Resolve()
	if err != nil {
		c.AddError(err)
		return
	}
	c.mov(b.Addr, dst, int(gotypes.Sizes.Sizeof(b.Type)), int(dst.Bytes()), b.Type)
}
