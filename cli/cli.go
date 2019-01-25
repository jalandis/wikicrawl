package main

import (
	"flag"
	"fmt"

	"jalandis.com/wikicrawl"
)

func main() {
	wiki := flag.String("wiki", "wiki_url", "a string")
	flag.Parse()

	c := wikicrawl.NewCrawler(*wiki)
	result := c.Crawl(*wiki)

	for url, set := range result.Links {
		fmt.Println("Visited link: " + url)
		for key, _ := range set.Set {
			fmt.Println("Link referenced from: " + key)
		}
	}

	for key, _ := range result.Broken.Set {
		fmt.Println("Broken link :" + key)
	}
}
