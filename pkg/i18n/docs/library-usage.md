# Library Usage

The `i18n` package provides a simple and robust internationalization service for Go applications. It handles locale loading, language detection, and message translation.

## Installation

```bash
go get github.com/snider/i18n/i18n
```

## Basic Usage

### 1. Import the Package

```go
import "github.com/snider/i18n/i18n"
```

### 2. Initialize the Service

Create a new instance of the i18n service. This will load the embedded locales.

```go
i18nService, err := i18n.New()
if err != nil {
    log.Fatal(err)
}
```

### 3. Set the Language

Set the desired language for the service. The language tag should be a valid BCP 47 tag (e.g., "en", "es", "fr-CA").

```go
err = i18nService.SetLanguage("es")
if err != nil {
    log.Printf("Language not supported: %v", err)
}
```

### 4. Translate Messages

Use the `Translate` method to retrieve localized messages by their ID.

```go
// Simple translation
greeting := i18nService.Translate("greeting")
fmt.Println(greeting)

// Translation with template data
data := map[string]interface{}{
    "Name": "Alice",
}
welcome := i18nService.Translate("welcome_message", data)
fmt.Println(welcome)
```

## Managing Locales

Locales are stored as JSON files in the `locales` directory within the `i18n` package. The filenames correspond to the language tags (e.g., `en.json`, `es.json`).

**Example `locales/en.json`:**
```json
{
  "greeting": "Hello",
  "welcome_message": "Welcome, {{.Name}}!"
}
```

These files are embedded into the Go binary, so they are available at runtime without needing external file access.

## Language Support

The service automatically detects available languages based on the files present in the `locales` directory. You can use any language defined there.
