package fst

import (
	_ "embed"
	"sync"

	"github.com/blevesearch/vellum"
)

//go:embed fst.gz
var fstData []byte

var once sync.Once
var fst *vellum.FST

func Get() *vellum.FST {
	once.Do(func() {
		var err error
		fst, err = vellum.Load(fstData)
		if err != nil {
			// Highly unlikely because we generate the archive separately.
			panic("corrupt fst archive " + err.Error())
		}
	})
	return fst
}
