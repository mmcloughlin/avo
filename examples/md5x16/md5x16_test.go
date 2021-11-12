package md5x16

import (
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
		// Place the same data in every lane.
		var data [Lanes][]byte
		for l := range data {
			data[l] = []byte(c.Data)
		}

		if err := Validate(data); err != nil {
			t.Fatal(err)
		}

		// Hash and check.
		digest := Sum(data)

		for l := range digest {
			got := hex.EncodeToString(digest[l][:])
			if got != c.HexDigest {
				t.Errorf("Sum(%#v) lane %02d = %s; expect %s", c.Data, l, got, c.HexDigest)
			}
		}
	}
}
