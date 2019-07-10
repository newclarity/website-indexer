package pages

//import (
//	"github.com/gearboxworks/go-status/only"
//	"website-indexer/global"
//)
//
//type Buffer struct {
//	CurrentUrl global.Url
//	PageMap    PageMap
//}
//
//func NewBuffer() *Buffer {
//	return &Buffer{
//		PageMap: make(PageMap, 0),
//	}
//}
//
//// Add a Page to the Page PageMap if not already mapped
//func (me *Buffer) MaybeMapUrl(url global.Url) (p *Page) {
//	for range only.Once {
//		if me.HasUrl(url) {
//			p = me.GetByUrl(url)
//			break
//		}
//		p = NewPage(url)
//		me.PageMap[p.GetHash()] = p
//	}
//	return p
//}
//
//func (me *Buffer) HasUrl(url global.Url) bool {
//	return me.HasPageHash(*NewHash(url))
//}
//
//func (me *Buffer) HasPage(page *Page) bool {
//	return me.HasPageHash(page.GetHash())
//}
//
//func (me *Buffer) HasPageHash(hash Hash) bool {
//	_, has := me.PageMap[hash]
//	return has
//}
//
//func (me *Buffer) GetByUrl(url global.Url) (p *Page) {
//	for range only.Once {
//		hash := NewHash(url)
//		if me.HasPageHash(hash) {
//			p = me.PageMap[hash]
//			break
//		}
//		p = me.PageMap[hash]
//	}
//	return p
//}
