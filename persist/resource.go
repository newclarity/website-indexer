package persist

import (
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"net/url"
	"strings"
	"website-indexer/global"
)

type Resource struct {
	Id       SqlId
	Hash     Hash
	UrlPath  global.UrlPath
	Fragment global.Fragment
	url      global.Url
	host     *Host
}

func NewResource(u global.Url) *Resource {
	return &Resource{
		url: u,
	}
}

//type rawResource Resource
//
//func(me *Resource) UnmarshalJSON(b []byte) (err error) {
//	for range only.Once {
//		rr := rawResource{}
//		rr = rawResource(*me)
//		err = json.Unmarshal(b,&rr)
//		if err != nil {
//			continue
//		}
//
//	}
//	return err
//}

func (me *Resource) Init(h *Host) (err error) {
	for range only.Once {
		if h == nil {
			h = NewHost(me.url)
		}
		err = h.Init()
		if err != nil {
			break
		}
		var uptr *url.URL
		uptr, err = parseUrl(me.url)
		if err != nil {
			break
		}
		me.Hash = NewHash(me.url)
		me.UrlPath = getParsedPath(*uptr)
		me.Fragment = uptr.Fragment
		me.host = h
	}
	return err
}

func (me *Resource) Initialized() bool {
	return me.Hash != 0 && me.host != nil && me.url != ""
}

func (me *Resource) Host() *Host {
	return me.host
}

func (me *Resource) String() string {
	s, _ := me.Url()
	return s
}

func (me *Resource) Url() (u global.Url, err error) {
	for range only.Once {
		if me.host == nil {
			sh := NewHost(me.url)
			err = sh.Init()
			if err != nil {
				break
			}
			me.host = sh
		}
		h := me.host
		if h.Domain == "" {
			err = ErrNonIndexableUrl
			break
		}
		if h.Domain == "blank" {
			err = ErrNonIndexableUrl
			break
		}
		u = fmt.Sprintf("%s/%s",
			strings.TrimRight(h.String(), "/"),
			strings.TrimLeft(me.UrlPath, "/"),
		)
		if me.Fragment != "" {
			u = fmt.Sprintf("%s#%s", u, me.Fragment)
		}
	}
	return u, err
}
