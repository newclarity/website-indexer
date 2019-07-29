package test

import (
	"testing"
	"website-indexer/persist"
)

var hostTestCases = []struct {
	Name    string
	Url     string
	HostUrl string
	Match   bool
}{
	{"strip-urlpath", "https://www.mckissock.com/foo/bar", "https://www.mckissock.com", true},
	{"self_http", "http://www.mckissock.com", "http://www.mckissock.com", true},
	{"self_https", "https://www.mckissock.com", "https://www.mckissock.com", true},
	{"no-protocol-change", "http://www.mckissock.com", "https://www.mckissock.com", false},
}

func TestHosts(t *testing.T) {
	var err error
	for _, tc := range hostTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			h := persist.NewHost(tc.Url)
			err = h.Init()
			if err != nil {
				t.Errorf("host failed to initialize with URL '%s': %s", tc.Url, err)
			}
			t.Run("initialized", func(t *testing.T) {
				if tc.Match && !h.Initialized() {
					t.Errorf("host not initialized for URL '%s'", tc.Url)
				}
			})
			switch tc.Match {
			case true:
				if h.Url() != tc.HostUrl {
					t.Errorf("mismatch: wanted '%s', got '%s'",
						tc.HostUrl,
						h.Url(),
					)
				}
			case false:
				if h.Url() == tc.HostUrl {
					t.Errorf("mismatch: wanted NOT '%s', got '%s'",
						tc.HostUrl,
						h.Url(),
					)
				}
			}
		})
	}
	CheckErr(t, err)
}
