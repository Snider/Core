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
	listCollections  bool
	showStats        bool
	deleteCollection string
)

var collectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: i18n.T("cmd.rag.collections.short"),
	Long:  i18n.T("cmd.rag.collections.long"),
	RunE:  runCollections,
}

func runCollections(cmd *cobra.Command, args []string) error {
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
	defer qdrantClient.Close()

	// Handle delete
	if deleteCollection != "" {
		exists, err := qdrantClient.CollectionExists(ctx, deleteCollection)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("collection not found: %s", deleteCollection)
		}
		if err := qdrantClient.DeleteCollection(ctx, deleteCollection); err != nil {
			return err
		}
		fmt.Printf("Deleted collection: %s\n", deleteCollection)
		return nil
	}

	// List collections
	collections, err := qdrantClient.ListCollections(ctx)
	if err != nil {
		return err
	}

	if len(collections) == 0 {
		fmt.Println("No collections found.")
		return nil
	}

	fmt.Printf("%s\n\n", cli.TitleStyle.Render("Collections"))

	for _, name := range collections {
		if showStats {
			info, err := qdrantClient.CollectionInfo(ctx, name)
			if err != nil {
				fmt.Printf("  %s (error: %v)\n", name, err)
				continue
			}
			fmt.Printf("  %s\n", cli.ValueStyle.Render(name))
			fmt.Printf("    Points:  %d\n", info.PointsCount)
			fmt.Printf("    Status:  %s\n", info.Status.String())
			fmt.Println()
		} else {
			fmt.Printf("  %s\n", name)
		}
	}

	return nil
}
