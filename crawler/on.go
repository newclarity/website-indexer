package crawler

import (
	"github.com/gearboxworks/go-status/only"
	"github.com/sirupsen/logrus"
	"strings"
	"website-indexer/global"
	"website-indexer/util"
)

func (me *Crawler) onLink(o global.Object, e *global.HTMLElement) {
	for range only.Once {
		if !me.HasElementRel(e, global.LinkElemsType) {
			break
		}
		o[e.Attr("rel")] = strings.TrimSpace(e.Attr("href"))
	}
}

func (me *Crawler) onTitle(o global.Object, e *global.HTMLElement) {
	texts := strings.Split(e.Text+"|", "|")
	o["title"] = strings.TrimSpace(texts[0])
}

func (me *Crawler) onA(o global.Object, e *global.HTMLElement) {
	for range only.Once {
		u := util.Cleanurl(e.Attr("href"))
		if u == "" {
			break
		}
		me.RequestUrlVisit(u, e)
	}
}

func (me *Crawler) onIFrame(o global.Object, e *global.HTMLElement) {
	for range only.Once {
		u := util.Cleanurl(e.Attr("src"))
		if u == "" {
			break
		}
		me.RequestUrlVisit(u, e)
	}
}

func (me *Crawler) RequestUrlVisit(u string, e *global.HTMLElement) {
	for range only.Once {
		err := e.Request.Visit(u)
		if err != nil {
			switch err.Error() {
			case "URL already visited":
			case "Forbidden domain":
			case "Missing URL":
			case "Not Found":
				break
			default:
				logrus.Errorf("On <%s>: %s", e.Name, err)
			}
		}
	}
}
func (me *Crawler) onMeta(o global.Object, e *global.HTMLElement) {
	for range only.Once {
		if !me.HasElementMeta(e) {
			break
		}
		o[e.Attr(global.MetaName)] = strings.TrimSpace(e.Attr(global.MetaContent))
	}
}

func (me *Crawler) onBody(o global.Object, e *global.HTMLElement) {
	o["body"] = util.GetHtml(e)
}
