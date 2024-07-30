package browser

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
)

var (
	//go:embed name.bsi.gz
	nameBSIData []byte
	nameBSI     = roaring64.NewDefaultBSI()

	//go:embed version.bsi.gz
	versionBSIData []byte
	versionBSI     = roaring64.NewDefaultBSI()

	//go:embed name_translate.json.gz
	nameTranslateData []byte
	nameTranslate     []string

	//go:embed version_translate.json.gz
	versionTranslateData []byte
	versionTranslate     []string
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
		unpackBSI(nameBSIData, nameBSI)
		unpackBSI(versionBSIData, versionBSI)
		unpackJSON(nameTranslateData, &nameTranslate)
		unpackJSON(versionTranslateData, &versionTranslate)
	})
}

func GetName(id uint64) string {
	setup()
	value, ok := nameBSI.GetValue(id)
	if !ok {
		return ""
	}
	return nameTranslate[value]
}

func GetVersion(id uint64) string {
	setup()
	value, ok := versionBSI.GetValue(id)
	if !ok {
		return ""
	}
	return versionTranslate[value]
}
