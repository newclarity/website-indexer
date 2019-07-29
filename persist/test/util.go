package test

import (
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"os"
	"testing"
	"website-indexer/global"
	"website-indexer/util"
)

func FmtDir(t *testing.T, p global.Path) (d global.Dir, err error) {
	for range only.Once {
		d, err = os.Getwd()
		if err != nil {
			t.Error("unable to get working directory")
			break
		}
		d = fmt.Sprintf("%s%cfixtures%c%s",
			d,
			os.PathSeparator,
			os.PathSeparator,
			p,
		)
		if !util.DirExists(d) {
			err = os.Mkdir(d, os.ModePerm)
			if err != nil {
				t.Errorf("unable to make directory '%s'", d)
			}
		}
	}
	return d, err
}

func CheckErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
