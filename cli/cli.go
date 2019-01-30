package main

import (
	"flag"
	"fmt"

	"jalandis.com/wikicrawl"
)

func main() {
	wiki := flag.String("wiki", "wiki_url", "a string")
	session := flag.String("session", "session", "a string")
	flag.Parse()

	c := wikicrawl.NewCrawler(*wiki, *session)
	result := c.Crawl(*wiki)

	for key, _ := range result.Visited.Set {
		fmt.Println("Visited link: " + key)
	}

	for key, _ := range result.Broken.Set {
		fmt.Println("Broken link :" + key)
	}
}
