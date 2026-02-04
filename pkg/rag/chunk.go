package rag

import (
	"crypto/md5"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// ChunkConfig holds chunking configuration.
type ChunkConfig struct {
	Size    int // Characters per chunk
	Overlap int // Overlap between chunks
}

// DefaultChunkConfig returns default chunking configuration.
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		Size:    500,
		Overlap: 50,
	}
}

// Chunk represents a text chunk with metadata.
type Chunk struct {
	Text    string
	Section string
	Index   int
}

// ChunkMarkdown splits markdown text into chunks by sections and paragraphs.
// Preserves context with configurable overlap.
func ChunkMarkdown(text string, cfg ChunkConfig) []Chunk {
	if cfg.Size <= 0 {
		cfg.Size = 500
	}
	if cfg.Overlap < 0 || cfg.Overlap >= cfg.Size {
		cfg.Overlap = 0
	}

	var chunks []Chunk

	// Split by ## headers
	sections := splitBySections(text)

	chunkIndex := 0
	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		// Extract section title
		lines := strings.SplitN(section, "\n", 2)
		title := ""
		if strings.HasPrefix(lines[0], "#") {
			title = strings.TrimLeft(lines[0], "#")
			title = strings.TrimSpace(title)
		}

		// If section is small enough, yield as-is
		if len(section) <= cfg.Size {
			chunks = append(chunks, Chunk{
				Text:    section,
				Section: title,
				Index:   chunkIndex,
			})
			chunkIndex++
			continue
		}

		// Otherwise, chunk by paragraphs
		paragraphs := splitByParagraphs(section)
		currentChunk := ""

		for _, para := range paragraphs {
			para = strings.TrimSpace(para)
			if para == "" {
				continue
			}

			if len(currentChunk)+len(para)+2 <= cfg.Size {
				if currentChunk != "" {
					currentChunk += "\n\n" + para
				} else {
					currentChunk = para
				}
			} else {
				if currentChunk != "" {
					chunks = append(chunks, Chunk{
						Text:    strings.TrimSpace(currentChunk),
						Section: title,
						Index:   chunkIndex,
					})
					chunkIndex++
				}
				// Start new chunk with overlap from previous (rune-safe for UTF-8)
				runes := []rune(currentChunk)
				if cfg.Overlap > 0 && len(runes) > cfg.Overlap {
					overlapText := string(runes[len(runes)-cfg.Overlap:])
					currentChunk = overlapText + "\n\n" + para
				} else {
					currentChunk = para
				}
			}
		}

		// Don't forget the last chunk
		if strings.TrimSpace(currentChunk) != "" {
			chunks = append(chunks, Chunk{
				Text:    strings.TrimSpace(currentChunk),
				Section: title,
				Index:   chunkIndex,
			})
			chunkIndex++
		}
	}

	return chunks
}

// splitBySections splits text by ## headers while preserving the header with its content.
func splitBySections(text string) []string {
	var sections []string
	lines := strings.Split(text, "\n")

	var currentSection strings.Builder
	for _, line := range lines {
		// Check if this line is a ## header
		if strings.HasPrefix(line, "## ") {
			// Save previous section if exists
			if currentSection.Len() > 0 {
				sections = append(sections, currentSection.String())
				currentSection.Reset()
			}
		}
		currentSection.WriteString(line)
		currentSection.WriteString("\n")
	}

	// Don't forget the last section
	if currentSection.Len() > 0 {
		sections = append(sections, currentSection.String())
	}

	return sections
}

// splitByParagraphs splits text by double newlines.
func splitByParagraphs(text string) []string {
	// Replace multiple newlines with a marker, then split
	normalized := text
	for strings.Contains(normalized, "\n\n\n") {
		normalized = strings.ReplaceAll(normalized, "\n\n\n", "\n\n")
	}
	return strings.Split(normalized, "\n\n")
}

// Category determines the document category from file path.
func Category(path string) string {
	lower := strings.ToLower(path)

	switch {
	case strings.Contains(lower, "flux") || strings.Contains(lower, "ui/component"):
		return "ui-component"
	case strings.Contains(lower, "brand") || strings.Contains(lower, "mascot"):
		return "brand"
	case strings.Contains(lower, "brief"):
		return "product-brief"
	case strings.Contains(lower, "help") || strings.Contains(lower, "draft"):
		return "help-doc"
	case strings.Contains(lower, "task") || strings.Contains(lower, "plan"):
		return "task"
	case strings.Contains(lower, "architecture") || strings.Contains(lower, "migration"):
		return "architecture"
	default:
		return "documentation"
	}
}

// ChunkID generates a unique ID for a chunk.
func ChunkID(path string, index int, text string) string {
	// Use first 100 runes of text for uniqueness (rune-safe for UTF-8)
	runes := []rune(text)
	if len(runes) > 100 {
		runes = runes[:100]
	}
	textPart := string(runes)
	data := fmt.Sprintf("%s:%d:%s", path, index, textPart)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// FileExtensions returns the file extensions to process.
func FileExtensions() []string {
	return []string{".md", ".markdown", ".txt"}
}

// ShouldProcess checks if a file should be processed based on extension.
func ShouldProcess(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(FileExtensions(), ext)
}
