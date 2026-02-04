package rag

import (
	"context"
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/rag"
	"github.com/spf13/cobra"
)

var (
	collection   string
	recreate     bool
	chunkSize    int
	chunkOverlap int
)

var ingestCmd = &cobra.Command{
	Use:   "ingest [directory]",
	Short: i18n.T("cmd.rag.ingest.short"),
	Long:  i18n.T("cmd.rag.ingest.long"),
	Args:  cobra.MaximumNArgs(1),
	RunE:  runIngest,
}

func runIngest(cmd *cobra.Command, args []string) error {
	directory := "."
	if len(args) > 0 {
		directory = args[0]
	}

	ctx := context.Background()

	// Connect to Qdrant
	fmt.Printf("Connecting to Qdrant at %s:%d...\n", qdrantHost, qdrantPort)
	qdrantClient, err := rag.NewQdrantClient(rag.QdrantConfig{
		Host:   qdrantHost,
		Port:   qdrantPort,
		UseTLS: false,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Qdrant: %w", err)
	}
	defer qdrantClient.Close()

	if err := qdrantClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("Qdrant health check failed: %w", err)
	}

	// Connect to Ollama
	fmt.Printf("Using embedding model: %s (via %s:%d)\n", model, ollamaHost, ollamaPort)
	ollamaClient, err := rag.NewOllamaClient(rag.OllamaConfig{
		Host:  ollamaHost,
		Port:  ollamaPort,
		Model: model,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}

	if err := ollamaClient.VerifyModel(ctx); err != nil {
		return err
	}

	// Configure ingestion
	if chunkSize <= 0 {
		return fmt.Errorf("chunk-size must be > 0")
	}
	if chunkOverlap < 0 || chunkOverlap >= chunkSize {
		return fmt.Errorf("chunk-overlap must be >= 0 and < chunk-size")
	}

	cfg := rag.IngestConfig{
		Directory:  directory,
		Collection: collection,
		Recreate:   recreate,
		Verbose:    verbose,
		BatchSize:  100,
		Chunk: rag.ChunkConfig{
			Size:    chunkSize,
			Overlap: chunkOverlap,
		},
	}

	// Progress callback
	progress := func(file string, chunks int, total int) {
		if verbose {
			fmt.Printf("  Processed: %s (%d chunks total)\n", file, chunks)
		} else {
			fmt.Printf("\r  %s (%d chunks)    ", cli.DimStyle.Render(file), chunks)
		}
	}

	// Run ingestion
	fmt.Printf("\nIngesting from: %s\n", directory)
	if recreate {
		fmt.Printf("  (recreating collection: %s)\n", collection)
	}

	stats, err := rag.Ingest(ctx, qdrantClient, ollamaClient, cfg, progress)
	if err != nil {
		return err
	}

	// Summary
	fmt.Printf("\n\n%s\n", cli.TitleStyle.Render("Ingestion complete!"))
	fmt.Printf("  Files processed: %d\n", stats.Files)
	fmt.Printf("  Chunks created:  %d\n", stats.Chunks)
	if stats.Errors > 0 {
		fmt.Printf("  Errors:          %s\n", cli.ErrorStyle.Render(fmt.Sprintf("%d", stats.Errors)))
	}
	fmt.Printf("  Collection:      %s\n", collection)

	return nil
}

// IngestDirectory is exported for use by other packages (e.g., MCP).
func IngestDirectory(ctx context.Context, directory, collectionName string, recreateCollection bool) error {
	qdrantClient, err := rag.NewQdrantClient(rag.DefaultQdrantConfig())
	if err != nil {
		return err
	}
	defer qdrantClient.Close()

	if err := qdrantClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("Qdrant health check failed: %w", err)
	}

	ollamaClient, err := rag.NewOllamaClient(rag.DefaultOllamaConfig())
	if err != nil {
		return err
	}

	if err := ollamaClient.VerifyModel(ctx); err != nil {
		return err
	}

	cfg := rag.DefaultIngestConfig()
	cfg.Directory = directory
	cfg.Collection = collectionName
	cfg.Recreate = recreateCollection

	_, err = rag.Ingest(ctx, qdrantClient, ollamaClient, cfg, nil)
	return err
}

// IngestFile is exported for use by other packages (e.g., MCP).
func IngestFile(ctx context.Context, filePath, collectionName string) (int, error) {
	qdrantClient, err := rag.NewQdrantClient(rag.DefaultQdrantConfig())
	if err != nil {
		return 0, err
	}
	defer qdrantClient.Close()

	if err := qdrantClient.HealthCheck(ctx); err != nil {
		return 0, fmt.Errorf("Qdrant health check failed: %w", err)
	}

	ollamaClient, err := rag.NewOllamaClient(rag.DefaultOllamaConfig())
	if err != nil {
		return 0, err
	}

	if err := ollamaClient.VerifyModel(ctx); err != nil {
		return 0, err
	}

	return rag.IngestFile(ctx, qdrantClient, ollamaClient, collectionName, filePath, rag.DefaultChunkConfig())
}
