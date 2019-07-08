package config

import (
	"regexp"
	"strings"
	"website-indexer/global"
)

type UrlPatterns []UrlPattern
type UrlPattern string

var templVarRegex = regexp.MustCompile(`^\s*{(.+)}\s*$`)

func (me UrlPatterns) ExtractStringMap(url global.Url) global.StringMap {
	sm := make(global.StringMap, 0)
	url = strings.Trim(url, "/")
	urls := strings.Split(url, "/")
	for _, urlpat := range me {
		patparts := strings.Split(strings.Trim(string(urlpat), "/"), "/")
		if len(urls) != len(patparts) {
			continue
		}
		for i, patpart := range patparts {
			tv := templVarRegex.ReplaceAllString(patpart, "$1")
			sm[tv] = urls[i]
		}
	}
	return sm
}
