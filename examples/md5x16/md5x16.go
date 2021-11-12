package md5x16

import (
	"encoding/binary"
	"errors"
	"math"
	"unsafe"
)

//go:generate go run asm.go -out md5x16.s -stubs stub.go

// Size of a MD5 checksum in bytes.
const Size = 16

// BlockSize is the block size of MD5 in bytes.
const BlockSize = 64

// Lanes is the maximum number of parallel MD5 computations.
const Lanes = 16

// Validate checks whether the preconditions required by Sum() are met.
func Validate(data [Lanes][]byte) error {
	_, err := config(data)
	return err
}

// Sum returns the MD5 checksum of up to Lanes data of the same length.
//
// Non-nil inputs must all have the same length, and occupy a memory span not
// exceeding 32 bits.
func Sum(data [Lanes][]byte) [Lanes][Size]byte {
	// Determine lane configuration.
	cfg, err := config(data)
	if err != nil {
		panic(err)
	}

	// Initialize hash.
	var h [4 * Lanes]uint32
	for _, l := range cfg.active {
		h[l] = 0x67452301
		h[l+16] = 0xefcdab89
		h[l+32] = 0x98badcfe
		h[l+48] = 0x10325476
	}

	// Consume full blocks.
	n := cfg.n
	for n >= BlockSize {
		block(&h, cfg.base, &cfg.offsets, cfg.mask)
		n -= BlockSize
	}

	// Final block.
	var last [Lanes][]byte
	var buffer [Lanes * BlockSize]byte
	base := uintptr(unsafe.Pointer(&buffer[0]))
	var offsets [Lanes]uint32
	for _, l := range cfg.active {
		last[l] = buffer[l*BlockSize : (l+1)*BlockSize]
		offsets[l] = uint32(l * BlockSize)
		copy(last[l], data[l])
		last[l][n] = 0x80
	}

	if n >= 56 {
		block(&h, base, &offsets, cfg.mask)
		for i := range buffer {
			buffer[i] = 0
		}
	}

	for _, l := range cfg.active {
		binary.LittleEndian.PutUint64(last[l][56:], uint64(8*cfg.n))
	}
	block(&h, base, &offsets, cfg.mask)

	// Write into byte array.
	var digest [Lanes][Size]byte
	for _, l := range cfg.active {
		for i := 0; i < 4; i++ {
			binary.LittleEndian.PutUint32(digest[l][4*i:], h[16*i+l])
		}
	}

	return digest
}

type lanes struct {
	n       int
	active  []int
	mask    uint16
	base    uintptr
	offsets [Lanes]uint32
}

func config(data [Lanes][]byte) (*lanes, error) {
	cfg := &lanes{}

	// Populate active lanes, and ensure they're all the same length.
	for l, d := range data {
		if d != nil {
			cfg.active = append(cfg.active, l)
		}
	}

	if len(cfg.active) == 0 {
		return nil, errors.New("no active lanes")
	}

	cfg.n = len(data[cfg.active[0]])
	for _, l := range cfg.active {
		cfg.mask |= 1 << l
		if len(data[l]) != cfg.n {
			return nil, errors.New("length mismatch")
		}
	}

	// Compute base pointer and lane offsets.
	cfg.base = ^uintptr(0)
	for _, l := range cfg.active {
		ptr := uintptr(unsafe.Pointer(&data[l]))
		if ptr < cfg.base {
			cfg.base = ptr
		}
	}

	for _, l := range cfg.active {
		ptr := uintptr(unsafe.Pointer(&data[l]))
		offset := ptr - cfg.base
		if offset > math.MaxUint32 {
			return nil, errors.New("input data exceed 32-bit memory region")
		}
		cfg.offsets[l] = uint32(offset)
	}

	return cfg, nil
}
