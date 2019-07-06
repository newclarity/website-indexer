package config

import (
	"regexp"
	"strings"
	"website-indexer/global"
)

type UrlPatterns []UrlPattern
type UrlPattern string

var templVarRegex = regexp.MustCompile(`^\s*{(.+)}\s*$`)

func (me UrlPatterns) ExtractStringMap(urlpath global.UrlPath) global.StringMap {
	sm := make(global.StringMap, 0)
	urlpath = strings.Trim(urlpath, "/")
	upps := strings.Split(urlpath, "/")
	for _, urlpat := range me {
		patparts := strings.Split(strings.Trim(string(urlpat), "/"), "/")
		if len(upps) != len(patparts) {
			continue
		}
		for i, patpart := range patparts {
			tv := templVarRegex.ReplaceAllString(patpart, "$1")
			sm[tv] = upps[i]
		}
	}
	return sm
}
