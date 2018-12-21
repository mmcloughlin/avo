// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

func main() {
	TEXT("block", "func(h *[5]uint32, m []byte)")
	h := Mem{Base: Load(Param("h"), GP64v())}
	m := Mem{Base: Load(Param("m").Base(), GP64v())}

	// Store message values on the stack.
	w := AllocLocal(64)
	W := func(r int) Mem { return w.Idx((r % 16) * 4) }

	// Load initial hash.
	h0, h1, h2, h3, h4 := GP32v(), GP32v(), GP32v(), GP32v(), GP32v()

	MOVL(h.Idx(0), h0)
	MOVL(h.Idx(4), h1)
	MOVL(h.Idx(8), h2)
	MOVL(h.Idx(12), h3)
	MOVL(h.Idx(16), h4)

	// Initialize registers.
	a, b, c, d, e := GP32v(), GP32v(), GP32v(), GP32v(), GP32v()

	MOVL(h0, a)
	MOVL(h1, b)
	MOVL(h2, c)
	MOVL(h3, d)
	MOVL(h4, e)

	// Generate round updates.
	quarter := []struct {
		F func(reg.Register, reg.Register, reg.Register) reg.Register
		K uint32
	}{
		{choose, 0x5a827999},
		{xor, 0x6ed9eba1},
		{majority, 0x8f1bbcdc},
		{xor, 0xca62c1d6},
	}

	for r := 0; r < 80; r++ {
		q := quarter[r/20]

		// Load message value.
		u := GP32v()
		if r < 16 {
			MOVL(m.Idx(4*r), u)
			BSWAPL(u)
		} else {
			MOVL(W(r-3), u)
			XORL(W(r-8), u)
			XORL(W(r-14), u)
			XORL(W(r-16), u)
			ROLL(Imm(1), u)
		}
		MOVL(u, W(r))

		// Compute the next state register.
		t := GP32v()
		MOVL(a, t)
		ROLL(Imm(5), t)
		ADDL(q.F(b, c, d), t)
		ADDL(e, t)
		ADDL(Imm(q.K), t)
		ADDL(u, t)

		// Update registers.
		ROLL(Imm(30), b)
		a, b, c, d, e = t, a, b, c, d
	}

	// Final add.
	ADDL(a, h0)
	ADDL(b, h1)
	ADDL(c, h2)
	ADDL(d, h3)
	ADDL(e, h4)

	// Store results back.
	MOVL(h0, h.Idx(0))
	MOVL(h1, h.Idx(4))
	MOVL(h2, h.Idx(8))
	MOVL(h3, h.Idx(12))
	MOVL(h4, h.Idx(16))
	RET()

	Generate()
}

func choose(b, c, d reg.Register) reg.Register {
	r := GP32v()
	MOVL(d, r)
	XORL(c, r)
	ANDL(b, r)
	XORL(d, r)
	return r
}

func xor(b, c, d reg.Register) reg.Register {
	r := GP32v()
	MOVL(b, r)
	XORL(c, r)
	XORL(d, r)
	return r
}

func majority(b, c, d reg.Register) reg.Register {
	t, r := GP32v(), GP32v()
	MOVL(b, t)
	ORL(c, t)
	ANDL(d, t)
	MOVL(b, r)
	ANDL(c, r)
	ORL(t, r)
	return r
}
