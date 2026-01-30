# Semantic i18n System Design

## Overview

Extend the i18n system beyond simple key-value translation to support **semantic intents** that encode meaning, enabling:

- Composite translations from reusable fragments
- Grammatical awareness (gender, plurality, formality)
- CLI prompt integration with localized options
- Reduced calling code complexity

## Goals

1. **Simple cases stay simple** - `_("key")` works as expected
2. **Complex cases become declarative** - Intent drives output, not caller logic
3. **Translators have power** - Grammar rules live in translations, not code
4. **CLI integration** - Questions, confirmations, choices are first-class

## API Design

### Translation Functions

```go
// Simple lookup (gettext-style) - unchanged
i18n._("cli.success")
i18n._("common.label.error")

// Opinionated lookup with namespace awareness
i18n.T("core.edit.question", i18n.Subject("file", path))

// Transform with semantic intent
i18n.Transmute("core.edit", map[string]any{
    "Subject": path,
    "Count":   1,
})
```

### CLI Integration

```go
// Simple yes/no with localized options
confirmed := cli.Confirm("core.delete", i18n.Subject("file", path))
// Displays: "Delete /path/to/file.txt? [y/N]"
// Returns: bool

// Question with custom options
choice := cli.Question("core.save", i18n.Subject("changes", 3), cli.Options{
    Default: "yes",
    Extra:   []string{"all"},  // Adds [a] option
})
// Displays: "Save 3 changes? [a/y/N]"
// Returns: "yes" | "no" | "all"

// Choice from list
selected := cli.Choose("core.select.branch", branches)
// Displays localized prompt with arrow selection
```

## Reserved Namespaces

### `common.*` - Reusable Fragments

Atomic translation units that can be composed:

```json
{
  "common": {
    "verb": {
      "edit": "edit",
      "delete": "delete",
      "create": "create",
      "save": "save",
      "update": "update"
    },
    "noun": {
      "file": { "one": "file", "other": "files" },
      "commit": { "one": "commit", "other": "commits" },
      "change": { "one": "change", "other": "changes" }
    },
    "article": {
      "the": "the",
      "a": { "one": "a", "vowel": "an" }
    },
    "prompt": {
      "yes": "y",
      "no": "n",
      "all": "a",
      "skip": "s",
      "quit": "q"
    }
  }
}
```

### `core.*` - Semantic Intents

Intents encode meaning and behavior:

```json
{
  "core": {
    "edit": {
      "_meta": {
        "type": "action",
        "verb": "common.verb.edit",
        "dangerous": false
      },
      "question": "Should I {{.Verb}} {{.Subject}}?",
      "confirm": "{{.Verb | title}} {{.Subject}}?",
      "success": "{{.Subject | title}} {{.Verb | past}}",
      "failure": "Failed to {{.Verb}} {{.Subject}}"
    },
    "delete": {
      "_meta": {
        "type": "action",
        "verb": "common.verb.delete",
        "dangerous": true,
        "default": "no"
      },
      "question": "Delete {{.Subject}}? This cannot be undone.",
      "confirm": "Really delete {{.Subject}}?",
      "success": "{{.Subject | title}} deleted",
      "failure": "Failed to delete {{.Subject}}"
    },
    "save": {
      "_meta": {
        "type": "action",
        "verb": "common.verb.save",
        "supports": ["all", "skip"]
      },
      "question": "Save {{.Subject}}?",
      "success": "{{.Subject | title}} saved"
    }
  }
}
```

## Transmute Function

`Transmute()` combines intent metadata with input to produce contextually correct output:

```go
// Signature
func Transmute(intent string, data map[string]any) TransmuteResult

// TransmuteResult provides multiple output forms
type TransmuteResult struct {
    Question string          // "Delete the file?"
    Confirm  string          // "Really delete the file?"
    Success  string          // "File deleted"
    Failure  string          // "Failed to delete the file"
    Meta     IntentMeta      // Dangerous, default, supports, etc.
}

// Usage
result := i18n.Transmute("core.delete", map[string]any{
    "Subject": "/path/to/file.txt",
})

if result.Meta.Dangerous {
    // Show warning styling
}
fmt.Println(result.Question)  // "Delete /path/to/file.txt? This cannot be undone."
```

## Template Functions

Available in translation templates:

