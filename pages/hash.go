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
func NewHash(urlpath global.UrlPath) (hash *Hash) {
	urlpath = strings.Trim(urlpath, "/")
	ih, ok := hashes.Get(urlpath)
	if ok {
		hash = ih.(*Hash)
	} else {
		h := Hash(sha256.Sum256([]byte(urlpath)))
		hash = &h
		hashes.Add(urlpath, hash)
	}
	return hash
}
