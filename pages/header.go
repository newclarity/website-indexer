package pages

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"website-indexer/global"
)

type HeaderMap map[global.HtmlName]Header
type Header = string

func (me HeaderMap) ToJson() (b []byte) {
	b, err := json.Marshal(me)
	if err != nil {
		logrus.Errorf("unable to marshal HeaderMap to JSON")
	}
	return b
}
func (me HeaderMap) ExtractStringMap() global.StringMap {
	sm := make(global.StringMap, 0)
	for n, v := range me {
		sm[n] = v
	}
	return sm
}
