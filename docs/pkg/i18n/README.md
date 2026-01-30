# i18n Package

The `pkg/i18n` package provides internationalisation and localisation for Go CLI applications. It features a grammar engine for automatic verb conjugation and noun pluralisation, CLDR plural support, and an extensible handler chain for dynamic key patterns.

## Quick Start

```go
import "github.com/host-uk/core/pkg/i18n"

func main() {
    // Initialise with embedded locales
    svc, err := i18n.New()
    if err != nil {
        log.Fatal(err)
    }
    i18n.SetDefault(svc)

    // Translate messages
    fmt.Println(i18n.T("cli.success"))                    // "Operation completed"
    fmt.Println(i18n.T("i18n.count.file", 5))             // "5 files"
    fmt.Println(i18n.T("i18n.progress.build"))            // "Building..."
    fmt.Println(i18n.T("i18n.done.delete", "config.yaml")) // "Config.yaml deleted"
}
```

## Table of Contents

- [Basic Translation](#basic-translation)
- [Template Variables](#template-variables)
- [Pluralisation](#pluralisation)
- [Magic Namespaces](#magic-namespaces)
- [Subjects](#subjects)
- [Grammar Engine](#grammar-engine)
- [Formality](#formality)
- [Modes](#modes)
- [Custom Loaders](#custom-loaders)
- [Custom Handlers](#custom-handlers)
- [Locale File Format](#locale-file-format)

## Basic Translation

The `T()` function translates message keys:

```go
// Simple translation
msg := i18n.T("cli.success")

// With template variables
msg := i18n.T("error.not_found", map[string]any{
    "Name": "config.yaml",
})
```

Use `Raw()` to bypass magic namespace handling:

```go
// T() handles i18n.* magic
i18n.T("i18n.label.status")  // "Status:"

// Raw() does direct lookup only
i18n.Raw("i18n.label.status") // Returns key as-is (not in JSON)
```

## Template Variables

Translation strings support Go templates:

```json
{
  "greeting": "Hello, {{.Name}}!",
  "summary": "Found {{.Count}} {{if eq .Count 1}}item{{else}}items{{end}}"
}
```

```go
i18n.T("greeting", map[string]any{"Name": "World"})  // "Hello, World!"
i18n.T("summary", map[string]any{"Count": 3})        // "Found 3 items"
```

### Available Template Functions

| Function | Description | Example |
|----------|-------------|---------|
| `title` | Title case | `{{title .Name}}` |
| `lower` | Lowercase | `{{lower .Name}}` |
| `upper` | Uppercase | `{{upper .Name}}` |
| `past` | Past tense | `{{past "delete"}}` → "deleted" |
| `gerund` | -ing form | `{{gerund "build"}}` → "building" |
| `plural` | Pluralise | `{{plural "file" .Count}}` |
| `article` | Add article | `{{article "apple"}}` → "an apple" |
| `quote` | Add quotes | `{{quote .Name}}` → `"name"` |

## Pluralisation

The package supports full CLDR plural categories:

```json
{
  "item_count": {
    "zero": "No items",
    "one": "{{.Count}} item",
    "two": "{{.Count}} items",
    "few": "{{.Count}} items",
    "many": "{{.Count}} items",
    "other": "{{.Count}} items"
  }
}
```

```go
i18n.T("item_count", map[string]any{"Count": 0})  // "No items" (if zero defined)
i18n.T("item_count", map[string]any{"Count": 1})  // "1 item"
i18n.T("item_count", map[string]any{"Count": 5})  // "5 items"
```

For simple cases, use `i18n.count.*`:

```go
i18n.T("i18n.count.file", 1)  // "1 file"
i18n.T("i18n.count.file", 5)  // "5 files"
```

## Magic Namespaces

The `i18n.*` namespace provides automatic message composition:

### Labels (`i18n.label.*`)

```go
i18n.T("i18n.label.status")   // "Status:"
i18n.T("i18n.label.version")  // "Version:"
```

### Progress (`i18n.progress.*`)

```go
i18n.T("i18n.progress.build")                // "Building..."
i18n.T("i18n.progress.check", "config")      // "Checking config..."
```

### Counts (`i18n.count.*`)

```go
i18n.T("i18n.count.file", 1)   // "1 file"
i18n.T("i18n.count.file", 5)   // "5 files"
i18n.T("i18n.count.repo", 10)  // "10 repos"
```

### Done (`i18n.done.*`)

```go
i18n.T("i18n.done.delete", "file")     // "File deleted"
i18n.T("i18n.done.create", "project")  // "Project created"
```

### Fail (`i18n.fail.*`)

```go
i18n.T("i18n.fail.delete", "file")  // "Failed to delete file"
i18n.T("i18n.fail.save", "config")  // "Failed to save config"
```

### Numeric (`i18n.numeric.*`)

```go
i18n.N("number", 1234567)   // "1,234,567"
i18n.N("percent", 0.85)     // "85%"
i18n.N("bytes", 1536000)    // "1.46 MB"
i18n.N("ordinal", 1)        // "1st"
```

## Subjects

Subjects provide typed context for translations:

```go
// Create a subject
subj := i18n.S("file", "config.yaml")

// Chain methods for additional context
subj := i18n.S("file", files).
    Count(len(files)).
    In("workspace").
    Formal()

// Use in translations
i18n.T("i18n.done.delete", subj.String())
```

### Subject Methods

| Method | Description |
|--------|-------------|
| `Count(n)` | Set count for pluralisation |
| `Gender(g)` | Set grammatical gender |
| `In(loc)` | Set location context |
| `Formal()` | Set formal address |
| `Informal()` | Set informal address |

## Grammar Engine

The grammar engine handles verb conjugation and noun forms:

```go
// Verb conjugation
i18n.PastTense("delete")   // "deleted"
i18n.PastTense("run")      // "ran" (irregular)
i18n.Gerund("build")       // "building"
i18n.Gerund("run")         // "running"

// Noun pluralisation
i18n.Pluralize("file", 1)  // "file"
i18n.Pluralize("file", 5)  // "files"
i18n.Pluralize("child", 2) // "children" (irregular)

// Articles
i18n.Article("apple")      // "an apple"
i18n.Article("banana")     // "a banana"

// Composed messages
i18n.Label("status")               // "Status:"
i18n.Progress("build")             // "Building..."
i18n.ProgressSubject("check", "cfg") // "Checking cfg..."
i18n.ActionResult("delete", "file")  // "File deleted"
i18n.ActionFailed("save", "config")  // "Failed to save config"
```

### Customising Grammar

Add irregular forms in your locale JSON:

```json
{
  "gram": {
    "verb": {
      "deploy": { "past": "deployed", "gerund": "deploying" }
    },
    "noun": {
      "repository": { "one": "repository", "other": "repositories" }
    },
    "punct": {
      "label": ":",
      "progress": "..."
    }
  }
}
```

## Formality

For languages with formal/informal address (German Sie/du, French vous/tu):

```go
// Set service-wide formality
svc.SetFormality(i18n.FormalityFormal)

// Per-translation formality via Subject
i18n.T("greeting", i18n.S("user", name).Formal())
i18n.T("greeting", i18n.S("user", name).Informal())

// Per-translation via TranslationContext
i18n.T("greeting", i18n.C("customer support").Formal())
```

Define formality variants in JSON:

```json
{
  "greeting": "Hello",
  "greeting._formal": "Good morning, sir",
  "greeting._informal": "Hey there"
}
```

## Modes

Three modes control missing key behaviour:

```go
// Normal (default): Returns key as-is
i18n.SetMode(i18n.ModeNormal)
i18n.T("missing.key")  // "missing.key"

// Strict: Panics on missing keys (dev/CI)
i18n.SetMode(i18n.ModeStrict)
i18n.T("missing.key")  // panic!

// Collect: Dispatches to handler (QA testing)
i18n.SetMode(i18n.ModeCollect)
i18n.OnMissingKey(func(m i18n.MissingKey) {
    log.Printf("MISSING: %s at %s:%d", m.Key, m.CallerFile, m.CallerLine)
})
```

## Custom Loaders

Implement the `Loader` interface for custom storage:

```go
type Loader interface {
    Load(lang string) (map[string]Message, *GrammarData, error)
    Languages() []string
}
```

Example database loader:

```go
type DBLoader struct {
    db *sql.DB
}

func (l *DBLoader) Languages() []string {
    // Query available languages from database
}

func (l *DBLoader) Load(lang string) (map[string]i18n.Message, *i18n.GrammarData, error) {
    // Load translations from database
}

// Use custom loader
svc, err := i18n.NewWithLoader(&DBLoader{db: db})
```

## Custom Handlers

Add custom key handlers for dynamic patterns:

```go
type MyHandler struct{}

func (h MyHandler) Match(key string) bool {
    return strings.HasPrefix(key, "my.prefix.")
}

func (h MyHandler) Handle(key string, args []any, next func() string) string {
    // Handle the key or call next() to delegate
    return "custom result"
}

// Add to handler chain
svc.AddHandler(MyHandler{})      // Append (lower priority)
svc.PrependHandler(MyHandler{})  // Prepend (higher priority)
```

## Locale File Format

Locale files use nested JSON with dot-notation access:

```json
{
  "cli": {
    "success": "Operation completed",
    "error": {
      "not_found": "{{.Name}} not found"
    }
  },
  "cmd": {
    "build": {
      "short": "Build the project",
      "long": "Build compiles source files into an executable"
    }
  },
  "gram": {
    "verb": {
      "build": { "past": "built", "gerund": "building" }
    },
    "noun": {
      "file": { "one": "file", "other": "files" }
    },
    "punct": {
      "label": ":",
      "progress": "..."
    }
  }
}
```

Access keys with dot notation:

```go
i18n.T("cli.success")           // "Operation completed"
i18n.T("cli.error.not_found")   // "{{.Name}} not found"
i18n.T("cmd.build.short")       // "Build the project"
```

## Configuration Options

Use functional options when creating a service:

```go
svc, err := i18n.New(
    i18n.WithFallback("de-DE"),           // Fallback language
    i18n.WithFormality(i18n.FormalityFormal),  // Default formality
    i18n.WithMode(i18n.ModeStrict),       // Missing key mode
    i18n.WithDebug(true),                 // Show [key] prefix
)
```

## Thread Safety

The package is fully thread-safe:

- `Service` uses `sync.RWMutex` for state
- Global `Default()` uses `atomic.Pointer`
- `OnMissingKey` uses `atomic.Value`
- `FSLoader.Languages()` uses `sync.Once`

Safe for concurrent use from multiple goroutines.

## Debug Mode

Enable debug mode to see translation keys:

```go
i18n.SetDebug(true)
i18n.T("cli.success")  // "[cli.success] Operation completed"
```

Useful for identifying which keys are used where.
