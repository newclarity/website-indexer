package main

import (
	"encoding/json"
	"github.com/gearboxworks/go-status/only"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type Config struct {
	AppId         string         `json:"app_id"`
	ApiKey        string         `json:"api_key"`
	Index         string         `json:"index"`
	Domain        string         `json:"domain"`
	SearchAttrs   Strings        `json:"search_attrs"`
	ElementsIndex ElemsTypeIndex `json:"elements"`
	LookupIndex   LookupIndex    `json:"ignore"`
}

func LoadConfig() *Config {
	cfg := Config{}
	b, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		panic(err)
	}
	cfg.LookupIndex = make(LookupIndex, len(cfg.ElementsIndex))
	for typ, es := range cfg.ElementsIndex {
		lookup := make(LookupMap, len(es))
		for _, e := range es {
			lookup[e] = true
		}
		cfg.LookupIndex[typ] = lookup
	}
	cfg.ElementsIndex = nil
	return &cfg
}

func (me *Config) hasElementName(ele *HTMLElement, typ ElemsType) (ok bool) {
	return me.hasElement(NameValue, ele, typ)
}

func (me *Config) hasElementRel(ele *HTMLElement, typ ElemsType) (ok bool) {
	return me.hasElement(RelValue, ele, typ)
}

func (me *Config) hasElementMeta(ele *HTMLElement) (ok bool) {
	return me.hasElement(MetaValue, ele, MetaElemsType)
}

func (me *Config) hasElement(v ValueType, ele *HTMLElement, typ ElemsType) (ok bool) {
	for range only.Once {
		var m LookupMap
		m, ok = me.LookupIndex[typ]
		if !ok {
			logrus.Fatalf("Invalid elements type in config.json: %s", typ)
		}
		switch v {
		case NameValue:
			_, ok = m[ele.Name]
		case RelValue:
			_, ok = m[ele.Attr(RelValue)]
		case MetaValue:
			_, ok = m[ele.Attr(MetaName)]
		default:
			logrus.Fatalf("Invalid value type: %s", v)
		}
	}
	return ok
}
