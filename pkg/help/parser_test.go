package help

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateID_Good(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple title",
			input:    "Getting Started",
			expected: "getting-started",
		},
		{
			name:     "already lowercase",
			input:    "installation",
			expected: "installation",
		},
		{
			name:     "multiple spaces",
			input:    "Quick   Start   Guide",
			expected: "quick-start-guide",
		},
		{
			name:     "with numbers",
			input:    "Chapter 1 Introduction",
			expected: "chapter-1-introduction",
		},
		{
			name:     "special characters",
			input:    "What's New? (v2.0)",
			expected: "whats-new-v20",
		},
		{
			name:     "underscores",
			input:    "config_file_reference",
			expected: "config-file-reference",
		},
		{
			name:     "hyphens preserved",
			input:    "pre-commit hooks",
			expected: "pre-commit-hooks",
		},
		{
			name:     "leading trailing spaces",
			input:    "  Trimmed Title  ",
			expected: "trimmed-title",
		},
		{
			name:     "unicode letters",
			input:    "Configuración Básica",
			expected: "configuración-básica",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFrontmatter_Good(t *testing.T) {
	content := `---
title: Getting Started
tags: [intro, setup]
order: 1
related:
  - installation
  - configuration
---

# Welcome

This is the content.
`

	fm, body := ExtractFrontmatter(content)

	assert.NotNil(t, fm)
	assert.Equal(t, "Getting Started", fm.Title)
	assert.Equal(t, []string{"intro", "setup"}, fm.Tags)
	assert.Equal(t, 1, fm.Order)
	assert.Equal(t, []string{"installation", "configuration"}, fm.Related)
	assert.Contains(t, body, "# Welcome")
	assert.Contains(t, body, "This is the content.")
}

func TestExtractFrontmatter_Good_NoFrontmatter(t *testing.T) {
	content := `# Just a Heading

Some content here.
`

	fm, body := ExtractFrontmatter(content)

	assert.Nil(t, fm)
	assert.Equal(t, content, body)
}

func TestExtractFrontmatter_Good_CRLF(t *testing.T) {
	// Content with CRLF line endings (Windows-style)
	content := "---\r\ntitle: CRLF Test\r\n---\r\n\r\n# Content"

	fm, body := ExtractFrontmatter(content)

	assert.NotNil(t, fm)
	assert.Equal(t, "CRLF Test", fm.Title)
	assert.Contains(t, body, "# Content")
}

func TestExtractFrontmatter_Good_Empty(t *testing.T) {
	// Empty frontmatter block
	content := "---\n---\n# Content"

	fm, body := ExtractFrontmatter(content)

	// Empty frontmatter should parse successfully
	assert.NotNil(t, fm)
	assert.Equal(t, "", fm.Title)
	assert.Contains(t, body, "# Content")
}

func TestExtractFrontmatter_Bad_InvalidYAML(t *testing.T) {
	content := `---
title: [invalid yaml
---

# Content
`

	fm, body := ExtractFrontmatter(content)

	// Invalid YAML should return nil frontmatter and original content
	assert.Nil(t, fm)
	assert.Equal(t, content, body)
}

func TestExtractSections_Good(t *testing.T) {
	content := `# Main Title

Introduction paragraph.

## Installation

Install instructions here.
More details.

### Prerequisites

You need these things.

## Configuration

Config info here.
`

	sections := ExtractSections(content)

	assert.Len(t, sections, 4)

	// Main Title (H1)
	assert.Equal(t, "main-title", sections[0].ID)
	assert.Equal(t, "Main Title", sections[0].Title)
	assert.Equal(t, 1, sections[0].Level)
	assert.Equal(t, 1, sections[0].Line)
	assert.Contains(t, sections[0].Content, "Introduction paragraph.")

	// Installation (H2)
	assert.Equal(t, "installation", sections[1].ID)
	assert.Equal(t, "Installation", sections[1].Title)
	assert.Equal(t, 2, sections[1].Level)
	assert.Contains(t, sections[1].Content, "Install instructions here.")
	assert.Contains(t, sections[1].Content, "More details.")

	// Prerequisites (H3)
	assert.Equal(t, "prerequisites", sections[2].ID)
	assert.Equal(t, "Prerequisites", sections[2].Title)
	assert.Equal(t, 3, sections[2].Level)
	assert.Contains(t, sections[2].Content, "You need these things.")

	// Configuration (H2)
	assert.Equal(t, "configuration", sections[3].ID)
	assert.Equal(t, "Configuration", sections[3].Title)
	assert.Equal(t, 2, sections[3].Level)
}

func TestExtractSections_Good_AllHeadingLevels(t *testing.T) {
	content := `# H1
## H2
### H3
#### H4
##### H5
###### H6
`

	sections := ExtractSections(content)

	assert.Len(t, sections, 6)
	for i, level := range []int{1, 2, 3, 4, 5, 6} {
		assert.Equal(t, level, sections[i].Level)
	}
}

func TestExtractSections_Good_Empty(t *testing.T) {
	content := `Just plain text.
No headings here.
`

	sections := ExtractSections(content)

	assert.Empty(t, sections)
}

func TestParseTopic_Good(t *testing.T) {
	content := []byte(`---
title: Quick Start Guide
tags: [intro, quickstart]
order: 5
related:
  - installation
---

# Quick Start Guide

Welcome to the guide.

## First Steps

Do this first.

## Next Steps

Then do this.
`)

	topic, err := ParseTopic("docs/quick-start.md", content)

	assert.NoError(t, err)
	assert.NotNil(t, topic)

	// Check metadata from frontmatter
	assert.Equal(t, "quick-start-guide", topic.ID)
	assert.Equal(t, "Quick Start Guide", topic.Title)
	assert.Equal(t, "docs/quick-start.md", topic.Path)
	assert.Equal(t, []string{"intro", "quickstart"}, topic.Tags)
	assert.Equal(t, []string{"installation"}, topic.Related)
	assert.Equal(t, 5, topic.Order)

	// Check sections
	assert.Len(t, topic.Sections, 3)
	assert.Equal(t, "quick-start-guide", topic.Sections[0].ID)
	assert.Equal(t, "first-steps", topic.Sections[1].ID)
	assert.Equal(t, "next-steps", topic.Sections[2].ID)

	// Content should not include frontmatter
	assert.NotContains(t, topic.Content, "---")
	assert.Contains(t, topic.Content, "# Quick Start Guide")
}

func TestParseTopic_Good_NoFrontmatter(t *testing.T) {
	content := []byte(`# Getting Started

This is a simple doc.

## Installation

Install it here.
`)

	topic, err := ParseTopic("getting-started.md", content)

	assert.NoError(t, err)
	assert.NotNil(t, topic)

	// Title should come from first H1
	assert.Equal(t, "Getting Started", topic.Title)
	assert.Equal(t, "getting-started", topic.ID)

	// Sections extracted
	assert.Len(t, topic.Sections, 2)
}

func TestParseTopic_Good_NoHeadings(t *testing.T) {
	content := []byte(`---
title: Plain Content
---

Just some text without any headings.
`)

	topic, err := ParseTopic("plain.md", content)

	assert.NoError(t, err)
	assert.NotNil(t, topic)
	assert.Equal(t, "Plain Content", topic.Title)
	assert.Equal(t, "plain-content", topic.ID)
	assert.Empty(t, topic.Sections)
}

func TestParseTopic_Good_IDFromPath(t *testing.T) {
	content := []byte(`Just content, no frontmatter or headings.`)

	topic, err := ParseTopic("commands/dev-workflow.md", content)

	assert.NoError(t, err)
	assert.NotNil(t, topic)

	// ID and title should be derived from path
	assert.Equal(t, "dev-workflow", topic.ID)
	assert.Equal(t, "", topic.Title) // No title available
}

func TestPathToTitle_Good(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"getting-started.md", "Getting Started"},
		{"commands/dev.md", "Dev"},
		{"path/to/file_name.md", "File Name"},
		{"UPPERCASE.md", "Uppercase"},
		{"no-extension", "No Extension"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := pathToTitle(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
