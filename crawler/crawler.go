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
	"website-indexer/util"
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
		if err.Error() == "No pattern defined in LimitRule" {
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
			if me.HasElementName(e, global.IgnoreElemsType) {
				break
			}
			if !me.HasElementName(e, global.CollectElemsType) {
				break
			}
			p := pages.NewPage(e.Request.URL.Path)
			pb.MaybeMapPage(p).AppendHtml(e)
		}
	})

	me.Collector.OnHTML("*", func(e *global.HtmlElement) {
		for range only.Once {
			if me.HasElementName(e, global.IgnoreElemsType) {
				break
			}
			p := pb.GetByUrl(e.Request.URL.Path)
			if p == nil {
				pb.MaybeMapPage(pages.NewPage(e.Request.URL.Path))
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
			p := pb.GetByUrl(up)
			if p == nil {
				logrus.Warnf("No attributes collected for %s", up)
				break
			}
			if !host.IndexPage(p) {
				break
			}
			// Reset pause
			me.Config.OnErrPause = config.InitialPause
		}
	})

	u := fmt.Sprintf("https://www.%s/", cfg.Domain)
	err := me.Collector.Visit(u)
	if err != nil {
		me.Config.OnFailedVisit(err, u, "queuing visit", true)
	}
}
func (me *Crawler) RequestVisit(u string, e *global.HtmlElement) {
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
