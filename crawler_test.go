package wikicrawl

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type expectedCounts struct {
	linkCount    int
	brokenCount  int
	requestCount int
}

func validateCrawl(t *testing.T, expected expectedCounts, handler func(http.ResponseWriter, *http.Request)) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requests++
		handler(rw, req)
	}))
	defer server.Close()

	result := NewCrawler(server.URL, "").Crawl(server.URL)
	if len(result.Visited.Set) != expected.linkCount {
		t.Errorf(`Visited links do not match expected.
			Expected %d, found %d.`, expected.linkCount, len(result.Visited.Set))
	}

	if len(result.Broken.Set) != expected.brokenCount {
		t.Errorf(`Broken links do not match expected.
			Expected %d, found %d.`, expected.brokenCount, len(result.Broken.Set))
	}

	if requests != expected.requestCount {
		t.Errorf(`Crawler should only request every internal link once.
			Expected %d, found %d.`, expected.requestCount, requests)
	}
}

func TestCrawl(t *testing.T) {

	t.Run("Mocking web crawl tests", func(t *testing.T) {
		t.Run("Validate single page with multiple links", func(t *testing.T) {
			t.Parallel()
			ex := expectedCounts{linkCount: 3, brokenCount: 0, requestCount: 3}
			validateCrawl(t, ex, func(rw http.ResponseWriter, req *http.Request) {
				fmt.Fprintf(rw, `<html><body><a href="/path1" /><a href="/path2" /></body></html>`)
			})
		})

		t.Run("Avoid duplicate requests", func(t *testing.T) {
			t.Parallel()
			ex := expectedCounts{linkCount: 2, brokenCount: 0, requestCount: 2}
			validateCrawl(t, ex, func(rw http.ResponseWriter, req *http.Request) {
				fmt.Fprintf(rw, `<html><body><a href="/path" /><a href="/path" /></body></html>`)
			})
		})

		t.Run("Record broken links", func(t *testing.T) {
			t.Parallel()
			ex := expectedCounts{linkCount: 3, brokenCount: 1, requestCount: 3}
			validateCrawl(t, ex, func(rw http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/error" {
					rw.WriteHeader(500)
					return
				}

				fmt.Fprintf(rw, `<html><body><a href="/path" /><a href="/error" /></body></html>`)
			})
		})
	})
}

func TestWikiPageTitle(t *testing.T) {
	t.Run("Validate getting Wikimedia page title", func(t *testing.T) {
		t.Run("Validate successful link", func(t *testing.T) {
			t.Parallel()
			title := "Page_Title"
			link, _ := url.Parse("http://testing.com?title=" + title)
			if WikiPageTitle(link) != "Page_Title" {
				t.Errorf("Wikimedia page title mismatch - url: %s, title: %s", link, title)
			}
		})
	})
}

func TestValidateLink(t *testing.T) {
	t.Run("Validate Wikimedia links", func(t *testing.T) {
		t.Run("Validate successful link", func(t *testing.T) {
			t.Parallel()
			link, _ := url.Parse("http://testing.com?title=Accept")
			c := NewCrawler("http://testing.com", "")
			if !c.ValidateLink(link) {
				t.Errorf("Url incorrectly marked as invalid: %s.", link)
			}
		})

		t.Run("Validate link with missing title", func(t *testing.T) {
			t.Parallel()
			link, _ := url.Parse("http://testing.com?notitle=1")
			c := NewCrawler("http://testing.com", "")
			if !c.ValidateLink(link) {
				t.Errorf("Url incorrectly marked as invalid: %s.", link)
			}
		})

		t.Run("Skip outside link", func(t *testing.T) {
			t.Parallel()
			link, _ := url.Parse("http://otherdomain.com?title=Accept")
			c := NewCrawler("http://testing.com", "")
			if c.ValidateLink(link) {
				t.Errorf("Url incorrectly marked as valid: %s.", link)
			}
		})

		t.Run("Skip forbidden pages", func(t *testing.T) {
			t.Parallel()
			link, _ := url.Parse("http://testing.com?title=Help:Skip")
			c := NewCrawler("http://testing.com", "")
			if c.ValidateLink(link) {
				t.Errorf("Url incorrectly marked as valid: %s.", link)
			}
		})
	})
}

func validateParseLinks(t *testing.T, html string, expected LinkSet) {
	found := ParseLinks(strings.NewReader(html))

	if !reflect.DeepEqual(found, expected) {
		t.Errorf("Parsing links failed, got: %v, want: %v.", found, expected)
	}
}

func TestParseLinks(t *testing.T) {
	t.Run("Parse HTML links", func(t *testing.T) {
		t.Run("Parsing single link", func(t *testing.T) {
			t.Parallel()
			html := `<html><body><a href="testing"></body></html>`

			expected := NewLinkSet()
			expected.Add("testing")

			validateParseLinks(t, html, expected)
		})

		t.Run("Malformed HTML missing closing body tag", func(t *testing.T) {
			t.Parallel()
			html := `<html><body><a href="testing"></html>`

			expected := NewLinkSet()
			expected.Add("testing")

			validateParseLinks(t, html, expected)
		})

		t.Run("Parsing self closing tag", func(t *testing.T) {
			t.Parallel()
			html := `<html><body><a href="testing" /></body></html>`

			expected := NewLinkSet()
			expected.Add("testing")

			validateParseLinks(t, html, expected)
		})
	})
}

func TestNormalizeUrl(t *testing.T) {
	t.Run("Normalize URL's", func(t *testing.T) {
		t.Run("Relative Url", func(t *testing.T) {
			t.Parallel()
			base, _ := url.Parse("http://testing.com")
			link, _ := url.Parse("/path")

			found := NormalizeUrl(link, base).String()
			expected := "http://testing.com/path"
			if found != expected {
				t.Errorf("Url malformed, got: %s, want: %s.", found, expected)
			}
		})

		t.Run("Mismatched protocol", func(t *testing.T) {
			t.Parallel()
			base, _ := url.Parse("http://testing.com")
			link, _ := url.Parse("https://testing.com/path")

			found := NormalizeUrl(link, base).String()
			expected := "http://testing.com/path"
			if found != expected {
				t.Errorf("Url malformed, got: %s, want: %s.", found, expected)
			}
		})

		t.Run("Trim unwanted parameters", func(t *testing.T) {
			t.Parallel()
			base, _ := url.Parse("http://testing.com")
			link, _ := url.Parse("http://testing.com/path?title=title&bad=2")

			found := NormalizeUrl(link, base).String()
			expected := "http://testing.com/path?title=title"
			if found != expected {
				t.Errorf("Url malformed, got: %s, want: %s.", found, expected)
			}
		})

		t.Run("Lowercase host and scheme", func(t *testing.T) {
			t.Parallel()
			base, _ := url.Parse("HTTP://Testing.Com")
			link, _ := url.Parse("HTTP://Testing.Com/path")

			found := NormalizeUrl(link, base).String()
			expected := "http://testing.com/path"
			if found != expected {
				t.Errorf("Url malformed, got: %s, want: %s.", found, expected)
			}
		})
	})
}
