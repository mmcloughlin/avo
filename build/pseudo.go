package build

import (
	"github.com/mmcloughlin/avo/reg"

	"github.com/mmcloughlin/avo/gotypes"
)

//go:generate avogen -output zmov.go mov

func (c *Context) Param(name string) gotypes.Component {
	return c.activefunc().Signature.Params().Lookup(name)
}

func (c *Context) ParamIndex(i int) gotypes.Component {
	return c.activefunc().Signature.Params().At(i)
}

func (c *Context) Return(name string) gotypes.Component {
	return c.activefunc().Signature.Results().Lookup(name)
}

func (c *Context) ReturnIndex(i int) gotypes.Component {
	return c.activefunc().Signature.Results().At(i)
}

func (c *Context) Load(src gotypes.Component, dst reg.Register) reg.Register {
	b, err := src.Resolve()
	if err != nil {
		c.AddError(err)
		return dst
	}
	c.mov(b.Addr, dst, int(gotypes.Sizes.Sizeof(b.Type)), int(dst.Bytes()), b.Type)
	return dst
}

func (c *Context) Store(src reg.Register, dst gotypes.Component) {
	b, err := dst.Resolve()
	if err != nil {
		c.AddError(err)
		return
	}
	c.mov(src, b.Addr, int(src.Bytes()), int(gotypes.Sizes.Sizeof(b.Type)), b.Type)
}
