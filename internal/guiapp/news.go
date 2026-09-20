package guiapp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/grobmeier/humblebee/internal/paths"
)

const newsOrigin = "https://www.timeandbill.de"
const newsMaxBytes = 512 * 1024
const newsMaxItems = 20

type NewsItem struct {
	GUID        string `json:"guid"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	PublishedAt string `json:"publishedAt"`
	Summary     string `json:"summary"`
}

type NewsResult struct {
	Items            []NewsItem `json:"items"`
	FetchedAt        string     `json:"fetchedAt"`
	Cached           bool       `json:"cached"`
	Unavailable      bool       `json:"unavailable"`
	CacheWriteFailed bool       `json:"cacheWriteFailed"`
}

// Independent of database operations; serializes atomic cache updates only.
var newsCacheMu sync.Mutex

func validNewsURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Host == "www.timeandbill.de" && u.User == nil && u.Opaque == "" && u.Path != "" && u.Fragment == "" && u.RawQuery == ""
}

func newNewsClient() *http.Client {
	return &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 || !validNewsURL(req.URL.String()) {
			return fmt.Errorf("news redirect rejected")
		}
		return nil
	}}
}

// GetNews performs a manual request. Neither construction nor startup calls it.
func (a *App) GetNews(language string) (NewsResult, error) {
	if language != "de" && language != "en" {
		return NewsResult{}, fmt.Errorf("unsupported news language")
	}
	dir, err := paths.DataDir()
	if err != nil {
		dir = ""
	}
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return loadNews(ctx, newNewsClient(), dir, language), nil
}

func loadNews(ctx context.Context, client *http.Client, dir, language string) NewsResult {
	result := NewsResult{Items: []NewsItem{}}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	items, err := fetchNews(ctx, client, language)
	if err != nil {
		if cached, ok := readNewsCache(dir, language); ok {
			result = cached
			result.Cached = true
		}
		result.Unavailable = true
		return result
	}
	result.Items = items
	result.FetchedAt = time.Now().UTC().Format(time.RFC3339)
	result.CacheWriteFailed = writeNewsCache(dir, language, result) != nil
	return result
}

func fetchNews(ctx context.Context, client *http.Client, language string) ([]NewsItem, error) {
	if language != "de" && language != "en" {
		return nil, fmt.Errorf("unsupported news language")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, newsOrigin+"/"+language+"/humblebee/news/feed.xml", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")
	req.Header.Set("User-Agent", "HumbleBee-News")
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("news unavailable")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, newsMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > newsMaxBytes {
		return nil, fmt.Errorf("news response too large")
	}
	return parseNews(body, language)
}

func readNewsCache(dir, language string) (NewsResult, bool) {
	if dir == "" {
		return NewsResult{}, false
	}
	newsCacheMu.Lock()
	defer newsCacheMu.Unlock()
	f, err := os.Open(filepath.Join(dir, "news-"+language+".json"))
	if err != nil {
		return NewsResult{}, false
	}
	defer f.Close()
	body, err := io.ReadAll(io.LimitReader(f, newsMaxBytes+1))
	var result NewsResult
	if err != nil || len(body) > newsMaxBytes || json.Unmarshal(body, &result) != nil {
		return NewsResult{}, false
	}
	if _, err := time.Parse(time.RFC3339, result.FetchedAt); err != nil || len(result.Items) > newsMaxItems {
		return NewsResult{}, false
	}
	for _, item := range result.Items {
		if !validNewsURL(item.URL) {
			return NewsResult{}, false
		}
		if _, err := time.Parse(time.RFC3339, item.PublishedAt); err != nil {
			return NewsResult{}, false
		}
	}
	if result.Items == nil {
		result.Items = []NewsItem{}
	}
	return result, true
}

func writeNewsCache(dir, language string, result NewsResult) error {
	if dir == "" {
		return fmt.Errorf("cache directory unavailable")
	}
	newsCacheMu.Lock()
	defer newsCacheMu.Unlock()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if len(body) > newsMaxBytes {
		return fmt.Errorf("news cache too large")
	}
	f, err := os.CreateTemp(dir, ".news-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(body); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filepath.Join(dir, "news-"+language+".json"))
}
