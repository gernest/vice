package bot

import (
	"bytes"

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

func setup() {
	once.Do(func() {
		botBSI.ReadFrom(bytes.NewReader(botBSIData))
	})
}

func GetBot(id uint64) bool {
	setup()
	value, _ := botBSI.GetValue(id)
	return value == 1
}
