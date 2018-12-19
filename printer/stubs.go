package printer

import (
	"github.com/mmcloughlin/avo"
	"github.com/mmcloughlin/avo/internal/prnt"
)

type stubs struct {
	cfg Config
	prnt.Generator
}

func NewStubs(cfg Config) Printer {
	return &stubs{cfg: cfg}
}

func (s *stubs) Print(f *avo.File) ([]byte, error) {
	s.Comment(s.cfg.GeneratedWarning())
	s.NL()
	s.Printf("package %s\n", s.cfg.Pkg)
	for _, fn := range f.Functions {
		s.NL()
		s.Printf("%s\n", fn.Stub())
	}
	return s.Result()
}
