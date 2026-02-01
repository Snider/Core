package help

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestTokenize_Good(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple words",
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "mixed case",
			input:    "Hello World",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with punctuation",
			input:    "Hello, world! How are you?",
			expected: []string{"hello", "world", "how", "are", "you"},
		},
		{
			name:     "single characters filtered",
			input:    "a b c hello d",
			expected: []string{"hello"},
		},
		{
			name:     "numbers included",
			input:    "version 2 release",
			expected: []string{"version", "release"},
		},
		{
			name:     "alphanumeric",
			input:    "v2.0 and config123",
			expected: []string{"v2", "and", "config123"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSearchIndex_Add_Good(t *testing.T) {
	idx := newSearchIndex()

	topic := &Topic{
		ID:      "getting-started",
		Title:   "Getting Started",
		Content: "Welcome to the guide.",
		Tags:    []string{"intro", "setup"},
		Sections: []Section{
			{ID: "installation", Title: "Installation", Content: "Install the CLI."},
		},
	}

	idx.Add(topic)

	// Verify topic is stored
	assert.NotNil(t, idx.topics["getting-started"])

	// Verify words are indexed
	assert.Contains(t, idx.index["getting"], "getting-started")
	assert.Contains(t, idx.index["started"], "getting-started")
	assert.Contains(t, idx.index["welcome"], "getting-started")
	assert.Contains(t, idx.index["guide"], "getting-started")
	assert.Contains(t, idx.index["intro"], "getting-started")
	assert.Contains(t, idx.index["setup"], "getting-started")
	assert.Contains(t, idx.index["installation"], "getting-started")
	assert.Contains(t, idx.index["cli"], "getting-started")
}

func TestSearchIndex_Search_Good(t *testing.T) {
	idx := newSearchIndex()

	// Add test topics
	idx.Add(&Topic{
		ID:      "getting-started",
		Title:   "Getting Started",
		Content: "Welcome to the CLI guide. This covers installation and setup.",
		Tags:    []string{"intro"},
	})

	idx.Add(&Topic{
		ID:      "configuration",
		Title:   "Configuration",
		Content: "Configure the CLI using environment variables.",
	})

	idx.Add(&Topic{
		ID:      "commands",
		Title:   "Commands Reference",
		Content: "List of all available commands.",
	})

	t.Run("single word query", func(t *testing.T) {
		results := idx.Search("configuration")
		assert.NotEmpty(t, results)
		assert.Equal(t, "configuration", results[0].Topic.ID)
	})

	t.Run("multi-word query", func(t *testing.T) {
		results := idx.Search("cli guide")
		assert.NotEmpty(t, results)
		// Should match getting-started (has both "cli" and "guide")
		assert.Equal(t, "getting-started", results[0].Topic.ID)
	})

	t.Run("title boost", func(t *testing.T) {
		results := idx.Search("commands")
		assert.NotEmpty(t, results)
		// "commands" appears in title of commands topic
		assert.Equal(t, "commands", results[0].Topic.ID)
	})

	t.Run("partial word matching", func(t *testing.T) {
		results := idx.Search("config")
		assert.NotEmpty(t, results)
		// Should match "configuration" and "configure"
		foundConfig := false
		for _, r := range results {
			if r.Topic.ID == "configuration" {
				foundConfig = true
				break
			}
		}
		assert.True(t, foundConfig, "Should find configuration topic with prefix match")
	})

	t.Run("no results", func(t *testing.T) {
		results := idx.Search("nonexistent")
		assert.Empty(t, results)
	})

	t.Run("empty query", func(t *testing.T) {
		results := idx.Search("")
		assert.Nil(t, results)
	})
}

func TestSearchIndex_Search_Good_WithSections(t *testing.T) {
	idx := newSearchIndex()

	idx.Add(&Topic{
		ID:      "installation",
		Title:   "Installation Guide",
		Content: "Overview of installation process.",
		Sections: []Section{
			{
				ID:      "linux",
				Title:   "Linux Installation",
				Content: "Run apt-get install core on Debian.",
			},
			{
				ID:      "macos",
				Title:   "macOS Installation",
				Content: "Use brew install core on macOS.",
			},
			{
				ID:      "windows",
				Title:   "Windows Installation",
				Content: "Download the installer from the website.",
			},
		},
	})

	t.Run("matches section content", func(t *testing.T) {
		results := idx.Search("debian")
		assert.NotEmpty(t, results)
		assert.Equal(t, "installation", results[0].Topic.ID)
		// Should identify the Linux section as best match
		if results[0].Section != nil {
			assert.Equal(t, "linux", results[0].Section.ID)
		}
	})

	t.Run("matches section title", func(t *testing.T) {
		results := idx.Search("windows")
		assert.NotEmpty(t, results)
		assert.Equal(t, "installation", results[0].Topic.ID)
	})
}

func TestExtractSnippet_Good(t *testing.T) {
	content := `This is the first paragraph with some introduction text.

Here is more content that talks about installation and setup.
The installation process is straightforward.

Finally, some closing remarks about the configuration.`

	t.Run("finds match and extracts context", func(t *testing.T) {
		snippet := extractSnippet(content, []string{"installation"})
		assert.Contains(t, snippet, "installation")
		assert.True(t, len(snippet) <= 200, "Snippet should be reasonably short")
	})

	t.Run("no query words returns start", func(t *testing.T) {
		snippet := extractSnippet(content, nil)
		assert.Contains(t, snippet, "first paragraph")
	})

	t.Run("empty content", func(t *testing.T) {
		snippet := extractSnippet("", []string{"test"})
		assert.Empty(t, snippet)
	})
}

func TestExtractSnippet_Good_UTF8(t *testing.T) {
	// Content with multi-byte UTF-8 characters
	content := "日本語のテキストです。This contains Japanese text. 検索機能をテストします。"

	t.Run("handles multi-byte characters without corruption", func(t *testing.T) {
		snippet := extractSnippet(content, []string{"japanese"})
		// Should not panic or produce invalid UTF-8
		assert.True(t, len(snippet) > 0)
		// Verify the result is valid UTF-8
		assert.True(t, isValidUTF8(snippet), "Snippet should be valid UTF-8")
	})

	t.Run("truncates multi-byte content safely", func(t *testing.T) {
		// Long content that will be truncated
		longContent := strings.Repeat("日本語", 100) // 300 characters
		snippet := extractSnippet(longContent, nil)
		assert.True(t, isValidUTF8(snippet), "Truncated snippet should be valid UTF-8")
	})
}

// isValidUTF8 checks if a string is valid UTF-8
func isValidUTF8(s string) bool {
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			return false
		}
		i += size
	}
	return true
}

func TestCountMatches_Good(t *testing.T) {
	tests := []struct {
		text     string
		words    []string
		expected int
	}{
		{"Hello world", []string{"hello"}, 1},
		{"Hello world", []string{"hello", "world"}, 2},
		{"Hello world", []string{"foo", "bar"}, 0},
		{"The quick brown fox", []string{"quick", "fox", "dog"}, 2},
	}

	for _, tt := range tests {
		result := countMatches(tt.text, tt.words)
		assert.Equal(t, tt.expected, result)
	}
}

func TestSearchResult_Score_Good(t *testing.T) {
	idx := newSearchIndex()

	// Topic with query word in title should score higher
	idx.Add(&Topic{
		ID:      "topic-in-title",
		Title:   "Installation Guide",
		Content: "Some content here.",
	})

	idx.Add(&Topic{
		ID:      "topic-in-content",
		Title:   "Some Other Topic",
		Content: "This covers installation steps.",
	})

	results := idx.Search("installation")
	assert.Len(t, results, 2)

	// Title match should score higher
	assert.Equal(t, "topic-in-title", results[0].Topic.ID)
	assert.Greater(t, results[0].Score, results[1].Score)
}
