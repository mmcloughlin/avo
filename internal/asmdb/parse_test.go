package asmdb

import "testing"

func TestParseRawFile(t *testing.T) {
	db, err := ParseRawFile("testdata/x86data.js")
	if err != nil {
		t.Fatal(err)
	}
	if len(db.Instructions) == 0 {
		t.FailNow()
	}
}
