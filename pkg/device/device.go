package device

import (
	"bytes"
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

func setup() {
	once.Do(func() {
		typeBSI.ReadFrom(bytes.NewReader(typeBSIData))
		json.Unmarshal(typeTranslateData, &typeTranslate)
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
