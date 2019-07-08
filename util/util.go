package util

import (
	"github.com/gearboxworks/go-status/only"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func SecondsDuration(n float64) time.Duration {
	d, _ := time.ParseDuration(strconv.FormatFloat(n, 'f', 3, 64) + "s")
	return d
}

func StripWhitespace(str string) string {
	ctr := 0
	f := func(r rune) rune {
		for range only.Once {
			if !unicode.IsSpace(r) {
				break
			}
			r = ' '
			if ctr == 1 {
				r = -1
				ctr = 0
				break
			}
			ctr++
		}
		return r
	}
	return strings.Map(f, str)
}

func noop(i ...interface{}) interface{} { return i }
