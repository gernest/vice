package device

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
)

var (
	//go:embed type.bsi.gz
	typeBSIData []byte
	typeBSI     = roaring64.NewDefaultBSI()

	//go:embed type_translate.json.gz
	typeTranslateData []byte
	typeTranslate     []string
)

var once sync.Once

func unpackBSI(data []byte, b *roaring64.BSI) error {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	_, err = b.ReadFrom(r)
	return err
}

func unpackJSON(data []byte, b any) error {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	return json.NewDecoder(r).Decode(b)
}

func setup() {
	once.Do(func() {
		typeBSI.ReadFrom(bytes.NewReader(typeBSIData))
		unpackBSI(typeBSIData, typeBSI)
		unpackJSON(typeTranslateData, &typeTranslate)
	})
}

func GetType(id uint64) string {
	setup()
	value, ok := typeBSI.GetValue(id)
	if !ok {
		return ""
	}
	return typeTranslate[value]
}
