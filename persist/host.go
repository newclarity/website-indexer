package persist

import (
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"net/url"
	"website-indexer/global"
)

type Host struct {
	Id     SqlId
	Scheme global.Protocol
	Domain global.Domain
	Port   global.Port
	url    global.Url
}

func NewHost(u global.Url) (sh *Host) {
	u, _ = getRootUrl(u)
	return &Host{
		url: u,
	}
}

func (me *Host) Init() (err error) {
	for range only.Once {
		var uptr *url.URL
		uptr, err = url.Parse(me.url)
		if err != nil {
			err = fmt.Errorf("unable to parse URL '%s'", me.url)
			break
		}
		s := uptr.Scheme
		d := uptr.Hostname()
		p := uptr.Port()
		if p == "" {
			p = "80"
		}
		me.Scheme = global.Protocol(s)
		me.Domain = d
		me.Port = p
		me.url = me.String()
	}
	return err
}

func (me *Host) Initialized() bool {
	return me.Domain != ""
}

func (me *Host) Exists() bool {
	return me.Id != 0
}

func (me *Host) Url() global.Domain {
	return me.String()
}

func (me *Host) String() string {
	for range only.Once {
		if me.url != "" {
			break
		}
		if me.Port == "80" {
			me.url = fmt.Sprintf("%s://%s", me.Scheme, me.Domain)
		} else {
			me.url = fmt.Sprintf("%s://%s:%s", me.Scheme, me.Domain, me.Port)
		}
	}
	return me.url
}
