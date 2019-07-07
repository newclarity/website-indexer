package persist

import (
	"encoding/json"
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/jtacoma/uritemplates"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"website-indexer/config"
	"website-indexer/global"
	"website-indexer/pages"
	"website-indexer/util"
)

var template *uritemplates.UriTemplate

func init() {
	template, _ = uritemplates.Parse(JsonFileTemplate)
}
func QueuedUrlPath(cfg *config.Config, urlpath global.UrlPath) (found bool) {
	for range only.Once {

		fp, err := GetSubdirFilepath(cfg, QueuedDir, urlpath)
		if err != nil {
			break
		}
		b := []byte(urlpath)
		if WriteFile(fp, b, urlpath, CannotExist) != nil {
			found = true
			break
		}
	}
	return found
}

func IndexedPage(cfg *config.Config, page *pages.Page) (err error) {
	return Persist(cfg, IndexedDir, page)
}
func ErroredPage(cfg *config.Config, page *pages.Page) (err error) {
	return Persist(cfg, ErroredDir, page)
}
func Persist(cfg *config.Config, subdir global.Dir, page *pages.Page) (err error) {
	for range only.Once {
		var b []byte
		b, err = json.Marshal(page)
		if err != nil {
			err = fmt.Errorf("unable to marshal page '%s': %s", page.UrlPath, err.Error())
			break
		}
		fp, err := GetSubdirFilepath(cfg, subdir, page.UrlPath)
		if err != nil {
			break
		}
		err = WriteFile(fp, b, page.UrlPath, CanExist)
		if err != nil {
			break
		}
	}
	return err
}
func WriteFile(fp global.Filepath, content []byte, urlpath global.UrlPath, exists Existence) (err error) {
	for range only.Once {
		switch exists {
		case CannotExist:
			if util.FileExists(fp) {
				err = fmt.Errorf("file '%s' for URL path '%s' already exists", fp, urlpath)
			}
		case MustExist:
			if !util.FileExists(fp) {
				err = fmt.Errorf("file '%s' for URL path '%s' does not exists", fp, urlpath)
			}
		case CanExist:
			// Do nothing
		}
		if err != nil {
			break
		}
		err = os.MkdirAll(filepath.Dir(fp), os.ModePerm)
		if err != nil {
			logrus.Errorf("unable to make directory '%s'", fp)
			break
		}
		err = ioutil.WriteFile(fp, content, os.ModePerm)
		if err != nil {
			err = fmt.Errorf("unable to write file '%s' for URL path %s: %s",
				fp,
				urlpath,
				err.Error(),
			)
			break
		}
	}
	return err
}

func GetUrlPathFilename(urlpath global.UrlPath) (fn global.Filename, err error) {
	h := pages.NewHash(urlpath)
	fn, err = template.Expand(map[string]interface{}{
		"hash": h,
	})
	if err != nil {
		logrus.Errorf("unable to expand template '%s' with hash='%s'",
			JsonFileTemplate,
			h,
		)
	}
	return fn, err
}

func GetSubdirFilepath(cfg *config.Config, subdir global.Dir, urlpath global.UrlPath) (fp global.Filepath, err error) {
	for range only.Once {
		var fn global.Filename
		fn, err = GetUrlPathFilename(urlpath)
		if err != nil {
			break
		}
		fp = fmt.Sprintf("%s%c%s%c%s",
			cfg.DataDir,
			os.PathSeparator,
			subdir,
			os.PathSeparator,
			fn,
		)
	}
	return fp, err
}
