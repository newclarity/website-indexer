package pages

import (
	"github.com/gearboxworks/go-status/only"
	"website-indexer/global"
)

type Buffer struct {
	CurrentUrl global.Url
	PageMap    Map
}

func NewBuffer() *Buffer {
	return &Buffer{
		PageMap: make(Map, 0),
	}
}

// Add a Page to the Page Map if not already mapped
func (me *Buffer) MaybeMapUrl(url *Url) (p *Page) {
	for range only.Once {
		if me.HasUrl(url) {
			p = me.GetByUrl(url)
			break
		}
		u := url.GetUrl()
		ru := me.CurrentUrl
		if u == ru {
			p = NewPage(url.GetUrl())
		} else {
			referer := me.CurrentUrl
			noop(referer)
			p = NewPage(url.GetUrl(), me.CurrentUrl)
		}
		me.PageMap[p.GetHash()] = p
	}
	return p
}

func (me *Buffer) HasUrl(url *Url) bool {
	return me.HasPageHash(*url.Hash())
}

func (me *Buffer) HasPage(page *Page) bool {
	return me.HasPageHash(page.GetHash())
}

func (me *Buffer) HasPageHash(hash Hash) bool {
	_, has := me.PageMap[hash]
	return has
}

func (me *Buffer) GetByUrl(url *Url) (p *Page) {
	for range only.Once {
		hash := *url.Hash()
		if me.HasPageHash(hash) {
			p = me.PageMap[hash]
			break
		}
		p = me.PageMap[hash]
	}
	return p
}
