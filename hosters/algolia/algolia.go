package algolia

import (
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/gearboxworks/go-status/only"
	"github.com/sirupsen/logrus"
	"website-indexer/config"
	"website-indexer/global"
	"website-indexer/hosters"
	"website-indexer/pages"
)

var _ hosters.IndexHoster = (*Algolia)(nil)

type Algolia struct {
	Index  global.Index
	Client algoliasearch.Client
	Config *config.Config
}

func NewAlgolia(c *config.Config) *Algolia {
	a := Algolia{}
	a.Config = c
	client := algoliasearch.NewClient(c.AppId, c.ApiKey)
	a.Index = client.InitIndex(c.IndexName)
	return &a
}

func (me *Algolia) Initialize() error {
	settings := algoliasearch.Map{
		"searchableAttributes": me.Config.SearchAttrs,
	}
	_, err := me.Index.SetSettings(settings)
	if err != nil {
		logrus.Fatalf("Unable to set index settings: %s", err)
	}
	return err
}

func (me *Algolia) IndexPage(p *pages.Page) bool {
	var err error
	for range only.Once {
		o := hosters.NewObject(global.Object{
			"objectID": p.Id.String(),
			"title":    p.Title,
			"urlpath":  p.UrlPath,
			"body":     p.Body.String(),
		})

		o.AppendProperties(p.HeaderMap.ExtractStringMap())

		o.AppendProperties(p.ElementsMap.ExtractStringMap())

		ps := me.Config.UrlPatterns.ExtractStringMap(p.UrlPath)
		o.AppendProperties(ps)
		_, err = me.Index.AddObject(o.Object)
		if err != nil {
			me.Config.OnFailedVisit(err, p.UrlPath, "adding page to index")
			break
		}
	}
	return err == nil
}

func noop(i ...interface{}) interface{} { return i }
