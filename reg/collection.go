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

func (c *Collection) GP8v() GPVirtual { return c.GPv(B8) }

func (c *Collection) GP16v() GPVirtual { return c.GPv(B16) }

func (c *Collection) GP32v() GPVirtual { return c.GPv(B32) }

func (c *Collection) GP64v() GPVirtual { return c.GPv(B64) }

func (c *Collection) GPv(s Size) GPVirtual { return newgpv(c.VirtualRegister(KindGP, s)) }

func (c *Collection) Xv() VecVirtual { return c.Vecv(B128) }

func (c *Collection) Yv() VecVirtual { return c.Vecv(B256) }

func (c *Collection) Zv() VecVirtual { return c.Vecv(B512) }

func (c *Collection) Vecv(s Size) VecVirtual { return newvecv(c.VirtualRegister(KindVector, s)) }
