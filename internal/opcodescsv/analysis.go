package opcodescsv

import (
	"reflect"
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
		s, err := strconv.Atoi("0" + i.DataSize)
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

// BuildIntelOrderSet builds the set of instructions that use intel order rather than the usual GNU/AT&T order.
func BuildIntelOrderSet(is []*x86csv.Inst) map[string]bool {
	s := map[string]bool{}
	for _, i := range is {
		if !reflect.DeepEqual(i.GoArgs(), i.GNUArgs()) {
			s[i.GoOpcode()] = true
		}
	}
	return s
}
