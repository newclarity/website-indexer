package persist

import (
	"website-indexer/global"
)

type Visited struct {
	Id           SqlId
	ResourceHash Hash
	Timestamp    global.UnixTime
	Headers      string
	Body         string
	Cookies      string
}
