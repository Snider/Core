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

### Function Reference (Stable API)

These function names are **permanent** - choose carefully, they cannot change.

| Function | Alias | Purpose |
|----------|-------|---------|
| `_()` | - | Simple gettext-style lookup |
| `T()` | `C()` | Compose - semantic intent resolution |
| `S()` | `Subject()` | Create typed subject with metadata |

### Simple Translation: `_()`

Standard gettext-style lookup. No magic, just key → value.

```go
i18n._("cli.success")                    // "Success"
i18n._("common.label.error")             // "Error:"
i18n._("common.error.failed", map[string]any{"Action": "load"})  // "Failed to load"
```

### Compose: `T()` / `C()`

Semantic intent resolution. Takes an intent key from `core.*` namespace and returns a `Composed` result with multiple output forms.

```go
// Full form
result := i18n.T("core.delete", i18n.S("file", path))
result := i18n.C("core.delete", i18n.S("file", path))  // Alias

// Result contains all forms
result.Question  // "Delete /path/to/file.txt?"
result.Confirm   // "Really delete /path/to/file.txt?"
result.Success   // "File deleted"
result.Failure   // "Failed to delete file"
result.Meta      // IntentMeta{Dangerous: true, Default: "no", ...}
```

### Subject: `S()` / `Subject()`

Creates a typed subject with optional metadata for grammar rules.

```go
// Simple
i18n.S("file", "/path/to/file.txt")

// With count (plurality)
i18n.S("commit", commits).Count(len(commits))

// With gender (for gendered languages)
i18n.S("user", name).Gender("female")

// Chained
i18n.S("file", path).Count(3).In("/project")
```

### Type Signatures

```go
// Simple lookup
func _(key string, args ...any) string

// Compose (T and C are aliases)
func T(intent string, subject *Subject) *Composed
func C(intent string, subject *Subject) *Composed

// Subject builder
func S(noun string, value any) *Subject
func Subject(noun string, value any) *Subject

// Composed result
type Composed struct {
    Question string
    Confirm  string
    Success  string
    Failure  string
    Meta     IntentMeta
}

// Subject with metadata
type Subject struct {
    Noun   string
    Value  any
    count  int
    gender string
    // ... other metadata
}

func (s *Subject) Count(n int) *Subject
func (s *Subject) Gender(g string) *Subject
func (s *Subject) In(location string) *Subject

// Intent metadata
type IntentMeta struct {
    Type      string   // "action", "question", "info"
    Verb      string   // Reference to common.verb.*
    Dangerous bool     // Requires confirmation
    Default   string   // "yes" or "no"
    Supports  []string // Extra options like "all", "skip"
}
```

## CLI Integration

The CLI package uses `T()` internally for prompts:

```go
// Confirm uses T() internally
confirmed := cli.Confirm("core.delete", i18n.S("file", path))
// Internally: result := i18n.T("core.delete", subject)
// Displays: result.Question + localized [y/N]
// Returns: bool

// Question with options
choice := cli.Question("core.save", i18n.S("changes", 3).Count(3), cli.Options{
    Default: "yes",
    Extra:   []string{"all"},
})
// Displays: "Save 3 changes? [a/y/N]"
// Returns: "yes" | "no" | "all"

// Choice from list
selected := cli.Choose("core.select.branch", branches)
// Displays localized prompt with arrow selection
```

### cli.Confirm()

```go
func Confirm(intent string, subject *i18n.Subject, opts ...ConfirmOption) bool

// Options
cli.DefaultYes()     // Default to yes instead of no
cli.DefaultNo()      // Explicit default no
cli.Required()       // No default, must choose
cli.Timeout(30*time.Second)  // Auto-select default after timeout
```

### cli.Question()

