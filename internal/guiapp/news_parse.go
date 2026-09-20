package guiapp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func parseNews(body []byte, language string) ([]NewsItem, error) {
	if len(body) > newsMaxBytes {
		return nil, fmt.Errorf("news response too large")
	}
	// Reject DTDs explicitly; encoding/xml never resolves external entities.
	d := xml.NewDecoder(bytes.NewReader(body))
	depth, roots := 0, 0
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if _, ok := token.(xml.Directive); ok {
			return nil, fmt.Errorf("XML directives are not allowed")
		}
		switch value := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(value)) != "" {
				return nil, fmt.Errorf("unexpected XML content")
			}
		}
	}
	if roots != 1 {
		return nil, fmt.Errorf("expected one RSS document")
	}
	var feed struct {
		XMLName  xml.Name `xml:"rss"`
		Version  string   `xml:"version,attr"`
		Channels []struct {
			Language string `xml:"language"`
			Items    []struct {
				GUID        string `xml:"guid"`
				Title       string `xml:"title"`
				Link        string `xml:"link"`
				Date        string `xml:"pubDate"`
				Description string `xml:"description"`
			} `xml:"item"`
		} `xml:"channel"`
	}
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}
	if feed.Version != "2.0" || len(feed.Channels) != 1 {
		return nil, fmt.Errorf("invalid RSS feed")
	}
	channel := feed.Channels[0]
	if channel.Language != "" && strings.Split(strings.ToLower(channel.Language), "-")[0] != language {
		return nil, fmt.Errorf("unexpected news language")
	}
	items := []NewsItem{}
	seen := map[string]bool{}
	for _, raw := range channel.Items {
		link := strings.TrimSpace(raw.Link)
		if !validNewsURL(link) {
			continue
		}
		date, err := httpNewsDate(raw.Date)
		if err != nil || date.After(time.Now()) {
			continue
		}
		title := newsPlainText(raw.Title, 300)
		if title == "" {
			continue
		}
		guid := strings.TrimSpace(raw.GUID)
		if guid == "" || len(guid) > 512 {
			guid = link
		}
		if seen[guid] {
			continue
		}
		seen[guid] = true
		items = append(items, NewsItem{GUID: guid, Title: title, URL: link, PublishedAt: date.UTC().Format(time.RFC3339), Summary: newsPlainText(raw.Description, 2000)})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].PublishedAt > items[j].PublishedAt })
	if len(items) > newsMaxItems {
		items = items[:newsMaxItems]
	}
	return items, nil
}

func httpNewsDate(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822, time.RFC3339} {
		if date, err := time.Parse(layout, strings.TrimSpace(value)); err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid publication date")
}

func newsPlainText(value string, limit int) string {
	z := html.NewTokenizer(strings.NewReader(value))
	var out strings.Builder
	blocked := ""
	for {
		switch z.Next() {
		case html.ErrorToken:
			text := []rune(strings.Join(strings.Fields(out.String()), " "))
			if len(text) > limit {
				text = text[:limit]
			}
			return string(text)
		case html.StartTagToken:
			name, _ := z.TagName()
			if string(name) == "script" || string(name) == "style" {
				blocked = string(name)
			}
			out.WriteByte(' ')
		case html.EndTagToken:
			name, _ := z.TagName()
			if string(name) == blocked {
				blocked = ""
			}
			out.WriteByte(' ')
		case html.TextToken:
			if blocked == "" {
				out.Write(z.Text())
			}
		}
	}
}
