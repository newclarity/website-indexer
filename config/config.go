package config

import (
	"encoding/json"
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
	"website-indexer/global"
	"website-indexer/util"
)

const (
	Dir      = "~/.config/website-indexer"
	Filename = "config.json"
)
const (
	InitialPause  time.Duration = 1 * time.Second
	PauseIncrease float64       = 1.1
)
const (
	DefaultRevisit  = "1d"
	TimeoutErr      = "net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)"
	UnsuccessfulErr = "All hosts have been contacted unsuccessfully"
)

type DurationString = string

type Config struct {
	AppId           string                `json:"app_id"`
	ApiKey          string                `json:"api_key"`
	IndexName       string                `json:"index"`
	Domain          string                `json:"domain"`
	SearchAttrs     global.Strings        `json:"search_attrs"`
	UrlPatterns     UrlPatterns           `json:"url_patterns"`
	ElementsIndex   global.ElemsTypeIndex `json:"elements"`
	LookupIndex     global.LookupIndex    `json:"ignore"`
	DataDir         global.Dir            `json:"data_dir"`
	CacheDir        global.Dir            `json:"cache_dir"`
	Revisit         DurationString        `json:"revisit"`
	RevisitDuration time.Duration         `json:"-"`
	HomeDir         global.Dir            `json:"-"`
	ConfigDir       global.Dir            `json:"-"`
	OnErrPause      time.Duration         `json:"-"`
}

func LoadConfig() *Config {
	cfg := Config{
		ConfigDir: Dir,
	}
	cfg.HomeDir = getHomeDir()
	cfg.ConfigDir = cfg.expandConfigDir()
	b := cfg.loadConfigFile()
	err := json.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatalf("Config file '%s' cannot be processed. It is likely invalid JSON or is not using the correct schema: %s.",
			cfg.GetFilepath(),
			err,
		)
	}
	if cfg.DataDir == "" {
		cfg.DataDir = fmt.Sprintf("%s%c%s",
			cfg.HomeDir,
			os.PathSeparator,
			global.AppName,
		)
	} else {
		cfg.DataDir = cfg.maybeExpandDataDir()
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = getCacheDir()
	}
	if cfg.Revisit == "" {
		cfg.Revisit = DefaultRevisit
	}
	cfg.RevisitDuration, err = time.ParseDuration(cfg.Revisit)
	if err != nil {
		logrus.Errorf("unable to parse revisit duration '%s': %s", cfg.Revisit, err)
		cfg.RevisitDuration, _ = time.ParseDuration(DefaultRevisit)
	}
	cfg.InitLookupIndex()
	cfg.OnErrPause = InitialPause
	return &cfg
}

func (me *Config) InitLookupIndex() {
	me.LookupIndex = make(global.LookupIndex, len(me.ElementsIndex))
	for typ, es := range me.ElementsIndex {
		lookup := make(global.LookupMap, len(es))
		for _, e := range es {
			lookup[e] = true
		}
		me.LookupIndex[typ] = lookup
	}
	me.ElementsIndex = nil
}

func (me *Config) GetFilepath() global.Filepath {
	return fmt.Sprintf("%s%c%s",
		me.ConfigDir,
		os.PathSeparator,
		Filename,
	)
}

var homeDirRegexp *regexp.Regexp

func init() {
	homeDirRegexp = regexp.MustCompile(`^~/`)
}
func (me *Config) HasElementName(ele *global.HtmlElement, typ global.ElemsType) (ok bool) {
	return me.HasElement(global.NameValue, ele, typ)
}

func (me *Config) HasElementRel(ele *global.HtmlElement, typ global.ElemsType) (ok bool) {
	return me.HasElement(global.RelValue, ele, typ)
}

func (me *Config) HasElementMeta(ele *global.HtmlElement) (ok bool) {
	return me.HasElement(global.MetaValue, ele, global.MetaElemsType)
}

func (me *Config) HasElement(v global.ValueType, ele *global.HtmlElement, typ global.ElemsType) (ok bool) {
	for range only.Once {
		var m global.LookupMap
		m, ok = me.LookupIndex[typ]
		if !ok {
			logrus.Fatalf("Invalid elements type in config.json: %s", typ)
		}
		switch v {
		case global.NameValue:
			_, ok = m[ele.Name]
		case global.RelValue:
			_, ok = m[ele.Attr(global.RelValue)]
		case global.MetaValue:
			_, ok = m[ele.Attr(global.MetaName)]
		default:
			logrus.Fatalf("Invalid value type: %s", v)
		}
	}
	return ok
}

func (me *Config) maybeExpandDataDir() global.Filepath {
	hd := fmt.Sprintf("%s%c", me.HomeDir, os.PathSeparator)
	dd := homeDirRegexp.ReplaceAllString(me.DataDir, hd)
	return dd
}

func (me *Config) expandConfigDir() global.Filepath {
	hd := fmt.Sprintf("%s%c", me.HomeDir, os.PathSeparator)
	cd := homeDirRegexp.ReplaceAllString(me.ConfigDir, hd)
	return cd
}

