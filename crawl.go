package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
)

func (me *Crawler) Crawl() {
	cfg := me.Config

	om := make(ObjectMap, 0)
	index := me.makeIndex()
	noop(index)

	me.Collector.OnHTML("*", func(e *HTMLElement) {
		for range only.Once {
			if me.hasElementName(e, IgnoreElemsType) {
				break
			}
			if !me.hasElementName(e, CollectElemsType) {
				break
			}
			om = appendHtml(om, e, e.Name)
		}
	})

	me.Collector.OnHTML("*", func(e *HTMLElement) {
		for range only.Once {
			if me.hasElementName(e, IgnoreElemsType) {
				break
			}
			p := getPage(om, e)
			if _, ok := om[p]["urlpath"]; !ok {
				om[p]["urlpath"] = e.Request.URL.Path
			}

			switch e.Name {
			case "a":
				me.onA(om[p], e)
			case "title":
				me.onTitle(om[p], e)
			case "link":
				me.onLink(om[p], e)
			case "meta":
				me.onMeta(om[p], e)
			case "body":
				me.onBody(om[p], e)
			default:
				logrus.Warnf("Unhandled HTML element <%s> in %s: %s",
					e.Name,
					e.Request.URL.Path,
					stripWhitespace(e.Text),
				)
			}
		}
	})

	me.Collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", cleanurl(r.URL.String()))
	})

	me.Collector.OnScraped(func(response *colly.Response) {
		for range only.Once {
			hash := sha256.Sum256([]byte(response.Request.URL.Path))
			p := hashToString(hash)
			if len(om[p]) <= 1 {
				logrus.Warnf("No attributes collected for %s", response.Request.URL.Path)
				break
			}
			om[p]["p"] = p
			indexObject(index, om[p])
		}
	})

	u := fmt.Sprintf("https://www.%s/", cfg.Domain)
	err := me.Collector.Visit(u)
	if err != nil {
		logrus.Errorf("On queuing visit to %s: %s", u, err)
	}
}
