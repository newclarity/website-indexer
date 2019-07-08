package crawler

import (
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
	"website-indexer/config"
	"website-indexer/global"
	"website-indexer/hosters"
	"website-indexer/hosters/algolia"
	"website-indexer/pages"
	"website-indexer/persist"
	"website-indexer/util"
)

const (
	NoPatternDefined = "No pattern defined in LimitRule"
)

type Crawler struct {
	*config.Config
	Collector *colly.Collector
	Host      hosters.IndexHoster
}

func NewCrawler(cfg *config.Config) (c *Crawler) {
	for range only.Once {
		c = &Crawler{
			Config: cfg,
			Collector: colly.NewCollector(
				colly.AllowedDomains("www."+cfg.Domain, cfg.Domain),
				//colly.Async(true),
			),
		}
		err := c.Collector.Limit(&colly.LimitRule{
			Delay: 250 * time.Millisecond,
		})
		if err == nil {
			break
		}
		if err.Error() == NoPatternDefined {
			break
		}
		logrus.Fatalf("Unable to set crawl delay: %s", err)
	}
	return c
}

func (me *Crawler) Crawl() {
	for range only.Once {
		cfg := me.Config

		pb := pages.NewBuffer()

		host := algolia.NewAlgolia(cfg)
		noop(host)

		me.Collector.OnHTML("*", func(e *global.HtmlElement) {
			me.onHtml(pb, e)
		})

		me.Collector.OnRequest(func(r *colly.Request) {
			me.onRequest(pb, r)
		})

		me.Collector.OnScraped(func(r *colly.Response) {
			me.onScraped(pb, host, r)
		})

		if !persist.HasQueuedUrls(cfg) {
			me.RequestVisits(global.Urls{
				fmt.Sprintf("https://www.%s/", cfg.Domain),
			})
		} else {
			queued, err := persist.GetQueuedUrls(cfg)
			if err != nil {
				logrus.Fatal(err.Error())
			}
			me.RequestVisits(queued)

		}
	}
}

func (me *Crawler) RequestVisits(urls global.Urls) {
	for _, u := range urls {
		err := me.Collector.Visit(u)
		if err == nil {
			continue
		}
		me.Config.OnFailedVisit(err, u, "queuing visit", true)
	}
}

func (me *Crawler) onRequest(pb *pages.Buffer, r *colly.Request) {
	for range only.Once {
		u := pages.NewUrl(r.URL.String())
		pb.CurrentUrl = u.GetUrl()
		p := pb.MaybeMapUrl(u)
		fmt.Print("\nVisiting ", p.Url)
	}
}

func (me *Crawler) onHtml(pb *pages.Buffer, e *global.HtmlElement) {
	for range only.Once {

		url := pages.NewUrl(e.Request.URL.String(), pb.CurrentUrl)
		p := pb.GetByUrl(url)
		if p == nil {
			logrus.Errorf("page not registered for URL '%s'", e.Request.URL.String())
			break
		}

		if me.HasElementName(e, global.IgnoreElemsType) {
			break
		}

		if me.HasElementName(e, global.CollectElemsType) {
			p.AppendElement(e)
		}

		switch e.Name {
		case "a":
			me.onA(pb, p, e)
		case "iframe":
			me.onIFrame(pb, p, e)
		case "title":
			me.onTitle(pb, p, e)
		case "link":
			me.onLink(pb, p, e)
		case "meta":
			me.onMeta(pb, p, e)
		case "body":
			me.onBody(pb, p, e)
		default:
			logrus.Warnf("Unhandled HTML element <%s> in %s: %s",
				e.Name,
				e.Request.URL.String(),
				util.StripWhitespace(e.Text),
			)
		}
	}
}

func (me *Crawler) onScraped(pb *pages.Buffer, host hosters.IndexHoster, r *colly.Response) {
	for range only.Once {
		url := pages.NewUrl(r.Request.URL.String(), pb.CurrentUrl)
		p := pb.GetByUrl(url)
		if p == nil {
			logrus.Warnf("No attributes collected for %s", url)
			break
		}
		if !host.IndexPage(p) {
			// @TODO Move file from /tmp/website-indexer/queued
			//                   to /tmp/website-indexer/error
			break
		}
		// @TODO Move file from /tmp/website-indexer/queued
		//                   to /tmp/website-indexer/indexed
		//                Write /tmp/website-indexer/indexed/{ha}/{sh}/{hash}/json

		// Reset pause
		me.Config.OnErrPause = config.InitialPause
	}
}

func (me *Crawler) onLink(pb *pages.Buffer, p *pages.Page, e *global.HtmlElement) {
	for range only.Once {
		if !me.HasElementRel(e, global.LinkElemsType) {
			break
		}
		url := pages.NewUrl(e.Attr("href"), p.GetUrl())
		p.AddHeader(e.Attr("rel"), url.String())
	}
}

func (me *Crawler) onTitle(pb *pages.Buffer, p *pages.Page, e *global.HtmlElement) {
	texts := strings.Split(e.Text+"|", "|")
	p.Title = strings.TrimSpace(texts[0])
}

func (me *Crawler) onA(pb *pages.Buffer, p *pages.Page, e *global.HtmlElement) {
	for range only.Once {
		if pages.GetRelativeness(e.Attr("href")) != pages.AbsoluteUrl {
			referer := p.GetUrl()
			noop(referer)
		}
		url := pages.NewUrl(e.Attr("href"), p.GetUrl())
		if url == nil {
			break
		}
		me.requestVisit(pb, url, e)
	}
}

func (me *Crawler) onIFrame(pb *pages.Buffer, p *pages.Page, e *global.HtmlElement) {
	for range only.Once {
		if pages.GetRelativeness(e.Attr("src")) != pages.AbsoluteUrl {
			referer := p.GetUrl()
			noop(referer)
		}
		url := pages.NewUrl(e.Attr("src"), p.GetUrl())
		if url == nil {
			break
		}
		me.requestVisit(pb, url, e)
	}
}

func (me *Crawler) onMeta(pb *pages.Buffer, p *pages.Page, e *global.HtmlElement) {
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

func (me *Crawler) onBody(pb *pages.Buffer, p *pages.Page, e *global.HtmlElement) {
	p.Body = append(p.Body, pages.NewElement(e).GetHtml())
}

func (me *Crawler) requestVisit(pb *pages.Buffer, url *pages.Url, e *global.HtmlElement) {
	for range only.Once {
		persist.QueuedUrl(me.Config, url)
		err := e.Request.Visit(url.GetUrl())
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
