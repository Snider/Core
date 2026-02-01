// Package i18n provides internationalization for the CLI.
package i18n

import (
	"fmt"
	"strings"
)

// --- Built-in Handlers ---

// LabelHandler handles i18n.label.{word} → "Status:" patterns.
type LabelHandler struct{}

// Match returns true for keys starting with "i18n.label.".
func (h LabelHandler) Match(key string) bool {
	return strings.HasPrefix(key, "i18n.label.")
}

// Handle transforms label keys into formatted labels with colons.
func (h LabelHandler) Handle(key string, args []any, next func() string) string {
	word := strings.TrimPrefix(key, "i18n.label.")
	return Label(word)
}

// ProgressHandler handles i18n.progress.{verb} → "Building..." patterns.
type ProgressHandler struct{}

// Match returns true for keys starting with "i18n.progress.".
func (h ProgressHandler) Match(key string) bool {
	return strings.HasPrefix(key, "i18n.progress.")
}

// Handle transforms progress keys into gerund phrases like "Building...".
func (h ProgressHandler) Handle(key string, args []any, next func() string) string {
	verb := strings.TrimPrefix(key, "i18n.progress.")
	if len(args) > 0 {
		if subj, ok := args[0].(string); ok {
			return ProgressSubject(verb, subj)
		}
	}
	return Progress(verb)
}

// CountHandler handles i18n.count.{noun} → "5 files" patterns.
type CountHandler struct{}

// Match returns true for keys starting with "i18n.count.".
func (h CountHandler) Match(key string) bool {
	return strings.HasPrefix(key, "i18n.count.")
}

// Handle transforms count keys into pluralized phrases like "5 files".
func (h CountHandler) Handle(key string, args []any, next func() string) string {
	noun := strings.TrimPrefix(key, "i18n.count.")
	if len(args) > 0 {
		count := toInt(args[0])
		return fmt.Sprintf("%d %s", count, Pluralize(noun, count))
	}
	return noun
}

// DoneHandler handles i18n.done.{verb} → "File deleted" patterns.
type DoneHandler struct{}

// Match returns true for keys starting with "i18n.done.".
func (h DoneHandler) Match(key string) bool {
	return strings.HasPrefix(key, "i18n.done.")
}

// Handle transforms done keys into past-tense completion messages.
func (h DoneHandler) Handle(key string, args []any, next func() string) string {
	verb := strings.TrimPrefix(key, "i18n.done.")
	if len(args) > 0 {
		if subj, ok := args[0].(string); ok {
			return ActionResult(verb, subj)
		}
	}
	return Title(PastTense(verb))
}

// FailHandler handles i18n.fail.{verb} → "Failed to delete file" patterns.
type FailHandler struct{}

// Match returns true for keys starting with "i18n.fail.".
func (h FailHandler) Match(key string) bool {
	return strings.HasPrefix(key, "i18n.fail.")
}

// Handle transforms fail keys into failure messages like "Failed to delete".
func (h FailHandler) Handle(key string, args []any, next func() string) string {
	verb := strings.TrimPrefix(key, "i18n.fail.")
	if len(args) > 0 {
		if subj, ok := args[0].(string); ok {
			return ActionFailed(verb, subj)
		}
	}
	return ActionFailed(verb, "")
}

// NumericHandler handles i18n.numeric.{format} → formatted numbers.
type NumericHandler struct{}

// Match returns true for keys starting with "i18n.numeric.".
func (h NumericHandler) Match(key string) bool {
	return strings.HasPrefix(key, "i18n.numeric.")
}

// Handle transforms numeric keys into locale-formatted numbers.
func (h NumericHandler) Handle(key string, args []any, next func() string) string {
	if len(args) == 0 {
		return next()
	}

	format := strings.TrimPrefix(key, "i18n.numeric.")
	switch format {
	case "number", "int":
		return FormatNumber(toInt64(args[0]))
	case "decimal", "float":
		return FormatDecimal(toFloat64(args[0]))
	case "percent", "pct":
		return FormatPercent(toFloat64(args[0]))
	case "bytes", "size":
		return FormatBytes(toInt64(args[0]))
	case "ordinal", "ord":
		return FormatOrdinal(toInt(args[0]))
	case "ago":
		if len(args) >= 2 {
			if unit, ok := args[1].(string); ok {
				return FormatAgo(toInt(args[0]), unit)
			}
		}
	}
	return next()
}

// --- Handler Chain ---

// DefaultHandlers returns the built-in i18n.* namespace handlers.
func DefaultHandlers() []KeyHandler {
	return []KeyHandler{
		LabelHandler{},
		ProgressHandler{},
		CountHandler{},
		DoneHandler{},
		FailHandler{},
		NumericHandler{},
	}
}

// RunHandlerChain executes a chain of handlers for a key.
// Returns empty string if no handler matched (caller should use standard lookup).
func RunHandlerChain(handlers []KeyHandler, key string, args []any, fallback func() string) string {
	for i, h := range handlers {
		if h.Match(key) {
			// Create next function that tries remaining handlers
			next := func() string {
				remaining := handlers[i+1:]
				if len(remaining) > 0 {
					return RunHandlerChain(remaining, key, args, fallback)
				}
				return fallback()
			}
			return h.Handle(key, args, next)
		}
	}
	return fallback()
}

// --- Compile-time interface checks ---

var (
	_ KeyHandler = LabelHandler{}
	_ KeyHandler = ProgressHandler{}
	_ KeyHandler = CountHandler{}
	_ KeyHandler = DoneHandler{}
	_ KeyHandler = FailHandler{}
	_ KeyHandler = NumericHandler{}
)
