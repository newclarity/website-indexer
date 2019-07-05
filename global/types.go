package global

import (
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/gocolly/colly"
)

type Object = algoliasearch.Object
type Map = algoliasearch.Map
type Index = algoliasearch.Index

type ObjectMap = map[string]Object

type Strings = []string

type ElemsType = string
type ElemsTypeIndex = map[ElemsType]Strings

type LookupMap = map[string]bool
type LookupIndex = map[ElemsType]LookupMap

type HTMLElement = colly.HTMLElement

type (
	Dir      = string
	Path     = string
	Filepath = string
	Entry    = string
)

type IndexHoster interface {
	Initialize() error
	IndexObject(Object)
}
