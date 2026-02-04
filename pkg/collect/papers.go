package collect

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	core "github.com/host-uk/core/pkg/framework/core"
	"golang.org/x/net/html"
)

// Paper source identifiers.
const (
	PaperSourceIACR  = "iacr"
	PaperSourceArXiv = "arxiv"
	PaperSourceAll   = "all"
)

// PapersCollector collects papers from IACR and arXiv.
type PapersCollector struct {
	// Source is one of PaperSourceIACR, PaperSourceArXiv, or PaperSourceAll.
	Source string

	// Category is the arXiv category (e.g. "cs.CR" for cryptography).
	Category string

	// Query is the search query string.
	Query string
}

// Name returns the collector name.
func (p *PapersCollector) Name() string {
	return fmt.Sprintf("papers:%s", p.Source)
}

// paper represents a parsed academic paper.
type paper struct {
	ID       string
	Title    string
	Authors  []string
	Abstract string
	Date     string
	URL      string
	Source   string
}

// Collect gathers papers from the configured sources.
func (p *PapersCollector) Collect(ctx context.Context, cfg *Config) (*Result, error) {
	result := &Result{Source: p.Name()}

	if p.Query == "" {
		return result, core.E("collect.Papers.Collect", "query is required", nil)
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(p.Name(), fmt.Sprintf("Starting paper collection for %q", p.Query))
	}

	if cfg.DryRun {
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(p.Name(), fmt.Sprintf("[dry-run] Would search papers for %q", p.Query), nil)
		}
		return result, nil
	}

	switch p.Source {
	case PaperSourceIACR:
		return p.collectIACR(ctx, cfg)
	case PaperSourceArXiv:
		return p.collectArXiv(ctx, cfg)
	case PaperSourceAll:
		iacrResult, iacrErr := p.collectIACR(ctx, cfg)
		arxivResult, arxivErr := p.collectArXiv(ctx, cfg)

		if iacrErr != nil && arxivErr != nil {
			return result, core.E("collect.Papers.Collect", "all sources failed", iacrErr)
		}

		merged := MergeResults(p.Name(), iacrResult, arxivResult)
		if iacrErr != nil {
			merged.Errors++
		}
		if arxivErr != nil {
			merged.Errors++
		}

		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitComplete(p.Name(), fmt.Sprintf("Collected %d papers", merged.Items), merged)
		}

		return merged, nil
	default:
		return result, core.E("collect.Papers.Collect",
			fmt.Sprintf("unknown source: %s (use iacr, arxiv, or all)", p.Source), nil)
	}
}

// collectIACR fetches papers from the IACR ePrint archive.
func (p *PapersCollector) collectIACR(ctx context.Context, cfg *Config) (*Result, error) {
	result := &Result{Source: "papers:iacr"}

	if cfg.Limiter != nil {
		if err := cfg.Limiter.Wait(ctx, "iacr"); err != nil {
			return result, err
		}
	}

	searchURL := fmt.Sprintf("https://eprint.iacr.org/search?q=%s", url.QueryEscape(p.Query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return result, core.E("collect.Papers.collectIACR", "failed to create request", err)
	}
	req.Header.Set("User-Agent", "CoreCollector/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return result, core.E("collect.Papers.collectIACR", "request failed", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return result, core.E("collect.Papers.collectIACR",
			fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return result, core.E("collect.Papers.collectIACR", "failed to parse HTML", err)
	}

	papers := extractIACRPapers(doc)

	baseDir := filepath.Join(cfg.OutputDir, "papers", "iacr")
	if err := cfg.Output.EnsureDir(baseDir); err != nil {
		return result, core.E("collect.Papers.collectIACR", "failed to create output directory", err)
	}

	for _, ppr := range papers {
		filePath := filepath.Join(baseDir, ppr.ID+".md")
		content := formatPaperMarkdown(ppr)

		if err := cfg.Output.Write(filePath, content); err != nil {
			result.Errors++
			continue
		}

		result.Items++
		result.Files = append(result.Files, filePath)

		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitItem(p.Name(), fmt.Sprintf("Paper: %s", ppr.Title), nil)
		}
	}

	return result, nil
}

// arxivFeed represents the Atom feed returned by the arXiv API.
type arxivFeed struct {
	XMLName xml.Name     `xml:"feed"`
	Entries []arxivEntry `xml:"entry"`
}

type arxivEntry struct {
	ID        string        `xml:"id"`
	Title     string        `xml:"title"`
	Summary   string        `xml:"summary"`
	Published string        `xml:"published"`
	Authors   []arxivAuthor `xml:"author"`
	Links     []arxivLink   `xml:"link"`
}

type arxivAuthor struct {
	Name string `xml:"name"`
}

type arxivLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

