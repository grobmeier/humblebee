package guiapp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type newsTransport func(*http.Request) (*http.Response, error)

func (f newsTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newsFixture(items string) string {
	return `<rss version="2.0"><channel><language>en</language>` + items + `</channel></rss>`
}
func newsFixtureItem(id int, link string) string {
	return fmt.Sprintf(`<item><guid>%d</guid><title>News &amp; updates</title><link>%s</link><pubDate>Sat, 19 Sep 2026 10:00:00 +0000</pubDate><description><![CDATA[<p>Hello <b>world</b></p><script>bad()</script><img src="https://evil.invalid/x">]]></description></item>`, id, link)
}
func newsFixtureClient(body string) *http.Client {
	c := newNewsClient()
	c.Transport = newsTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}, Request: r}, nil
	})
	return c
}

func TestNewsParsing(t *testing.T) {
	items, err := parseNews([]byte(newsFixture(newsFixtureItem(1, newsOrigin+"/article/"))), "en")
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%v err=%v", items, err)
	}
	if items[0].Title != "News & updates" || items[0].Summary != "Hello world" {
		t.Fatalf("unsafe or incorrect text: %+v", items[0])
	}
	if items[0].GUID != "1" || items[0].PublishedAt != "2026-09-19T10:00:00Z" {
		t.Fatal(items[0])
	}
	if _, err := parseNews([]byte(newsFixture("")), "de"); err == nil {
		t.Fatal("accepted wrong language")
	}
}

func TestNewsRejectsUnsafeInput(t *testing.T) {
	for _, raw := range []string{"http://www.timeandbill.de/a", "https://evil.invalid/a", "https://www.timeandbill.de.evil.invalid/a", "https://user@www.timeandbill.de/a", "https://www.timeandbill.de:444/a", "//www.timeandbill.de/a", "javascript:alert(1)", newsOrigin + "/a?tracking=1", newsOrigin + "/a#fragment"} {
		t.Run(raw, func(t *testing.T) {
			if validNewsURL(raw) {
				t.Fatal("accepted URL")
			}
			items, err := parseNews([]byte(newsFixture(newsFixtureItem(1, raw))), "en")
			if err == nil && len(items) != 0 {
				t.Fatal("unsafe item returned")
			}
		})
	}
	for _, body := range []string{"<rss>", `<html/>`, `<!DOCTYPE rss [<!ENTITY x SYSTEM "file:///etc/passwd">]><rss version="2.0"><channel>&x;</channel></rss>`, strings.Repeat("x", newsMaxBytes+1)} {
		if _, err := parseNews([]byte(body), "en"); err == nil {
			t.Fatal("accepted invalid feed")
		}
	}
}

func TestNewsBoundedAndDeduplicated(t *testing.T) {
	var body strings.Builder
	for i := 0; i < 30; i++ {
		body.WriteString(newsFixtureItem(i, newsOrigin+"/article/"))
	}
	items, err := parseNews([]byte(newsFixture(body.String())), "en")
	if err != nil || len(items) != newsMaxItems {
		t.Fatalf("%d %v", len(items), err)
	}
	item := newsFixtureItem(1, newsOrigin+"/article/")
	items, err = parseNews([]byte(newsFixture(item+item)), "en")
	if err != nil || len(items) != 1 {
		t.Fatal("duplicate GUID")
	}
	items, err = parseNews([]byte(newsFixture(strings.ReplaceAll(item, "2026", "2099"))), "en")
	if err != nil || len(items) != 0 {
		t.Fatal("future article")
	}
}

