package asmdb

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
)

// Extension specifies an architecture extension.
type Extension struct {
	Name string `json:"name"`
}

// Attribute is specifies an instruction attribute.
type Attribute struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Doc  string `json:"doc"`
}

// SpecialReg specifies a special register or flag.
type SpecialReg struct {
	Name  string `json:"name"`
	Group string `json:"group"`
	Doc   string `json:"doc"`
}

// Shortcut specifies a shorthand for a collection of attributes.
type Shortcut struct {
	Name   string `json:"name"`
	Expand string `json:"expand"`
}

// Register defines a class of registers.
type Register struct {
	Kind  string   `json:"kind"`
	Any   string   `json:"any"`
	Names []string `json:"names"`
}

// Raw is an asmdb instruction database in its unprocessed form.
type Raw struct {
	Architectures []string            `json:"architectures"`
	Extensions    []Extension         `json:"extensions"`
	Attributes    []Attribute         `json:"attributes"`
	SpecialRegs   []SpecialReg        `json:"specialRegs"`
	Shortcuts     []Shortcut          `json:"shortcuts"`
	Registers     map[string]Register `json:"registers"`
	Instructions  [][]string          `json:"instructions"`
}

// ReadRaw reads the asmdb x86data.js file as a raw unprocessed data structure.
func ReadRaw(r io.Reader) (*Raw, error) {
	// Extract the JSON blob from the javascript file.
	b, err := extractjson(r)
	if err != nil {
		return nil, err
	}

	// Parse JSON.
	db := &Raw{}
	if err := json.Unmarshal(b, db); err != nil {
		return nil, err
	}

	return db, nil
}

// ReadRawFile parses an asmdb x86data.js file.
func ReadRawFile(filename string) (*Raw, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadRaw(f)
}

// extractjson extracts the pure JSON part of the x86data.js file.
func extractjson(r io.Reader) ([]byte, error) {
	dividers := []string{
		`// ${JSON:BEGIN}`,
		`// ${JSON:END}`,
		`/*`,
		`*/`,
	}
	s := bufio.NewScanner(r)
	var b []byte
	take := false
	for s.Scan() {
		line := s.Text()

		if contains(strings.TrimSpace(line), dividers) {
			take = !take
			continue
		}

		if take {
			b = append(b, []byte(line)...)
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return b, nil
}

func contains(needle string, haystack []string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
