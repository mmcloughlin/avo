# fnv1a

[FNV-1a](https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function#FNV-1a_hash) in `avo`.

[embedmd]:# (asm.go /const/ $)
```go
const (
	OffsetBasis = 0xcbf29ce484222325
	Prime       = 0x100000001b3
)

func main() {
	TEXT("Hash64", "func(data []byte) uint64")
	Doc("Hash64 computes the FNV-1a hash of data.")
	ptr := Load(Param("data").Base(), GP64v())
	n := Load(Param("data").Len(), GP64v())

	h := reg.RAX
	MOVQ(operand.Imm(OffsetBasis), h)
	p := GP64v()
	MOVQ(operand.Imm(Prime), p)

	LABEL("loop")
	CMPQ(n, operand.Imm(0))
	JE(operand.LabelRef("done"))
	b := GP64v()
	MOVBQZX(operand.Mem{Base: ptr}, b)
	XORQ(b, h)
	MULQ(p)
	INCQ(ptr)
	DECQ(n)

	JMP(operand.LabelRef("loop"))
	LABEL("done")
	Store(h, ReturnIndex(0))
	RET()
	Generate()
}
```
