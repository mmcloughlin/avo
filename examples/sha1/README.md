# sha1

[SHA-1](https://en.wikipedia.org/wiki/SHA-1) in `avo`.

Compare to the [`crypto/sha1`](https://github.com/golang/go/blob/204a8f55dc2e0ac8d27a781dab0da609b98560da/src/crypto/sha1/sha1block_amd64.s#L16-L225) assembly.

[embedmd]:# (asm.go /func main/ /^}/)
```go
func main() {
	TEXT("block", "func(h *[5]uint32, m []byte)")
	Doc("block SHA-1 hashes the 64-byte message m into the running state h.")
	h := Mem{Base: Load(Param("h"), GP64())}
	m := Mem{Base: Load(Param("m").Base(), GP64())}

	// Store message values on the stack.
	w := AllocLocal(64)
	W := func(r int) Mem { return w.Offset((r % 16) * 4) }

	// Load initial hash.
	h0, h1, h2, h3, h4 := GP32(), GP32(), GP32(), GP32(), GP32()

	MOVL(h.Offset(0), h0)
	MOVL(h.Offset(4), h1)
	MOVL(h.Offset(8), h2)
	MOVL(h.Offset(12), h3)
	MOVL(h.Offset(16), h4)

	// Initialize registers.
	a, b, c, d, e := GP32(), GP32(), GP32(), GP32(), GP32()

	MOVL(h0, a)
	MOVL(h1, b)
	MOVL(h2, c)
	MOVL(h3, d)
	MOVL(h4, e)

	// Generate round updates.
	quarter := []struct {
		F func(Register, Register, Register) Register
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
		u := GP32()
		if r < 16 {
			MOVL(m.Offset(4*r), u)
			BSWAPL(u)
		} else {
			MOVL(W(r-3), u)
			XORL(W(r-8), u)
			XORL(W(r-14), u)
			XORL(W(r-16), u)
			ROLL(U8(1), u)
		}
		MOVL(u, W(r))

		// Compute the next state register.
		t := GP32()
		MOVL(a, t)
		ROLL(U8(5), t)
		ADDL(q.F(b, c, d), t)
		ADDL(e, t)
		ADDL(U32(q.K), t)
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
	MOVL(h0, h.Offset(0))
	MOVL(h1, h.Offset(4))
	MOVL(h2, h.Offset(8))
	MOVL(h3, h.Offset(12))
	MOVL(h4, h.Offset(16))
	RET()

	Generate()
}
```
