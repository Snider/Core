package rag

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/host-uk/core/pkg/log"
)

// QueryConfig holds query configuration.
type QueryConfig struct {
	Collection string
	Limit      uint64
	Threshold  float32 // Minimum similarity score (0-1)
	Category   string  // Filter by category
}

// DefaultQueryConfig returns default query configuration.
func DefaultQueryConfig() QueryConfig {
	return QueryConfig{
		Collection: "hostuk-docs",
		Limit:      5,
		Threshold:  0.5,
	}
}

// QueryResult represents a query result with metadata.
type QueryResult struct {
	Text       string
	Source     string
	Section    string
	Category   string
	ChunkIndex int
	Score      float32
}

// Query searches for similar documents in Qdrant.
func Query(ctx context.Context, qdrant *QdrantClient, ollama *OllamaClient, query string, cfg QueryConfig) ([]QueryResult, error) {
	// Generate embedding for query
	embedding, err := ollama.Embed(ctx, query)
	if err != nil {
		return nil, log.E("rag.Query", "error generating query embedding", err)
	}

	// Build filter
	var filter map[string]string
	if cfg.Category != "" {
		filter = map[string]string{"category": cfg.Category}
	}

	// Search Qdrant
	results, err := qdrant.Search(ctx, cfg.Collection, embedding, cfg.Limit, filter)
	if err != nil {
		return nil, log.E("rag.Query", "error searching", err)
	}

	// Convert and filter by threshold
	var queryResults []QueryResult
	for _, r := range results {
		if r.Score < cfg.Threshold {
			continue
		}

		qr := QueryResult{
			Score: r.Score,
		}

		// Extract payload fields
		if text, ok := r.Payload["text"].(string); ok {
			qr.Text = text
		}
		if source, ok := r.Payload["source"].(string); ok {
			qr.Source = source
		}
		if section, ok := r.Payload["section"].(string); ok {
			qr.Section = section
		}
		if category, ok := r.Payload["category"].(string); ok {
			qr.Category = category
		}
		// Handle chunk_index from various types (JSON unmarshaling produces float64)
		switch idx := r.Payload["chunk_index"].(type) {
		case int64:
			qr.ChunkIndex = int(idx)
		case float64:
			qr.ChunkIndex = int(idx)
		case int:
			qr.ChunkIndex = idx
		}

		queryResults = append(queryResults, qr)
	}

	return queryResults, nil
}

// FormatResultsText formats query results as plain text.
func FormatResultsText(results []QueryResult) string {
	if len(results) == 0 {
		return "No results found."
	}

	var sb strings.Builder
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("\n--- Result %d (score: %.2f) ---\n", i+1, r.Score))
		sb.WriteString(fmt.Sprintf("Source: %s\n", r.Source))
		if r.Section != "" {
			sb.WriteString(fmt.Sprintf("Section: %s\n", r.Section))
		}
		sb.WriteString(fmt.Sprintf("Category: %s\n\n", r.Category))
		sb.WriteString(r.Text)
		sb.WriteString("\n")
	}
	return sb.String()
}

// FormatResultsContext formats query results for LLM context injection.
func FormatResultsContext(results []QueryResult) string {
	if len(results) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<retrieved_context>\n")
	for _, r := range results {
		// Escape XML special characters to prevent malformed output
		fmt.Fprintf(&sb, "<document source=\"%s\" section=\"%s\" category=\"%s\">\n",
			html.EscapeString(r.Source),
			html.EscapeString(r.Section),
			html.EscapeString(r.Category))
		sb.WriteString(html.EscapeString(r.Text))
		sb.WriteString("\n</document>\n\n")
	}
	sb.WriteString("</retrieved_context>")
	return sb.String()
}

// FormatResultsJSON formats query results as JSON-like output.
func FormatResultsJSON(results []QueryResult) string {
	if len(results) == 0 {
		return "[]"
	}

	var sb strings.Builder
	sb.WriteString("[\n")
	for i, r := range results {
		sb.WriteString("  {\n")
		sb.WriteString(fmt.Sprintf("    \"source\": %q,\n", r.Source))
		sb.WriteString(fmt.Sprintf("    \"section\": %q,\n", r.Section))
		sb.WriteString(fmt.Sprintf("    \"category\": %q,\n", r.Category))
		sb.WriteString(fmt.Sprintf("    \"score\": %.4f,\n", r.Score))
		sb.WriteString(fmt.Sprintf("    \"text\": %q\n", r.Text))
		if i < len(results)-1 {
			sb.WriteString("  },\n")
		} else {
			sb.WriteString("  }\n")
		}
	}
	sb.WriteString("]")
	return sb.String()
}
