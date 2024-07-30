package main

import (
	"bytes"
	"testing"
)

func TestModule(t *testing.T) {
	f, err := module("test", droar, dvellum)
	if err != nil {
		t.Fatal(err)
	}
	want := `module test

require (
	github.com/RoaringBitmap/roaring/v2 v2.3.1
	github.com/blevesearch/vellum v1.0.10
)
`
	if !bytes.Equal([]byte(want), f) {
		t.Fatal("invalid module")
	}
}
