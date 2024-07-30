package fst

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"sync"
	
	"github.com/blevesearch/vellum"
)

//go:embed fst.gz
var fstData []byte

var once sync.Once
var fst *vellum.FST

func Get() *vellum.FST {
	once.Do(func() {
		r, _ := gzip.NewReader(bytes.NewReader(fstData))
		all, _ := io.ReadAll(r)
		fst, _ = vellum.Load(all)
	})
	return fst
}
