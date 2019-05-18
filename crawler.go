package main

import (
	"github.com/gearboxworks/go-status/only"
	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
	"time"
)

type Crawler struct {
	*Config
	Collector *colly.Collector
}

func NewCrawler(cfg *Config) (c *Crawler) {
	for range only.Once {
		c = &Crawler{
			Config: cfg,
			Collector: colly.NewCollector(
				colly.AllowedDomains("www."+cfg.Domain, cfg.Domain),
				//colly.Async(true),
			),
		}
		err := c.Collector.Limit(&colly.LimitRule{
			Delay: 1 * time.Second,
		})
		if err == nil {
			break
		}
		if err.Error() == "No pattern defined in LimitRule" {
			break
		}
		logrus.Fatalf("Unable to set crawl delay: %s", err)
	}
	return c
}
