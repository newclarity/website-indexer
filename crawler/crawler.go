package crawler

import (
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
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
		for range only.Once {

			pb.MaybeMapPathUrl(e.Request.URL.Path)

			if me.HasElementName(e, global.IgnoreElemsType) {
				break
			}

			if !me.HasElementName(e, global.CollectElemsType) {
				break
			}

			pb.GetByUrlPath(e.Request.URL.Path).AppendElement(e)

		}
	})

	me.Collector.OnHTML("*", func(e *global.HtmlElement) {
		for range only.Once {
			if me.HasElementName(e, global.IgnoreElemsType) {
				break
			}
			p := pb.GetByUrlPath(e.Request.URL.Path)
			if p == nil {
				logrus.Warnf("page '%s' not mapped initially", e.Request.URL.Path)
				break
				//p = pb.MaybeMapElement(e)
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
	})

	me.Collector.OnRequest(func(r *colly.Request) {
		fmt.Print("\nVisiting ", util.Cleanurl(r.URL.String()))
	})

	me.Collector.OnScraped(func(r *colly.Response) {
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
	})

	// @TODO Read URLs from /tmp/website-indexer/queued.
	//       Only set to root if no queued files.
	u := fmt.Sprintf("https://www.%s/", cfg.Domain)
	err := me.Collector.Visit(u)
	if err != nil {
		me.Config.OnFailedVisit(err, u, "queuing visit", true)
	}
}
func (me *Crawler) RequestVisit(u string, e *global.HtmlElement) {
	for range only.Once {
		persist.QueueUrlPath(me.Config, u)
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
