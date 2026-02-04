package rag

import (
	"context"
	"fmt"

	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/rag"
	"github.com/spf13/cobra"
)

var (
	queryCollection string
	limit           int
	threshold       float32
	category        string
	format          string
)

var queryCmd = &cobra.Command{
	Use:   "query [question]",
	Short: i18n.T("cmd.rag.query.short"),
	Long:  i18n.T("cmd.rag.query.long"),
	Args:  cobra.ExactArgs(1),
	RunE:  runQuery,
}

func runQuery(cmd *cobra.Command, args []string) error {
	question := args[0]
	ctx := context.Background()

	// Connect to Qdrant
	qdrantClient, err := rag.NewQdrantClient(rag.QdrantConfig{
		Host:   qdrantHost,
		Port:   qdrantPort,
		UseTLS: false,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Qdrant: %w", err)
	}
	defer func() { _ = qdrantClient.Close() }()

	// Connect to Ollama
	ollamaClient, err := rag.NewOllamaClient(rag.OllamaConfig{
		Host:  ollamaHost,
		Port:  ollamaPort,
		Model: model,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}

	// Configure query
	if limit < 0 {
		limit = 0
	}
	cfg := rag.QueryConfig{
		Collection: queryCollection,
		Limit:      uint64(limit),
		Threshold:  threshold,
		Category:   category,
	}

	// Run query
	results, err := rag.Query(ctx, qdrantClient, ollamaClient, question, cfg)
	if err != nil {
		return err
	}

	// Format output
	switch format {
	case "json":
		fmt.Println(rag.FormatResultsJSON(results))
	case "context":
		fmt.Println(rag.FormatResultsContext(results))
	default:
		fmt.Println(rag.FormatResultsText(results))
	}

	return nil
}

// QueryDocs is exported for use by other packages (e.g., MCP).
func QueryDocs(ctx context.Context, question, collectionName string, topK int) ([]rag.QueryResult, error) {
	qdrantClient, err := rag.NewQdrantClient(rag.DefaultQdrantConfig())
	if err != nil {
		return nil, err
	}
	defer func() { _ = qdrantClient.Close() }()

	ollamaClient, err := rag.NewOllamaClient(rag.DefaultOllamaConfig())
	if err != nil {
		return nil, err
	}

	cfg := rag.DefaultQueryConfig()
	cfg.Collection = collectionName
	cfg.Limit = uint64(topK)

	return rag.Query(ctx, qdrantClient, ollamaClient, question, cfg)
}

// QueryDocsContext is exported and returns context-formatted results.
func QueryDocsContext(ctx context.Context, question, collectionName string, topK int) (string, error) {
	results, err := QueryDocs(ctx, question, collectionName, topK)
	if err != nil {
		return "", err
	}
	return rag.FormatResultsContext(results), nil
}
