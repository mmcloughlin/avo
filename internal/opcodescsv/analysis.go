package opcodescsv

import (
	"strconv"
	"strings"

	"golang.org/x/arch/x86/x86csv"
)

type Alias struct {
	Opcode      string
	DataSize    int
	NumOperands int
}

// BuildAliasMap constructs a map from AT&T/GNU/Intel to Go syntax.
func BuildAliasMap(is []*x86csv.Inst) (map[Alias]string, error) {
	m := map[Alias]string{}
	for _, i := range is {
		s, err := datasize(i.DataSize)
		if err != nil {
			return nil, err
		}

		if strings.Contains(i.GoOpcode(), "/") {
			continue
		}

		for _, alt := range []string{i.IntelOpcode(), i.GNUOpcode()} {
			if strings.ToUpper(alt) != i.GoOpcode() {
				a := Alias{
					Opcode:      strings.ToLower(alt),
					DataSize:    s,
					NumOperands: len(i.GoArgs()),
				}
				m[a] = i.GoOpcode()
			}
		}
	}
	return m, nil
}

func datasize(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}
