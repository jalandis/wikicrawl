package wikicrawl

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/html"
)

// Wikimedia namespaces to ignore.
var ignore = []string{"User:", "Talk:", "Help:", "Help_talk:"}

// Results of crawling wiki.
//  1. Links: Map of links to pages referencing them.
//  2. Broken: List of Broken links.
type CrawlResult struct {
	Links map[Link]LinkSet
	Broken LinkSet
}

// Crawler type holds state and methods for exploring a wiki.
type Crawler struct {
	base   *url.URL
	Client *http.Client
}

func parseUrlOrPanic(link Link) *url.URL {
	result, err := url.Parse(link)
	if err != nil {
		panic(err)
	}

	return result
}

func parseQueryOrPanic(query string) url.Values {
	result, err := url.ParseQuery(query)
	if err != nil {
		panic(err)
	}

	return result
}

// Simple constructor for Crawler type.
func NewCrawler(base Link) *Crawler {
	c := new(Crawler)
	c.base = parseUrlOrPanic(base)
	c.Client = &http.Client{
		Timeout: time.Second * 10,
	}

	return c
}

// Crawls all valid links that can be found from the initial url.
func (c *Crawler) Crawl(source Link) *CrawlResult {
	result := &CrawlResult{Links: make(map[Link]LinkSet), Broken: NewLinkSet()}
	c.followLink(source, result)

	return result
}

func (c *Crawler) followLink(source Link, result *CrawlResult) {
	log.WithFields(log.Fields{"source": source}).Debug("Crawling new url")

	resp, err := c.Client.Get(source)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Warn("GET returned with non 200 response")
		result.Broken.Add(source)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.WithFields(log.Fields{
			"source": source,
			"status": resp.Status,
		}).Warn("GET returned with non 200 response")
		result.Broken.Add(source)
		return
	}

	// TODO: Handle redirection properly to prevent duplicate HTTP requests.
	if source != resp.Request.URL.String() {
		log.WithFields(log.Fields{
			"requested": source,
			"redirect":  resp.Request.URL,
		}).Warn("Redirect detected.")
	}

	for _, raw := range ParseLinks(resp.Body) {
		href := NormalizeUrl(parseUrlOrPanic(raw), c.base).String()
		if _, ok := result.Links[href]; !ok {
			result.Links[href] = NewLinkSet()
			if c.ValidateLink(href) {
				c.followLink(href, result)
			}
		}

		result.Links[href].Add(source)
	}
}

// Validates if link should be followed.
//
//  1. Only crawls internal links.
//  2. Skips trivial Wikimedia namespaces.
func (c *Crawler) ValidateLink(link Link) bool {
	if !strings.Contains(link, c.base.String()) {
		return false
	}

	title := WikiPageTitle(parseUrlOrPanic(link))
	for _, trivial := range ignore {
		if strings.HasPrefix(title, trivial) {
			return false
		}
	}

	return true
}

// Parse WikiMedia page title with namespace.
// WikiMedia short url's not supported.
func WikiPageTitle(link *url.URL) string {
	query := parseQueryOrPanic(link.RawQuery)
	if title, found := query["title"]; found {
		return title[0]
	}

	return ""
}

// Normalize a url to facilitate comparison.
//
//  1. Resolve url from known base (/relative => http://base/relative)
//  2. Cleanup query by filtering unnecessary parameters
//  3. Remove any URL fragment (#junk)
//  4. Force protocol to match base
//  5. Unify case of host and protocol
func NormalizeUrl(link *url.URL, base *url.URL) *url.URL {
	clean := base.ResolveReference(link)

	if title := WikiPageTitle(clean); len(title) != 0 {
		clean.RawQuery = url.Values{"title": []string{title}}.Encode()
	}

	clean.Fragment = ""

	clean.Scheme = strings.ToLower(base.Scheme)
	clean.Host = strings.ToLower(clean.Host)

	return clean
}

// Parses HTML and returns a list of all href values found.
func ParseLinks(reader io.Reader) []Link {
	links := make([]Link, 0)
	z := html.NewTokenizer(reader)
	for {
		tokenType := z.Next()

		switch {
		case tokenType == html.ErrorToken:
			return links
		case tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken:
			token := z.Token()

			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
						break
					}
				}
			}
		}
	}
}