// collectArXiv fetches papers from the arXiv API.
func (p *PapersCollector) collectArXiv(ctx context.Context, cfg *Config) (*Result, error) {
	result := &Result{Source: "papers:arxiv"}

	if cfg.Limiter != nil {
		if err := cfg.Limiter.Wait(ctx, "arxiv"); err != nil {
			return result, err
		}
	}

	query := url.QueryEscape(p.Query)
	if p.Category != "" {
		query = fmt.Sprintf("cat:%s+AND+%s", url.QueryEscape(p.Category), query)
	}

	searchURL := fmt.Sprintf("https://export.arxiv.org/api/query?search_query=%s&max_results=50", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return result, core.E("collect.Papers.collectArXiv", "failed to create request", err)
	}
	req.Header.Set("User-Agent", "CoreCollector/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return result, core.E("collect.Papers.collectArXiv", "request failed", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return result, core.E("collect.Papers.collectArXiv",
			fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	var feed arxivFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return result, core.E("collect.Papers.collectArXiv", "failed to parse XML", err)
	}

	baseDir := filepath.Join(cfg.OutputDir, "papers", "arxiv")
	if err := cfg.Output.EnsureDir(baseDir); err != nil {
		return result, core.E("collect.Papers.collectArXiv", "failed to create output directory", err)
	}

	for _, entry := range feed.Entries {
		ppr := arxivEntryToPaper(entry)

		filePath := filepath.Join(baseDir, ppr.ID+".md")
		content := formatPaperMarkdown(ppr)

		if err := cfg.Output.Write(filePath, content); err != nil {
			result.Errors++
			continue
		}

		result.Items++
		result.Files = append(result.Files, filePath)

		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitItem(p.Name(), fmt.Sprintf("Paper: %s", ppr.Title), nil)
		}
	}

	return result, nil
}

// arxivEntryToPaper converts an arXiv Atom entry to a paper.
func arxivEntryToPaper(entry arxivEntry) paper {
	authors := make([]string, len(entry.Authors))
	for i, a := range entry.Authors {
		authors[i] = a.Name
	}

	// Extract the arXiv ID from the URL
	id := entry.ID
	if idx := strings.LastIndex(id, "/abs/"); idx != -1 {
		id = id[idx+5:]
	}
	// Replace characters that are not valid in file names
	id = strings.ReplaceAll(id, "/", "-")
	id = strings.ReplaceAll(id, ":", "-")

	paperURL := entry.ID
	for _, link := range entry.Links {
		if link.Rel == "alternate" {
			paperURL = link.Href
			break
		}
	}

	return paper{
		ID:       id,
		Title:    strings.TrimSpace(entry.Title),
		Authors:  authors,
		Abstract: strings.TrimSpace(entry.Summary),
		Date:     entry.Published,
		URL:      paperURL,
		Source:   "arxiv",
	}
}

// extractIACRPapers extracts paper metadata from an IACR search results page.
func extractIACRPapers(doc *html.Node) []paper {
	var papers []paper
	var walk func(*html.Node)

	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "paperentry") {
					ppr := parseIACREntry(n)
					if ppr.Title != "" {
						papers = append(papers, ppr)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return papers
}

// parseIACREntry extracts paper data from an IACR paper entry div.
func parseIACREntry(node *html.Node) paper {
	ppr := paper{Source: "iacr"}
	var walk func(*html.Node)

	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" && strings.Contains(attr.Val, "/eprint/") {
						ppr.URL = "https://eprint.iacr.org" + attr.Val
						// Extract ID from URL
						parts := strings.Split(attr.Val, "/")
						if len(parts) >= 2 {
							ppr.ID = parts[len(parts)-2] + "-" + parts[len(parts)-1]
						}
					}
				}
				if ppr.Title == "" {
					ppr.Title = strings.TrimSpace(extractText(n))
				}
			case "span":
				for _, attr := range n.Attr {
					if attr.Key == "class" {
						switch {
						case strings.Contains(attr.Val, "author"):
							author := strings.TrimSpace(extractText(n))
							if author != "" {
								ppr.Authors = append(ppr.Authors, author)
							}
						case strings.Contains(attr.Val, "date"):
							ppr.Date = strings.TrimSpace(extractText(n))
						}
					}
				}
			case "p":
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "abstract") {
						ppr.Abstract = strings.TrimSpace(extractText(n))
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(node)
	return ppr
}

// formatPaperMarkdown formats a paper as markdown.
func formatPaperMarkdown(ppr paper) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", ppr.Title)

	if len(ppr.Authors) > 0 {
		fmt.Fprintf(&b, "- **Authors:** %s\n", strings.Join(ppr.Authors, ", "))
	}
	if ppr.Date != "" {
		fmt.Fprintf(&b, "- **Published:** %s\n", ppr.Date)
	}
	if ppr.URL != "" {
		fmt.Fprintf(&b, "- **URL:** %s\n", ppr.URL)
	}
	if ppr.Source != "" {
		fmt.Fprintf(&b, "- **Source:** %s\n", ppr.Source)
	}

	if ppr.Abstract != "" {
		fmt.Fprintf(&b, "\n## Abstract\n\n%s\n", ppr.Abstract)
	}

	return b.String()
}

// FormatPaperMarkdown is exported for testing.
func FormatPaperMarkdown(title string, authors []string, date, paperURL, source, abstract string) string {
	return formatPaperMarkdown(paper{
		Title:    title,
		Authors:  authors,
		Date:     date,
		URL:      paperURL,
		Source:   source,
		Abstract: abstract,
	})
}
