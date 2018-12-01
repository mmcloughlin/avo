package build

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/mmcloughlin/avo"
)

// ctx provides a global build context.
var ctx = NewContext()

func TEXT(name string)  { ctx.Function(name) }
func LABEL(name string) { ctx.Label(avo.Label(name)) }

var (
	output = flag.String("output", "", "output filename (default stdout)")
)

func EOF() {
	if !flag.Parsed() {
		flag.Parse()
	}

	var w io.Writer = os.Stdout
	if *output != "" {
		if f, err := os.Create(*output); err != nil {
			log.Fatal(err)
		} else {
			defer f.Close()
			w = f
		}
	}

	os.Exit(ctx.Main(w, os.Stderr))
}
