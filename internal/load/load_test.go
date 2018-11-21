package load_test

import (
	"bytes"
	"testing"

	"github.com/mmcloughlin/avo/internal/gen"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/load"
	"github.com/mmcloughlin/avo/internal/test"
)

func Load(t *testing.T) []*inst.Instruction {
	t.Helper()
	l := load.NewLoaderFromDataDir("testdata")
	is, err := l.Load()
	if err != nil {
		t.Fatal(err)
	}
	return is
}

func TestAssembles(t *testing.T) {
	is := Load(t)

	g := &gen.LoaderTest{}
	var buf bytes.Buffer
	g.Generate(&buf, is)

	test.Assembles(t, buf.Bytes())
}
