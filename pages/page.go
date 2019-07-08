package pages

import (
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"strings"
	"website-indexer/global"
)

type HtmlBody global.Strings

func (me HtmlBody) String() string {
	return strings.Join(me, "\n")
}

type Map map[Hash]*Page
type Pages []*Page
type Page struct {
	Id          *Hash       `json:"id"`
	HeaderMap   HeaderMap   `json:"header_map"`
	Title       string      `json:"title"`
	Body        HtmlBody    `json:"body"`
	ElementsMap ElementsMap `json:"elements_map"`
	PropertyMap PropertyMap `json:"property_map"`

	*Url `json:"url"`
}

func NewPage(url global.Url, referer ...global.Url) *Page {
	if len(referer) == 0 {
		referer = []string{""}
	}
	u := NewUrl(url, referer[0])
	return &Page{
		Id:          u.Hash(),
		Url:         u,
		Body:        make(HtmlBody, 0),
		ElementsMap: make(ElementsMap, 0),
		HeaderMap:   make(HeaderMap, 0),
	}
}

func (me *Page) AddHeader(name global.HtmlName, value global.Content) {
	me.HeaderMap[name] = strings.TrimSpace(value)
}

func (me *Page) AppendElement(ele *global.HtmlElement) {
	e := NewElement(ele)
	for range only.Once {
		me.Body = append(me.Body, e.GetHtml())
		var ok bool
		if _, ok = me.ElementsMap[e.Name]; ok {
			me.ElementsMap[e.Name] = make(Elements, 0)
			break
		}
		me.ElementsMap[e.Name] = append(me.ElementsMap[e.Name], e)
	}
}

func (me *Page) GetHash() Hash {
	return *me.Id
}

func (me *Page) GetHashString() string {
	return string(fmt.Sprintf("%x", *me.Id))
}
