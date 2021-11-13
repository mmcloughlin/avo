# md5x16

AVX-512 accelerated 16-lane [MD5](https://en.wikipedia.org/wiki/MD5) in `avo`.

Inspired by [`minio/md5-simd`](https://github.com/minio/md5-simd) and
[`igneous-systems/md5vec`](https://github.com/igneous-systems/md5vec).

Note that the focus of this example is the core assembly `block` function. The
`Sum` function can only handle parallel hashes of exactly the same length. In
practice you'd likely need hash server functionality provided by
[`md5-simd`](https://github.com/minio/md5-simd) to multiplex independent hashes
of different lengths into the 16 SIMD lanes.

[embedmd]:# (asm.go /func main/ /^}/)
```go
func main() {
	// Define round constants data section.
	//
	// These may be computed as the integer part of abs(sin(i+1))*2^32.
	T := GLOBL("consts", RODATA|NOPTR)
	for i := 0; i < 64; i++ {
		k := uint32(math.Floor(math.Ldexp(math.Abs(math.Sin(float64(i+1))), 32)))
		DATA(4*i, U32(k))
	}

	// MD5 16-lane block function.
	TEXT("block", 0, "func(h *[4][16]uint32, base uintptr, offsets *[16]uint32, mask uint16)")
	Doc(
		"block MD5 hashes 16 messages into the running hash states h. Messages are",
		"at the given offsets from the base pointer. The 16-bit mask specifies",
		"which lanes are active: when bit i is not set loads will be disabled and",
		"the value of the resulting hash is undefined.",
	)
	h := Mem{Base: Load(Param("h"), GP64())}
	base := Mem{Base: Load(Param("base"), GP64())}
	offsetsptr := Mem{Base: Load(Param("offsets"), GP64())}
	mask := Load(Param("mask"), K())

	Comment("Load offsets.")
	offsets := ZMM()
	VMOVUPD(offsetsptr, offsets)

	Comment("Load initial hash.")
	hash := [4]Register{ZMM(), ZMM(), ZMM(), ZMM()}
	for i, r := range hash {
		VMOVUPD(h.Offset(64*i), r)
	}

	Comment("Initialize registers.")
	a, b, c, d := ZMM(), ZMM(), ZMM(), ZMM()
	for i, r := range []Register{a, b, c, d} {
		VMOVUPD(hash[i], r)
	}

	// Allocate message registers.
	m := make([]Register, 16)
	for i := range m {
		m[i] = ZMM()
	}

	// Generate round updates.
	//
	// Each 16-round block is parameterized based on the btiwise function,
	// message indexes and shift amounts. Constants B, C, D are helpers in
	// computing the logic table required by VPTERNLOGD.
	const (
		B = uint8(0b10101010)
		C = uint8(0b11001100)
		D = uint8(0b11110000)
	)
	quarter := []struct {
		F uint8         // ternary logic table
		i func(int) int // message index at round r
		s []int         // shift amounts
	}{
		{
			F: (B & C) | (^B & D),
			i: func(r int) int { return r % 16 },
			s: []int{7, 12, 17, 22},
		},
		{
			F: (D & B) | (^D & C),
			i: func(r int) int { return (5*r + 1) % 16 },
			s: []int{5, 9, 14, 20},
		},
		{
			F: B ^ C ^ D,
			i: func(r int) int { return (3*r + 5) % 16 },
			s: []int{4, 11, 16, 23},
		},
		{
			F: C ^ (B | ^D),
			i: func(r int) int { return (7 * r) % 16 },
			s: []int{6, 10, 15, 21},
		},
	}

	for r := 0; r < 64; r++ {
		Commentf("Round %d.", r)
		q := quarter[r/16]

		// Load message words.
		if r < 16 {
			k := K()
			KMOVW(mask, k)
			VPGATHERDD(base.Offset(4*r).Idx(offsets, 1), k, m[r])
		}

		VPADDD(m[q.i(r)], a, a)
		VPADDD_BCST(T.Offset(4*r), a, a)
		f := ZMM()
		VMOVUPD(d, f)
		VPTERNLOGD(U8(q.F), b, c, f)
		VPADDD(f, a, a)
		VPROLD(U8(q.s[r%4]), a, a)
		VPADDD(b, a, a)
		a, b, c, d = d, a, b, c
	}

	Comment("Final add.")
	for i, r := range []Register{a, b, c, d} {
		VPADDD(r, hash[i], hash[i])
	}

	Comment("Store results back.")
	for i, r := range hash {
		VMOVUPD(r, h.Offset(64*i))
	}

	VZEROUPPER()
	RET()

	Generate()
}
```
