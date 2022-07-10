<p align="center">
  <img src="logo.svg" width="40%" border="0" alt="avo" />
  <br />
  <img src="https://img.shields.io/github/workflow/status/mmcloughlin/avo/ci/master.svg?style=flat-square" alt="Build Status" />
  <a href="https://pkg.go.dev/github.com/mmcloughlin/avo"><img src="https://img.shields.io/badge/doc-reference-007d9b?logo=go&style=flat-square" alt="go.dev" /></a>
  <a href="https://goreportcard.com/report/github.com/mmcloughlin/avo"><img src="https://goreportcard.com/badge/github.com/mmcloughlin/avo?style=flat-square" alt="Go Report Card" /></a>
</p>

<p align="center">Generate x86 Assembly with Go</p>

`avo` makes high-performance Go assembly easier to write, review and maintain. The `avo` package presents a familiar assembly-like interface that simplifies development without sacrificing performance:

* **Use Go control structures** for assembly generation; `avo` programs _are_ Go programs
* **Register allocation**: write functions with virtual registers and `avo` assigns physical registers for you
* **Automatically load arguments and store return values**: ensure memory offsets are correct for complex structures
* **Generation of stub files** to interface with your Go package

For more about `avo`:

* Introductory talk ["Better `x86` Assembly Generation with Go"](https://www.youtube.com/watch?v=6Y5CZ7_tyA4) at [dotGo 2019](https://2019.dotgo.eu/) ([slides](https://speakerdeck.com/mmcloughlin/better-x86-assembly-generation-with-go))
* [Longer tutorial at Gophercon 2019](https://www.youtube.com/watch?v=WaD8sNqroAw) showing a highly-optimized dot product ([slides](https://speakerdeck.com/mmcloughlin/better-x86-assembly-generation-with-go-gophercon-2019))
* Watch [Filippo Valsorda](https://filippo.io/) live code the [rewrite of `filippo.io/edwards25519` assembly with `avo`](https://vimeo.com/679848853)
* Explore [projects using `avo`](doc/adopters.md)
* Discuss `avo` and general Go assembly topics in the [#assembly](https://gophers.slack.com/archives/C6WDZJ70S) channel of [Gophers Slack](https://invite.slack.golangbridge.org/)

_Note: APIs subject to change while `avo` is still in an experimental phase. You can use it to build [real things](examples) but we suggest you pin a version with your package manager of choice._

## Quick Start

Install `avo` with `go get`:

```
$ go get -u github.com/mmcloughlin/avo
```

`avo` assembly generators are pure Go programs. Here's a function that adds two `uint64` values:

```go
//go:build ignore
// +build ignore

package main

import . "github.com/mmcloughlin/avo/build"

func main() {
	TEXT("Add", NOSPLIT, "func(x, y uint64) uint64")
	Doc("Add adds x and y.")
	x := Load(Param("x"), GP64())
	y := Load(Param("y"), GP64())
	ADDQ(x, y)
	Store(y, ReturnIndex(0))
	RET()
	Generate()
}
```

`go run` this code to see the assembly output. To integrate this into the rest of your Go package we recommend a [`go:generate`](https://blog.golang.org/generate) line to produce the assembly and the corresponding Go stub file.

```go
//go:generate go run asm.go -out add.s -stubs stub.go
```

After running `go generate` the [`add.s`](examples/add/add.s) file will contain the Go assembly.

```s
// Code generated by command: go run asm.go -out add.s -stubs stub.go. DO NOT EDIT.

#include "textflag.h"

// func Add(x uint64, y uint64) uint64
TEXT ·Add(SB), NOSPLIT, $0-24
	MOVQ x+0(FP), AX
	MOVQ y+8(FP), CX
	ADDQ AX, CX
	MOVQ CX, ret+16(FP)
	RET
```

The same call will produce the stub file [`stub.go`](examples/add/stub.go) which will enable the function to be called from your Go code.

```go
// Code generated by command: go run asm.go -out add.s -stubs stub.go. DO NOT EDIT.

package add

// Add adds x and y.
func Add(x uint64, y uint64) uint64
```

See the [`examples/add`](examples/add) directory for the complete working example.

## Examples

See [`examples`](examples) for the full suite of examples.

### Slice Sum

Sum a slice of `uint64`s:

```go
func main() {
	TEXT("Sum", NOSPLIT, "func(xs []uint64) uint64")
	Doc("Sum returns the sum of the elements in xs.")
	ptr := Load(Param("xs").Base(), GP64())
	n := Load(Param("xs").Len(), GP64())

	Comment("Initialize sum register to zero.")
	s := GP64()
	XORQ(s, s)

	Label("loop")
	Comment("Loop until zero bytes remain.")
	CMPQ(n, Imm(0))
	JE(LabelRef("done"))

	Comment("Load from pointer and add to running sum.")
	ADDQ(Mem{Base: ptr}, s)

	Comment("Advance pointer, decrement byte count.")
	ADDQ(Imm(8), ptr)
	DECQ(n)
	JMP(LabelRef("loop"))

	Label("done")
	Comment("Store sum to return value.")
	Store(s, ReturnIndex(0))
	RET()
	Generate()
}
```

The result from this code generator is:

```s
// Code generated by command: go run asm.go -out sum.s -stubs stub.go. DO NOT EDIT.

#include "textflag.h"

// func Sum(xs []uint64) uint64
TEXT ·Sum(SB), NOSPLIT, $0-32
	MOVQ xs_base+0(FP), AX
	MOVQ xs_len+8(FP), CX

	// Initialize sum register to zero.
	XORQ DX, DX

loop:
	// Loop until zero bytes remain.
	CMPQ CX, $0x00
	JE   done

	// Load from pointer and add to running sum.
	ADDQ (AX), DX

	// Advance pointer, decrement byte count.
	ADDQ $0x08, AX
	DECQ CX
	JMP  loop

done:
	// Store sum to return value.
	MOVQ DX, ret+24(FP)
	RET
```

Full example at [`examples/sum`](examples/sum).

### Features

For demonstrations of `avo` features:

* **[args](examples/args):** Loading function arguments.
* **[returns](examples/returns):** Building return values.
* **[complex](examples/complex):** Working with `complex{64,128}` types.
* **[data](examples/data):** Defining `DATA` sections.
* **[ext](examples/ext):** Interacting with types from external packages.
* **[pragma](examples/pragma):** Apply compiler directives to generated functions.

### Real Examples

Implementations of full algorithms:

* **[sha1](examples/sha1):** [SHA-1](https://en.wikipedia.org/wiki/SHA-1) cryptographic hash.
* **[fnv1a](examples/fnv1a):** [FNV-1a](https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function#FNV-1a_hash) hash function.
* **[dot](examples/dot):** Vector dot product.
* **[md5x16](examples/md5x16):** AVX-512 accelerated [MD5](https://en.wikipedia.org/wiki/MD5).
* **[geohash](examples/geohash):** Integer [geohash](https://en.wikipedia.org/wiki/Geohash) encoding.
* **[stadtx](examples/stadtx):** [`StadtX` hash](https://github.com/demerphq/BeagleHash) port from [dgryski/go-stadtx](https://github.com/dgryski/go-stadtx).

## Adopters

Popular projects[^projects] using `avo`:

[^projects]: Projects drawn from the `avo` third-party test suite. Popularity
estimated from Github star count collected on Jul 10, 2022.

<img src="https://images.weserv.nl?fit=cover&h=24&mask=circle&maxage=7d&url=https%3A%2F%2Fgithub.com%2Fgolang.png&w=24" width="24" height="24" hspace="4" valign="middle" /> [golang / **go**](https://github.com/golang/go)
:star: 101.5k
> The Go programming language

<img src="https://images.weserv.nl?fit=cover&h=24&mask=circle&maxage=7d&url=https%3A%2F%2Fgithub.com%2Fklauspost.png&w=24" width="24" height="24" hspace="4" valign="middle" /> [klauspost / **compress**](https://github.com/klauspost/compress)
:star: 3k
> Optimized Go Compression Packages

<img src="https://images.weserv.nl?fit=cover&h=24&mask=circle&maxage=7d&url=https%3A%2F%2Fgithub.com%2Fgolang.png&w=24" width="24" height="24" hspace="4" valign="middle" /> [golang / **crypto**](https://github.com/golang/crypto)
:star: 2.5k
> [mirror] Go supplementary cryptography libraries

<img src="https://images.weserv.nl?fit=cover&h=24&mask=circle&maxage=7d&url=https%3A%2F%2Fgithub.com%2Fklauspost.png&w=24" width="24" height="24" hspace="4" valign="middle" /> [klauspost / **reedsolomon**](https://github.com/klauspost/reedsolomon)
:star: 1.5k
> Reed-Solomon Erasure Coding in Go

<img src="https://images.weserv.nl?fit=cover&h=24&mask=circle&maxage=7d&url=https%3A%2F%2Fgithub.com%2Fsegmentio.png&w=24" width="24" height="24" hspace="4" valign="middle" /> [segmentio / **asm**](https://github.com/segmentio/asm)
:star: 750
> Go library providing algorithms optimized to leverage the characteristics of modern CPUs

<img src="https://images.weserv.nl?fit=cover&h=24&mask=circle&maxage=7d&url=https%3A%2F%2Fgithub.com%2Fcloudflare.png&w=24" width="24" height="24" hspace="4" valign="middle" /> [cloudflare / **circl**](https://github.com/cloudflare/circl)
:star: 683
> CIRCL: Cloudflare Interoperable Reusable Cryptographic Library

<img src="https://images.weserv.nl?fit=cover&h=24&mask=circle&maxage=7d&url=https%3A%2F%2Fgithub.com%2Fzeebo.png&w=24" width="24" height="24" hspace="4" valign="middle" /> [zeebo / **blake3**](https://github.com/zeebo/blake3)
:star: 302
> Pure Go implementation of BLAKE3 with AVX2 and SSE4.1 acceleration

<img src="https://images.weserv.nl?fit=cover&h=24&mask=circle&maxage=7d&url=https%3A%2F%2Fgithub.com%2Flukechampine.png&w=24" width="24" height="24" hspace="4" valign="middle" /> [lukechampine / **blake3**](https://github.com/lukechampine/blake3)
:star: 280
> A pure-Go implementation of the BLAKE3 cryptographic hash function

<img src="https://images.weserv.nl?fit=cover&h=24&mask=circle&maxage=7d&url=https%3A%2F%2Fgithub.com%2Fzeebo.png&w=24" width="24" height="24" hspace="4" valign="middle" /> [zeebo / **xxh3**](https://github.com/zeebo/xxh3)
:star: 266
> XXH3 algorithm in Go

<img src="https://images.weserv.nl?fit=cover&h=24&mask=circle&maxage=7d&url=https%3A%2F%2Fgithub.com%2Fminio.png&w=24" width="24" height="24" hspace="4" valign="middle" /> [minio / **md5-simd**](https://github.com/minio/md5-simd)
:star: 121
> Accelerate aggregated MD5 hashing performance up to 8x for AVX512 and 4x for AVX2. Useful for server applications that need to compute many MD5 sums in parallel.

See the [full list of projects using `avo`](doc/adopters.md).

## Contributing

Contributions to `avo` are welcome:

* Feedback from using `avo` in a real project is incredibly valuable. Consider [porting an existing project to `avo`](https://github.com/mmcloughlin/avo/issues/40).
* [Submit bug reports](https://github.com/mmcloughlin/avo/issues/new) to the issues page.
* Pull requests accepted. Take a look at outstanding [issues](https://github.com/mmcloughlin/avo/issues) for ideas (especially the ["good first issue"](https://github.com/mmcloughlin/avo/labels/good%20first%20issue) label).
* Join us in the [#assembly](https://gophers.slack.com/archives/C6WDZJ70S) channel of [Gophers Slack](https://invite.slack.golangbridge.org/).

## Credits

Inspired by the [PeachPy](https://github.com/Maratyszcza/PeachPy) and [asmjit](https://github.com/asmjit/asmjit) projects. Thanks to [Damian Gryski](https://github.com/dgryski) for advice, and his [extensive library of PeachPy Go projects](https://github.com/mmcloughlin/avo/issues/40).

## License

`avo` is available under the [BSD 3-Clause License](LICENSE).
