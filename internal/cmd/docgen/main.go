package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/mmcloughlin/avo/tests/thirdparty"
)

func main() {
	log.SetPrefix("docgen: ")
	log.SetFlags(0)
	if err := mainerr(); err != nil {
		log.Fatal(err)
	}
}

var (
	tmpl         = flag.String("tmpl", "", "template file")
	output       = flag.String("output", "", "path to output file (default stdout)")
	pkgsfilename = flag.String("pkgs", "", "packages configuration")
)

func mainerr() (err error) {
	flag.Parse()

	// Initialize template.
	t := template.New("doc")

	t.Funcs(template.FuncMap{
		"include": include,
		"snippet": snippet,
	})

	// Load template.
	if *tmpl == "" {
		return errors.New("missing template file")
	}

	b, err := ioutil.ReadFile(*tmpl)
	if err != nil {
		return err
	}

	if _, err := t.Parse(string(b)); err != nil {
		return err
	}

	// Load third-party packages.
	if *pkgsfilename == "" {
		return errors.New("missing packages configuration")
	}

	pkgs, err := thirdparty.LoadPackagesFile(*pkgsfilename)
	if err != nil {
		return err
	}

	// Execute.
	data := map[string]interface{}{
		"Packages": pkgs,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}
	body := buf.Bytes()

	// Output.
	if *output != "" {
		err = ioutil.WriteFile(*output, body, 0640)
	} else {
		_, err = os.Stdout.Write(body)
	}

	if err != nil {
		return err
	}

	return nil
}

// include template function.
func include(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// snippet of a file between start and end regular expressions.
func snippet(filename, start, end string) (string, error) {
	// Parse regular expressions.
	startx, err := regexp.Compile(start)
	if err != nil {
		return "", err
	}

	endx, err := regexp.Compile(end)
	if err != nil {
		return "", err
	}

	// Read the full file.
	data, err := include(filename)
	if err != nil {
		return "", err
	}

	// Collect matched lines.
	var buf bytes.Buffer
	output := false
	s := bufio.NewScanner(strings.NewReader(data))
	for s.Scan() {
		line := s.Text()
		if startx.MatchString(line) {
			output = true
		}
		if output {
			fmt.Fprintln(&buf, line)
		}
		if endx.MatchString(line) {
			output = false
		}
	}

	return buf.String(), nil
}
