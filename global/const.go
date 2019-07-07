package global

const AppName = "website-indexer"

const (
	MetaName    = "name"
	MetaContent = "content"
)

const (
	MetaElemsType    ElemsType = "meta"
	LinkElemsType    ElemsType = "links"
	CollectElemsType ElemsType = "collect"
	IgnoreElemsType  ElemsType = "ignore"
)

type ValueType = string

const (
	RelValue  ValueType = "rel"
	NameValue ValueType = "name"
	MetaValue ValueType = "meta"
)
