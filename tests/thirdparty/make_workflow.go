//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/mmcloughlin/avo/tests/thirdparty"
)

var (
	pkgsfilename = flag.String("pkgs", "", "packages configuration")
	output       = flag.String("output", "", "path to output file (default stdout)")
)

func main() {
	if err := mainerr(); err != nil {
		log.Fatal(err)
	}
}

func mainerr() error {
	flag.Parse()

	// Read packages.
	pkgs, err := thirdparty.LoadPackagesFile(*pkgsfilename)
	if err != nil {
		return err
	}

	if err := pkgs.Validate(); err != nil {
		return err
	}

	// Determine output.
	w := os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	// Genenerate workflow file.
	buf := bytes.NewBuffer(nil)
	PrintWorkflow(buf, pkgs)

	// Write output.
	if _, err = io.Copy(w, buf); err != nil {
		return err
	}

	return nil
}

func PrintWorkflow(w io.Writer, pkgs thirdparty.Packages) {
	_, self, _, _ := runtime.Caller(0)
	fmt.Fprintf(w, "# Code generated by %s. DO NOT EDIT.\n\n", filepath.Base(self))

	// Header.
	fmt.Fprint(w, `name: packages

permissions:
  contents: read

on:
  push:
    branches:
      - master
  pull_request:

`)

	// One job per package.

	fmt.Fprintln(w, "jobs:")
	for _, pkg := range pkgs {
		fmt.Fprintf(w, "  %s:\n", pkg.ID())
		fmt.Fprintf(w, "    runs-on: ubuntu-latest\n")
		fmt.Fprintf(w, "    steps:\n")

		// Install Go.
		fmt.Fprintf(w, "    - name: Install Go\n")
		fmt.Fprintf(w, "      uses: actions/setup-go@37335c7bb261b353407cff977110895fa0b4f7d8 # v2.1.3\n")
		fmt.Fprintf(w, "      with:\n")
		fmt.Fprintf(w, "        go-version: 1.17.x\n")

		// Checkout avo.
		avodir := "avo"
		fmt.Fprintf(w, "    - name: Checkout avo\n")
		fmt.Fprintf(w, "      uses: actions/checkout@5a4ac9002d0be2fb38bd78e4b4dbde5606d7042f # v2.3.4\n")
		fmt.Fprintf(w, "      with:\n")
		fmt.Fprintf(w, "        path: %s\n", avodir)
		fmt.Fprintf(w, "        persist-credentials: false\n")

		// Checkout the third-party package.
		pkgdir := pkg.Repository.Name
		fmt.Fprintf(w, "    - name: Checkout %s\n", pkg.Repository)
		fmt.Fprintf(w, "      uses: actions/checkout@5a4ac9002d0be2fb38bd78e4b4dbde5606d7042f # v2.3.4\n")
		fmt.Fprintf(w, "      with:\n")
		fmt.Fprintf(w, "        repository: %s\n", pkg.Repository)
		fmt.Fprintf(w, "        ref: %s\n", pkg.Version)
		fmt.Fprintf(w, "        path: %s\n", pkgdir)
		fmt.Fprintf(w, "        persist-credentials: false\n")

		// Build steps.
		c := &thirdparty.Context{
			AvoDirectory:        path.Join("${{ github.workspace }}", avodir),
			RepositoryDirectory: path.Join("${{ github.workspace }}", pkgdir),
		}

		for _, step := range pkg.Steps(c) {
			fmt.Fprintf(w, "    - name: %s\n", step.Name)
			fmt.Fprintf(w, "      working-directory: %s\n", path.Join(pkgdir, step.WorkingDirectory))
			if len(step.Commands) == 1 {
				fmt.Fprintf(w, "      run: %s\n", step.Commands[0])
			} else {
				fmt.Fprintf(w, "      run: |\n")
				for _, cmd := range step.Commands {
					fmt.Fprintf(w, "        %s\n", cmd)
				}

			}
		}

		fmt.Fprintf(w, "\n")
	}
}
