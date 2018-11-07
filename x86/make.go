// +build ignore

package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"golang.org/x/arch/x86/x86csv"
)

const csvname = "x86.v0.2.csv"

var csvpath = flag.String(
	"csv",
	filepath.Join(build.Default.GOPATH, "src/golang.org/x/arch/x86", csvname),
	"path to "+csvname,
)

func UniqueOpCodes(instrs []*x86csv.Inst) []string {
	set := map[string]bool{}
	for _, i := range instrs {
		set[i.GoOpcode()] = true
	}

	opcodes := make([]string, 0, len(set))
	for opcode := range set {
		opcodes = append(opcodes, opcode)
	}

	sort.Strings(opcodes)

	return opcodes
}

var functionRegex = regexp.MustCompile(`^[A-Z0-9]+$`)

func main() {
	flag.Parse()

	f, err := os.Open(*csvpath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := x86csv.NewReader(f)

	instrs, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("package x86\n\nimport \"github.com/mmcloughlin/avo\"\n")

	for _, opcode := range UniqueOpCodes(instrs) {
		if !functionRegex.MatchString(opcode) {
			log.Printf("skip %s", opcode)
			continue
		}
		fmt.Printf("\nfunc %s(f *avo.Function, operands ...string) {\n", opcode)
		fmt.Printf("\tf.AddInstruction(avo.Instruction{\n")
		fmt.Printf("\t\tMnemonic: \"%s\",\n", opcode)
		fmt.Printf("\t\tOperands: operands,\n")
		fmt.Printf("\t})\n")
		fmt.Printf("}\n")
	}
}
