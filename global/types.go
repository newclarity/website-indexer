package global

import (
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/gocolly/colly"
)

type Object = algoliasearch.Object
type Map = algoliasearch.Map
type Index = algoliasearch.Index

type ObjectMap = map[string]Object

type StringMap = map[string]string
type Strings = []string

type ElemsType = string
type ElemsTypeIndex = map[ElemsType]Strings

type LookupMap = map[string]bool
type LookupIndex = map[ElemsType]LookupMap

type HtmlElement = colly.HTMLElement

type (
	Url      = string
	UrlPath  = string
	Dir      = string
	Path     = string
	Filename = string
	Filepath = string
	Entry    = string

	HtmlName = string
	Content  = string
)
