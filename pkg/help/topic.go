// Package help provides display-agnostic help content management.
package help

// Topic represents a help topic/page.
type Topic struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Path     string    `json:"path"`
	Content  string    `json:"content"`
	Sections []Section `json:"sections"`
	Tags     []string  `json:"tags"`
	Related  []string  `json:"related"`
	Order    int       `json:"order"` // For sorting
}

// Section represents a heading within a topic.
type Section struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Level   int    `json:"level"`
	Line    int    `json:"line"`    // Start line in content (1-indexed)
	Content string `json:"content"` // Content under heading
}

// Frontmatter represents YAML frontmatter metadata.
type Frontmatter struct {
	Title   string   `yaml:"title"`
	Tags    []string `yaml:"tags"`
	Related []string `yaml:"related"`
	Order   int      `yaml:"order"`
}
