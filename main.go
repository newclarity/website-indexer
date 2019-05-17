package main

import (
	"bytes"
	"colibri-crawler/only"
	"crypto/sha256"
	"fmt"
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/gocolly/colly"
	"golang.org/x/net/html"
	"log"
	"strings"
	"time"
)

// @see https://benjamincongdon.me/blog/2018/03/01/Scraping-the-Web-in-Golang-with-Colly-and-Goquery/
type Object = algoliasearch.Object
type Map = algoliasearch.Map

type ObjectMap = map[string]Object

type Args struct {
	AppId  string
	ApiKey string
	Index  string
	Domain string
}

func main() {
	crawl(&Args{
		AppId:  "IJJSDKOR05",
		ApiKey: "f3fbdc83509aa03361a242318d1b7c01",
		Index:  "local_McKissock",
		Domain: "mckissock.com",
	})
}

func makeIndex(args *Args) (index algoliasearch.Index) {
	client := algoliasearch.NewClient(args.AppId, args.ApiKey)
	index = client.InitIndex(args.Index)
	settings := algoliasearch.Map{
		"searchableAttributes": []string{
			"title",
			"h1",
			"h2",
			"h3",
			"li",
			"article",
			"body",
		},
	}
	_, err := index.SetSettings(settings)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}
	return index
}

func getHtml(e *colly.HTMLElement) (h string) {
	for range only.Once {
		if e == nil {
			break
		}
		if e.DOM == nil {
			break
		}
		if len(e.DOM.Nodes) == 0 {
			break
		}
		n := e.DOM.Nodes[0]
		var buf bytes.Buffer
		_ = html.Render(&buf, n)
		h = buf.String()
	}
	return h
}

func getId(om ObjectMap, e *colly.HTMLElement) string {
	hash := sha256.Sum256([]byte(e.Request.URL.Path))
	id := hashToString(hash)
	if _, ok := om[id]; !ok {
		om[id] = make(Object, 0)
		om[id]["urlpath"] = e.Request.URL.Path
	}
	return id
}

func appendHtml(om ObjectMap, e *colly.HTMLElement, k string) ObjectMap {
	id := getId(om, e)
	_, ok := om[id][k]
	if !ok {
		om[id][k] = make([]string, 0)
	}
	om[id][k] = append(om[id][k].([]string), getHtml(e))
	return om
}

func crawl(args *Args) {
	c := colly.NewCollector(
		colly.AllowedDomains("www."+args.Domain, args.Domain),
		//colly.Async(true),
	)
	err := c.Limit(&colly.LimitRule{
		Delay: 1 * time.Second,
	})
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}

	om := make(ObjectMap, 0)
	index := makeIndex(args)
	noop(index)

	c.OnHTML("*", func(e *colly.HTMLElement) {
		switch e.Name {
		case "svg", "img", "h1", "h2", "h3", "li", "button":
		case "section", "nav", "header", "article", "main":
			om = appendHtml(om, e, e.Name)
		}
	})

	c.OnHTML("*", func(e *colly.HTMLElement) {
		id := getId(om, e)
		if _, ok := om[id]["urlpath"]; !ok {
			om[id]["urlpath"] = e.Request.URL.Path
		}

		switch e.Name {
		case "a":
			onA(om[id], e)
		case "title":
			onTitle(om[id], e)
		case "link":
			onLink(om[id], e)
		case "meta":
			onMeta(om[id], e)
		case "body":
			om[id]["body"] = getHtml(e)
		default:
			switch e.Name {
			case "html", "head", "script", "style", "noscript", "path", "defs", "symbol", "clipPath":
			case "svg", "use", "circle", "rect", "text", "g", "image", "ul", "li", "section", "nav":
			case "header", "div", "span", "button", "main", "article", "img", "h1", "h2", "h3", "p":
			case "form":
				// Do nothing
			default:
				//fmt.Printf("\t[%d] %s: %s\n",e.Index, e.Name, e.Text)
				noop()
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", cleanurl(r.URL.String()))
	})

	c.OnScraped(func(response *colly.Response) {
		for range only.Once {
			hash := sha256.Sum256([]byte(response.Request.URL.Path))
			id := hashToString(hash)
			if len(om[id]) <= 1 {
				break
			}
			om[id]["objectID"] = id
			_, err := index.AddObject(om[id])
			if err != nil {
				log.Printf("ERROR: %s", err.Error())
			}
			om[id] = make(Object, 0)
		}
	})

	err = c.Visit(fmt.Sprintf("https://www.%s/", args.Domain))
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}
}

func hashToString(hash [sha256.Size]byte) string {
	h := sha256.New()
	h.Write(hash[:])
	bs := h.Sum(nil)
	sh := string(fmt.Sprintf("%x", bs))
	return sh
}

func onLink(o Object, e *colly.HTMLElement) {
	switch e.Attr("rel") {
	case "shortlink", "canonical":
		o[e.Attr("rel")] = strings.TrimSpace(e.Attr("href"))

	default:
		noop()
	}
}
func onTitle(o Object, e *colly.HTMLElement) {
	texts := strings.Split(e.Text+"|", "|")
	o["title"] = strings.TrimSpace(texts[0])
}

func onA(o Object, e *colly.HTMLElement) {
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
				log.Printf("ERROR: %s", err.Error())
			}
		}
	}
}

func onMeta(o Object, e *colly.HTMLElement) {
	if e.Attr("name") == "description" {
		o["description"] = strings.TrimSpace(e.Attr("content"))
	}
}

func cleanurl(u string) string {
	for range only.Once {
		if len(u) == 0 {
			u = "/"
			break
		}
		if strings.Contains(u, "?") {
			break
		}
		if u[0] == '#' {
			u = ""
			break
		}
		if u[:4] == "tel:" {
			u = ""
			break
		}
		if u[len(u)-1] != '/' {
			u += "/"
			break
		}
		if u[len(u)-2:] == "//" {
			u = u[:len(u)-2]
			break
		}
	}
	return u
}

func noop(i ...interface{}) {}
