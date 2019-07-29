package test

import (
	"testing"
	"website-indexer/persist"
)

var resourceTestCases = []struct {
	Name  string
	Url   string
	Path  string
	Match bool
}{
	{"parse", "https://www.mckissock.com/foo/bar#baz", "/foo/bar", true},
}

func TestResources(t *testing.T) {
	var err error
	for _, tc := range resourceTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			r := persist.NewResource(tc.Url)
			err = r.Init(nil)
			if err != nil {
				t.Errorf("unable to initialize resource with URL '%s': %s", tc.Url, err)
			}
			t.Run("initialized", func(t *testing.T) {
				if tc.Match && !r.Initialized() {
					t.Errorf("resource not initialized for URL '%s'", tc.Url)
				}
				if !tc.Match && r.Initialized() {
					t.Errorf("resource initialized for URL '%s'", tc.Url)
				}
			})
			u, err := r.Url()
			if err != nil {
				t.Errorf("unable to initialize URL '%s': %s", tc.Url, err)
			}
			if tc.Match && u != tc.Url {
				t.Errorf("resource parse mismatch: wanted '%s', got '%s'",
					tc.Url,
					u,
				)
			}
			if !tc.Match && u != tc.Url {
				t.Errorf("resource parse mismatch: wanted NOT '%s', got '%s'",
					tc.Url,
					u,
				)
			}
			t.Run("urlpath", func(t *testing.T) {
				if tc.Match && r.UrlPath != tc.Path {
					t.Errorf("urlpath parse mismatch: wanted '%s', got '%s'", r.UrlPath, tc.Path)
				}
				if !tc.Match && r.UrlPath == tc.Path {
					t.Errorf("urlpath parse mismatch: wanted NOT '%s', got '%s'", r.UrlPath, tc.Path)
				}
			})
		})
	}
	CheckErr(t, err)
}
