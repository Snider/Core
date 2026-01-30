// Package i18n provides internationalization for the CLI.
package i18n

// TranslationContext provides disambiguation for translations.
// Use this when the same word translates differently in different contexts.
//
// Example: "right" can mean direction or correctness:
//
//	T("direction.right", C("navigation")) // → "rechts" (German)
//	T("status.right", C("correctness"))   // → "richtig" (German)
type TranslationContext struct {
	Context   string         // Semantic context (e.g., "navigation", "correctness")
	Gender    string         // Grammatical gender hint (e.g., "masculine", "feminine")
	Formality Formality      // Formality level override
	Extra     map[string]any // Additional context-specific data
}

// C creates a TranslationContext with the given context string.
// Chain methods to add more context:
//
//	C("navigation").Gender("masculine").Formal()
func C(context string) *TranslationContext {
	return &TranslationContext{
		Context: context,
	}
}

// WithGender sets the grammatical gender hint.
func (c *TranslationContext) WithGender(gender string) *TranslationContext {
	if c == nil {
		return nil
	}
	c.Gender = gender
	return c
}

// Formal sets the formality level to formal.
func (c *TranslationContext) Formal() *TranslationContext {
	if c == nil {
		return nil
	}
	c.Formality = FormalityFormal
	return c
}

// Informal sets the formality level to informal.
func (c *TranslationContext) Informal() *TranslationContext {
	if c == nil {
		return nil
	}
	c.Formality = FormalityInformal
	return c
}

// WithFormality sets an explicit formality level.
func (c *TranslationContext) WithFormality(f Formality) *TranslationContext {
	if c == nil {
		return nil
	}
	c.Formality = f
	return c
}

// Set adds a key-value pair to the extra context data.
func (c *TranslationContext) Set(key string, value any) *TranslationContext {
	if c == nil {
		return nil
	}
	if c.Extra == nil {
		c.Extra = make(map[string]any)
	}
	c.Extra[key] = value
	return c
}

// Get retrieves a value from the extra context data.
func (c *TranslationContext) Get(key string) any {
	if c == nil || c.Extra == nil {
		return nil
	}
	return c.Extra[key]
}

// ContextString returns the context string (nil-safe).
func (c *TranslationContext) ContextString() string {
	if c == nil {
		return ""
	}
	return c.Context
}

// GenderString returns the gender hint (nil-safe).
func (c *TranslationContext) GenderString() string {
	if c == nil {
		return ""
	}
	return c.Gender
}

// FormalityValue returns the formality level (nil-safe).
func (c *TranslationContext) FormalityValue() Formality {
	if c == nil {
		return FormalityNeutral
	}
	return c.Formality
}
