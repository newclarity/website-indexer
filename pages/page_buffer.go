package pages

import (
	"github.com/gearboxworks/go-status/only"
	"website-indexer/global"
)

type Buffer struct {
	PageMap Map
}

func NewBuffer() *Buffer {
	return &Buffer{
		PageMap: make(Map, 0),
	}
}

// Add a Page to the Page Map if not already mapped
func (me *Buffer) MaybeMapPage(page *Page) *Page {
	for range only.Once {
		if me.HasPage(page) {
			break
		}
		me.PageMap[page.GetHash()] = page
	}
	return page
}

func (me *Buffer) HasPage(page *Page) bool {
	return me.HasPageHash(page.GetHash())
}

func (me *Buffer) HasPageHash(hash Hash) bool {
	_, has := me.PageMap[hash]
	return has
}

func (me *Buffer) GetByElement(ele *global.HtmlElement) (p *Page) {
	return me.GetByUrl(ele.Request.URL.Path)
}

func (me *Buffer) GetByUrl(urlpath global.UrlPath) (p *Page) {
	for range only.Once {
		hash := *NewHash(urlpath)
		if me.HasPageHash(hash) {
			p = me.PageMap[hash]
			break
		}
		p = me.PageMap[hash]
	}
	return p
}
