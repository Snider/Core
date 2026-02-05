package mcp

import (
	"context"
	"fmt"

	ragcmd "github.com/host-uk/core/internal/cmd/rag"
	"github.com/host-uk/core/pkg/log"
	"github.com/host-uk/core/pkg/rag"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Default values for RAG operations.
const (
	DefaultRAGCollection = "hostuk-docs"
	DefaultRAGTopK       = 5
)

// RAGQueryInput contains parameters for querying the RAG vector database.
type RAGQueryInput struct {
	Question   string `json:"question"`             // The question or search query
	Collection string `json:"collection,omitempty"` // Collection name (default: hostuk-docs)
	TopK       int    `json:"topK,omitempty"`       // Number of results to return (default: 5)
}

// RAGQueryResult represents a single query result.
type RAGQueryResult struct {
	Content    string  `json:"content"`
	Source     string  `json:"source"`
	Section    string  `json:"section,omitempty"`
	Category   string  `json:"category,omitempty"`
	ChunkIndex int     `json:"chunkIndex,omitempty"`
	Score      float32 `json:"score"`
}

// RAGQueryOutput contains the results of a RAG query.
type RAGQueryOutput struct {
	Results    []RAGQueryResult `json:"results"`
	Query      string           `json:"query"`
	Collection string           `json:"collection"`
	Context    string           `json:"context"`
}

// RAGIngestInput contains parameters for ingesting documents into the RAG database.
type RAGIngestInput struct {
	Path       string `json:"path"`                 // File or directory path to ingest
	Collection string `json:"collection,omitempty"` // Collection name (default: hostuk-docs)
	Recreate   bool   `json:"recreate,omitempty"`   // Whether to recreate the collection
}

// RAGIngestOutput contains the result of a RAG ingest operation.
type RAGIngestOutput struct {
	Success    bool   `json:"success"`
	Path       string `json:"path"`
	Collection string `json:"collection"`
	Chunks     int    `json:"chunks"`
	Message    string `json:"message,omitempty"`
}

// RAGCollectionsInput contains parameters for listing collections.
type RAGCollectionsInput struct {
	ShowStats bool `json:"show_stats,omitempty"` // Include collection stats (point count, status)
}

// CollectionInfo contains information about a collection.
type CollectionInfo struct {
	Name        string `json:"name"`
	PointsCount uint64 `json:"points_count"`
	Status      string `json:"status"`
}

// RAGCollectionsOutput contains the list of available collections.
type RAGCollectionsOutput struct {
	Collections []CollectionInfo `json:"collections"`
}

// registerRAGTools adds RAG tools to the MCP server.
func (s *Service) registerRAGTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rag_query",
		Description: "Query the RAG vector database for relevant documentation. Returns semantically similar content based on the query.",
	}, s.ragQuery)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "rag_ingest",
		Description: "Ingest documents into the RAG vector database. Supports both single files and directories.",
	}, s.ragIngest)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "rag_collections",
		Description: "List all available collections in the RAG vector database.",
	}, s.ragCollections)
}

// ragQuery handles the rag_query tool call.
func (s *Service) ragQuery(ctx context.Context, req *mcp.CallToolRequest, input RAGQueryInput) (*mcp.CallToolResult, RAGQueryOutput, error) {
	// Apply defaults
	collection := input.Collection
	if collection == "" {
		collection = DefaultRAGCollection
	}
	topK := input.TopK
	if topK <= 0 {
		topK = DefaultRAGTopK
	}

	s.logger.Info("MCP tool execution", "tool", "rag_query", "question", input.Question, "collection", collection, "topK", topK, "user", log.Username())

	// Validate input
	if input.Question == "" {
		return nil, RAGQueryOutput{}, fmt.Errorf("question cannot be empty")
	}

	// Call the RAG query function
	results, err := ragcmd.QueryDocs(ctx, input.Question, collection, topK)
	if err != nil {
		log.Error("mcp: rag query failed", "question", input.Question, "collection", collection, "err", err)
		return nil, RAGQueryOutput{}, fmt.Errorf("failed to query RAG: %w", err)
	}

	// Convert results
	output := RAGQueryOutput{
		Results:    make([]RAGQueryResult, len(results)),
		Query:      input.Question,
		Collection: collection,
		Context:    rag.FormatResultsContext(results),
	}
	for i, r := range results {
		output.Results[i] = RAGQueryResult{
			Content:    r.Text,
			Source:     r.Source,
			Section:    r.Section,
			Category:   r.Category,
			ChunkIndex: r.ChunkIndex,
			Score:      r.Score,
		}
	}

	return nil, output, nil
}

