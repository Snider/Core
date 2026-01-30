package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		data     map[string]any
		expected map[string]Message
	}{
		{
			name:   "simple string",
			prefix: "",
			data:   map[string]any{"hello": "world"},
			expected: map[string]Message{
				"hello": {Text: "world"},
			},
		},
		{
			name:   "nested object",
			prefix: "",
			data: map[string]any{
				"cli": map[string]any{
					"success": "Done",
					"error":   "Failed",
				},
			},
			expected: map[string]Message{
				"cli.success": {Text: "Done"},
				"cli.error":   {Text: "Failed"},
			},
		},
		{
			name:   "with prefix",
			prefix: "app",
			data:   map[string]any{"key": "value"},
			expected: map[string]Message{
				"app.key": {Text: "value"},
			},
		},
		{
			name:   "deeply nested",
			prefix: "",
			data: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "deep value",
					},
				},
			},
			expected: map[string]Message{
				"a.b.c": {Text: "deep value"},
			},
		},
		{
			name:   "plural object",
			prefix: "",
			data: map[string]any{
				"items": map[string]any{
					"one":   "{{.Count}} item",
					"other": "{{.Count}} items",
				},
			},
			expected: map[string]Message{
				"items": {One: "{{.Count}} item", Other: "{{.Count}} items"},
			},
		},
		{
			name:   "full CLDR plural",
			prefix: "",
			data: map[string]any{
				"files": map[string]any{
					"zero":  "no files",
					"one":   "one file",
					"two":   "two files",
					"few":   "a few files",
					"many":  "many files",
					"other": "{{.Count}} files",
				},
			},
			expected: map[string]Message{
				"files": {
					Zero:  "no files",
					One:   "one file",
					Two:   "two files",
					Few:   "a few files",
					Many:  "many files",
					Other: "{{.Count}} files",
				},
			},
		},
		{
			name:   "mixed content",
			prefix: "",
			data: map[string]any{
				"simple": "text",
				"plural": map[string]any{
					"one":   "singular",
					"other": "plural",
				},
				"nested": map[string]any{
					"child": "nested value",
				},
			},
			expected: map[string]Message{
				"simple":       {Text: "text"},
				"plural":       {One: "singular", Other: "plural"},
				"nested.child": {Text: "nested value"},
			},
		},
		{
			name:     "empty data",
			prefix:   "",
			data:     map[string]any{},
			expected: map[string]Message{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := make(map[string]Message)
			flatten(tt.prefix, tt.data, out)
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestFlattenWithGrammar(t *testing.T) {
	t.Run("extracts verb forms", func(t *testing.T) {
		data := map[string]any{
			"gram": map[string]any{
				"verb": map[string]any{
					"run": map[string]any{
						"base":   "run",
						"past":   "ran",
						"gerund": "running",
					},
				},
			},
		}
		out := make(map[string]Message)
		grammar := &GrammarData{
			Verbs: make(map[string]VerbForms),
			Nouns: make(map[string]NounForms),
		}
		flattenWithGrammar("", data, out, grammar)

		assert.Contains(t, grammar.Verbs, "run")
		assert.Equal(t, "ran", grammar.Verbs["run"].Past)
		assert.Equal(t, "running", grammar.Verbs["run"].Gerund)
	})

	t.Run("extracts noun forms", func(t *testing.T) {
		data := map[string]any{
			"gram": map[string]any{
				"noun": map[string]any{
					"file": map[string]any{
						"one":    "file",
						"other":  "files",
						"gender": "neuter",
					},
				},
			},
		}
		out := make(map[string]Message)
		grammar := &GrammarData{
			Verbs: make(map[string]VerbForms),
			Nouns: make(map[string]NounForms),
		}
		flattenWithGrammar("", data, out, grammar)

		assert.Contains(t, grammar.Nouns, "file")
		assert.Equal(t, "file", grammar.Nouns["file"].One)
		assert.Equal(t, "files", grammar.Nouns["file"].Other)
		assert.Equal(t, "neuter", grammar.Nouns["file"].Gender)
	})

	t.Run("extracts articles", func(t *testing.T) {
		data := map[string]any{
			"gram": map[string]any{
				"article": map[string]any{
					"indefinite": map[string]any{
						"default": "a",
						"vowel":   "an",
					},
					"definite": "the",
				},
			},
		}
		out := make(map[string]Message)
		grammar := &GrammarData{
			Verbs: make(map[string]VerbForms),
			Nouns: make(map[string]NounForms),
		}
		flattenWithGrammar("", data, out, grammar)

		assert.Equal(t, "a", grammar.Articles.IndefiniteDefault)
		assert.Equal(t, "an", grammar.Articles.IndefiniteVowel)
		assert.Equal(t, "the", grammar.Articles.Definite)
	})

	t.Run("extracts punctuation rules", func(t *testing.T) {
		data := map[string]any{
			"gram": map[string]any{
				"punct": map[string]any{
					"label":    ":",
					"progress": "...",
				},
			},
		}
		out := make(map[string]Message)
		grammar := &GrammarData{
			Verbs: make(map[string]VerbForms),
			Nouns: make(map[string]NounForms),
		}
		flattenWithGrammar("", data, out, grammar)

		assert.Equal(t, ":", grammar.Punct.LabelSuffix)
		assert.Equal(t, "...", grammar.Punct.ProgressSuffix)
	})

	t.Run("nil grammar skips extraction", func(t *testing.T) {
		data := map[string]any{
			"gram": map[string]any{
				"verb": map[string]any{
					"run": map[string]any{
						"past":   "ran",
						"gerund": "running",
					},
				},
			},
			"simple": "text",
		}
		out := make(map[string]Message)
		flattenWithGrammar("", data, out, nil)

		// Without grammar, verb forms are recursively processed as nested objects
		assert.Contains(t, out, "simple")
		assert.Equal(t, "text", out["simple"].Text)
	})
}
