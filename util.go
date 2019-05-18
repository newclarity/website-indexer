package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"golang.org/x/net/html"
	"strings"
	"unicode"
)

func getHtml(e *HTMLElement) (h string) {
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

func getPage(om ObjectMap, e *HTMLElement) string {
	hash := sha256.Sum256([]byte(e.Request.URL.Path))
	p := hashToString(hash)
	if _, ok := om[p]; !ok {
		om[p] = make(Object, 0)
		om[p]["urlpath"] = e.Request.URL.Path
	}
	return p
}

func appendHtml(om ObjectMap, e *HTMLElement, k string) ObjectMap {
	p := getPage(om, e)
	_, ok := om[p][k]
	if !ok {
		om[p][k] = make([]string, 0)
	}
	om[p][k] = append(om[p][k].([]string), getHtml(e))
	return om
}

func stripWhitespace(str string) string {
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

func hashToString(hash [sha256.Size]byte) string {
	h := sha256.New()
	h.Write(hash[:])
	bs := h.Sum(nil)
	sh := string(fmt.Sprintf("%x", bs))
	return sh
}

func cleanurl(u string) string {
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