// ragIngest handles the rag_ingest tool call.
func (s *Service) ragIngest(ctx context.Context, req *mcp.CallToolRequest, input RAGIngestInput) (*mcp.CallToolResult, RAGIngestOutput, error) {
	// Apply defaults
	collection := input.Collection
	if collection == "" {
		collection = DefaultRAGCollection
	}

	s.logger.Security("MCP tool execution", "tool", "rag_ingest", "path", input.Path, "collection", collection, "recreate", input.Recreate, "user", log.Username())

	// Validate input
	if input.Path == "" {
		return nil, RAGIngestOutput{}, fmt.Errorf("path cannot be empty")
	}

	// Check if path is a file or directory using the medium
	info, err := s.medium.Stat(input.Path)
	if err != nil {
		log.Error("mcp: rag ingest stat failed", "path", input.Path, "err", err)
		return nil, RAGIngestOutput{}, fmt.Errorf("failed to access path: %w", err)
	}

	var message string
	var chunks int
	if info.IsDir() {
		// Ingest directory
		err = ragcmd.IngestDirectory(ctx, input.Path, collection, input.Recreate)
		if err != nil {
			log.Error("mcp: rag ingest directory failed", "path", input.Path, "collection", collection, "err", err)
			return nil, RAGIngestOutput{}, fmt.Errorf("failed to ingest directory: %w", err)
		}
		message = fmt.Sprintf("Successfully ingested directory %s into collection %s", input.Path, collection)
	} else {
		// Ingest single file
		chunks, err = ragcmd.IngestFile(ctx, input.Path, collection)
		if err != nil {
			log.Error("mcp: rag ingest file failed", "path", input.Path, "collection", collection, "err", err)
			return nil, RAGIngestOutput{}, fmt.Errorf("failed to ingest file: %w", err)
		}
		message = fmt.Sprintf("Successfully ingested file %s (%d chunks) into collection %s", input.Path, chunks, collection)
	}

	return nil, RAGIngestOutput{
		Success:    true,
		Path:       input.Path,
		Collection: collection,
		Chunks:     chunks,
		Message:    message,
	}, nil
}

// ragCollections handles the rag_collections tool call.
func (s *Service) ragCollections(ctx context.Context, req *mcp.CallToolRequest, input RAGCollectionsInput) (*mcp.CallToolResult, RAGCollectionsOutput, error) {
	s.logger.Info("MCP tool execution", "tool", "rag_collections", "show_stats", input.ShowStats, "user", log.Username())

	// Create Qdrant client with default config
	qdrantClient, err := rag.NewQdrantClient(rag.DefaultQdrantConfig())
	if err != nil {
		log.Error("mcp: rag collections connect failed", "err", err)
		return nil, RAGCollectionsOutput{}, fmt.Errorf("failed to connect to Qdrant: %w", err)
	}
	defer func() { _ = qdrantClient.Close() }()

	// List collections
	collectionNames, err := qdrantClient.ListCollections(ctx)
	if err != nil {
		log.Error("mcp: rag collections list failed", "err", err)
		return nil, RAGCollectionsOutput{}, fmt.Errorf("failed to list collections: %w", err)
	}

	// Build collection info list
	collections := make([]CollectionInfo, len(collectionNames))
	for i, name := range collectionNames {
		collections[i] = CollectionInfo{Name: name}

		// Fetch stats if requested
		if input.ShowStats {
			info, err := qdrantClient.CollectionInfo(ctx, name)
			if err != nil {
				log.Error("mcp: rag collection info failed", "collection", name, "err", err)
				// Continue with defaults on error
				continue
			}
			if info.PointsCount != nil {
				collections[i].PointsCount = *info.PointsCount
			}
			collections[i].Status = info.Status.String()
		}
	}

	return nil, RAGCollectionsOutput{
		Collections: collections,
	}, nil
}
