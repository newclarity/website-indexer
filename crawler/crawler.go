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
	cfg := me.Config

	pb := pages.NewBuffer()

	host := algolia.NewAlgolia(cfg)
	noop(host)

	me.Collector.OnHTML("*", func(e *global.HtmlElement) {
		me.onHtml(pb, e)
	})

	me.Collector.OnRequest(func(r *colly.Request) {
		fmt.Print("\nVisiting ", util.Cleanurl(r.URL.String()))
	})

	me.Collector.OnScraped(func(r *colly.Response) {
		me.onScraped(pb, host, r)
	})

	// @TODO Read URLs from /tmp/website-indexer/queued.
	//       Only set to root if no queued files.
	u := fmt.Sprintf("https://www.%s/", cfg.Domain)
	err := me.Collector.Visit(u)
	if err != nil {
		me.Config.OnFailedVisit(err, u, "queuing visit", true)
	}
}

func (me *Crawler) onHtml(pb *pages.Buffer, e *global.HtmlElement) {
	for range only.Once {

		p := pb.MaybeMapPathUrl(e.Request.URL.Path)

		if me.HasElementName(e, global.IgnoreElemsType) {
			break
		}

		if me.HasElementName(e, global.CollectElemsType) {
			p.AppendElement(e)
		}

		switch e.Name {
		case "a":
			me.onA(p, e)
		case "iframe":
			me.onIFrame(p, e)
		case "title":
			me.onTitle(p, e)
		case "link":
			me.onLink(p, e)
		case "meta":
			me.onMeta(p, e)
		case "body":
			me.onBody(p, e)
		default:
			logrus.Warnf("Unhandled HTML element <%s> in %s: %s",
				e.Name,
				e.Request.URL.Path,
				util.StripWhitespace(e.Text),
			)
		}
	}
}

func (me *Crawler) onScraped(pb *pages.Buffer, host hosters.IndexHoster, r *colly.Response) {
	for range only.Once {
		up := r.Request.URL.Path
		p := pb.GetByUrlPath(up)
		if p == nil {
			logrus.Warnf("No attributes collected for %s", up)
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

func (me *Crawler) RequestVisit(u string, e *global.HtmlElement) {
	for range only.Once {
		persist.QueuedUrlPath(me.Config, u)
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
