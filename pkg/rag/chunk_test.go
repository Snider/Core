package rag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunkMarkdown_Good_SmallSection(t *testing.T) {
	text := `# Title

This is a small section that fits in one chunk.
`
	chunks := ChunkMarkdown(text, DefaultChunkConfig())

	assert.Len(t, chunks, 1)
	assert.Contains(t, chunks[0].Text, "small section")
}

func TestChunkMarkdown_Good_MultipleSections(t *testing.T) {
	text := `# Main Title

Introduction paragraph.

## Section One

Content for section one.

## Section Two

Content for section two.
`
	chunks := ChunkMarkdown(text, DefaultChunkConfig())

	assert.GreaterOrEqual(t, len(chunks), 2)
}

func TestChunkMarkdown_Good_LargeSection(t *testing.T) {
	// Create a section larger than chunk size
	text := `## Large Section

` + repeatString("This is a test paragraph with some content. ", 50)

	cfg := ChunkConfig{Size: 200, Overlap: 20}
	chunks := ChunkMarkdown(text, cfg)

	assert.Greater(t, len(chunks), 1)
	for _, chunk := range chunks {
		assert.NotEmpty(t, chunk.Text)
		assert.Equal(t, "Large Section", chunk.Section)
	}
}

func TestChunkMarkdown_Good_ExtractsTitle(t *testing.T) {
	text := `## My Section Title

Some content here.
`
	chunks := ChunkMarkdown(text, DefaultChunkConfig())

	assert.Len(t, chunks, 1)
	assert.Equal(t, "My Section Title", chunks[0].Section)
}

func TestCategory_Good_UIComponent(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"docs/flux/button.md", "ui-component"},
		{"ui/components/modal.md", "ui-component"},
		{"brand/vi-personality.md", "brand"},
		{"mascot/expressions.md", "brand"},
		{"product-brief.md", "product-brief"},
		{"tasks/2024-01-15-feature.md", "task"},
		{"plans/architecture.md", "task"},
		{"architecture/migration.md", "architecture"},
		{"docs/api.md", "documentation"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.expected, Category(tc.path))
		})
	}
}

func TestChunkID_Good_Deterministic(t *testing.T) {
	id1 := ChunkID("test.md", 0, "hello world")
	id2 := ChunkID("test.md", 0, "hello world")

	assert.Equal(t, id1, id2)
}

func TestChunkID_Good_DifferentForDifferentInputs(t *testing.T) {
	id1 := ChunkID("test.md", 0, "hello world")
	id2 := ChunkID("test.md", 1, "hello world")
	id3 := ChunkID("other.md", 0, "hello world")

	assert.NotEqual(t, id1, id2)
	assert.NotEqual(t, id1, id3)
}

func TestShouldProcess_Good_MarkdownFiles(t *testing.T) {
	assert.True(t, ShouldProcess("doc.md"))
	assert.True(t, ShouldProcess("doc.markdown"))
	assert.True(t, ShouldProcess("doc.txt"))
	assert.False(t, ShouldProcess("doc.go"))
	assert.False(t, ShouldProcess("doc.py"))
	assert.False(t, ShouldProcess("doc"))
}

// Helper function
func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
