# Extending the i18n Package

This guide covers how to extend the i18n package with custom loaders, handlers, and integrations.

## Custom Loaders

The `Loader` interface allows loading translations from any source:

```go
type Loader interface {
    Load(lang string) (map[string]Message, *GrammarData, error)
    Languages() []string
}
```

### Database Loader Example

```go
type PostgresLoader struct {
    db *sql.DB
}

func (l *PostgresLoader) Languages() []string {
    rows, err := l.db.Query("SELECT DISTINCT lang FROM translations")
    if err != nil {
        return nil
    }
    defer rows.Close()

    var langs []string
    for rows.Next() {
        var lang string
        rows.Scan(&lang)
        langs = append(langs, lang)
    }
    return langs
}

func (l *PostgresLoader) Load(lang string) (map[string]i18n.Message, *i18n.GrammarData, error) {
    rows, err := l.db.Query(
        "SELECT key, text, plural_one, plural_other FROM translations WHERE lang = $1",
        lang,
    )
    if err != nil {
        return nil, nil, err
    }
    defer rows.Close()

    messages := make(map[string]i18n.Message)
    for rows.Next() {
        var key, text string
        var one, other sql.NullString
        rows.Scan(&key, &text, &one, &other)

        if one.Valid || other.Valid {
            messages[key] = i18n.Message{One: one.String, Other: other.String}
        } else {
            messages[key] = i18n.Message{Text: text}
        }
    }

    return messages, nil, nil
}

// Usage
svc, err := i18n.NewWithLoader(&PostgresLoader{db: db})
```

### Remote API Loader Example

```go
type APILoader struct {
    baseURL string
    client  *http.Client
}

func (l *APILoader) Languages() []string {
    resp, _ := l.client.Get(l.baseURL + "/languages")
    defer resp.Body.Close()

    var langs []string
    json.NewDecoder(resp.Body).Decode(&langs)
    return langs
}

func (l *APILoader) Load(lang string) (map[string]i18n.Message, *i18n.GrammarData, error) {
    resp, err := l.client.Get(l.baseURL + "/translations/" + lang)
    if err != nil {
        return nil, nil, err
    }
    defer resp.Body.Close()

    var data struct {
        Messages map[string]i18n.Message `json:"messages"`
        Grammar  *i18n.GrammarData       `json:"grammar"`
    }
    json.NewDecoder(resp.Body).Decode(&data)

    return data.Messages, data.Grammar, nil
}
```

### Multi-Source Loader

Combine multiple loaders with fallback:

```go
type FallbackLoader struct {
    primary   i18n.Loader
    secondary i18n.Loader
}

func (l *FallbackLoader) Languages() []string {
    // Merge languages from both sources
    langs := make(map[string]bool)
    for _, lang := range l.primary.Languages() {
        langs[lang] = true
    }
    for _, lang := range l.secondary.Languages() {
        langs[lang] = true
    }

    result := make([]string, 0, len(langs))
    for lang := range langs {
        result = append(result, lang)
    }
    return result
}

func (l *FallbackLoader) Load(lang string) (map[string]i18n.Message, *i18n.GrammarData, error) {
    msgs, grammar, err := l.primary.Load(lang)
    if err != nil {
        return l.secondary.Load(lang)
    }

    // Merge with secondary for missing keys
    secondary, secGrammar, _ := l.secondary.Load(lang)
    for k, v := range secondary {
        if _, exists := msgs[k]; !exists {
            msgs[k] = v
        }
    }

    if grammar == nil {
        grammar = secGrammar
    }

    return msgs, grammar, nil
}
```

## Custom Handlers

Handlers process keys before standard lookup. Use for dynamic patterns.

### Handler Interface

```go
type KeyHandler interface {
    Match(key string) bool
    Handle(key string, args []any, next func() string) string
}
```

### Emoji Handler Example

```go
type EmojiHandler struct{}

func (h EmojiHandler) Match(key string) bool {
    return strings.HasPrefix(key, "emoji.")
}

func (h EmojiHandler) Handle(key string, args []any, next func() string) string {
    name := strings.TrimPrefix(key, "emoji.")
    emojis := map[string]string{
        "success": "✅",
        "error":   "❌",
        "warning": "⚠️",
        "info":    "ℹ️",
    }
    if emoji, ok := emojis[name]; ok {
        return emoji
    }
    return next() // Delegate to next handler
}

// Usage
i18n.AddHandler(EmojiHandler{})
i18n.T("emoji.success")  // "✅"
```

### Conditional Handler Example