```go
func Question(intent string, subject *i18n.Subject, opts ...QuestionOption) string

// Options
cli.Extra("all", "skip")    // Extra options beyond y/n
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
      "update": "update",
      "commit": "commit"
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
    },
    "commit": {
      "_meta": {
        "type": "action",
        "verb": "common.verb.commit",
        "dangerous": false
      },
      "question": "Commit {{.Subject}}?",
      "success": "{{.Subject | title}} committed",
      "failure": "Failed to commit {{.Subject}}"
    }
  }
}
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

## Implementation Plan

### Phase 1: Foundation
1. Define `Composed` and `Subject` types
2. Add `S()` / `Subject()` builder
3. Add `T()` / `C()` with intent resolution
4. Parse `_meta` from JSON
5. Add template functions (title, lower, past, etc.)

### Phase 2: CLI Integration
1. Implement `cli.Confirm()` using intents
2. Implement `cli.Question()` with options
3. Implement `cli.Choose()` for lists
4. Localize prompt characters [y/N] → [j/N] etc.

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
if !cli.Confirm("core.commit", i18n.S("file", path).Count(len(files))) {
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

## Grammar Fundamentals

Parts of speech we need to handle:

| Part | Role | Example | Transforms |
|------|------|---------|------------|
| **Verb** | Action | delete, save, commit | tense (past/present), mood (imperative) |
| **Noun** | Subject/Object | file, commit, user | plurality, gender, case |
| **Article** | Determiner | a/an, the | vowel-awareness, gender agreement |
| **Adjective** | Describes noun | modified, new, deleted | gender/number agreement |
| **Preposition** | Relation | in, from, to | - |

### Verb Conjugation

```json
{
  "common": {
    "verb": {
      "delete": {
        "base": "delete",
        "past": "deleted",
        "gerund": "deleting",
        "imperative": "delete"
      }
    }
  }
}
```

For most English verbs, derive automatically:
- `past`: base + "ed" (or irregular lookup)
- `gerund`: base + "ing"

### Noun Handling

```json
{
  "common": {
    "noun": {
      "file": {
        "one": "file",
        "other": "files",
        "gender": "neuter"
      }
    }
  }
}
```

### Article Selection

English: a/an based on next word's sound (not letter)
- "a file", "an item", "a user", "an hour"

Other languages: gender agreement (der/die/das, le/la, etc.)

## DX Improvements

### 1. Compile-Time Validation
- `go generate` checks all `T("core.X")` calls have matching JSON keys
- Warns on missing `_meta` fields
- Type-checks template variables

### 2. IDE Support
- JSON schema for autocomplete in translation files
- Go constants generated from JSON keys: `i18n.CoreDelete` instead of `"core.delete"`

### 3. Fallback Chain
```
T("core.delete", subject)
  → try core.delete.question
  → try core.delete (plain string)
  → try common.action.delete
  → return "Delete {{.Subject}}?" (hardcoded fallback)
```

### 4. Debug Mode
```go
i18n.Debug(true)  // Shows: [core.delete] Delete file.txt?
```

### 5. Short Subject Syntax
```go
// Instead of:
i18n.T("core.delete", i18n.S("file", path))

// Allow:
i18n.T("core.delete", path)  // Infers subject type from intent's expected noun
```

### 6. Fluent Chaining
```go
i18n.T("core.delete").
    Subject("file", path).
    Count(3).
    Question()  // Returns just the question string
```

## Notes for Future Implementation

- Use `github.com/gertd/go-pluralize` for English plurality
- Consider `github.com/nicksnyder/go-i18n` patterns for CLDR plural rules
- Store compiled templates in sync.Map for caching
- `_meta` parsing happens once at load time, not per-call
- CLI prompt chars from `common.prompt.*` - allows `[j/N]` for German

## Open Questions

1. **Verb conjugation library** - Use existing Go library or build custom?
2. **Gender detection** - How to infer gender for subjects in gendered languages?
3. **Fallback behavior** - What happens when intent metadata is missing?
4. **Caching** - Should compiled templates be cached?
5. **Validation** - How to validate intent definitions at build time?
