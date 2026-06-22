package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const maxOpenGraphBodyBytes = 1 << 20 // 1 MiB

// OpenGraphMeta — распарсенные OG/HTML title метаданные.
type OpenGraphMeta struct {
	Title       string
	Description string
	ImageURL    string
}

// FetchOpenGraphMetadata качает страницу через SafeHTTP и извлекает OG.
// Заменяет opengraph.Fetch — SafeDialContext против SSRF (NEW-02).
func FetchOpenGraphMetadata(ctx context.Context, rawURL string) (*OpenGraphMeta, error) {
	if err := ValidateOutboundURL(rawURL); err != nil {
		return nil, err
	}

	client := NewSafeHTTPClient(10 * time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("opengraph request: %w", err)
	}
	req.Header.Set("User-Agent", "Hub-LinkPreview/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("opengraph fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("opengraph fetch: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxOpenGraphBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("opengraph read body: %w", err)
	}

	return parseOpenGraphHTML(string(body)), nil
}

func parseOpenGraphHTML(doc string) *OpenGraphMeta {
	meta := &OpenGraphMeta{}
	z := html.NewTokenizer(strings.NewReader(doc))
	var inTitle bool

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if meta.Title == "" && meta.Description == "" {
				return nil
			}
			return meta
		case html.StartTagToken, html.SelfClosingTagToken:
			tn, hasAttr := z.TagName()
			name := string(tn)
			if name == "title" {
				inTitle = true
				continue
			}
			if name != "meta" || !hasAttr {
				continue
			}
			var prop, content string
			for {
				key, val, more := z.TagAttr()
				k := string(key)
				v := string(val)
				switch k {
				case "property", "name":
					prop = strings.ToLower(v)
				case "content":
					content = v
				}
				if !more {
					break
				}
			}
			switch prop {
			case "og:title":
				if content != "" {
					meta.Title = content
				}
			case "og:description":
				if content != "" {
					meta.Description = content
				}
			case "og:image":
				if content != "" && meta.ImageURL == "" {
					meta.ImageURL = content
				}
			}
		case html.TextToken:
			if inTitle && meta.Title == "" {
				meta.Title = strings.TrimSpace(string(z.Text()))
			}
		case html.EndTagToken:
			tn, _ := z.TagName()
			if string(tn) == "title" {
				inTitle = false
			}
		}
	}
}
