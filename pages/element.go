package pages

import (
	"bytes"
	"encoding/json"
	"github.com/gearboxworks/go-status/only"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"website-indexer/global"
)

type ElementsMap map[global.HtmlName]Elements
type Elements []*Element
type Element struct {
	*global.HtmlElement
	AttributeMap AttributeMap
}

func NewElement(ele *global.HtmlElement) *Element {
	return &Element{
		HtmlElement:  ele,
		AttributeMap: make(AttributeMap, 0),
	}
}

func (me *Element) GetHtml() (h string) {
	for range only.Once {
		if me == nil {
			break
		}
		if me.DOM == nil {
			break
		}
		if len(me.DOM.Nodes) == 0 {
			break
		}
		n := me.DOM.Nodes[0]
		var buf bytes.Buffer
		_ = html.Render(&buf, n)
		h = buf.String()
	}
	return h
}

func (me *Element) AddAttribute(attr *Attribute) *Attribute {
	me.AttributeMap[attr.Name] = attr
	return attr
}

func (me ElementsMap) ToJson() (b []byte) {
	b, err := json.Marshal(me)
	if err != nil {
		logrus.Errorf("unable to marshal ElementsMap to JSON")
	}
	return b
}

func (me ElementsMap) ExtractStringMap() global.StringMap {
	sm := make(global.StringMap, 0)
	for n, vs := range me {
		s := ""
		for _, v := range vs {
			s += v.GetHtml()
		}
		sm[n] = s
	}
	return sm
}