func (me *Config) loadConfigFile() (b []byte) {
	var isnew bool
	var err error
	var f *os.File
	fp := me.GetFilepath()
	for range only.Once {
		if util.FileExists(fp) {
			c, err := ioutil.ReadFile(fp)
			if err != nil {
				log.Fatalf("Config file '%s' exists but cannot be read: %s.", fp, err)
			}
			if string(c) == DefaultJson() {
				isnew = true
			}
			b = []byte(c)
			break
		}
		if !util.DirExists(me.ConfigDir) {
			err := os.MkdirAll(me.ConfigDir, os.ModePerm)
			if err != nil {
				log.Fatalf("Cannot make directory '%s'; Check permissions: %s.", me.ConfigDir, err)
			}
		}
		f, err = os.Create(fp)
		if err != nil {
			log.Fatalf("Cannot create config file '%s'; Check permissions: %s.", fp, err)
		}
		var n int
		dc := DefaultJson()
		n, err = f.WriteString(dc)
		if err != nil || n != len(dc) {
			log.Fatalf("Cannot create config file '%s'; Check permissions: %s.", fp, err)
		}
		var size int64
		size, err = f.Seek(0, 2)
		if err != nil || size != int64(len(dc)) {
			log.Fatalf("Cannot determine length of config file just written '%s'; Check permissions: %s", fp, err)
		}
		var n64 int64
		n64, err = f.Seek(0, 0)
		if err != nil || n64 != 0 {
			log.Fatalf("Cannot reset config file just written '%s'; Check permissions: %s.", fp, err)
		}
		b, err = ioutil.ReadAll(f)
		if err != nil || string(b) != dc {
			log.Fatalf("Config read does not equal config file just written '%s': %s.", fp, err)
		}
		isnew = true
	}
	closeFile(f)
	if isnew {
		fmt.Printf("\nYour config file '%s' is newly initialized.\nPlease EDIT to configure appropriate settings.\n", fp)
		os.Exit(1)
	}
	return b
}

func (me *Config) OnFailedVisit(err error, url global.Url, descr string, nosleep ...bool) {
	msg := err.Error()
	nointernet := true
	for range only.Once {
		if strings.HasSuffix(msg, TimeoutErr) {
			break
		}
		if strings.HasSuffix(msg, UnsuccessfulErr) {
			break
		}
		nointernet = false
	}
	for range only.Once {
		if nointernet {
			if len(nosleep) == 0 {
				me.OnErrPause = util.SecondsDuration(me.OnErrPause.Seconds() * PauseIncrease)
				fmt.Printf("\nInternet connection unavailable; pausing %d seconds...",
					int(me.OnErrPause.Seconds()),
				)
				time.Sleep(me.OnErrPause)
				break
			}
			fmt.Printf("\nInternet connection unavailable; terminating.")
			break
		}
		fmt.Print("\n")
		logrus.Errorf("On %s to %s: %s",
			descr,
			url,
			err.Error(),
		)
	}
}

func closeFile(f *os.File) {
	_ = f.Close()
}

func getCacheDir() (cd global.Dir) {
	cd, err := os.UserCacheDir()
	if err != nil {
		if runtime.GOOS == "windows" {
			cd = "C:\\tmp"
		} else {
			cd = "/tmp"
		}
	}
	return fmt.Sprintf("%s%c%s",
		cd,
		os.PathSeparator,
		global.AppName,
	)
}

func getHomeDir() (hd global.Dir) {
	hd, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("User home directory not found. Set environment variable HOME and retry.")
		os.Exit(1)
	}
	return hd
}

func DefaultJson() string {
	return `{
   "index_host": "algolia_or_elastic"
   "app_id": "ABC123XYZ9",
   "api_key": "abcdef123456789abcdef123456789ab",
   "index": "text_Example",
   "domain": "example.com",
   "data_dir": "~/Desktop/website-indexer",
   "revisit": "1d",
   "search_attrs": [
      "article",
      "body",
      "h1",
      "h2",
      "h3",
      "li",
      "title"
   ],
   "url_patterns": [
      "/{profession}/{edu_type}/{state}/{course}"
   ],
   "elements": {
      "meta": [
         "description"
      ],
      "links": [
         "shortlink",
         "canonical"
      ],
      "collect": [
         "article",
         "b",
         "blockquote",
         "button",
         "em",
         "h1",
         "h2",
         "h3",
         "h4",
         "h5",
         "h6",
         "header",
         "i",
         "li",
         "main",
         "label",
         "nav",
         "section",
         "strong",
         "figcaption"
      ],
      "ignore": [
         "area",
         "article",
         "aside",
         "b",
         "blockquote",
         "br",
         "button",
         "circle",
         "center",
         "clipPath",
         "colgroup",
         "defs",
         "desc",
         "div",
         "em",
         "figure",
         "figcaption",
         "font",
         "footer",
         "form",
         "g",
         "h1",
         "h2",
         "h3",
         "h4",
         "h5",
         "h6",
         "head",
         "header",
         "hr",
         "html",
         "i",
         "image",
         "img",
         "input",
         "label",
         "li",
         "main",
         "map",
         "nav",
         "noscript",
         "o:p",
         "ol",
         "option",
         "p",
         "pre",
         "path",
         "rect",
         "script",
         "section",
         "select",
         "span",
         "strong",
         "style",
         "sup",
         "svg",
         "symbol",
         "table",
         "tbody",
         "td",
         "text",
         "textarea",
         "time",
         "tr",
         "u",
         "ul",
         "use"
      ]
   }
}`
}
