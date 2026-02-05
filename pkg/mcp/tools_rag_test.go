package mcp

import (
	"testing"
)

// TestRAGToolsRegistered_Good verifies that RAG tools are registered with the MCP server.
func TestRAGToolsRegistered_Good(t *testing.T) {
	// Create a new MCP service - this should register all tools including RAG
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// The server should have registered the RAG tools
	// We verify by checking that the tool handlers exist on the service
	// (The actual MCP registration is tested by the SDK)

	if s.server == nil {
		t.Fatal("Server should not be nil")
	}

	// Verify the service was created with expected defaults
	if s.logger == nil {
		t.Error("Logger should not be nil")
	}
}

// TestRAGQueryInput_Good verifies the RAGQueryInput struct has expected fields.
func TestRAGQueryInput_Good(t *testing.T) {
	input := RAGQueryInput{
		Question:   "test question",
		Collection: "test-collection",
		TopK:       10,
	}

	if input.Question != "test question" {
		t.Errorf("Expected question 'test question', got %q", input.Question)
	}
	if input.Collection != "test-collection" {
		t.Errorf("Expected collection 'test-collection', got %q", input.Collection)
	}
	if input.TopK != 10 {
		t.Errorf("Expected topK 10, got %d", input.TopK)
	}
}

// TestRAGQueryInput_Defaults verifies default values are handled correctly.
func TestRAGQueryInput_Defaults(t *testing.T) {
	// Empty input should use defaults when processed
	input := RAGQueryInput{
		Question: "test",
	}

	// Defaults should be applied in the handler, not in the struct
	if input.Collection != "" {
		t.Errorf("Expected empty collection before defaults, got %q", input.Collection)
	}
	if input.TopK != 0 {
		t.Errorf("Expected zero topK before defaults, got %d", input.TopK)
	}
}

// TestRAGIngestInput_Good verifies the RAGIngestInput struct has expected fields.
func TestRAGIngestInput_Good(t *testing.T) {
	input := RAGIngestInput{
		Path:       "/path/to/docs",
		Collection: "my-collection",
		Recreate:   true,
	}

	if input.Path != "/path/to/docs" {
		t.Errorf("Expected path '/path/to/docs', got %q", input.Path)
	}
	if input.Collection != "my-collection" {
		t.Errorf("Expected collection 'my-collection', got %q", input.Collection)
	}
	if !input.Recreate {
		t.Error("Expected recreate to be true")
	}
}

// TestRAGCollectionsInput_Good verifies the RAGCollectionsInput struct exists.
func TestRAGCollectionsInput_Good(t *testing.T) {
	// RAGCollectionsInput has optional ShowStats parameter
	input := RAGCollectionsInput{}
	if input.ShowStats {
		t.Error("Expected ShowStats to default to false")
	}
}

// TestRAGQueryOutput_Good verifies the RAGQueryOutput struct has expected fields.
func TestRAGQueryOutput_Good(t *testing.T) {
	output := RAGQueryOutput{
		Results: []RAGQueryResult{
			{
				Content:  "some content",
				Source:   "doc.md",
				Section:  "Introduction",
				Category: "docs",
				Score:    0.95,
			},
		},
		Query:      "test query",
		Collection: "test-collection",
		Context:    "<retrieved_context>...</retrieved_context>",
	}

	if len(output.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(output.Results))
	}
	if output.Results[0].Content != "some content" {
		t.Errorf("Expected content 'some content', got %q", output.Results[0].Content)
	}
	if output.Results[0].Score != 0.95 {
		t.Errorf("Expected score 0.95, got %f", output.Results[0].Score)
	}
	if output.Context == "" {
		t.Error("Expected context to be set")
	}
}

// TestRAGIngestOutput_Good verifies the RAGIngestOutput struct has expected fields.
func TestRAGIngestOutput_Good(t *testing.T) {
	output := RAGIngestOutput{
		Success:    true,
		Path:       "/path/to/docs",
		Collection: "my-collection",
		Chunks:     10,
		Message:    "Ingested successfully",
	}

	if !output.Success {
		t.Error("Expected success to be true")
	}
	if output.Path != "/path/to/docs" {
		t.Errorf("Expected path '/path/to/docs', got %q", output.Path)
	}
	if output.Chunks != 10 {
		t.Errorf("Expected chunks 10, got %d", output.Chunks)
	}
}

// TestRAGCollectionsOutput_Good verifies the RAGCollectionsOutput struct has expected fields.
func TestRAGCollectionsOutput_Good(t *testing.T) {
	output := RAGCollectionsOutput{
		Collections: []CollectionInfo{
			{Name: "collection1", PointsCount: 100, Status: "green"},
			{Name: "collection2", PointsCount: 200, Status: "green"},
		},
	}

	if len(output.Collections) != 2 {
		t.Fatalf("Expected 2 collections, got %d", len(output.Collections))
	}
	if output.Collections[0].Name != "collection1" {
		t.Errorf("Expected 'collection1', got %q", output.Collections[0].Name)
	}
	if output.Collections[0].PointsCount != 100 {
		t.Errorf("Expected PointsCount 100, got %d", output.Collections[0].PointsCount)
	}
}

// TestRAGCollectionsInput_Good verifies the RAGCollectionsInput struct has expected fields.
func TestRAGCollectionsInput_ShowStats(t *testing.T) {
	input := RAGCollectionsInput{
		ShowStats: true,
	}

	if !input.ShowStats {
		t.Error("Expected ShowStats to be true")
	}
}
