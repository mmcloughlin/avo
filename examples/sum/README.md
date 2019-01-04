# sum

Sum a slice of `uint64`s.

[embedmd]:# (asm.go go /func main/ /^}/)
```go
func main() {
	TEXT("Sum", "func(xs []uint64) uint64")
	Doc("Sum returns the sum of the elements in xs.")
	ptr := Load(Param("xs").Base(), GP64v())
	n := Load(Param("xs").Len(), GP64v())
	s := GP64v()
	XORQ(s, s)
	LABEL("loop")
	CMPQ(n, operand.Imm(0))
	JE(operand.LabelRef("done"))
	ADDQ(operand.Mem{Base: ptr}, s)
	ADDQ(operand.Imm(8), ptr)
	DECQ(n)
	JMP(operand.LabelRef("loop"))
	LABEL("done")
	Store(s, ReturnIndex(0))
	RET()
	Generate()
}
```
