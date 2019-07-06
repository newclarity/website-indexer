package crawler

import (
	"github.com/gearboxworks/go-status/only"
	"strings"
	"website-indexer/global"
	"website-indexer/pages"
	"website-indexer/util"
)

func (me *Crawler) onLink(p *pages.Page, e *global.HtmlElement) {
	for range only.Once {
		if !me.HasElementRel(e, global.LinkElemsType) {
			break
		}
		p.AddHeader(e.Attr("rel"), e.Attr("href"))
	}
}

func (me *Crawler) onTitle(p *pages.Page, e *global.HtmlElement) {
	texts := strings.Split(e.Text+"|", "|")
	p.Title = strings.TrimSpace(texts[0])
}

func (me *Crawler) onA(p *pages.Page, e *global.HtmlElement) {
	for range only.Once {
		u := util.Cleanurl(e.Attr("href"))
		if u == "" {
			break
		}
		me.RequestVisit(u, e)
	}
}

func (me *Crawler) onIFrame(p *pages.Page, e *global.HtmlElement) {
	for range only.Once {
		u := util.Cleanurl(e.Attr("src"))
		if u == "" {
			break
		}
		me.RequestVisit(u, e)
	}
}

func (me *Crawler) onMeta(p *pages.Page, e *global.HtmlElement) {
	for range only.Once {
		if !me.HasElementMeta(e) {
			break
		}
		p.AddHeader(
			e.Attr(global.MetaName),
			e.Attr(global.MetaContent),
		)
	}
}

func (me *Crawler) onBody(p *pages.Page, e *global.HtmlElement) {
	p.Body = append(p.Body, pages.NewElement(e).GetHtml())
}
