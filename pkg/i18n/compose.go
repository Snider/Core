// Package i18n provides internationalization for the CLI.
package i18n

import (
	"fmt"
)

// S creates a new Subject with the given noun and value.
// The noun is used for grammar rules, the value for display.
//
//	S("file", "config.yaml")     // "config.yaml"
//	S("repo", repo)              // Uses repo.String() or fmt.Sprint()
func S(noun string, value any) *Subject {
	return &Subject{
		Noun:  noun,
		Value: value,
		count: 1, // Default to singular
	}
}

// NewSubject is an alias for S() for readability in longer expressions.
//
//	NewSubject("file", path).Count(3).In("workspace")
func NewSubject(noun string, value any) *Subject {
	return S(noun, value)
}

// Count sets the count for pluralization.
// Used to determine singular/plural forms in templates.
// Panics if called on nil receiver; use S() to create subjects.
//
//	S("file", files).Count(len(files))
func (s *Subject) Count(n int) *Subject {
	if s == nil {
		return nil
	}
	s.count = n
	return s
}

// Gender sets the grammatical gender for languages that require it.
// Common values: "masculine", "feminine", "neuter"
// Panics if called on nil receiver; use S() to create subjects.
//
//	S("user", user).Gender("female")
func (s *Subject) Gender(g string) *Subject {
	if s == nil {
		return nil
	}
	s.gender = g
	return s
}

// In sets the location context for the subject.
// Used in templates to provide spatial context.
// Panics if called on nil receiver; use S() to create subjects.
//
//	S("file", "config.yaml").In("workspace")
func (s *Subject) In(location string) *Subject {
	if s == nil {
		return nil
	}
	s.location = location
	return s
}

// Formal sets the formality level to formal (Sie, vous, usted).
// Use for polite/professional address in languages that distinguish formality.
// Panics if called on nil receiver; use S() to create subjects.
//
//	S("colleague", name).Formal()
func (s *Subject) Formal() *Subject {
	if s == nil {
		return nil
	}
	s.formality = FormalityFormal
	return s
}

// Informal sets the formality level to informal (du, tu, tú).
// Use for casual/friendly address in languages that distinguish formality.
// Panics if called on nil receiver; use S() to create subjects.
//
//	S("friend", name).Informal()
func (s *Subject) Informal() *Subject {
	if s == nil {
		return nil
	}
	s.formality = FormalityInformal
	return s
}

// Formality sets the formality level explicitly.
// Panics if called on nil receiver; use S() to create subjects.
//
//	S("user", name).Formality(FormalityFormal)
func (s *Subject) Formality(f Formality) *Subject {
	if s == nil {
		return nil
	}
	s.formality = f
	return s
}

// String returns the display value of the subject.
func (s *Subject) String() string {
	if s == nil {
		return ""
	}
	if stringer, ok := s.Value.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprint(s.Value)
}

// IsPlural returns true if count != 1.
func (s *Subject) IsPlural() bool {
	return s != nil && s.count != 1
}

// CountInt returns the count value.
func (s *Subject) CountInt() int {
	if s == nil {
		return 1
	}
	return s.count
}

// GenderString returns the grammatical gender.
func (s *Subject) GenderString() string {
	if s == nil {
		return ""
	}
	return s.gender
}

// LocationString returns the location context.
func (s *Subject) LocationString() string {
	if s == nil {
		return ""
	}
	return s.location
}

// NounString returns the noun type.
func (s *Subject) NounString() string {
	if s == nil {
		return ""
	}
	return s.Noun
}

// FormalityString returns the formality level.
// Returns FormalityNeutral if not explicitly set.
func (s *Subject) FormalityString() Formality {
	if s == nil {
		return FormalityNeutral
	}
	return s.formality
}

// IsFormal returns true if formal address should be used.
func (s *Subject) IsFormal() bool {
	return s != nil && s.formality == FormalityFormal
}

// IsInformal returns true if informal address should be used.
func (s *Subject) IsInformal() bool {
	return s != nil && s.formality == FormalityInformal
}

// newTemplateData creates templateData from a Subject.
func newTemplateData(s *Subject) templateData {
	if s == nil {
		return templateData{Count: 1}
	}
	return templateData{
		Subject:   s.String(),
		Noun:      s.Noun,
		Count:     s.count,
		Gender:    s.gender,
		Location:  s.location,
		Formality: s.formality,
		IsFormal:  s.formality == FormalityFormal,
		IsPlural:  s.count != 1,
		Value:     s.Value,
	}
}