| Function | Description | Example |
|----------|-------------|---------|
| `title` | Title case | `{{.Name \| title}}` → "Hello World" |
| `lower` | Lower case | `{{.Name \| lower}}` → "hello world" |
| `upper` | Upper case | `{{.Name \| upper}}` → "HELLO WORLD" |
| `past` | Past tense verb | `{{.Verb \| past}}` → "edited" |
| `plural` | Pluralize noun | `{{.Noun \| plural .Count}}` → "files" |
| `article` | Add article | `{{.Noun \| article}}` → "a file" |
| `quote` | Wrap in quotes | `{{.Path \| quote}}` → `"/path/to/file"` |

## Subject Helper

`Subject()` creates a typed subject with metadata:

```go
// Simple subject
i18n.Subject("file", "/path/to/file.txt")

// With count (for plurality)
i18n.Subject("commit", commits, i18n.Count(len(commits)))

// With gender (for languages that need it)
i18n.Subject("user", userName, i18n.Gender("female"))
```

## CLI Integration Details

### cli.Confirm()

```go
func Confirm(intent string, subject Subject, opts ...ConfirmOption) bool

// Options
cli.DefaultYes()     // Default to yes instead of no
cli.DefaultNo()      // Explicit default no
cli.Required()       // No default, must choose
cli.Timeout(30*time.Second)  // Auto-select default after timeout
```

### cli.Question()

```go
func Question(intent string, subject Subject, opts ...QuestionOption) string

// Options
cli.Options{"all", "skip"}  // Extra options beyond y/n
cli.Default("yes")          // Which option is default
cli.Validate(func(s string) bool)  // Custom validation
```

### cli.Choose()

```go
func Choose[T any](intent string, items []T, opts ...ChooseOption) T

// Options
cli.Display(func(T) string)  // How to display each item
cli.Filter()                 // Enable fuzzy filtering
cli.Multi()                  // Allow multiple selection
```

## Implementation Plan

### Phase 1: Foundation
1. Add `_meta` parsing to JSON loader
2. Implement `Transmute()` with basic templates
3. Add template functions (title, lower, past, etc.)
4. Add `Subject()` helper

### Phase 2: CLI Integration
1. Implement `cli.Confirm()` using intents
2. Implement `cli.Question()` with options
3. Implement `cli.Choose()` for lists
4. Localize prompt characters [y/N]

### Phase 3: Grammar Engine
1. Verb conjugation (past tense, etc.)
2. Noun plurality with irregular forms
3. Article selection (a/an, gender)
4. Language-specific rules

### Phase 4: Extended Languages
1. Gender agreement (French, German, etc.)
2. Formality levels (Japanese, Korean, etc.)
3. Right-to-left support
4. Plural forms beyond one/other (Russian, Arabic, etc.)

## Example: Full Flow

```go
// In cmd/dev/dev_commit.go
path := "/Users/dev/project"
files := []string{"main.go", "config.yaml"}

// Old way (hardcoded English, manual prompt handling)
fmt.Printf("Commit %d files in %s? [y/N] ", len(files), path)
var response string
fmt.Scanln(&response)
if response != "y" && response != "Y" {
    return
}

// New way (semantic, localized, integrated)
subject := i18n.Subject("file", path, i18n.Count(len(files)))
if !cli.Confirm("core.commit", subject) {
    return
}

// For German user, displays:
// "2 Dateien in /Users/dev/project committen? [j/N]"
// (note: "j" for "ja" instead of "y" for "yes")
```

## JSON Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "common": {
      "description": "Reusable translation fragments",
      "type": "object"
    },
    "core": {
      "description": "Semantic intents with metadata",
      "type": "object",
      "additionalProperties": {
        "type": "object",
        "properties": {
          "_meta": {
            "type": "object",
            "properties": {
              "type": { "enum": ["action", "question", "info"] },
              "verb": { "type": "string" },
              "dangerous": { "type": "boolean" },
              "default": { "enum": ["yes", "no"] },
              "supports": { "type": "array", "items": { "type": "string" } }
            }
          },
          "question": { "type": "string" },
          "confirm": { "type": "string" },
          "success": { "type": "string" },
          "failure": { "type": "string" }
        }
      }
    }
  }
}
```

## Open Questions

1. **Verb conjugation library** - Use existing Go library or build custom?
2. **Gender detection** - How to infer gender for subjects in gendered languages?
3. **Fallback behavior** - What happens when intent metadata is missing?
4. **Caching** - Should compiled templates be cached?
5. **Validation** - How to validate intent definitions at build time?
