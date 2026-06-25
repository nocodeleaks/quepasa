package library

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/nocodeleaks/quepasa/qplog"
	"golang.org/x/net/html"
)

// OpenGraphData holds Open Graph metadata fetched from a URL
type OpenGraphData struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

// ogCacheEntry wraps OpenGraphData with a TTL timestamp
type ogCacheEntry struct {
	data      *OpenGraphData
	timestamp time.Time
}

var (
	ogCache    sync.Map
	ogCacheTTL = 5 * time.Minute
	ogTimeout  = 5 * time.Second
	urlRegex   = regexp.MustCompile(`https?://[^\s<>"]+`)
	httpClient *http.Client
	clientOnce sync.Once
)

func getHTTPClient() *http.Client {
	clientOnce.Do(func() {
		httpClient = &http.Client{
			Timeout: ogTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		}
	})
	return httpClient
}

// ExtractURLFromText extracts the first URL found in a text string
func ExtractURLFromText(text string) string {
	return urlRegex.FindString(text)
}

// ExtractAllURLsFromText extracts all URLs found in a text string
func ExtractAllURLsFromText(text string) []string {
	return urlRegex.FindAllString(text, -1)
}

// FetchOpenGraph fetches Open Graph metadata from a URL with in-memory caching
func FetchOpenGraph(url string) (*OpenGraphData, error) {
	// Check cache first
	if cached, ok := ogCache.Load(url); ok {
		entry := cached.(*ogCacheEntry)
		if time.Since(entry.timestamp) < ogCacheTTL {
			log.Debugf("opengraph cache hit for: %s", url)
			return entry.data, nil
		}
		ogCache.Delete(url)
	}

	log.Debugf("fetching opengraph for: %s", url)

	client := getHTTPClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Limit reading to 1 MB to prevent memory issues
	limitedReader := io.LimitReader(resp.Body, 1024*1024)

	data, err := parseOpenGraph(limitedReader, url)
	if err != nil {
		return nil, err
	}

	ogCache.Store(url, &ogCacheEntry{
		data:      data,
		timestamp: time.Now(),
	})

	return data, nil
}

// parseOpenGraph parses HTML and extracts Open Graph meta tags
func parseOpenGraph(r io.Reader, originalURL string) (*OpenGraphData, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	data := &OpenGraphData{URL: originalURL}

	var titleFromTag string

	var parseNode func(*html.Node)
	parseNode = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "meta":
				var property, content, name string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "property":
						property = attr.Val
					case "name":
						name = attr.Val
					case "content":
						content = attr.Val
					}
				}

				switch property {
				case "og:title":
					if data.Title == "" {
						data.Title = content
					}
				case "og:description":
					if data.Description == "" {
						data.Description = content
					}
				case "og:image":
					if data.ImageURL == "" {
						data.ImageURL = content
					}
				case "og:url":
					if content != "" {
						data.URL = content
					}
				}

				switch name {
				case "description":
					if data.Description == "" {
						data.Description = content
					}
				case "twitter:title":
					if data.Title == "" {
						data.Title = content
					}
				case "twitter:description":
					if data.Description == "" {
						data.Description = content
					}
				case "twitter:image":
					if data.ImageURL == "" {
						data.ImageURL = content
					}
				}

			case "title":
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					titleFromTag = strings.TrimSpace(n.FirstChild.Data)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseNode(c)
		}
	}

	parseNode(doc)

	if data.Title == "" && titleFromTag != "" {
		data.Title = titleFromTag
	}

	return data, nil
}

// DownloadImage downloads an image from a URL (max 5 MB) and returns its bytes
func DownloadImage(url string) ([]byte, error) {
	if url == "" {
		return nil, fmt.Errorf("empty image URL")
	}

	log.Debugf("downloading image from: %s", url)

	client := getHTTPClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create image request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected image status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	return data, nil
}

// DecodeBase64 decodes a base64-encoded string to bytes
func DecodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// ClearOpenGraphCache evicts all cached Open Graph entries
func ClearOpenGraphCache() {
	ogCache.Range(func(key, value any) bool {
		ogCache.Delete(key)
		return true
	})
}
