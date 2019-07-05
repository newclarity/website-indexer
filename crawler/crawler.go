package crawler

import (
	"crypto/sha256"
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
	"time"
	"website-indexer/config"
	"website-indexer/global"
	"website-indexer/hosters/algolia"
	"website-indexer/util"
)

type Crawler struct {
	*config.Config
	Collector *colly.Collector
	Host      global.IndexHoster
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
			Delay: 1 * time.Second,
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

	om := make(global.ObjectMap, 0)
	host := algolia.NewAlgolia(cfg)
	noop(host)

	me.Collector.OnHTML("*", func(e *global.HTMLElement) {
		for range only.Once {
			if me.HasElementName(e, global.IgnoreElemsType) {
				break
			}
			if !me.HasElementName(e, global.CollectElemsType) {
				break
			}
			om = util.AppendHtml(om, e, e.Name)
		}
	})

	me.Collector.OnHTML("*", func(e *global.HTMLElement) {
		for range only.Once {
			if me.HasElementName(e, global.IgnoreElemsType) {
				break
			}
			p := util.GetPage(om, e)

			if _, ok := om[p]["urlpath"].(string); !ok {
				om[p]["urlpath"] = e.Request.URL.Path
			}

			switch e.Name {
			case "a":
				me.onA(om[p], e)
			case "iframe":
				me.onIFrame(om[p], e)
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
					util.StripWhitespace(e.Text),
				)
			}
		}
	})

	me.Collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", util.Cleanurl(r.URL.String()))
	})

	me.Collector.OnScraped(func(response *colly.Response) {
		for range only.Once {
			hash := sha256.Sum256([]byte(response.Request.URL.Path))
			p := util.HashToString(hash)
			if len(om[p]) <= 1 {
				logrus.Warnf("No attributes collected for %s", response.Request.URL.Path)
				break
			}
			om[p]["id"] = p
			host.IndexObject(om[p])
		}
	})

	u := fmt.Sprintf("https://www.%s/", cfg.Domain)
	err := me.Collector.Visit(u)
	if err != nil {
		logrus.Errorf("On queuing visit to %s: %s", u, err)
	}
}
