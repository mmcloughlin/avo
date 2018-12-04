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

func (c *Collection) GP8v() Virtual { return c.GPv(B8) }

func (c *Collection) GP16v() Virtual { return c.GPv(B16) }

func (c *Collection) GP32v() Virtual { return c.GPv(B32) }

func (c *Collection) GP64v() Virtual { return c.GPv(B64) }

func (c *Collection) GPv(s Size) Virtual { return c.VirtualRegister(GP, s) }

func (c *Collection) Xv() Virtual { return c.VirtualRegister(SSEAVX, B128) }

func (c *Collection) Yv() Virtual { return c.VirtualRegister(SSEAVX, B256) }

func (c *Collection) Zv() Virtual { return c.VirtualRegister(SSEAVX, B512) }
