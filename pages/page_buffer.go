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

func (me *Buffer) MaybeMapElement(e *global.HtmlElement) *Page {
	p := NewPage(e.Request.URL.Path)
	me.MaybeMapPage(p).AppendElement(e)
	return p
}

// Add a Page to the Page Map if not already mapped
func (me *Buffer) MaybeMapPathUrl(urlpath global.UrlPath) (p *Page) {
	for range only.Once {
		if me.HasUrlPath(urlpath) {
			p = me.GetByUrlPath(urlpath)
			break
		}
		p = NewPage(urlpath)
		me.PageMap[*NewHash(urlpath)] = p
	}
	return p
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

func (me *Buffer) HasUrlPath(urlpath global.UrlPath) bool {
	return me.HasPageHash(*NewHash(urlpath))
}

func (me *Buffer) HasPage(page *Page) bool {
	return me.HasPageHash(page.GetHash())
}

func (me *Buffer) HasPageHash(hash Hash) bool {
	_, has := me.PageMap[hash]
	return has
}

func (me *Buffer) GetByElement(ele *global.HtmlElement) (p *Page) {
	return me.GetByUrlPath(ele.Request.URL.Path)
}

func (me *Buffer) GetByUrlPath(urlpath global.UrlPath) (p *Page) {
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