```go
type FeatureFlagHandler struct {
    flags map[string]bool
}

func (h FeatureFlagHandler) Match(key string) bool {
    return strings.HasPrefix(key, "feature.")
}

func (h FeatureFlagHandler) Handle(key string, args []any, next func() string) string {
    feature := strings.TrimPrefix(key, "feature.")
    parts := strings.SplitN(feature, ".", 2)

    if len(parts) < 2 {
        return next()
    }

    flag, subkey := parts[0], parts[1]
    if h.flags[flag] {
        // Feature enabled - translate the subkey
        return i18n.T(subkey, args...)
    }

    // Feature disabled - return empty or fallback
    return ""
}
```

### Handler Chain Priority

```go
// Prepend for highest priority (runs first)
svc.PrependHandler(CriticalHandler{})

// Append for lower priority (runs after defaults)
svc.AddHandler(FallbackHandler{})

// Clear all handlers
svc.ClearHandlers()

// Add back defaults
svc.AddHandler(i18n.DefaultHandlers()...)
```

## Integrating with Frameworks

### Cobra CLI

```go
func init() {
    // Initialise i18n before command setup
    if err := i18n.Init(); err != nil {
        log.Fatal(err)
    }
}

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: i18n.T("cmd.root.short"),
    Long:  i18n.T("cmd.root.long"),
}

var buildCmd = &cobra.Command{
    Use:   "build",
    Short: i18n.T("cmd.build.short"),
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Println(i18n.T("i18n.progress.build"))
        // ...
        fmt.Println(i18n.T("i18n.done.build", "project"))
        return nil
    },
}
```

### Error Messages

```go
type LocalisedError struct {
    Key  string
    Args map[string]any
}

func (e LocalisedError) Error() string {
    return i18n.T(e.Key, e.Args)
}

// Usage
return LocalisedError{
    Key:  "error.file_not_found",
    Args: map[string]any{"Name": filename},
}
```

### Structured Logging

```go
func LogInfo(key string, args ...any) {
    msg := i18n.T(key, args...)
    slog.Info(msg, "i18n_key", key)
}

func LogError(key string, err error, args ...any) {
    msg := i18n.T(key, args...)
    slog.Error(msg, "i18n_key", key, "error", err)
}
```

## Testing

### Mock Loader for Tests

```go
type MockLoader struct {
    messages map[string]map[string]i18n.Message
}

func (l *MockLoader) Languages() []string {
    langs := make([]string, 0, len(l.messages))
    for lang := range l.messages {
        langs = append(langs, lang)
    }
    return langs
}

func (l *MockLoader) Load(lang string) (map[string]i18n.Message, *i18n.GrammarData, error) {
    if msgs, ok := l.messages[lang]; ok {
        return msgs, nil, nil
    }
    return nil, nil, fmt.Errorf("language not found: %s", lang)
}

// Usage in tests
func TestMyFeature(t *testing.T) {
    loader := &MockLoader{
        messages: map[string]map[string]i18n.Message{
            "en-GB": {
                "test.greeting": {Text: "Hello"},
                "test.farewell": {Text: "Goodbye"},
            },
        },
    }

    svc, _ := i18n.NewWithLoader(loader)
    i18n.SetDefault(svc)

    // Test your code
    assert.Equal(t, "Hello", i18n.T("test.greeting"))
}
```

### Testing Missing Keys

```go
func TestMissingKeys(t *testing.T) {
    svc, _ := i18n.New(i18n.WithMode(i18n.ModeCollect))
    i18n.SetDefault(svc)

    var missing []string
    i18n.OnMissingKey(func(m i18n.MissingKey) {
        missing = append(missing, m.Key)
    })

    // Run your code that uses translations
    runMyFeature()

    // Check for missing keys
    assert.Empty(t, missing, "Found missing translation keys: %v", missing)
}
```

## Hot Reloading

Implement a loader that watches for file changes:

```go
type HotReloadLoader struct {
    base    *i18n.FSLoader
    service *i18n.Service
    watcher *fsnotify.Watcher
}

func (l *HotReloadLoader) Watch() {
    for {
        select {
        case event := <-l.watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                // Reload translations
                l.service.LoadFS(os.DirFS("."), "locales")
            }
        }
    }
}
```

## Performance Considerations

1. **Cache translations**: The service caches all loaded messages
2. **Template caching**: Parsed templates are cached in `sync.Map`
3. **Handler chain**: Keep chain short (6 default handlers is fine)
4. **Grammar cache**: Grammar lookups are cached per-language

For high-throughput applications:
- Pre-warm the cache by calling common translations at startup
- Consider using `Raw()` to bypass handler chain when not needed
- Profile with `go test -bench` if performance is critical
