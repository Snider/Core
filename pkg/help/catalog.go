package help

import (
	"fmt"
)

// Catalog manages help topics.
type Catalog struct {
	topics map[string]*Topic
	index  *searchIndex
}

// DefaultCatalog returns a catalog with built-in topics.
func DefaultCatalog() *Catalog {
	c := &Catalog{
		topics: make(map[string]*Topic),
		index:  newSearchIndex(),
	}

	// Add default topics
	c.Add(&Topic{
		ID:    "getting-started",
		Title: "Getting Started",
		Content: `# Getting Started

Welcome to Core! This CLI tool helps you manage development workflows.

## Common Commands

- core dev: Development workflows
- core setup: Setup repository
- core doctor: Check environment health
- core test: Run tests

## Next Steps

Run 'core help <topic>' to learn more about a specific topic.
`,
	})
	c.Add(&Topic{
		ID:    "config",
		Title: "Configuration",
		Content: `# Configuration

Core is configured via environment variables and config files.

## Environment Variables

- CORE_DEBUG: Enable debug logging
- GITHUB_TOKEN: GitHub API token

## Config Files

Config is stored in ~/.core/config.yaml
`,
	})
	return c
}

// Add adds a topic to the catalog.
func (c *Catalog) Add(t *Topic) {
	c.topics[t.ID] = t
	c.index.Add(t)
}

// List returns all topics.
func (c *Catalog) List() []*Topic {
	var list []*Topic
	for _, t := range c.topics {
		list = append(list, t)
	}
	return list
}

// Search searches for topics.
func (c *Catalog) Search(query string) []*SearchResult {
	return c.index.Search(query)
}

// Get returns a topic by ID.
func (c *Catalog) Get(id string) (*Topic, error) {
	t, ok := c.topics[id]
	if !ok {
		return nil, fmt.Errorf("topic not found: %s", id)
	}
	return t, nil
}
