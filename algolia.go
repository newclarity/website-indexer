package main

import (
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/sirupsen/logrus"
)

func (me *Crawler) makeIndex() (index Index) {
	cfg := me.Config
	client := algoliasearch.NewClient(cfg.AppId, cfg.ApiKey)
	index = client.InitIndex(cfg.Index)
	settings := algoliasearch.Map{
		"searchableAttributes": cfg.SearchAttrs,
	}
	_, err := index.SetSettings(settings)
	if err != nil {
		logrus.Fatalf("Unable to set index settings: %s", err)
	}
	return index
}

func indexObject(index Index, o Object) {
	id := o["id"]
	o["objectID"] = id
	delete(o, "id")
	_, err := index.AddObject(o)
	if err != nil {
		logrus.Errorf("On adding object to index %s: %s", id, err)
	}
}
