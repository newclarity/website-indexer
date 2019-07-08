package pages

import (
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"website-indexer/global"
)

type Url struct {
	url     global.Url
	referer global.Url
	clean   bool
	hash    *Hash
}

func NewUrl(url string, referer ...global.Url) (u *Url) {
	for range only.Once {
		if len(referer) == 0 {
			referer = []string{""}
		}
		if len(url) != 0 && url[0] == '#' {
			break
		}

		u = &Url{
			url:     url,
			referer: referer[0],
		}
	}
	return u
}

func (me *Url) Hash() *Hash {
	if me.hash == nil {
		me.hash = NewHash(me.GetUrl())
	}
	return me.hash
}

var rootDomainRegex *regexp.Regexp

func init() {
	rootDomainRegex = regexp.MustCompile(`^(https?://.+?)/?`)
}

func (me *Url) GetRootDomain() global.Url {
	return rootDomainRegex.ReplaceAllString(me.GetUrl(), "$1")
}

func (me *Url) String() string {
	return me.GetUrl()
}

func (me *Url) GetUrl() string {
	for range only.Once {
		if me.clean {
			break
		}
		me.clean = true
		r := GetRelativeness(me.url)
		if r == AbsoluteUrl {
			break
		}
		me.url = CleanPath(me.url)
		switch r {
		case RootRelativeUrl:
			me.url = fmt.Sprintf("%s%c%s",
				me.GetRootDomain(),
				os.PathSeparator,
				me.url,
			)
		case PathRelativeUrl:
			me.url = fmt.Sprintf("%s%c%s",
				filepath.Dir(me.url),
				os.PathSeparator,
				me.url,
			)
		}
	}
	return me.url
}

type UrlRelativeness int

const (
	AbsoluteUrl UrlRelativeness = iota
	RootRelativeUrl
	PathRelativeUrl
)

var absoluteRegex *regexp.Regexp

func init() {
	absoluteRegex = regexp.MustCompile(`^https?://`)
}
func IsAbsoluteUrl(url global.Url) bool {
	return absoluteRegex.MatchString(url)
}

func GetRelativeness(url global.Url) (ur UrlRelativeness) {
	switch {
	case len(url) != 0 && url[0] == '/':
		ur = RootRelativeUrl
	case IsAbsoluteUrl(url):
		ur = AbsoluteUrl
	default:
		ur = PathRelativeUrl
	}
	return ur
}

func CleanPath(urlpath global.UrlPath) global.UrlPath {
	for range only.Once {
		if len(urlpath) == 0 {
			urlpath = "/"
			break
		}
		if strings.Contains(urlpath, "?") {
			break
		}
		if urlpath[0] == '#' {
			urlpath = ""
			break
		}
		if urlpath[:4] == "tel:" {
			urlpath = ""
			break
		}
		//if u[len(u)-1] != '/' {
		//	u += "/"
		//	break
		//}
		if urlpath[len(urlpath)-2:] == "//" {
			urlpath = urlpath[:len(urlpath)-2]
			break
		}
	}
	return urlpath
}
