// Package i18n provides internationalization for the CLI.
package i18n

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
	// Only consider it a noun form if it has a gender field
	// This distinguishes noun forms from CLDR plural objects which use one/other
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
