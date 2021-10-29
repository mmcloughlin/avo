package buildtags

import (
	"bytes"
	"fmt"
	"testing"
)

func TestFormat(t *testing.T) {
	cases := []struct {
		PlusBuild []string
		GoBuild   string
	}{
		{
			PlusBuild: []string{"amd64"},
			GoBuild:   "amd64",
		},
		{
			PlusBuild: []string{
				"linux darwin",
				"amd64 arm64 mips64x ppc64x",
			},
			GoBuild: "(linux || darwin) && (amd64 || arm64 || mips64x || ppc64x)",
		},
		{
			PlusBuild: []string{"!linux,!darwin !amd64,!arm64,!mips64x,!ppc64x"},
			GoBuild:   "(!linux && !darwin) || (!amd64 && !arm64 && !mips64x && !ppc64x)",
		},
		{
			PlusBuild: []string{
				"linux,386 darwin,!cgo",
				"!noasm",
			},
			GoBuild: "((linux && 386) || (darwin && !cgo)) && !noasm",
		},
	}

	for _, c := range cases {
		// Parse constraints.
		var cs Constraints
		for _, expr := range c.PlusBuild {
			constraint, err := ParseConstraint(expr)
			if err != nil {
				t.Fatal(err)
			}
			cs = append(cs, constraint)
		}

		// Build expected output.
		var buf bytes.Buffer
		if GoBuildSyntaxSupported() {
			fmt.Fprintf(&buf, "//go:build %s\n", c.GoBuild)
		}
		if PlusBuildSyntaxSupported() {
			for _, expr := range c.PlusBuild {
				fmt.Fprintf(&buf, "// +build %s\n", expr)
			}
		}
		expect := buf.String()

		// Format and check.
		got, err := Format(cs)
		if err != nil {
			t.Fatal(err)
		}

		if got != expect {
			t.Errorf("got=\n%s\nexpect=\n%s\n", got, expect)
		}
	}
}
