package cmd

import (
	"github.com/gearboxworks/go-status/only"
	"github.com/spf13/cobra"
	"website-indexer/config"
	"website-indexer/crawler"
)

type CrawlArgs struct {
	Domain     string
	ConfigFile string
}

var crawlargs = CrawlArgs{}

var CrawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Crawl a website to generate an index",
	PreRun: func(cmd *cobra.Command, args []string) {
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		for range only.Once {
			crawler.NewCrawler(config.LoadConfig()).Crawl()
		}
		return err
	},
}

func init() {
	RootCmd.AddCommand(CrawlCmd)
	fs := CrawlCmd.Flags()
	fs.StringVar(&crawlargs.Domain, "domain", "", "Domain to crawl (assumes 'www' too if only 2nd level domain given)")
	fs.StringVar(&crawlargs.ConfigFile, "config", "", "Filepath to a config.json to load for crawling")
}
