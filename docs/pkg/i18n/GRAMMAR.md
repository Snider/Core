# Grammar Engine

The i18n grammar engine automatically handles verb conjugation, noun pluralisation, and article selection. It uses a combination of locale-defined rules and built-in English defaults.

## Verb Conjugation

### Past Tense

```go
i18n.PastTense("delete")   // "deleted"
i18n.PastTense("create")   // "created"
i18n.PastTense("run")      // "ran" (irregular)
i18n.PastTense("build")    // "built" (irregular)
```

**Rules applied (in order):**

1. Check locale JSON `gram.verb.{verb}.past`
2. Check built-in irregular verbs map
3. Apply regular conjugation rules:
   - Ends in 'e' → add 'd' (delete → deleted)
   - Ends in consonant + 'y' → change to 'ied' (try → tried)
   - Short verb ending in CVC → double consonant (stop → stopped)
   - Otherwise → add 'ed' (walk → walked)

### Gerund (-ing form)

```go
i18n.Gerund("build")   // "building"
i18n.Gerund("run")     // "running"
i18n.Gerund("make")    // "making"
i18n.Gerund("die")     // "dying"
```

**Rules applied:**

1. Check locale JSON `gram.verb.{verb}.gerund`
2. Check built-in irregular verbs map
3. Apply regular rules:
   - Ends in 'ie' → change to 'ying' (die → dying)
   - Ends in 'e' (not 'ee') → drop 'e', add 'ing' (make → making)
   - Short verb ending in CVC → double consonant (run → running)
   - Otherwise → add 'ing' (build → building)

## Noun Pluralisation

```go
i18n.Pluralize("file", 1)     // "file"
i18n.Pluralize("file", 5)     // "files"
i18n.Pluralize("child", 2)    // "children" (irregular)
i18n.Pluralize("analysis", 3) // "analyses" (Latin)
```

**Rules applied (in order):**

1. Check locale JSON `gram.noun.{noun}.other`
2. Check built-in irregular nouns map
3. Apply regular rules:
   - Ends in 's', 'x', 'z', 'ch', 'sh' → add 'es'
   - Ends in consonant + 'y' → change to 'ies'
   - Ends in 'f' or 'fe' → change to 'ves' (leaf → leaves)
   - Otherwise → add 's'

### Built-in Irregular Nouns

| Singular | Plural |
|----------|--------|
| child | children |
| person | people |
| man | men |
| woman | women |
| foot | feet |
| tooth | teeth |
| mouse | mice |
| datum | data |
| index | indices |
| crisis | crises |
| fish | fish |
| sheep | sheep |

## Articles

```go
i18n.Article("apple")     // "an apple"
i18n.Article("banana")    // "a banana"
i18n.Article("hour")      // "an hour" (silent h)
i18n.Article("user")      // "a user" (y sound)
i18n.Article("umbrella")  // "an umbrella"
```

**Rules:**

1. Vowel sound words get "an" (a, e, i, o, u start)
2. Consonant sound words get "a"
3. Exception lists handle:
   - Silent 'h' words: hour, honest, honour, heir, herb
   - 'Y' sound words: user, union, unique, unit, universe

## Composed Messages

### Labels

```go
i18n.Label("status")   // "Status:"
i18n.Label("version")  // "Version:"
```

Uses `gram.punct.label` suffix (default `:`) from locale.

### Progress Messages

```go
i18n.Progress("build")                  // "Building..."
i18n.ProgressSubject("check", "config") // "Checking config..."
```

Uses `gram.punct.progress` suffix (default `...`) from locale.

### Action Results

```go
i18n.ActionResult("delete", "file")    // "File deleted"
i18n.ActionResult("create", "project") // "Project created"
```

Pattern: `{Title(subject)} {past(verb)}`

### Action Failures

```go
i18n.ActionFailed("delete", "file")  // "Failed to delete file"
i18n.ActionFailed("save", "config")  // "Failed to save config"
```

Pattern: `Failed to {verb} {subject}`

## Locale Configuration

Define grammar in your locale JSON:

```json
{
  "gram": {
    "verb": {
      "deploy": {
        "past": "deployed",
        "gerund": "deploying"
      },
      "sync": {
        "past": "synced",
        "gerund": "syncing"
      }
    },
    "noun": {
      "repository": {
        "one": "repository",
        "other": "repositories"
      },
      "schema": {
        "one": "schema",
        "other": "schemata"
      }
    },
    "article": {
      "indefinite": {
        "default": "a",
        "vowel": "an"
      },
      "definite": "the"
    },
    "punct": {
      "label": ":",
      "progress": "..."
    },
    "word": {
      "status": "status",
      "version": "version"
    }
  }
}
```

## Template Functions

Use grammar functions in templates:

```go
template.New("").Funcs(i18n.TemplateFuncs())
```

| Function | Example | Result |
|----------|---------|--------|
| `past` | `{{past "delete"}}` | "deleted" |
| `gerund` | `{{gerund "build"}}` | "building" |
| `plural` | `{{plural "file" 5}}` | "files" |
| `article` | `{{article "apple"}}` | "an apple" |
| `title` | `{{title "hello world"}}` | "Hello World" |
| `lower` | `{{lower "HELLO"}}` | "hello" |
| `upper` | `{{upper "hello"}}` | "HELLO" |
| `quote` | `{{quote "text"}}` | `"text"` |

## Language-Specific Grammar

The grammar engine loads language-specific data when available:

```go
// Get grammar data for a language
data := i18n.GetGrammarData("de-DE")
if data != nil {
    // Access verb forms, noun forms, etc.
}

// Set grammar data programmatically
i18n.SetGrammarData("de-DE", &i18n.GrammarData{
    Verbs: map[string]i18n.VerbForms{
        "machen": {Past: "gemacht", Gerund: "machend"},
    },
})
```

## Performance

Grammar results are computed on-demand but templates are cached:

- First call: Parse template + apply grammar
- Subsequent calls: Reuse cached template

The template cache uses `sync.Map` for thread-safe concurrent access.
