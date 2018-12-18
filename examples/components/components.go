package components

type Struct struct {
	Byte       byte
	Int8       int8
	Uint16     uint16
	Int32      int32
	Uint64     uint64
	Float32    float32
	Float64    float64
	String     string
	Slice      []Sub
	Array      [5]Sub
	Complex64  complex64
	Complex128 complex128
}

type Sub struct {
	A uint64
	B [3]byte
	C uint16
}
