package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPastTense(t *testing.T) {
	tests := []struct {
		verb     string
		expected string
	}{
		// Irregular verbs
		{"be", "was"},
		{"have", "had"},
		{"do", "did"},
		{"go", "went"},
		{"make", "made"},
		{"get", "got"},
		{"run", "ran"},
		{"write", "wrote"},
		{"build", "built"},
		{"find", "found"},
		{"keep", "kept"},
		{"think", "thought"},

		// Regular verbs - ends in -e
		{"delete", "deleted"},
		{"save", "saved"},
		{"create", "created"},
		{"update", "updated"},
		{"remove", "removed"},

		// Regular verbs - consonant + y -> ied
		{"copy", "copied"},
		{"carry", "carried"},
		{"try", "tried"},

		// Regular verbs - vowel + y -> yed
		{"play", "played"},
		{"stay", "stayed"},
		{"enjoy", "enjoyed"},

		// Regular verbs - CVC doubling
		{"stop", "stopped"},
		{"drop", "dropped"},
		{"plan", "planned"},

		// Regular verbs - no doubling
		{"install", "installed"},
		{"open", "opened"},
		{"start", "started"},

		// Edge cases
		{"", ""},
		{" delete ", "deleted"},
	}

	for _, tt := range tests {
		t.Run(tt.verb, func(t *testing.T) {
			result := PastTense(tt.verb)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGerund(t *testing.T) {
	tests := []struct {
		verb     string
		expected string
	}{
		// Irregular verbs
		{"be", "being"},
		{"have", "having"},
		{"run", "running"},
		{"write", "writing"},

		// Regular verbs - drop -e
		{"delete", "deleting"},
		{"save", "saving"},
		{"create", "creating"},
		{"update", "updating"},

		// Regular verbs - ie -> ying
		{"die", "dying"},
		{"lie", "lying"},
		{"tie", "tying"},

		// Regular verbs - CVC doubling
		{"stop", "stopping"},
		{"run", "running"},
		{"plan", "planning"},

		// Regular verbs - no doubling
		{"install", "installing"},
		{"open", "opening"},
		{"start", "starting"},
		{"play", "playing"},

		// Edge cases
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.verb, func(t *testing.T) {
			result := Gerund(tt.verb)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		noun     string
		count    int
		expected string
	}{
		// Singular (count = 1)
		{"file", 1, "file"},
		{"repo", 1, "repo"},

		// Regular plurals
		{"file", 2, "files"},
		{"repo", 5, "repos"},
		{"user", 0, "users"},

		// -s, -ss, -sh, -ch, -x, -z -> -es
		{"bus", 2, "buses"},
		{"class", 3, "classes"},
		{"bush", 2, "bushes"},
		{"match", 2, "matches"},
		{"box", 2, "boxes"},

		// consonant + y -> -ies
		{"city", 2, "cities"},
		{"repository", 3, "repositories"},
		{"copy", 2, "copies"},

		// vowel + y -> -ys
		{"key", 2, "keys"},
		{"day", 2, "days"},
		{"toy", 2, "toys"},

		// Irregular nouns
		{"child", 2, "children"},
		{"person", 3, "people"},
		{"man", 2, "men"},
		{"woman", 2, "women"},
		{"foot", 2, "feet"},
		{"tooth", 2, "teeth"},
		{"mouse", 2, "mice"},
		{"index", 2, "indices"},

		// Unchanged plurals
		{"fish", 2, "fish"},
		{"sheep", 2, "sheep"},
		{"deer", 2, "deer"},
		{"species", 2, "species"},
	}

	for _, tt := range tests {
		t.Run(tt.noun, func(t *testing.T) {
			result := Pluralize(tt.noun, tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPluralForm(t *testing.T) {
	tests := []struct {
		noun     string
		expected string
	}{
		// Regular
		{"file", "files"},
		{"repo", "repos"},

		// -es endings
		{"box", "boxes"},
		{"class", "classes"},
		{"bush", "bushes"},
		{"match", "matches"},

		// -ies endings
		{"city", "cities"},
		{"copy", "copies"},

		// Irregular
		{"child", "children"},
		{"person", "people"},

		// Title case preservation
		{"Child", "Children"},
		{"Person", "People"},

		// Empty
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.noun, func(t *testing.T) {
			result := PluralForm(tt.noun)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestArticle(t *testing.T) {
	tests := []struct {
		word     string
		expected string
	}{
		// Regular vowels -> "an"
		{"error", "an"},
		{"apple", "an"},
		{"issue", "an"},
		{"update", "an"},
		{"item", "an"},
		{"object", "an"},

		// Regular consonants -> "a"
		{"file", "a"},
		{"repo", "a"},
		{"commit", "a"},
		{"branch", "a"},
		{"test", "a"},

		// Consonant sounds despite vowel start -> "a"
		{"user", "a"},
		{"union", "a"},
		{"unique", "a"},
		{"unit", "a"},
		{"universe", "a"},
		{"one", "a"},
		{"once", "a"},
		{"euro", "a"},

		// Vowel sounds despite consonant start -> "an"
		{"hour", "an"},
		{"honest", "an"},
		{"honour", "an"},
		{"heir", "an"},

		// Edge cases
		{"", ""},
		{" error ", "an"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result := Article(tt.word)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "Hello World"},
		{"file deleted", "File Deleted"},
		{"ALREADY CAPS", "ALREADY CAPS"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Title(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"file.txt", `"file.txt"`},
		{"", `""`},
		{"hello world", `"hello world"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Quote(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateFuncs(t *testing.T) {
	funcs := TemplateFuncs()

	// Check all expected functions are present
	expectedFuncs := []string{"title", "lower", "upper", "past", "gerund", "plural", "pluralForm", "article", "quote"}
	for _, name := range expectedFuncs {
		assert.Contains(t, funcs, name, "TemplateFuncs should contain %s", name)
	}
}
