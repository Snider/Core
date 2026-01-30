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
	"embed"
	"strings"
	"sync"
	"text/template"
)

//go:embed locales/*.json
var localeFS embed.FS

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

// _ is the raw gettext-style translation helper.
// Unlike T(), this does NOT handle core.* namespace magic.
// Use this for direct key lookups without auto-composition.
//
//	i18n._("cli.success")           // Raw lookup
//	i18n.T("i18n.label.status")     // Smart: returns "Status:"
func _(messageID string, args ...any) string {
	if svc := Default(); svc != nil {
		return svc.Raw(messageID, args...)
	}
	return messageID
}

// --- Template helpers ---

// templateCache stores compiled templates for reuse.
// Key is the template string, value is the compiled template.
var templateCache sync.Map

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
