package inst

//go:generate avogen -bootstrap -data ../data -output ztable.go godata
//go:generate avogen -bootstrap -data ../data -output ztable_test.go godatatest

func Lookup(opcode string) (Instruction, bool) {
	for _, i := range Instructions {
		if i.Opcode == opcode {
			return i, true
		}
	}
	return Instruction{}, false
}
