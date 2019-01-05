package reg

type Collection struct {
	vid map[Kind]VID
}

func NewCollection() *Collection {
	return &Collection{
		vid: map[Kind]VID{},
	}
}

func (c *Collection) VirtualRegister(k Kind, s Size) Virtual {
	vid := c.vid[k]
	c.vid[k]++
	return NewVirtual(vid, k, s)
}

func (c *Collection) GP8() GPVirtual { return c.GPv(B8) }

func (c *Collection) GP16() GPVirtual { return c.GPv(B16) }

func (c *Collection) GP32() GPVirtual { return c.GPv(B32) }

func (c *Collection) GP64() GPVirtual { return c.GPv(B64) }

func (c *Collection) GPv(s Size) GPVirtual { return newgpv(c.VirtualRegister(KindGP, s)) }

func (c *Collection) XMM() VecVirtual { return c.Vecv(B128) }

func (c *Collection) YMM() VecVirtual { return c.Vecv(B256) }

func (c *Collection) ZMM() VecVirtual { return c.Vecv(B512) }

func (c *Collection) Vecv(s Size) VecVirtual { return newvecv(c.VirtualRegister(KindVector, s)) }
