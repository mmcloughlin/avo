package main

import (
	"flag"
	"go/build"
	"log"
	"os"
	"path/filepath"

	"github.com/mmcloughlin/avo/internal/gen"
	"github.com/mmcloughlin/avo/internal/load"
)

var generators = map[string]gen.Interface{
	"asmtest": gen.NewAsmTest(),
	"godata":  gen.NewGoData(),
}

var datadir = flag.String(
	"data",
	filepath.Join(build.Default.GOPATH, "src/github.com/mmcloughlin/avo/internal/data"),
	"path to data directory",
)

var output = flag.String("output", "", "path to output file (default stdout)")

func main() {
	flag.Parse()

	// Build generator.
	t := flag.Arg(0)
	g := generators[t]
	if g == nil {
		log.Fatalf("unknown generator type '%s'", t)
	}

	// Determine output writer.
	w := os.Stdout
	if *output != "" {
		f, err := os.Open(*output)
		if err != nil {
			log.Fatal(err)
		}
		w = f
	}

	// Load instructions.
	l := load.NewLoaderFromDataDir(*datadir)
	is, err := l.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Generate output.
	b, err := g.Generate(is)
	if err != nil {
		log.Fatal(err)
	}

	// Write.
	if _, err := w.Write(b); err != nil {
		log.Fatal(err)
	}
}
