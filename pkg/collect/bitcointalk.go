package collect

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	core "github.com/host-uk/core/pkg/framework/core"
	"golang.org/x/net/html"
)

// httpClient is the HTTP client used for all collection requests.
// Use SetHTTPClient to override for testing.
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// BitcoinTalkCollector collects forum posts from BitcoinTalk.
type BitcoinTalkCollector struct {
	// TopicID is the numeric topic identifier.
	TopicID string

	// URL is a full URL to a BitcoinTalk topic page. If set, TopicID is
	// extracted from it.
	URL string

	// Pages limits collection to this many pages. 0 means all pages.
	Pages int
}

// Name returns the collector name.
func (b *BitcoinTalkCollector) Name() string {
	id := b.TopicID
	if id == "" && b.URL != "" {
		id = "url"
	}
	return fmt.Sprintf("bitcointalk:%s", id)
}

// Collect gathers posts from a BitcoinTalk topic.
func (b *BitcoinTalkCollector) Collect(ctx context.Context, cfg *Config) (*Result, error) {
	result := &Result{Source: b.Name()}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(b.Name(), "Starting BitcoinTalk collection")
	}

	topicID := b.TopicID
	if topicID == "" {
		return result, core.E("collect.BitcoinTalk.Collect", "topic ID is required", nil)
	}

	if cfg.DryRun {
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(b.Name(), fmt.Sprintf("[dry-run] Would collect topic %s", topicID), nil)
		}
		return result, nil
	}

	baseDir := filepath.Join(cfg.OutputDir, "bitcointalk", topicID, "posts")
	if err := cfg.Output.EnsureDir(baseDir); err != nil {
		return result, core.E("collect.BitcoinTalk.Collect", "failed to create output directory", err)
	}

	postNum := 0
	offset := 0
	pageCount := 0
	postsPerPage := 20

	for {
		if ctx.Err() != nil {
			return result, core.E("collect.BitcoinTalk.Collect", "context cancelled", ctx.Err())
		}

		if b.Pages > 0 && pageCount >= b.Pages {
			break
		}

		if cfg.Limiter != nil {
			if err := cfg.Limiter.Wait(ctx, "bitcointalk"); err != nil {
				return result, err
			}
		}

		pageURL := fmt.Sprintf("https://bitcointalk.org/index.php?topic=%s.%d", topicID, offset)

		posts, err := b.fetchPage(ctx, pageURL)
		if err != nil {
			result.Errors++
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(b.Name(), fmt.Sprintf("Failed to fetch page at offset %d: %v", offset, err), nil)
			}
			break
		}

		if len(posts) == 0 {
			break
		}

		for _, post := range posts {
			postNum++
			filePath := filepath.Join(baseDir, fmt.Sprintf("%d.md", postNum))
			content := formatPostMarkdown(postNum, post)

			if err := cfg.Output.Write(filePath, content); err != nil {
				result.Errors++
				continue
			}

			result.Items++
			result.Files = append(result.Files, filePath)

			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitItem(b.Name(), fmt.Sprintf("Post %d by %s", postNum, post.Author), nil)
			}
		}

		pageCount++
		offset += postsPerPage

		// If we got fewer posts than expected, we've reached the end
		if len(posts) < postsPerPage {
			break
		}
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(b.Name(), fmt.Sprintf("Collected %d posts", result.Items), result)
	}

	return result, nil
}

// btPost represents a parsed BitcoinTalk forum post.
type btPost struct {
	Author  string
	Date    string
	Content string
}

// fetchPage fetches and parses a single BitcoinTalk topic page.
func (b *BitcoinTalkCollector) fetchPage(ctx context.Context, pageURL string) ([]btPost, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, core.E("collect.BitcoinTalk.fetchPage", "failed to create request", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; CoreCollector/1.0)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, core.E("collect.BitcoinTalk.fetchPage", "request failed", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, core.E("collect.BitcoinTalk.fetchPage",
			fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, core.E("collect.BitcoinTalk.fetchPage", "failed to parse HTML", err)
	}

	return extractPosts(doc), nil
}

// extractPosts extracts post data from a parsed HTML document.
// It looks for the common BitcoinTalk post structure using div.post elements.
func extractPosts(doc *html.Node) []btPost {
	var posts []btPost
	var walk func(*html.Node)

	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "post") {
					post := parsePost(n)
					if post.Content != "" {
						posts = append(posts, post)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return posts
}

// parsePost extracts author, date, and content from a post div.
func parsePost(node *html.Node) btPost {
	post := btPost{}
	var walk func(*html.Node)

	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "class" {
					switch {
					case strings.Contains(attr.Val, "poster_info"):
						post.Author = extractText(n)
					case strings.Contains(attr.Val, "headerandpost"):
						// Look for date in smalltext
						for c := n.FirstChild; c != nil; c = c.NextSibling {
							if c.Type == html.ElementNode && c.Data == "div" {
								for _, a := range c.Attr {
									if a.Key == "class" && strings.Contains(a.Val, "smalltext") {
										post.Date = strings.TrimSpace(extractText(c))
									}
								}
							}
						}
					case strings.Contains(attr.Val, "inner"):
						post.Content = strings.TrimSpace(extractText(n))
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(node)
	return post
}

// extractText recursively extracts text content from an HTML node.
func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text := extractText(c)
		if text != "" {
			if b.Len() > 0 && c.Type == html.ElementNode && (c.Data == "br" || c.Data == "p" || c.Data == "div") {
				b.WriteString("\n")
			}
			b.WriteString(text)
		}
	}
	return b.String()
}

// formatPostMarkdown formats a BitcoinTalk post as markdown.
func formatPostMarkdown(num int, post btPost) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Post %d by %s\n\n", num, post.Author)

	if post.Date != "" {
		fmt.Fprintf(&b, "**Date:** %s\n\n", post.Date)
	}

	b.WriteString(post.Content)
	b.WriteString("\n")

	return b.String()
}

// ParsePostsFromHTML parses BitcoinTalk posts from raw HTML content.
// This is exported for testing purposes.
func ParsePostsFromHTML(htmlContent string) ([]btPost, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, core.E("collect.ParsePostsFromHTML", "failed to parse HTML", err)
	}
	return extractPosts(doc), nil
}

// FormatPostMarkdown is exported for testing purposes.
func FormatPostMarkdown(num int, author, date, content string) string {
	return formatPostMarkdown(num, btPost{Author: author, Date: date, Content: content})
}

// FetchPageFunc is an injectable function type for fetching pages, used in testing.
type FetchPageFunc func(ctx context.Context, url string) ([]btPost, error)

// BitcoinTalkCollectorWithFetcher wraps BitcoinTalkCollector with a custom fetcher for testing.
type BitcoinTalkCollectorWithFetcher struct {
	BitcoinTalkCollector
	Fetcher FetchPageFunc
}

// SetHTTPClient replaces the package-level HTTP client.
// Use this in tests to inject a custom transport or timeout.
func SetHTTPClient(c *http.Client) {
	httpClient = c
}
