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

func TestCrawl(t *testing.T) {

	t.Run("Mocking web crawl tests", func(t *testing.T) {
		t.Run("Validate single page", func(t *testing.T) {
			t.Parallel()

			// Mock server.
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				fmt.Fprintf(rw, `<html><body><a href="%s/path" /></body></html>`, req.URL.Path)
			}))
			defer server.Close()

			c := NewCrawler(server.URL)
			result := c.Crawl(server.URL)
			if len(result.Links) != 1 {
				t.Errorf("Visited links do not match expected. %s", server.URL)
			}
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
			link := "http://testing.com?title=Accept"
			c := NewCrawler("http://testing.com")
			if !c.ValidateLink(link) {
				t.Errorf("Url incorrectly marked as invalid: %s.", link)
			}
		})

		t.Run("Validate link with missing title", func(t *testing.T) {
			t.Parallel()
			link := "http://testing.com?notitle=1"
			c := NewCrawler("http://testing.com")
			if !c.ValidateLink(link) {
				t.Errorf("Url incorrectly marked as invalid: %s.", link)
			}
		})

		t.Run("Skip outside link", func(t *testing.T) {
			t.Parallel()
			link := "http://otherdomain.com?title=Accept"
			c := NewCrawler("http://testing.com")
			if c.ValidateLink(link) {
				t.Errorf("Url incorrectly marked as valid: %s.", link)
			}
		})

		t.Run("Skip forbidden pages", func(t *testing.T) {
			t.Parallel()
			link := "http://testing.com?title=Help:Skip"
			c := NewCrawler("http://testing.com")
			if c.ValidateLink(link) {
				t.Errorf("Url incorrectly marked as valid: %s.", link)
			}
		})
	})
}

func TestParseLinks(t *testing.T) {
	t.Run("Parse HTML links", func(t *testing.T) {
		t.Run("Parsing single link", func(t *testing.T) {
			t.Parallel()
			html := `<html><body><a href="testing"></body></html>`

			found := ParseLinks(strings.NewReader(html))
			expected := []string{"testing"}

			if !reflect.DeepEqual(found, expected) {
				t.Errorf("Parsing links failed, got: %s, want: %s.", found, expected)
			}
		})

		t.Run("Malformed HTML missing closing body tag", func(t *testing.T) {
			t.Parallel()
			html := `<html><body><a href="testing"></html>`

			found := ParseLinks(strings.NewReader(html))
			expected := []string{"testing"}

			if !reflect.DeepEqual(found, expected) {
				t.Errorf("Parsing links failed, got: %s, want: %s.", found, expected)
			}
		})

		t.Run("Parsing self closing tag", func(t *testing.T) {
			t.Parallel()
			html := `<html><body><a href="testing" /></body></html>`

			found := ParseLinks(strings.NewReader(html))
			expected := []string{"testing"}

			if !reflect.DeepEqual(found, expected) {
				t.Errorf("Parsing links failed, got: %s, want: %s.", found, expected)
			}
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
