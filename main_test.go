package main

import (
	"bytes"
	"testing"
)

func Test_Main(t *testing.T) {
	b := bytes.NewBufferString("")
	Main([]string{"", "test.tokn"}, b)

	if b.String() != "test foo\nbar\n" {
		t.Fail()
	}
}
