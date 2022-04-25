// Command docgen generates documentation from templates.
package main

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
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
	typ          = flag.String("type", "", "documentation type")
	tmpl         = flag.String("tmpl", "", "explicit template file (overrides -type)")
	output       = flag.String("output", "", "path to output file (default stdout)")
	prjsfilename = flag.String("prjs", "", "projects configuration")
)

func mainerr() (err error) {
	flag.Parse()

	// Initialize template.
	t := template.New("doc")

	t.Funcs(template.FuncMap{
		"include": include,
		"snippet": snippet,
		"avatar":  avatar,
	})

	// Load template.
	s, err := load()
	if err != nil {
		return err
	}

	if _, err := t.Parse(s); err != nil {
		return err
	}

	// Load third-party projects.
	if *prjsfilename == "" {
		return errors.New("missing projects configuration")
	}

	prjs, err := thirdparty.LoadProjectsFile(*prjsfilename)
	if err != nil {
		return err
	}

	// Execute.
	data := map[string]interface{}{
		"Projects": prjs,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}
	body := buf.Bytes()

	// Output.
	if *output != "" {
		err = ioutil.WriteFile(*output, body, 0o640)
	} else {
		_, err = os.Stdout.Write(body)
	}

	if err != nil {
		return err
	}

	return nil
}

//go:embed templates
var templates embed.FS

// load template.
func load() (string, error) {
	// Prefer explicit filename, if provided.
	if *tmpl != "" {
		b, err := ioutil.ReadFile(*tmpl)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	// Otherwise expect a named type.
	if *typ == "" {
		return "", errors.New("missing documentation type")
	}
	path := fmt.Sprintf("templates/%s.tmpl", *typ)
	b, err := templates.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("unknown documentation type %q", *typ)
	}

	return string(b), nil
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

// avatar returns HTML for a Github user avatar.
func avatar(owner string, size int) (string, error) {
	// Origin avatar URL from Github.
	u := fmt.Sprintf("https://github.com/%s.png", owner)

	// Use images.weserv.nl service to resize and apply circle mask.
	v := url.Values{}
	v.Set("url", u)
	v.Set("w", strconv.Itoa(size))
	v.Set("h", strconv.Itoa(size))
	v.Set("fit", "cover")
	v.Set("mask", "circle")
	v.Set("maxage", "7d")

	src := url.URL{
		Scheme:   "https",
		Host:     "images.weserv.nl",
		RawQuery: v.Encode(),
	}

	// Build <img> tag.
	format := `<img src="%s" width="%d" height="%d" hspace="4" valign="middle" />`
	return fmt.Sprintf(format, src.String(), size, size), nil
}
