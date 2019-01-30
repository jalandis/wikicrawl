package wikicrawl

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/html"
)

// Wikimedia namespaces to ignore.
var ignore = []string{
	"User:", "User_talk:",
	"Help:", "Help_talk:",
	"Talk:", "File_talk:", "Category_talk:",
	"Create_Article", "Special:",
}

// Results of crawling wiki.
//  1. Visited: List of visited links.
//  2. Broken: List of Broken links.
type CrawlResult struct {
	Visited LinkSet
	Broken  LinkSet
}

// Crawler type holds state and methods for exploring a wiki.
type Crawler struct {
	base   *url.URL
	Client *http.Client
}

// Simple constructor for Crawler type.
func NewCrawler(base Link, session string) *Crawler {
	c := new(Crawler)
	result, err := url.Parse(base)
	if err != nil {
		panic(err)
	}
	c.base = result

	jar, _ := cookiejar.New(nil)
	cookie := &http.Cookie{
		Name:   "wikidb2_is__session",
		Value:  session,
		Path:   "/",
		Domain: c.base.Host,
	}
	cookies := []*http.Cookie{cookie}
	jar.SetCookies(c.base, cookies)

	c.Client = &http.Client{
		Timeout: time.Second * 10,
		Jar:     jar,
	}

	return c
}

// Crawls all valid links that can be found from the initial url.
func (c *Crawler) Crawl(source Link) *CrawlResult {
	queue := NewWorkQueue(*c, 1000)
	queue.Start(10)
	queue.AddWork(source)
	queue.Wait()
	return queue.Result
}

func (c *Crawler) FollowLink(source Link, queue *WorkQueue) {

	// Avoid duplicate visits.
	if ok := queue.Result.Visited.Add(source); !ok {
		return
	}

	log.WithFields(log.Fields{"source": source}).Debug("Crawling new url")

	resp, err := c.Client.Get(source)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Warn("GET returned with error")
		queue.Result.Broken.Add(source)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.WithFields(log.Fields{
			"source": source,
			"status": resp.Status,
		}).Warn("GET returned with non 200 response")
		queue.Result.Broken.Add(source)
		return
	}

	if source != resp.Request.URL.String() {
		log.WithFields(log.Fields{
			"requested": source,
			"redirect":  resp.Request.URL,
		}).Warn("Redirect detected.")

		if ok := queue.Result.Visited.Add(resp.Request.URL.String()); !ok {
			return
		}
	}

	for raw := range ParseLinks(resp.Body).Set {
		result, err := url.Parse(raw)
		if err != nil {
			queue.Result.Broken.Add(raw)
			continue
		}

		href := NormalizeUrl(result, c.base)
		if c.ValidateLink(href) && !queue.Result.Visited.Contains(href.String()) {
			queue.AddWork(href.String())
		} else {
			log.WithFields(log.Fields{"href": href}).Debug("Skipping link.")
		}
	}
}

// Validates if link should be followed.
//
//  1. Only crawls internal links.
//  2. Skips trivial Wikimedia namespaces.
func (c *Crawler) ValidateLink(link *url.URL) bool {
	if !strings.Contains(link.String(), c.base.String()) {
		return false
	}

	if title := WikiPageTitle(link); len(title) > 0 {
		for _, trivial := range ignore {
			if strings.HasPrefix(title, trivial) {
				return false
			}
		}
	}

	return true
}

// Parse WikiMedia page title with namespace.
// WikiMedia short url's not supported.
func WikiPageTitle(link *url.URL) string {
	query, err := url.ParseQuery(link.RawQuery)
	if err != nil {
		panic(err)
	}

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

	log.WithFields(log.Fields{
		"base":     base.String(),
		"original": link.String(),
		"cleaned":  clean.String(),
	}).Debug("Normalized URL.")

	return clean
}

// Parses HTML and returns a list of all href values found.
func ParseLinks(reader io.Reader) LinkSet {
	links := NewLinkSet()
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
						links.Add(attr.Val)
						break
					}
				}
			}
		}
	}
}
