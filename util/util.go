package util

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"golang.org/x/net/html"
	"strings"
	"unicode"
	"website-indexer/global"
)

func GetHtml(e *global.HTMLElement) (h string) {
	for range only.Once {
		if e == nil {
			break
		}
		if e.DOM == nil {
			break
		}
		if len(e.DOM.Nodes) == 0 {
			break
		}
		n := e.DOM.Nodes[0]
		var buf bytes.Buffer
		_ = html.Render(&buf, n)
		h = buf.String()
	}
	return h
}

func GetPage(om global.ObjectMap, e *global.HTMLElement) string {
	hash := sha256.Sum256([]byte(e.Request.URL.Path))
	p := HashToString(hash)
	if _, ok := om[p]; !ok {
		om[p] = make(global.Object, 0)
		om[p]["urlpath"] = e.Request.URL.Path
	}
	return p
}

func AppendHtml(om global.ObjectMap, e *global.HTMLElement, k string) global.ObjectMap {
	p := GetPage(om, e)
	_, ok := om[p][k]
	if !ok {
		om[p][k] = make([]string, 0)
	}
	om[p][k] = append(om[p][k].([]string), GetHtml(e))
	return om
}

func StripWhitespace(str string) string {
	ctr := 0
	f := func(r rune) rune {
		for range only.Once {
			if !unicode.IsSpace(r) {
				break
			}
			r = ' '
			if ctr == 1 {
				r = -1
				ctr = 0
				break
			}
			ctr++
		}
		return r
	}
	return strings.Map(f, str)
}

func HashToString(hash [sha256.Size]byte) string {
	h := sha256.New()
	h.Write(hash[:])
	bs := h.Sum(nil)
	sh := string(fmt.Sprintf("%x", bs))
	return sh
}

func Cleanurl(u string) string {
	for range only.Once {
		if len(u) == 0 {
			u = "/"
			break
		}
		if strings.Contains(u, "?") {
			break
		}
		if u[0] == '#' {
			u = ""
			break
		}
		if u[:4] == "tel:" {
			u = ""
			break
		}
		if u[len(u)-1] != '/' {
			u += "/"
			break
		}
		if u[len(u)-2:] == "//" {
			u = u[:len(u)-2]
			break
		}
	}
	return u
}

func noop(i ...interface{}) interface{} { return i }
