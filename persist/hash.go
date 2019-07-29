package persist

import (
	"github.com/gearboxworks/go-status/only"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"hash/fnv"
	"strconv"
	"website-indexer/global"
)

type Hash uint64

const cacheSize = 1024

func (me Hash) String() string {
	return strconv.FormatUint(uint64(me), 10)
}

var hashes *lru.Cache

func init() {
	hashes, _ = lru.New(cacheSize)
}
func NewHash(url global.Url) (hash Hash) {
	for range only.Once {
		ih, ok := hashes.Get(url)
		if ok {
			hash = ih.(Hash)
			break
		}
		h := fnv.New64a()
		_, err := h.Write([]byte(url))
		if err != nil {
			logrus.Errorf("unable to hash URL '%s': %s", url, err)
		}
		hash = Hash(h.Sum64())
	}
	return hash
}
