package opcodesxml

import (
	"fmt"
	"testing"
)

func TestReadFile(t *testing.T) {
	is, err := ReadFile("testdata/x86_64.xml")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", is)
}
