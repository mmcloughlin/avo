// Package md5x16 implements 16-lane parallel MD5 with AVX-512 instructions.
package md5x16

import (
	"encoding/binary"
	"errors"
	"math"
	"reflect"
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
	var h [4][Lanes]uint32
	for _, l := range cfg.active {
		h[0][l] = 0x67452301
		h[1][l] = 0xefcdab89
		h[2][l] = 0x98badcfe
		h[3][l] = 0x10325476
	}

	// Consume full blocks.
	base, n := cfg.base, cfg.n
	for ; n >= BlockSize; n -= BlockSize {
		block(&h, base, &cfg.offsets, cfg.mask)
		base += BlockSize
	}

	// Final block.
	var last [Lanes][]byte
	var buffer [Lanes * BlockSize]byte
	base = dataptr(buffer[:])
	var offsets [Lanes]uint32
	for _, l := range cfg.active {
		last[l] = buffer[l*BlockSize : (l+1)*BlockSize]
		offsets[l] = uint32(l * BlockSize)
		copy(last[l], data[l][cfg.n-n:])
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
			binary.LittleEndian.PutUint32(digest[l][4*i:], h[i][l])
		}
	}

	return digest
}

// lanes represents the configuration of the 16 data lanes of an MD5
// computation.
type lanes struct {
	n       int           // length of all active (non-nil) lanes
	active  []int         // indexes of active lanes
	mask    uint16        // mask of active lanes
	base    uintptr       // base pointer
	offsets [Lanes]uint32 // offset of data lanes relative to base
}

// config determines the lane configuration for the provided data. Returns an
// error if there are no active lanes, there's a length mismatch among active
// lanes, or the data spans a memory region larger than 32-bits.
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
		ptr := dataptr(data[l])
		if ptr < cfg.base {
			cfg.base = ptr
		}
	}

	for _, l := range cfg.active {
		ptr := dataptr(data[l])
		offset := ptr - cfg.base
		if offset > math.MaxUint32 {
			return nil, errors.New("input data exceed 32-bit memory region")
		}
		cfg.offsets[l] = uint32(offset)
	}

	return cfg, nil
}

// dataptr extracts the data pointer from the given slice.
func dataptr(data []byte) uintptr {
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	return hdr.Data
}
