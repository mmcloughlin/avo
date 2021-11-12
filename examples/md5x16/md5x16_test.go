package md5x16

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"testing"
	"testing/quick"

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

func TestCmp(t *testing.T) {
	RequireISA(t)

	sum := func(data []byte) [Size]byte { return Single(t, data) }
	if err := quick.CheckEqual(sum, md5.Sum, nil); err != nil {
		t.Fatal(err)
	}
}

func TestLengths(t *testing.T) {
	RequireISA(t)

	const max = BlockSize << 6
	data := make([]byte, max)
	rand.Read(data)

	for n := 0; n <= max; n++ {
		got := Single(t, data[:n])
		expect := md5.Sum(data[:n])
		if got != expect {
			t.Fatalf("failed on length %d", n)
		}
	}
}

// Single hashes a single data buffer in all 16 lanes and returns the result,
// after asserting that all lanes are the same.
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

func TestActiveLanes(t *testing.T) {
	const trials = 1 << 10
	const maxlen = BlockSize << 6
	for trial := 0; trial < trials; trial++ {
		// Pick active lanes.
		lanes := 1 + rand.Intn(Lanes-1)
		active := rand.Perm(Lanes)[:lanes]

		// Fill active lanes with random data.
		n := rand.Intn(maxlen)
		buffer := make([]byte, lanes*n)
		rand.Read(buffer)

		var data [Lanes][]byte
		for i, l := range active {
			data[l] = buffer[i*n : (i+1)*n]
		}

		// Hash.
		digest := Sum(data)

		// Verify correct result in active lanes.
		for _, l := range active {
			expect := md5.Sum(data[l])
			if digest[l] != expect {
				t.Fatalf("lane %02d: mismatch", l)
			}
		}

		// Verify other lanes are zero.
		isactive := map[int]bool{}
		for _, l := range active {
			isactive[l] = true
		}
		for l := 0; l < Lanes; l++ {
			if !isactive[l] {
				var zero [Size]byte
				if digest[l] != zero {
					t.Fatalf("inactive lane %d is non-zero", l)
				}
			}
		}
	}
}
