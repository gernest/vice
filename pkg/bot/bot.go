package bot

import (
	"bytes"

	"compress/gzip"
	_ "embed"
	"sync"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
)

var (
	//go:embed bot.bsi.gz
	botBSIData []byte
	botBSI     = roaring64.NewDefaultBSI()
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

func setup() {
	once.Do(func() {
		unpackBSI(botBSIData, botBSI)
	})
}

func GetBot(id uint64) bool {
	setup()
	value, _ := botBSI.GetValue(id)
	return value == 1
}
