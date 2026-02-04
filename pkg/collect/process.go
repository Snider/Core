package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	core "github.com/host-uk/core/pkg/framework/core"
	"golang.org/x/net/html"
)

// Processor converts collected data to clean markdown.
type Processor struct {
	// Source identifies the data source directory to process.
	Source string

	// Dir is the directory containing files to process.
	Dir string
}

// Name returns the processor name.
func (p *Processor) Name() string {
	return fmt.Sprintf("process:%s", p.Source)
}

// Process reads files from the source directory, converts HTML or JSON
// to clean markdown, and writes the results to the output directory.
func (p *Processor) Process(ctx context.Context, cfg *Config) (*Result, error) {
	result := &Result{Source: p.Name()}

	if p.Dir == "" {
		return result, core.E("collect.Processor.Process", "directory is required", nil)
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(p.Name(), fmt.Sprintf("Processing files in %s", p.Dir))
	}

	if cfg.DryRun {
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(p.Name(), fmt.Sprintf("[dry-run] Would process files in %s", p.Dir), nil)
		}
		return result, nil
	}

	entries, err := cfg.Output.List(p.Dir)
	if err != nil {
		return result, core.E("collect.Processor.Process", "failed to list directory", err)
	}

	outputDir := filepath.Join(cfg.OutputDir, "processed", p.Source)
	if err := cfg.Output.EnsureDir(outputDir); err != nil {
		return result, core.E("collect.Processor.Process", "failed to create output directory", err)
	}

	for _, entry := range entries {
		if ctx.Err() != nil {
			return result, core.E("collect.Processor.Process", "context cancelled", ctx.Err())
		}

		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		srcPath := filepath.Join(p.Dir, name)

		content, err := cfg.Output.Read(srcPath)
		if err != nil {
			result.Errors++
			continue
		}

		var processed string
		ext := strings.ToLower(filepath.Ext(name))

		switch ext {
		case ".html", ".htm":
			processed, err = htmlToMarkdown(content)
			if err != nil {
				result.Errors++
				if cfg.Dispatcher != nil {
					cfg.Dispatcher.EmitError(p.Name(), fmt.Sprintf("Failed to convert %s: %v", name, err), nil)
				}
				continue
			}
		case ".json":
			processed, err = jsonToMarkdown(content)
			if err != nil {
				result.Errors++
				if cfg.Dispatcher != nil {
					cfg.Dispatcher.EmitError(p.Name(), fmt.Sprintf("Failed to convert %s: %v", name, err), nil)
				}
				continue
			}
		case ".md":
			// Already markdown, just clean up
			processed = strings.TrimSpace(content)
		default:
			result.Skipped++
			continue
		}

		// Write with .md extension
		outName := strings.TrimSuffix(name, ext) + ".md"
		outPath := filepath.Join(outputDir, outName)

		if err := cfg.Output.Write(outPath, processed); err != nil {
			result.Errors++
			continue
		}

		result.Items++
		result.Files = append(result.Files, outPath)

		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitItem(p.Name(), fmt.Sprintf("Processed: %s", name), nil)
		}
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(p.Name(), fmt.Sprintf("Processed %d files", result.Items), result)
	}

	return result, nil
}

// htmlToMarkdown converts HTML content to clean markdown.
func htmlToMarkdown(content string) (string, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return "", core.E("collect.htmlToMarkdown", "failed to parse HTML", err)
	}

	var b strings.Builder
	nodeToMarkdown(&b, doc, 0)
	return strings.TrimSpace(b.String()), nil
}

