package pages

import (
	"strings"
	"website-indexer/global"
)

type AttributeMap map[global.HtmlName]*Attribute
type Attributes []*Attribute
type Attribute struct {
	Name  global.HtmlName
	Value string
}

func NewAttribute(name global.HtmlName, val string) *Attribute {
	return &Attribute{
		Name:  strings.TrimSpace(name),
		Value: strings.TrimSpace(val),
	}
}
