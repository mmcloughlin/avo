# pragma

Apply [compiler directives](https://golang.org/pkg/cmd/compile/#hdr-Compiler_Directives) to `avo` functions.

The [code generator](asm.go) uses the `Pragma` function to apply the `//go:noescape` directive to the `Add` function:

[embedmd]:# (asm.go go /func main/ /^}/)
```go
func main() {
	TEXT("Add", NOSPLIT, "func(z, x, y *uint64)")
	Pragma("noescape")
	Doc("Add adds the values at x and y and writes the result to z.")
	zptr := Mem{Base: Load(Param("z"), GP64())}
	xptr := Mem{Base: Load(Param("x"), GP64())}
	yptr := Mem{Base: Load(Param("y"), GP64())}
	x, y := GP64(), GP64()
	MOVQ(xptr, x)
	MOVQ(yptr, y)
	ADDQ(x, y)
	MOVQ(y, zptr)
	RET()
	Generate()
}
```

Note the directive is applied in the generated stub file:

[embedmd]:# (stub.go go /\/\/ Add/ /func/)
```go
// Add adds the values at x and y and writes the result to z.
//go:noescape
func
```
