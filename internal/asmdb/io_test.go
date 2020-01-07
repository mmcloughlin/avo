package asmdb

import "testing"

func TestReadRawFile(t *testing.T) {
	db, err := ReadRawFile("testdata/x86data.js")
	if err != nil {
		t.Fatal(err)
	}
	if len(db.Instructions) == 0 {
		t.FailNow()
	}
}
