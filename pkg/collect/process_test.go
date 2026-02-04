package collect

import (
	"context"
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestProcessor_Name_Good(t *testing.T) {
	p := &Processor{Source: "github"}
	assert.Equal(t, "process:github", p.Name())
}

func TestProcessor_Process_Bad_NoDir(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")

	p := &Processor{Source: "test"}
	_, err := p.Process(context.Background(), cfg)
	assert.Error(t, err)
}

func TestProcessor_Process_Good_DryRun(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	p := &Processor{Source: "test", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Items)
}

func TestProcessor_Process_Good_HTMLFiles(t *testing.T) {
	m := io.NewMockMedium()
	m.Dirs["/input"] = true
	m.Files["/input/page.html"] = `<html><body><h1>Hello</h1><p>World</p></body></html>`

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &Processor{Source: "test", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Items)
	assert.Len(t, result.Files, 1)

	content, err := m.Read("/output/processed/test/page.md")
	assert.NoError(t, err)
	assert.Contains(t, content, "# Hello")
	assert.Contains(t, content, "World")
}

func TestProcessor_Process_Good_JSONFiles(t *testing.T) {
	m := io.NewMockMedium()
	m.Dirs["/input"] = true
	m.Files["/input/data.json"] = `{"name": "Bitcoin", "price": 42000}`

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &Processor{Source: "market", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Items)

	content, err := m.Read("/output/processed/market/data.md")
	assert.NoError(t, err)
	assert.Contains(t, content, "# Data")
	assert.Contains(t, content, "Bitcoin")
}

func TestProcessor_Process_Good_MarkdownPassthrough(t *testing.T) {
	m := io.NewMockMedium()
	m.Dirs["/input"] = true
	m.Files["/input/readme.md"] = "# Already Markdown\n\nThis is already formatted."

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &Processor{Source: "docs", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Items)

	content, err := m.Read("/output/processed/docs/readme.md")
	assert.NoError(t, err)
	assert.Contains(t, content, "# Already Markdown")
}

func TestProcessor_Process_Good_SkipUnknownTypes(t *testing.T) {
	m := io.NewMockMedium()
	m.Dirs["/input"] = true
	m.Files["/input/image.png"] = "binary data"
	m.Files["/input/doc.html"] = "<h1>Heading</h1>"

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &Processor{Source: "mixed", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Items)   // Only the HTML file
	assert.Equal(t, 1, result.Skipped) // The PNG file
}

func TestHTMLToMarkdown_Good(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "heading",
			input:    "<h1>Title</h1>",
			contains: []string{"# Title"},
		},
		{
			name:     "paragraph",
			input:    "<p>Hello world</p>",
			contains: []string{"Hello world"},
		},
		{
			name:     "bold",
			input:    "<p><strong>bold text</strong></p>",
			contains: []string{"**bold text**"},
		},
		{
			name:     "italic",
			input:    "<p><em>italic text</em></p>",
			contains: []string{"*italic text*"},
		},
		{
			name:     "code",
			input:    "<p><code>code</code></p>",
			contains: []string{"`code`"},
		},
		{
			name:     "link",
			input:    `<p><a href="https://example.com">Example</a></p>`,
			contains: []string{"[Example](https://example.com)"},
		},
		{
			name:     "nested headings",
			input:    "<h2>Section</h2><h3>Subsection</h3>",
			contains: []string{"## Section", "### Subsection"},
		},
		{
			name:     "pre block",
			input:    "<pre>func main() {}</pre>",
			contains: []string{"```", "func main() {}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HTMLToMarkdown(tt.input)
			assert.NoError(t, err)
			for _, s := range tt.contains {
				assert.Contains(t, result, s)
			}
		})
	}
}

func TestHTMLToMarkdown_Good_StripsScripts(t *testing.T) {
	input := `<html><head><script>alert('xss')</script></head><body><p>Clean</p></body></html>`
	result, err := HTMLToMarkdown(input)
	assert.NoError(t, err)
	assert.Contains(t, result, "Clean")
	assert.NotContains(t, result, "alert")
	assert.NotContains(t, result, "script")
}

func TestJSONToMarkdown_Good(t *testing.T) {
	input := `{"name": "test", "count": 42}`
	result, err := JSONToMarkdown(input)
	assert.NoError(t, err)
	assert.Contains(t, result, "# Data")
	assert.Contains(t, result, "test")
	assert.Contains(t, result, "42")
}

func TestJSONToMarkdown_Good_Array(t *testing.T) {
	input := `[{"id": 1}, {"id": 2}]`
	result, err := JSONToMarkdown(input)
	assert.NoError(t, err)
	assert.Contains(t, result, "# Data")
}

func TestJSONToMarkdown_Bad_InvalidJSON(t *testing.T) {
	_, err := JSONToMarkdown("not json")
	assert.Error(t, err)
}
