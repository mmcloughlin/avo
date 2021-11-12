package md5x16

import (
	"crypto/md5"
	"encoding/hex"
	"testing"

	"golang.org/x/sys/cpu"
)

func RequireISA(t *testing.T) {
	t.Helper()
	if !cpu.X86.HasAVX512F {
		t.Skip("requires AVX512F instruction set")
	}
}

func TestVectors(t *testing.T) {
	RequireISA(t)

	cases := []struct {
		Data      string
		HexDigest string
	}{
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"The quick brown fox jumps over the lazy dog", "9e107d9d372bb6826bd81d3542a419d6"},
		{"The quick brown fox jumps over the lazy dog.", "e4d909c290d0fb1ca068ffaddf22cbd0"},
	}
	for _, c := range cases {
		digest := Single(t, []byte(c.Data))
		got := hex.EncodeToString(digest[:])
		if got != c.HexDigest {
			t.Errorf("Sum(%#v) = %s; expect %s", c.Data, got, c.HexDigest)
		}
	}
}

func TestLengths(t *testing.T) {
	RequireISA(t)

	const max = BlockSize << 6
	data := make([]byte, max)
	for n := 0; n <= max; n++ {
		got := Single(t, data[:n])
		expect := md5.Sum(data[:n])
		if got != expect {
			t.Errorf("failed on length %d", n)
		}
	}
}

func Single(t *testing.T, d []byte) [Size]byte {
	// Place the same data in every lane.
	var data [Lanes][]byte
	for l := range data {
		data[l] = d
	}

	if err := Validate(data); err != nil {
		t.Fatal(err)
	}

	// Hash and check the lanes are the same.
	digest := Sum(data)
	for l := range data {
		if digest[0] != digest[l] {
			t.Logf("lane %02d: %x", 0, digest[0])
			t.Logf("lane %02d: %x", l, digest[l])
			t.Fatal("lane mismatch")
		}
	}

	return digest[0]
}
