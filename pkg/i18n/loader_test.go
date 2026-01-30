package i18n

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFSLoader_Load(t *testing.T) {
	t.Run("loads simple messages", func(t *testing.T) {
		fsys := fstest.MapFS{
			"locales/en.json": &fstest.MapFile{
				Data: []byte(`{"hello": "world", "nested": {"key": "value"}}`),
			},
		}
		loader := NewFSLoader(fsys, "locales")
		messages, grammar, err := loader.Load("en")
		require.NoError(t, err)
		assert.NotNil(t, grammar)
		assert.Equal(t, "world", messages["hello"].Text)
		assert.Equal(t, "value", messages["nested.key"].Text)
	})

	t.Run("handles underscore/hyphen variants", func(t *testing.T) {
		fsys := fstest.MapFS{
			"locales/en_GB.json": &fstest.MapFile{
				Data: []byte(`{"greeting": "Hello"}`),
			},
		}
		loader := NewFSLoader(fsys, "locales")
		messages, _, err := loader.Load("en-GB")
		require.NoError(t, err)
		assert.Equal(t, "Hello", messages["greeting"].Text)
	})

	t.Run("returns error for missing language", func(t *testing.T) {
		fsys := fstest.MapFS{}
		loader := NewFSLoader(fsys, "locales")
		_, _, err := loader.Load("fr")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("extracts grammar data", func(t *testing.T) {
		fsys := fstest.MapFS{
			"locales/en.json": &fstest.MapFile{
				Data: []byte(`{
					"gram": {
						"verb": {
							"run": {"past": "ran", "gerund": "running"}
						},
						"noun": {
							"file": {"one": "file", "other": "files", "gender": "neuter"}
						}
					}
				}`),
			},
		}
		loader := NewFSLoader(fsys, "locales")
		_, grammar, err := loader.Load("en")
		require.NoError(t, err)
		assert.Equal(t, "ran", grammar.Verbs["run"].Past)
		assert.Equal(t, "running", grammar.Verbs["run"].Gerund)
		assert.Equal(t, "files", grammar.Nouns["file"].Other)
	})
}

func TestFSLoader_Languages(t *testing.T) {
	t.Run("lists available languages", func(t *testing.T) {
		fsys := fstest.MapFS{
			"locales/en.json":    &fstest.MapFile{Data: []byte(`{}`)},
			"locales/de.json":    &fstest.MapFile{Data: []byte(`{}`)},
			"locales/fr_FR.json": &fstest.MapFile{Data: []byte(`{}`)},
		}
		loader := NewFSLoader(fsys, "locales")
		langs := loader.Languages()
		assert.Contains(t, langs, "en")
		assert.Contains(t, langs, "de")
		assert.Contains(t, langs, "fr-FR") // normalised
	})

	t.Run("caches result", func(t *testing.T) {
		fsys := fstest.MapFS{
			"locales/en.json": &fstest.MapFile{Data: []byte(`{}`)},
		}
		loader := NewFSLoader(fsys, "locales")
		langs1 := loader.Languages()
		langs2 := loader.Languages()
		assert.Equal(t, langs1, langs2)
	})

	t.Run("empty directory", func(t *testing.T) {
		fsys := fstest.MapFS{}
		loader := NewFSLoader(fsys, "locales")
		langs := loader.Languages()
		assert.Empty(t, langs)
	})
}

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

func TestIsVerbFormObject(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected bool
	}{
		{
			name:     "has base only",
			input:    map[string]any{"base": "run"},
			expected: true,
		},
		{
			name:     "has past only",
			input:    map[string]any{"past": "ran"},
			expected: true,
		},
		{
			name:     "has gerund only",
			input:    map[string]any{"gerund": "running"},
			expected: true,
		},
		{
			name:     "has all verb forms",
			input:    map[string]any{"base": "run", "past": "ran", "gerund": "running"},
			expected: true,
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: false,
		},
		{
			name:     "plural object not verb",
			input:    map[string]any{"one": "item", "other": "items"},
			expected: false,
		},
		{
			name:     "unrelated keys",
			input:    map[string]any{"foo": "bar", "baz": "qux"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVerbFormObject(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNounFormObject(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected bool
	}{
		{
			name:     "has gender",
			input:    map[string]any{"gender": "masculine", "one": "file", "other": "files"},
			expected: true,
		},
		{
			name:     "gender only",
			input:    map[string]any{"gender": "feminine"},
			expected: true,
		},
		{
			name:     "no gender",
			input:    map[string]any{"one": "item", "other": "items"},
			expected: false,
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNounFormObject(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasPluralCategories(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected bool
	}{
		{
			name:     "has zero",
			input:    map[string]any{"zero": "none", "one": "one", "other": "many"},
			expected: true,
		},
		{
			name:     "has two",
			input:    map[string]any{"one": "one", "two": "two", "other": "many"},
			expected: true,
		},
		{
			name:     "has few",
			input:    map[string]any{"one": "one", "few": "few", "other": "many"},
			expected: true,
		},
		{
			name:     "has many",
			input:    map[string]any{"one": "one", "many": "many", "other": "other"},
			expected: true,
		},
		{
			name:     "has all categories",
			input:    map[string]any{"zero": "0", "one": "1", "two": "2", "few": "few", "many": "many", "other": "other"},
			expected: true,
		},
		{
			name:     "only one and other",
			input:    map[string]any{"one": "item", "other": "items"},
			expected: false,
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: false,
		},
		{
			name:     "unrelated keys",
			input:    map[string]any{"foo": "bar"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPluralCategories(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPluralObject(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected bool
	}{
		{
			name:     "one and other",
			input:    map[string]any{"one": "item", "other": "items"},
			expected: true,
		},
		{
			name:     "all CLDR categories",
			input:    map[string]any{"zero": "0", "one": "1", "two": "2", "few": "few", "many": "many", "other": "other"},
			expected: true,
		},
		{
			name:     "only other",
			input:    map[string]any{"other": "items"},
			expected: true,
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: false,
		},
		{
			name:     "nested map is not plural",
			input:    map[string]any{"one": "item", "other": map[string]any{"nested": "value"}},
			expected: false,
		},
		{
			name:     "unrelated keys",
			input:    map[string]any{"foo": "bar", "baz": "qux"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPluralObject(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMessageIsPlural(t *testing.T) {
	tests := []struct {
		name     string
		msg      Message
		expected bool
	}{
		{
			name:     "has zero",
			msg:      Message{Zero: "none"},
			expected: true,
		},
		{
			name:     "has one",
			msg:      Message{One: "item"},
			expected: true,
		},
		{
			name:     "has two",
			msg:      Message{Two: "items"},
			expected: true,
		},
		{
			name:     "has few",
			msg:      Message{Few: "a few"},
			expected: true,
		},
		{
			name:     "has many",
			msg:      Message{Many: "lots"},
			expected: true,
		},
		{
			name:     "has other",
			msg:      Message{Other: "items"},
			expected: true,
		},
		{
			name:     "has all",
			msg:      Message{Zero: "0", One: "1", Two: "2", Few: "few", Many: "many", Other: "other"},
			expected: true,
		},
		{
			name:     "text only",
			msg:      Message{Text: "hello"},
			expected: false,
		},
		{
			name:     "empty message",
			msg:      Message{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.IsPlural()
			assert.Equal(t, tt.expected, result)
		})
	}
}
