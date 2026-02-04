package ai

import (
	"context"
	"time"

	"github.com/host-uk/core/pkg/rag"
)

// TaskInfo carries the minimal task data needed for RAG queries,
// avoiding a direct dependency on pkg/agentic (which imports pkg/ai).
type TaskInfo struct {
	Title       string
	Description string
}

// QueryRAGForTask queries Qdrant for documentation relevant to a task.
// It builds a query from the task title and description, queries with
// sensible defaults, and returns formatted context. Returns "" on any
// error (e.g. Qdrant/Ollama not running) for graceful degradation.
func QueryRAGForTask(task TaskInfo) string {
	query := task.Title + " " + task.Description

	// Truncate to 500 runes to keep the embedding focused.
	runes := []rune(query)
	if len(runes) > 500 {
		query = string(runes[:500])
	}

	qdrantCfg := rag.DefaultQdrantConfig()
	qdrantClient, err := rag.NewQdrantClient(qdrantCfg)
	if err != nil {
		return ""
	}
	defer func() { _ = qdrantClient.Close() }()

	ollamaCfg := rag.DefaultOllamaConfig()
	ollamaClient, err := rag.NewOllamaClient(ollamaCfg)
	if err != nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queryCfg := rag.QueryConfig{
		Collection: "hostuk-docs",
		Limit:      3,
		Threshold:  0.5,
	}

	results, err := rag.Query(ctx, qdrantClient, ollamaClient, query, queryCfg)
	if err != nil {
		return ""
	}

	return rag.FormatResultsContext(results)
}
