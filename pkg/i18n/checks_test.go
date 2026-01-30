package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
