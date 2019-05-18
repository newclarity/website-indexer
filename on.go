package main

import (
	"github.com/gearboxworks/go-status/only"
	"github.com/sirupsen/logrus"
	"strings"
)

func (me *Crawler) onLink(o Object, e *HTMLElement) {
	for range only.Once {
		if !me.hasElementRel(e, LinkElemsType) {
			break
		}
		o[e.Attr("rel")] = strings.TrimSpace(e.Attr("href"))
	}
}

func (me *Crawler) onTitle(o Object, e *HTMLElement) {
	texts := strings.Split(e.Text+"|", "|")
	o["title"] = strings.TrimSpace(texts[0])
}

func (me *Crawler) onA(o Object, e *HTMLElement) {
	for range only.Once {
		u := cleanurl(e.Attr("href"))
		if u == "" {
			break
		}
		err := e.Request.Visit(u)
		if err != nil {
			switch err.Error() {
			case "URL already visited":
			case "Forbidden domain":
			case "Missing URL":
			case "Not Found":
				break
			default:
				logrus.Errorf("On <a>: %s", err)
			}
		}
	}
}

func (me *Crawler) onMeta(o Object, e *HTMLElement) {
	for range only.Once {
		if !me.hasElementMeta(e) {
			break
		}
		o[e.Attr(MetaName)] = strings.TrimSpace(e.Attr(MetaContent))
	}
}

func (me *Crawler) onBody(o Object, e *HTMLElement) {
	o["body"] = getHtml(e)
}