func TestNewsOfflineCacheAndLanguageIsolation(t *testing.T) {
	dir := t.TempDir()
	client := newsFixtureClient(newsFixture(newsFixtureItem(1, newsOrigin+"/article/")))
	result := loadNews(context.Background(), client, dir, "en")
	if result.Unavailable || result.Cached || result.CacheWriteFailed || len(result.Items) != 1 {
		t.Fatal(result)
	}
	client.Transport = newsTransport(func(r *http.Request) (*http.Response, error) { return nil, errors.New("offline") })
	cached := loadNews(context.Background(), client, dir, "en")
	if !cached.Unavailable || !cached.Cached || cached.FetchedAt != result.FetchedAt || len(cached.Items) != 1 {
		t.Fatal(cached)
	}
	other := loadNews(context.Background(), client, dir, "de")
	if !other.Unavailable || other.Cached || len(other.Items) != 0 {
		t.Fatal(other)
	}
	if err := os.WriteFile(filepath.Join(dir, "news-en.json"), []byte(`{"fetchedAt":"2026-09-19T10:00:00Z","items":[{"url":"https://evil.invalid/a"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := readNewsCache(dir, "en"); ok {
		t.Fatal("trusted tampered cache")
	}
}

func TestNewsRequestsAndRedirects(t *testing.T) {
	client := newNewsClient()
	calls := 0
	client.Transport = newsTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.String() != newsOrigin+"/en/humblebee/news/feed.xml" || r.Method != "GET" || r.Header.Get("Cookie") != "" || r.Header.Get("Authorization") != "" {
			t.Fatalf("unexpected request: %v", r)
		}
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{"https://evil.invalid/"}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
	})
	if calls != 0 {
		t.Fatal("unsolicited fetch")
	}
	if _, err := fetchNews(context.Background(), client, "../../private"); err == nil || calls != 0 {
		t.Fatal("invalid language fetched")
	}
	if _, err := fetchNews(context.Background(), client, "en"); err == nil || calls != 1 {
		t.Fatal("redirect followed")
	}
	u, _ := url.Parse(newsOrigin + "/redirect/")
	if client.CheckRedirect(&http.Request{URL: u}, make([]*http.Request, 3)) == nil {
		t.Fatal("unbounded redirects")
	}
	if client.CheckRedirect(&http.Request{URL: u}, make([]*http.Request, 1)) != nil {
		t.Fatal("same origin rejected")
	}
}

func TestNewsLimitsTimeoutAndCacheFailure(t *testing.T) {
	if _, err := fetchNews(context.Background(), newsFixtureClient(strings.Repeat("x", newsMaxBytes+1)), "en"); err == nil {
		t.Fatal("oversized response accepted")
	}
	client := newNewsClient()
	client.Timeout = 10 * time.Millisecond
	client.Transport = newsTransport(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })
	start := time.Now()
	if _, err := fetchNews(context.Background(), client, "en"); err == nil {
		t.Fatal("timeout ignored")
	}
	if time.Since(start) > time.Second {
		t.Fatal("timeout exceeded")
	}
	result := loadNews(context.Background(), newsFixtureClient(newsFixture("")), "", "en")
	if result.Unavailable || !result.CacheWriteFailed {
		t.Fatal(result)
	}
	client = newsFixtureClient("")
	client.Transport = newsTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
	})
	if _, err := fetchNews(context.Background(), client, "en"); err == nil {
		t.Fatal("HTTP error accepted")
	}
}

func TestNewsOrderingAndMalformedDocuments(t *testing.T) {
	older := strings.ReplaceAll(newsFixtureItem(1, newsOrigin+"/old/"), "19 Sep", "18 Sep")
	newer := newsFixtureItem(2, newsOrigin+"/new/")
	items, err := parseNews([]byte(newsFixture(older+newer)), "en")
	if err != nil || len(items) != 2 || items[0].GUID != "2" {
		t.Fatalf("incorrect ordering: %v %v", items, err)
	}
	for _, body := range []string{newsFixture("") + newsFixture(""), newsFixture("") + "junk"} {
		if _, err := parseNews([]byte(body), "en"); err == nil {
			t.Fatal("accepted trailing content")
		}
	}
}

func TestNewsConcurrentCache(t *testing.T) {
	dir := t.TempDir()
	client := newsFixtureClient(newsFixture(newsFixtureItem(1, newsOrigin+"/article/")))
	var workers sync.WaitGroup
	for i := 0; i < 6; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			result := loadNews(context.Background(), client, dir, "en")
			if result.CacheWriteFailed || result.Unavailable {
				t.Error(result)
			}
		}()
	}
	workers.Wait()
	if result, ok := readNewsCache(dir, "en"); !ok || len(result.Items) != 1 {
		t.Fatal("invalid cache after concurrent fetch")
	}
}
