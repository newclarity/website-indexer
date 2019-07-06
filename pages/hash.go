package pages

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"website-indexer/global"
)

type Hash [sha256.Size]byte

func (me *Hash) String() string {

	return fmt.Sprintf("%x", me[:])
}

func NewHash(urlpath global.UrlPath) *Hash {
	urlpath = strings.Trim(urlpath, "/")
	hash := Hash(sha256.Sum256([]byte(urlpath)))
	return &hash
}
