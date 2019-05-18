package main

// @see https://benjamincongdon.me/blog/2018/03/01/Scraping-the-Web-in-Golang-with-Colly-and-Goquery/

func main() {
	NewCrawler(LoadConfig()).Crawl()
}
