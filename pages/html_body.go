package pages

import (
	"strings"
	"website-indexer/global"
)

type HtmlBody global.Strings

func (me HtmlBody) String() string {
	return strings.Join(me, "\n")
}
