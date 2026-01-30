// Package i18n provides internationalization for the CLI.
package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

// FSLoader loads translations from a filesystem (embedded or disk).
type FSLoader struct {
	fsys fs.FS
	dir  string

	// Cache of available languages (populated on first Languages() call)
	languages []string
	langOnce  sync.Once
}

// NewFSLoader creates a loader for the given filesystem and directory.
func NewFSLoader(fsys fs.FS, dir string) *FSLoader {
	return &FSLoader{
		fsys: fsys,
		dir:  dir,
	}
}

// Load implements Loader.Load - loads messages and grammar for a language.
func (l *FSLoader) Load(lang string) (map[string]Message, *GrammarData, error) {
	// Try both hyphen and underscore variants
	variants := []string{
		lang + ".json",
		strings.ReplaceAll(lang, "-", "_") + ".json",
		strings.ReplaceAll(lang, "_", "-") + ".json",
	}

	var data []byte
	var err error
	for _, filename := range variants {
		filePath := filepath.Join(l.dir, filename)
		data, err = fs.ReadFile(l.fsys, filePath)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("locale %q not found: %w", lang, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON in locale %q: %w", lang, err)
	}

	messages := make(map[string]Message)
	grammar := &GrammarData{
		Verbs: make(map[string]VerbForms),
		Nouns: make(map[string]NounForms),
		Words: make(map[string]string),
	}

	flattenWithGrammar("", raw, messages, grammar)

	return messages, grammar, nil
}

// Languages implements Loader.Languages - returns available language codes.
// Thread-safe: uses sync.Once to ensure the directory is scanned only once.
func (l *FSLoader) Languages() []string {
	l.langOnce.Do(func() {
		entries, err := fs.ReadDir(l.fsys, l.dir)
		if err != nil {
			return
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			lang := strings.TrimSuffix(entry.Name(), ".json")
			// Normalise underscore to hyphen (en_GB -> en-GB)
			lang = strings.ReplaceAll(lang, "_", "-")
			l.languages = append(l.languages, lang)
		}
	})

	return l.languages
}

// Ensure FSLoader implements Loader at compile time.
var _ Loader = (*FSLoader)(nil)

// --- Flatten helpers ---

// flatten recursively flattens nested maps into dot-notation keys.
func flatten(prefix string, data map[string]any, out map[string]Message) {
	flattenWithGrammar(prefix, data, out, nil)
}

// flattenWithGrammar recursively flattens nested maps and extracts grammar data.
func flattenWithGrammar(prefix string, data map[string]any, out map[string]Message, grammar *GrammarData) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			out[fullKey] = Message{Text: v}

		case map[string]any:
			// Check if this is a verb form object
			// Grammar data lives under "gram.*" (a nod to Gram - grandmother)
			if grammar != nil && isVerbFormObject(v) {
				verbName := key
				if strings.HasPrefix(fullKey, "gram.verb.") {
					verbName = strings.TrimPrefix(fullKey, "gram.verb.")
				}
				forms := VerbForms{}
				if past, ok := v["past"].(string); ok {
					forms.Past = past
				}
				if gerund, ok := v["gerund"].(string); ok {
					forms.Gerund = gerund
				}
				grammar.Verbs[strings.ToLower(verbName)] = forms
				continue
			}

			// Check if this is a noun form object
			if grammar != nil && isNounFormObject(v) {
				nounName := key
				if strings.HasPrefix(fullKey, "gram.noun.") {
					nounName = strings.TrimPrefix(fullKey, "gram.noun.")
				}
				forms := NounForms{}
				if one, ok := v["one"].(string); ok {
					forms.One = one
				}
				if other, ok := v["other"].(string); ok {
					forms.Other = other
				}
				if gender, ok := v["gender"].(string); ok {
					forms.Gender = gender
				}
				grammar.Nouns[strings.ToLower(nounName)] = forms
				continue
			}

			// Check if this is an article object
			if grammar != nil && fullKey == "gram.article" {
				if indef, ok := v["indefinite"].(map[string]any); ok {
					if def, ok := indef["default"].(string); ok {
						grammar.Articles.IndefiniteDefault = def
					}
					if vowel, ok := indef["vowel"].(string); ok {
						grammar.Articles.IndefiniteVowel = vowel
					}
				}
				if def, ok := v["definite"].(string); ok {
					grammar.Articles.Definite = def
				}
				continue
			}

			// Check if this is a punctuation rules object
			if grammar != nil && fullKey == "gram.punct" {
				if label, ok := v["label"].(string); ok {
					grammar.Punct.LabelSuffix = label
				}
				if progress, ok := v["progress"].(string); ok {
					grammar.Punct.ProgressSuffix = progress
				}
				continue
			}

			// Check if this is a base word in gram.word.*
			if grammar != nil && strings.HasPrefix(fullKey, "gram.word.") {
				wordKey := strings.TrimPrefix(fullKey, "gram.word.")
				// v could be a string or a nested object
				if str, ok := value.(string); ok {
					if grammar.Words == nil {
						grammar.Words = make(map[string]string)
					}
					grammar.Words[strings.ToLower(wordKey)] = str
				}
				continue
			}

			// Check if this is a plural object (has CLDR plural category keys)
			if isPluralObject(v) {
				msg := Message{}
				if zero, ok := v["zero"].(string); ok {
					msg.Zero = zero
				}
				if one, ok := v["one"].(string); ok {
					msg.One = one
				}
				if two, ok := v["two"].(string); ok {
					msg.Two = two
				}
				if few, ok := v["few"].(string); ok {
					msg.Few = few
				}
				if many, ok := v["many"].(string); ok {
					msg.Many = many
				}
				if other, ok := v["other"].(string); ok {
					msg.Other = other
				}
				out[fullKey] = msg
			} else {
				// Recurse into nested object
				flattenWithGrammar(fullKey, v, out, grammar)
			}
		}
	}
}

// --- Check helpers ---

// isVerbFormObject checks if a map represents verb conjugation forms.
func isVerbFormObject(m map[string]any) bool {
	_, hasBase := m["base"]
	_, hasPast := m["past"]
	_, hasGerund := m["gerund"]
	return (hasBase || hasPast || hasGerund) && !isPluralObject(m)
}

// isNounFormObject checks if a map represents noun forms (with gender).
// Noun form objects have "gender" field, distinguishing them from CLDR plural objects.
func isNounFormObject(m map[string]any) bool {
	_, hasGender := m["gender"]
	return hasGender
}

// hasPluralCategories checks if a map has CLDR plural categories beyond one/other.
func hasPluralCategories(m map[string]any) bool {
	_, hasZero := m["zero"]
	_, hasTwo := m["two"]
	_, hasFew := m["few"]
	_, hasMany := m["many"]
	return hasZero || hasTwo || hasFew || hasMany
}

// isPluralObject checks if a map represents plural forms.
// Recognizes all CLDR plural categories: zero, one, two, few, many, other.
func isPluralObject(m map[string]any) bool {
	_, hasZero := m["zero"]
	_, hasOne := m["one"]
	_, hasTwo := m["two"]
	_, hasFew := m["few"]
	_, hasMany := m["many"]
	_, hasOther := m["other"]

	// It's a plural object if it has any plural category key
	if !hasZero && !hasOne && !hasTwo && !hasFew && !hasMany && !hasOther {
		return false
	}
	// But not if it contains nested objects (those are namespace containers)
	for _, v := range m {
		if _, isMap := v.(map[string]any); isMap {
			return false
		}
	}
	return true
}
