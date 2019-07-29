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

type Urls = []Url
type UrlPaths = []UrlPath
type Filepaths = []Filepath
type (
	Name = string

	Domain   = string
	Port     = string
	Url      = string
	UrlPath  = string
	Fragment = string
	Dir      = string
	Path     = string
	Filename = string
	Filepath = string
	Entry    = string

	HtmlName  = string
	Content   = string
	Sql       = string
	Tablename = Name
)
type Protocol string

const (
	HttpScheme  Protocol = "http"
	HttpsScheme Protocol = "https"
)

type UnixTime = int64
