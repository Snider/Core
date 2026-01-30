// Package i18n provides internationalization for the CLI.
//
// Locale files use nested JSON for compatibility with translation tools:
//
//	{
//	    "cli": {
//	        "success": "Operation completed",
//	        "count": {
//	            "items": {
//	                "one": "{{.Count}} item",
//	                "other": "{{.Count}} items"
//	            }
//	        }
//	    }
//	}
//
// Keys are accessed with dot notation: T("cli.success"), T("cli.count.items")
//
// # Getting Started
//
//	svc, err := i18n.New()
//	fmt.Println(svc.T("cli.success"))
//	fmt.Println(svc.T("cli.count.items", map[string]any{"Count": 5}))
package i18n

import (
	"bytes"
	"errors"
	"strings"
	"text/template"
)

// --- Global convenience functions ---

// T translates a message using the default service.
// For semantic intents (core.* namespace), pass a Subject as the first argument.
//
//	T("cli.success")                           // Simple translation
//	T("core.delete", S("file", "config.yaml")) // Semantic intent
func T(messageID string, args ...any) string {
	if svc := Default(); svc != nil {
		return svc.T(messageID, args...)
	}
	return messageID
}

// Raw is the raw translation helper without i18n.* namespace magic.
// Unlike T(), this does NOT handle i18n.* namespace patterns.
// Use this for direct key lookups without auto-composition.
//
//	Raw("cli.success")              // Direct lookup
//	T("i18n.label.status")          // Smart: returns "Status:"
func Raw(messageID string, args ...any) string {
	if svc := Default(); svc != nil {
		return svc.Raw(messageID, args...)
	}
	return messageID
}

// ErrServiceNotInitialized is returned when the i18n service is not initialized.
var ErrServiceNotInitialized = errors.New("i18n: service not initialized")

// SetLanguage sets the language for the default service.
// Returns ErrServiceNotInitialized if the service has not been initialized,
// or an error if the language tag is invalid or unsupported.
//
// Unlike other Set* functions, this returns an error because it validates
// the language tag against available locales.
func SetLanguage(lang string) error {
	svc := Default()
	if svc == nil {
		return ErrServiceNotInitialized
	}
	return svc.SetLanguage(lang)
}

// CurrentLanguage returns the current language code from the default service.
// Returns "en-GB" (the fallback language) if the service is not initialized.
func CurrentLanguage() string {
	if svc := Default(); svc != nil {
		return svc.Language()
	}
	return "en-GB"
}

// SetMode sets the translation mode for the default service.
// Does nothing if the service is not initialized.
func SetMode(m Mode) {
	if svc := Default(); svc != nil {
		svc.SetMode(m)
	}
}

// CurrentMode returns the current translation mode from the default service.
func CurrentMode() Mode {
	if svc := Default(); svc != nil {
		return svc.Mode()
	}
	return ModeNormal
}

// N formats a number using the i18n.numeric.* namespace.
// Wrapper for T("i18n.numeric.{format}", value).
//
//	N("number", 1234567)   // T("i18n.numeric.number", 1234567)
//	N("percent", 0.85)     // T("i18n.numeric.percent", 0.85)
//	N("bytes", 1536000)    // T("i18n.numeric.bytes", 1536000)
//	N("ordinal", 1)        // T("i18n.numeric.ordinal", 1)
func N(format string, value any) string {
	return T("i18n.numeric."+format, value)
}

// AddHandler appends a handler to the default service's handler chain.
// Does nothing if the service is not initialized.
func AddHandler(h KeyHandler) {
	if svc := Default(); svc != nil {
		svc.AddHandler(h)
	}
}

// PrependHandler inserts a handler at the start of the default service's handler chain.
// Does nothing if the service is not initialized.
func PrependHandler(h KeyHandler) {
	if svc := Default(); svc != nil {
		svc.PrependHandler(h)
	}
}

// --- Template helpers ---

// executeIntentTemplate executes an intent template with the given data.
// Templates are cached for performance - repeated calls with the same template
// string will reuse the compiled template.
func executeIntentTemplate(tmplStr string, data templateData) string {
	if tmplStr == "" {
		return ""
	}

	// Check cache first
	if cached, ok := templateCache.Load(tmplStr); ok {
		var buf bytes.Buffer
		if err := cached.(*template.Template).Execute(&buf, data); err != nil {
			return tmplStr
		}
		return buf.String()
	}

	// Parse and cache
	tmpl, err := template.New("").Funcs(TemplateFuncs()).Parse(tmplStr)
	if err != nil {
		return tmplStr
	}

	// Store in cache (safe even if another goroutine stored it first)
	templateCache.Store(tmplStr, tmpl)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return tmplStr
	}
	return buf.String()
}

func applyTemplate(text string, data any) string {
	// Quick check for template syntax
	if !strings.Contains(text, "{{") {
		return text
	}

	// Check cache first
	if cached, ok := templateCache.Load(text); ok {
		var buf bytes.Buffer
		if err := cached.(*template.Template).Execute(&buf, data); err != nil {
			return text
		}
		return buf.String()
	}

	// Parse and cache
	tmpl, err := template.New("").Parse(text)
	if err != nil {
		return text
	}

	templateCache.Store(text, tmpl)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return text
	}
	return buf.String()
}
