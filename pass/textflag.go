package pass

import (
	"github.com/mmcloughlin/avo"
)

// IncludeTextFlagHeader includes textflag.h if necessary.
func IncludeTextFlagHeader(f *avo.File) error {
	const textflagheader = "textflag.h"

	// Check if we already have it.
	for _, path := range f.Includes {
		if path == textflagheader {
			return nil
		}
	}

	// Add it if necessary.
	if requirestextflags(f) {
		f.Includes = append(f.Includes, textflagheader)
	}

	return nil
}

// requirestextflags returns whether the file uses flags in the textflags.h header.
func requirestextflags(f *avo.File) bool {
	for _, s := range f.Sections {
		var a avo.Attribute
		switch s := s.(type) {
		case *avo.Function:
			a = s.Attributes
		case *avo.Global:
			a = s.Attributes
		}
		if a.ContainsTextFlags() {
			return true
		}
	}
	return false
}
