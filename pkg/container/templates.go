package container

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed templates/*.yml
var embeddedTemplates embed.FS

// Template represents a LinuxKit YAML template.
type Template struct {
	// Name is the template identifier (e.g., "core-dev", "server-php").
	Name string
	// Description is a human-readable description of the template.
	Description string
	// Path is the file path to the template (relative or absolute).
	Path string
}

// builtinTemplates defines the metadata for embedded templates.
var builtinTemplates = []Template{
	{
		Name:        "core-dev",
		Description: "Development environment with Go, Node.js, PHP, Docker-in-LinuxKit, and SSH access",
		Path:        "templates/core-dev.yml",
	},
	{
		Name:        "server-php",
		Description: "Production PHP server with FrankenPHP, Caddy reverse proxy, and health checks",
		Path:        "templates/server-php.yml",
	},
}

// ListTemplates returns all available LinuxKit templates.
// It combines embedded templates with any templates found in the user's
// .core/linuxkit directory.
func ListTemplates() []Template {
	templates := make([]Template, len(builtinTemplates))
	copy(templates, builtinTemplates)

	// Check for user templates in .core/linuxkit/
	userTemplatesDir := getUserTemplatesDir()
	if userTemplatesDir != "" {
		userTemplates := scanUserTemplates(userTemplatesDir)
		templates = append(templates, userTemplates...)
	}

	return templates
}

// GetTemplate returns the content of a template by name.
// It first checks embedded templates, then user templates.
func GetTemplate(name string) (string, error) {
	// Check embedded templates first
	for _, t := range builtinTemplates {
		if t.Name == name {
			content, err := embeddedTemplates.ReadFile(t.Path)
			if err != nil {
				return "", fmt.Errorf("failed to read embedded template %s: %w", name, err)
			}
			return string(content), nil
		}
	}

	// Check user templates
	userTemplatesDir := getUserTemplatesDir()
	if userTemplatesDir != "" {
		templatePath := filepath.Join(userTemplatesDir, name+".yml")
		if _, err := os.Stat(templatePath); err == nil {
			content, err := os.ReadFile(templatePath)
			if err != nil {
				return "", fmt.Errorf("failed to read user template %s: %w", name, err)
			}
			return string(content), nil
		}
	}

	return "", fmt.Errorf("template not found: %s", name)
}

// ApplyTemplate applies variable substitution to a template.
// It supports two syntaxes:
//   - ${VAR} - required variable, returns error if not provided
//   - ${VAR:-default} - variable with default value
func ApplyTemplate(name string, vars map[string]string) (string, error) {
	content, err := GetTemplate(name)
	if err != nil {
		return "", err
	}

	return ApplyVariables(content, vars)
}

// ApplyVariables applies variable substitution to content string.
// It supports two syntaxes:
//   - ${VAR} - required variable, returns error if not provided
//   - ${VAR:-default} - variable with default value
func ApplyVariables(content string, vars map[string]string) (string, error) {
	// Pattern for ${VAR:-default} syntax
	defaultPattern := regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*):-([^}]*)\}`)

	// Pattern for ${VAR} syntax (no default)
	requiredPattern := regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

	// Track missing required variables
	var missingVars []string

	// First pass: replace variables with defaults
	result := defaultPattern.ReplaceAllStringFunc(content, func(match string) string {
		submatch := defaultPattern.FindStringSubmatch(match)
		if len(submatch) != 3 {
			return match
		}
		varName := submatch[1]
		defaultVal := submatch[2]

		if val, ok := vars[varName]; ok {
			return val
		}
		return defaultVal
	})

	// Second pass: replace required variables and track missing ones
	result = requiredPattern.ReplaceAllStringFunc(result, func(match string) string {
		submatch := requiredPattern.FindStringSubmatch(match)
		if len(submatch) != 2 {
			return match
		}
		varName := submatch[1]

		if val, ok := vars[varName]; ok {
			return val
		}
		missingVars = append(missingVars, varName)
		return match // Keep original if missing
	})

	if len(missingVars) > 0 {
		return "", fmt.Errorf("missing required variables: %s", strings.Join(missingVars, ", "))
	}

	return result, nil
}

// ExtractVariables extracts all variable names from a template.
// Returns two slices: required variables and optional variables (with defaults).
func ExtractVariables(content string) (required []string, optional map[string]string) {
	optional = make(map[string]string)
	requiredSet := make(map[string]bool)

	// Pattern for ${VAR:-default} syntax
	defaultPattern := regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*):-([^}]*)\}`)

	// Pattern for ${VAR} syntax (no default)
	requiredPattern := regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

	// Find optional variables with defaults
	matches := defaultPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) == 3 {
			optional[match[1]] = match[2]
		}
	}

	// Find required variables
	matches = requiredPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) == 2 {
			varName := match[1]
			// Only add if not already in optional (with default)
			if _, hasDefault := optional[varName]; !hasDefault {
				requiredSet[varName] = true
			}
		}
	}

	// Convert set to slice
	for v := range requiredSet {
		required = append(required, v)
	}

	return required, optional
}

// getUserTemplatesDir returns the path to user templates directory.
// Returns empty string if the directory doesn't exist.
func getUserTemplatesDir() string {
	// Try workspace-relative .core/linuxkit first
	cwd, err := os.Getwd()
	if err == nil {
		wsDir := filepath.Join(cwd, ".core", "linuxkit")
		if info, err := os.Stat(wsDir); err == nil && info.IsDir() {
			return wsDir
		}
	}

	// Try home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	homeDir := filepath.Join(home, ".core", "linuxkit")
	if info, err := os.Stat(homeDir); err == nil && info.IsDir() {
		return homeDir
	}

	return ""
}

// scanUserTemplates scans a directory for .yml template files.
func scanUserTemplates(dir string) []Template {
	var templates []Template

	entries, err := os.ReadDir(dir)
	if err != nil {
		return templates
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}

		// Extract template name from filename
		templateName := strings.TrimSuffix(strings.TrimSuffix(name, ".yml"), ".yaml")

		// Skip if this is a builtin template name (embedded takes precedence)
		isBuiltin := false
		for _, bt := range builtinTemplates {
			if bt.Name == templateName {
				isBuiltin = true
				break
			}
		}
		if isBuiltin {
			continue
		}

		// Read file to extract description from comments
		description := extractTemplateDescription(filepath.Join(dir, name))
		if description == "" {
			description = "User-defined template"
		}

		templates = append(templates, Template{
			Name:        templateName,
			Description: description,
			Path:        filepath.Join(dir, name),
		})
	}

	return templates
}

// extractTemplateDescription reads the first comment block from a YAML file
// to use as a description.
func extractTemplateDescription(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	var descLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Remove the # and trim
			comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if comment != "" {
				descLines = append(descLines, comment)
				// Only take the first meaningful comment line as description
				if len(descLines) == 1 {
					return comment
				}
			}
		} else if trimmed != "" {
			// Hit non-comment content, stop
			break
		}
	}

	if len(descLines) > 0 {
		return descLines[0]
	}
	return ""
}