// nodeToMarkdown recursively converts an HTML node tree to markdown.
func nodeToMarkdown(b *strings.Builder, n *html.Node, depth int) {
	switch n.Type {
	case html.TextNode:
		text := n.Data
		if strings.TrimSpace(text) != "" {
			b.WriteString(text)
		}
	case html.ElementNode:
		switch n.Data {
		case "h1":
			b.WriteString("\n# ")
			writeChildrenText(b, n)
			b.WriteString("\n\n")
			return
		case "h2":
			b.WriteString("\n## ")
			writeChildrenText(b, n)
			b.WriteString("\n\n")
			return
		case "h3":
			b.WriteString("\n### ")
			writeChildrenText(b, n)
			b.WriteString("\n\n")
			return
		case "h4":
			b.WriteString("\n#### ")
			writeChildrenText(b, n)
			b.WriteString("\n\n")
			return
		case "h5":
			b.WriteString("\n##### ")
			writeChildrenText(b, n)
			b.WriteString("\n\n")
			return
		case "h6":
			b.WriteString("\n###### ")
			writeChildrenText(b, n)
			b.WriteString("\n\n")
			return
		case "p":
			b.WriteString("\n")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				nodeToMarkdown(b, c, depth)
			}
			b.WriteString("\n")
			return
		case "br":
			b.WriteString("\n")
			return
		case "strong", "b":
			b.WriteString("**")
			writeChildrenText(b, n)
			b.WriteString("**")
			return
		case "em", "i":
			b.WriteString("*")
			writeChildrenText(b, n)
			b.WriteString("*")
			return
		case "code":
			b.WriteString("`")
			writeChildrenText(b, n)
			b.WriteString("`")
			return
		case "pre":
			b.WriteString("\n```\n")
			writeChildrenText(b, n)
			b.WriteString("\n```\n")
			return
		case "a":
			var href string
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href = attr.Val
				}
			}
			text := getChildrenText(n)
			if href != "" {
				fmt.Fprintf(b, "[%s](%s)", text, href)
			} else {
				b.WriteString(text)
			}
			return
		case "ul":
			b.WriteString("\n")
		case "ol":
			b.WriteString("\n")
			counter := 1
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "li" {
					fmt.Fprintf(b, "%d. ", counter)
					for gc := c.FirstChild; gc != nil; gc = gc.NextSibling {
						nodeToMarkdown(b, gc, depth+1)
					}
					b.WriteString("\n")
					counter++
				}
			}
			return
		case "li":
			b.WriteString("- ")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				nodeToMarkdown(b, c, depth+1)
			}
			b.WriteString("\n")
			return
		case "blockquote":
			b.WriteString("\n> ")
			text := getChildrenText(n)
			b.WriteString(strings.ReplaceAll(text, "\n", "\n> "))
			b.WriteString("\n")
			return
		case "hr":
			b.WriteString("\n---\n")
			return
		case "script", "style", "head":
			return
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		nodeToMarkdown(b, c, depth)
	}
}

// writeChildrenText writes the text content of all children.
func writeChildrenText(b *strings.Builder, n *html.Node) {
	b.WriteString(getChildrenText(n))
}

// getChildrenText returns the concatenated text content of all children.
func getChildrenText(n *html.Node) string {
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		} else {
			b.WriteString(getChildrenText(c))
		}
	}
	return b.String()
}

// jsonToMarkdown converts JSON content to a formatted markdown document.
func jsonToMarkdown(content string) (string, error) {
	var data any
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", core.E("collect.jsonToMarkdown", "failed to parse JSON", err)
	}

	var b strings.Builder
	b.WriteString("# Data\n\n")
	jsonValueToMarkdown(&b, data, 0)
	return strings.TrimSpace(b.String()), nil
}

// jsonValueToMarkdown recursively formats a JSON value as markdown.
func jsonValueToMarkdown(b *strings.Builder, data any, depth int) {
	switch v := data.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			val := v[key]
			indent := strings.Repeat("  ", depth)
			switch child := val.(type) {
			case map[string]any, []any:
				fmt.Fprintf(b, "%s- **%s:**\n", indent, key)
				jsonValueToMarkdown(b, child, depth+1)
			default:
				fmt.Fprintf(b, "%s- **%s:** %v\n", indent, key, val)
			}
		}
	case []any:
		for i, item := range v {
			indent := strings.Repeat("  ", depth)
			switch child := item.(type) {
			case map[string]any, []any:
				fmt.Fprintf(b, "%s- Item %d:\n", indent, i+1)
				jsonValueToMarkdown(b, child, depth+1)
			default:
				fmt.Fprintf(b, "%s- %v\n", indent, item)
			}
		}
	default:
		indent := strings.Repeat("  ", depth)
		fmt.Fprintf(b, "%s%v\n", indent, data)
	}
}

// HTMLToMarkdown is exported for testing.
func HTMLToMarkdown(content string) (string, error) {
	return htmlToMarkdown(content)
}

// JSONToMarkdown is exported for testing.
func JSONToMarkdown(content string) (string, error) {
	return jsonToMarkdown(content)
}
