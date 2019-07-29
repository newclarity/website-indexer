package crawler

import (
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"github.com/gocolly/colly/storage"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
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

var _ debug.Debugger = (*Debugger)(nil)

type Debugger struct{}

func (me *Debugger) Init() error {
	return nil
}

func (me *Debugger) Event(e *debug.Event) {
	return
}

type Crawler struct {
	*config.Config
	Collector *colly.Collector
	Host      hosters.IndexHoster
	Storage   persist.Storager
	Page      *pages.Page
}

func NewCrawler(cfg *config.Config) (c *Crawler) {
	for range only.Once {
		cc := colly.NewCollector(
			colly.CacheDir(cfg.CacheDir),
			colly.ParseHTTPErrorResponse(),
			colly.AllowedDomains("www."+cfg.Domain, cfg.Domain),
			colly.Debugger(&Debugger{}),
		)

		cc.RedirectHandler = func(req *http.Request, via []*http.Request) error {
			return nil
		}

		c = &Crawler{
			Config:    cfg,
			Collector: cc,
		}

		fp := persist.GetDbFilepath(cfg)
		c.Storage = &persist.Storage{
			Config:   cfg,
			Filename: fp,
		}

		err := cc.SetStorage(c.Storage.(storage.Storage))
		if err != nil {
			logrus.Fatalf("Unable to open crawl DB: %s", fp, err)
			break
		}

		err = c.Collector.Limit(&colly.LimitRule{
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

func (me *Crawler) HasQueuedUrls() bool {
	n, _ := me.Storage.QueueSize()
	return n != 0
}

func (me *Crawler) Crawl() *Crawler {

	me.Host = algolia.NewAlgolia(me.Config)
	me.Collector.OnHTML("*", me.onHtml)
	me.Collector.OnRequest(me.onRequest)
	me.Collector.OnScraped(me.onScraped)

	if !me.HasQueuedUrls() {
		me.QueueUrl(me.RootUrl())
	}
	me.VisitQueued()

	return me
}

func (me *Crawler) RootUrl() global.Url {
	return fmt.Sprintf("https://www.%s/", me.Config.Domain)
}

func (me *Crawler) Close() {
	err := me.Storage.Close()
	if err != nil {
		logrus.Errorf("unable to close Sqlite storage")
	}
}

// AddURL adds a new URL to the queue
func (me *Crawler) AddUrl(URL string) (err error) {
	for range only.Once {
		var u *url.URL
		u, err = url.Parse(URL)
		if err != nil {
			break
		}
		r := &colly.Request{
			URL:    u,
			Method: "GET",
		}
		var b []byte
		b, err = r.Marshal()
		if err != nil {
			break
		}
		err = me.Storage.AddRequest(b)
		if err != nil {
			break
		}
	}
	return err
}

func (me *Crawler) QueueUrl(url global.Url) {
	err := me.AddUrl(url)
	if err != nil {
		logrus.Errorf("URL '%s' not added to queue", url)
	}
}

func (me *Crawler) VisitQueued() {
	var err error
	var retries int
	for {
		if err != nil {
			logrus.Errorf("unable to visit queued resources: %s", err)
			err = nil
			if retries > 3 {
				logrus.Fatal("too many errors; terminating")
				os.Exit(1)
			}
			retries++
		}
		var b []byte
		b, err = me.Storage.GetRequest()
		if err != nil {
			break
		}
		if b == nil {
			break
		}
		var res *persist.Resource
		res, err = me.Storage.UnmarshalResource(b)
		if err != nil {
			continue
		}
		var u global.Url
		u, err = res.Url()
		if err != nil {
			continue
		}
		err = me.Collector.Visit(u)
		if err == nil {
			continue
		}
		if err == colly.ErrAlreadyVisited {
			err = nil
			continue
		}
		if err == colly.ErrForbiddenDomain {
			err = nil
			continue
		}
	}
	if err != nil {
		me.Config.OnFailedVisit(err, "", "visiting queued", true)
	}

}

func (me *Crawler) onRequest(r *colly.Request) {
	fmt.Print("\nVisiting ", r.URL)
	me.Page = pages.NewPage(r.URL)
}

func (me *Crawler) onHtml(e *global.HtmlElement) {
	for range only.Once {
		if me.Page == nil {
			logrus.Errorf("page not registered for URL '%s'", e.Request.URL.String())
			break
		}

		if me.HasElementName(e, global.IgnoreElemsType) {
			break
		}

		if me.HasElementName(e, global.CollectElemsType) {
			me.Page.AppendElement(e)
		}

		switch e.Name {
		case "a":
			me.onA(e)
		case "iframe":
			me.onIFrame(e)
		case "title":
			me.onTitle(e)
		case "link":
			me.onLink(e)
		case "meta":
			me.onMeta(e)
		case "body":
			me.onBody(e)
		default:
			logrus.Warnf("Unhandled HTML element <%s> in %s: %s",
				e.Name,
				e.Request.URL.String(),
				util.StripWhitespace(e.Text),
			)
		}
	}
}

func (me *Crawler) onScraped(r *colly.Response) {
	for range only.Once {
		p := me.Page
		p.Url = r.Request.URL.String()
		p.Id = pages.NewHash(p.Url)
		if !me.Host.IndexPage(p) {
			logrus.Warnf("No attributes collected for %s", p.Url)
			break
		}
		// Reset pause
		me.Config.OnErrPause = config.InitialPause
	}
}

func (me *Crawler) onLink(e *global.HtmlElement) {
	for range only.Once {
		if !me.HasElementRel(e, global.LinkElemsType) {
			break
		}
		me.Page.AddHeader(
			e.Attr("rel"),
			e.Request.AbsoluteURL(e.Attr("href")),
		)
	}
}

func (me *Crawler) onTitle(e *global.HtmlElement) {
	texts := strings.Split(e.Text+"|", "|")
	me.Page.Title = strings.TrimSpace(texts[0])
}

func (me *Crawler) onA(e *global.HtmlElement) {
	for range only.Once {
		u := e.Request.AbsoluteURL(e.Attr("href"))
		if !pages.IsIndexable(u) {
			break
		}
		me.requestVisit(u, e)
	}
}

func (me *Crawler) onIFrame(e *global.HtmlElement) {
	for range only.Once {
		u := e.Request.AbsoluteURL(e.Attr("src"))
		if !pages.IsIndexable(u) {
			break
		}
		me.requestVisit(u, e)
	}
}

func (me *Crawler) onMeta(e *global.HtmlElement) {
	for range only.Once {
		if !me.HasElementMeta(e) {
			break
		}
		me.Page.AddHeader(
			e.Attr(global.MetaName),
			e.Attr(global.MetaContent),
		)
	}
}

func (me *Crawler) onBody(e *global.HtmlElement) {
	me.Page.Body = append(
		me.Page.Body,
		pages.NewElement(e).GetHtml(),
	)
}

func (me *Crawler) requestVisit(url global.Url, e *global.HtmlElement) {
	for range only.Once {
		//if ! persist.QueueUrl(me.Config, url) {
		//	break
		//}
		err := me.AddUrl(url)
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
