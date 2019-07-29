package mocks

import (
	"encoding/json"
	"github.com/gearboxworks/go-status/only"
	"testing"
	"website-indexer/config"
	"website-indexer/persist/test"
)

func NewConfig(t *testing.T) *config.Config {
	var err error
	var cfg config.Config
	for range only.Once {
		cfg = config.Config{
			ConfigDir: config.Dir,
		}
		err = json.Unmarshal([]byte(config.DefaultJson()), &cfg)
		if err != nil {
			t.Errorf("unable to marshal default config: %s", err)
			break
		}
		cfg.DataDir, err = test.FmtDir(t, "data")
		if err != nil {
			t.Errorf("unable to format DataDir: %s", err)
			break
		}
		cfg.CacheDir, err = test.FmtDir(t, "cache")
		if err != nil {
			t.Errorf("unable to format CacheDir: %s", err)
			break
		}
		cfg.InitLookupIndex()
		cfg.OnErrPause = config.InitialPause
	}
	test.CheckErr(t, err)
	return &cfg
}
