package pages

import (
	"crypto/sha256"
	"fmt"
	"github.com/hashicorp/golang-lru"
	"strings"
	"website-indexer/global"
)

type Hash [sha256.Size]byte

const cacheSize = 1024

func (me *Hash) String() string {
	return fmt.Sprintf("%x", me[:])
}

var hashes *lru.Cache

func init() {
	hashes, _ = lru.New(cacheSize)
}
func NewHash(url global.Url) (hash *Hash) {
	url = strings.Trim(url, "/")
	ih, ok := hashes.Get(url)
	if ok {
		hash = ih.(*Hash)
	} else {
		h := Hash(sha256.Sum256([]byte(url)))
		hash = &h
		hashes.Add(url, hash)
	}
	return hash
}
