package buildtags

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"strings"
)

func PlusBuildSyntaxSupported() bool { return plusbuild }

func GoBuildSyntaxSupported() bool { return gobuild }

func Format(t ConstraintsConvertable) (string, error) {
	// Print build tags to minimal Go source that can be passed to go/format.
	src := t.ToConstraints().GoString() + "\npackage stub"

	// Format them.
	formatted, err := format.Source([]byte(src))
	if err != nil {
		return "", fmt.Errorf("format build constraints: %w", err)
	}

	// Extract the comment lines.
	buf := bytes.NewReader(formatted)
	scanner := bufio.NewScanner(buf)
	output := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "//") {
			output += line + "\n"
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("parse formatted build constraints: %w", err)
	}

	return output, nil
}
