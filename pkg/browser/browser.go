package browser

import (
	"bytes"
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

func setup() {
	once.Do(func() {
		nameBSI.ReadFrom(bytes.NewReader(nameBSIData))
		versionBSI.ReadFrom(bytes.NewReader(versionBSIData))
		json.Unmarshal(nameTranslateData, &nameTranslate)
		json.Unmarshal(versionTranslateData, &versionTranslate)
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
