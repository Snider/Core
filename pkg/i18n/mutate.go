// Package i18n provides internationalization for the CLI.
package i18n

import "strings"

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
				if base, ok := v["base"].(string); ok {
					_ = base // base form stored but not used in VerbForms
				}
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
