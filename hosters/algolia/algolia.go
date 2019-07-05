package algolia

import (
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/sirupsen/logrus"
	"website-indexer/config"
	"website-indexer/global"
)

var _ global.IndexHoster = (*Algolia)(nil)

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
func (me *Algolia) IndexObject(o global.Object) {
	id, ok := o["id"]
	if ok && id != nil {
		o["objectID"] = id
	}
	delete(o, "id")
	_, err := me.Index.AddObject(o)
	if err != nil {
		logrus.Errorf("On adding object to index %s: %s", id, err)
	}
}
