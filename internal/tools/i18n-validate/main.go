// Command i18n-validate scans Go source files for i18n key usage and validates
// them against the locale JSON files.
//
// Usage:
//
//	go run ./cmd/i18n-validate ./...
//	go run ./cmd/i18n-validate ./pkg/cli ./cmd/dev
//
// The validator checks:
//   - T("key") calls - validates key exists in locale files
//   - C("intent", ...) calls - validates intent exists in registered intents
//   - i18n.T("key") and i18n.C("intent", ...) qualified calls
//
// Exit codes:
//   - 0: All keys valid
//   - 1: Missing keys found
//   - 2: Error during validation
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// KeyUsage records where a key is used in the source code.
type KeyUsage struct {
	Key      string
	File     string
	Line     int
	Function string // "T" or "C"
}

// ValidationResult holds the results of validation.
type ValidationResult struct {
	TotalKeys    int
	ValidKeys    int
	MissingKeys  []KeyUsage
	IntentKeys   int
	MessageKeys  int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: i18n-validate <packages...>")
		fmt.Fprintln(os.Stderr, "Example: i18n-validate ./...")
		os.Exit(2)
	}

	// Find the project root (where locales are)
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project root: %v\n", err)
		os.Exit(2)
	}

	// Load valid keys from locale files
	validKeys, err := loadValidKeys(filepath.Join(root, "pkg/i18n/locales"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading locale files: %v\n", err)
		os.Exit(2)
	}

	// Load valid intents
	validIntents := loadValidIntents()

	// Scan source files
	usages, err := scanPackages(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning packages: %v\n", err)
		os.Exit(2)
	}

	// Validate
	result := validate(usages, validKeys, validIntents)

	// Report
	printReport(result)

	if len(result.MissingKeys) > 0 {
		os.Exit(1)
	}
}

// findProjectRoot finds the project root by looking for go.mod.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod in any parent directory")
		}
		dir = parent
	}
}

// loadValidKeys loads all valid keys from locale JSON files.
func loadValidKeys(localesDir string) (map[string]bool, error) {
	keys := make(map[string]bool)

	entries, err := os.ReadDir(localesDir)
	if err != nil {
		return nil, fmt.Errorf("reading locales dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(localesDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		extractKeys("", raw, keys)
	}

	return keys, nil
}

// extractKeys recursively extracts flattened keys from nested JSON.
func extractKeys(prefix string, data map[string]any, out map[string]bool) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			out[fullKey] = true
		case map[string]any:
			// Check if it's a plural/verb/noun object (has specific keys)
			if isPluralOrGrammarObject(v) {
				out[fullKey] = true
			} else {
				extractKeys(fullKey, v, out)
			}
		}
	}
}

// isPluralOrGrammarObject checks if a map is a leaf object (plural forms, verb forms, etc).
func isPluralOrGrammarObject(m map[string]any) bool {
	// CLDR plural keys
	_, hasOne := m["one"]
	_, hasOther := m["other"]
	_, hasZero := m["zero"]
	_, hasTwo := m["two"]
	_, hasFew := m["few"]
	_, hasMany := m["many"]

	// Grammar keys
	_, hasPast := m["past"]
	_, hasGerund := m["gerund"]
	_, hasGender := m["gender"]
	_, hasBase := m["base"]

	// Article keys
	_, hasDefault := m["default"]
	_, hasVowel := m["vowel"]

	if hasOne || hasOther || hasZero || hasTwo || hasFew || hasMany {
		return true
	}
	if hasPast || hasGerund || hasGender || hasBase {
		return true
	}
	if hasDefault || hasVowel {
		return true
	}

	return false
}

// loadValidIntents returns the set of valid intent keys.
func loadValidIntents() map[string]bool {
	// Core intents - these match what's defined in intents.go
	return map[string]bool{
		// Destructive
		"core.delete":    true,
		"core.remove":    true,
		"core.discard":   true,
		"core.reset":     true,
		"core.overwrite": true,
		// Creation
		"core.create": true,
		"core.add":    true,
		"core.clone":  true,
		"core.copy":   true,
		// Modification
		"core.save":   true,
		"core.update": true,
		"core.rename": true,
		"core.move":   true,
		// Git
		"core.commit": true,
		"core.push":   true,
		"core.pull":   true,
		"core.merge":  true,
		"core.rebase": true,
		// Network
		"core.install":  true,
		"core.download": true,
		"core.upload":   true,
		"core.publish":  true,
		"core.deploy":   true,
		// Process
		"core.start":   true,
		"core.stop":    true,
		"core.restart": true,
		"core.run":     true,
		"core.build":   true,
		"core.test":    true,
		// Information
		"core.continue": true,
		"core.proceed":  true,
		"core.confirm":  true,
		// Additional
		"core.sync":     true,
		"core.boot":     true,
		"core.format":   true,
		"core.analyse":  true,
		"core.link":     true,
		"core.unlink":   true,
		"core.fetch":    true,
		"core.generate": true,
		"core.validate": true,
		"core.check":    true,
		"core.scan":     true,
	}
}

// scanPackages scans Go packages for i18n key usage.
func scanPackages(patterns []string) ([]KeyUsage, error) {
	var usages []KeyUsage

	for _, pattern := range patterns {
		// Expand pattern
		matches, err := expandPattern(pattern)
		if err != nil {
			return nil, fmt.Errorf("expanding pattern %q: %w", pattern, err)
		}

		for _, dir := range matches {
			dirUsages, err := scanDirectory(dir)
			if err != nil {
				return nil, fmt.Errorf("scanning %s: %w", dir, err)
			}
			usages = append(usages, dirUsages...)
		}
	}

	return usages, nil
}

// expandPattern expands a Go package pattern to directories.
func expandPattern(pattern string) ([]string, error) {
	// Handle ./... or ... pattern
	if strings.HasSuffix(pattern, "...") {
		base := strings.TrimSuffix(pattern, "...")
		base = strings.TrimSuffix(base, "/")
		if base == "" || base == "." {
			base = "."
		}
		return findAllGoDirs(base)
	}

	// Single directory
	return []string{pattern}, nil
}

// findAllGoDirs finds all directories containing .go files.
func findAllGoDirs(root string) ([]string, error) {
	var dirs []string
	seen := make(map[string]bool)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even on error
		}

		if info == nil {
			return nil
		}

		// Skip vendor, testdata, and hidden directories (but not . itself)
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == "testdata" || (strings.HasPrefix(name, ".") && name != ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for .go files
		if strings.HasSuffix(path, ".go") {
			dir := filepath.Dir(path)
			if !seen[dir] {
				seen[dir] = true
				dirs = append(dirs, dir)
			}
		}

		return nil
	})

	return dirs, err
}

// scanDirectory scans a directory for i18n key usage.
func scanDirectory(dir string) ([]KeyUsage, error) {
	var usages []KeyUsage

	fset := token.NewFileSet()
	// Parse all .go files except those ending exactly in _test.go
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		name := fi.Name()
		// Only exclude files that are actual test files (ending in _test.go)
		// Files like "go_test_cmd.go" should be included
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			fileUsages := scanFile(fset, filename, file)
			usages = append(usages, fileUsages...)
		}
	}

	return usages, nil
}

// scanFile scans a single file for i18n key usage.
func scanFile(fset *token.FileSet, filename string, file *ast.File) []KeyUsage {
	var usages []KeyUsage

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		funcName := getFuncName(call)
		if funcName == "" {
			return true
		}

		// Check for T(), C(), i18n.T(), i18n.C()
		if funcName == "T" || funcName == "i18n.T" || funcName == "_" || funcName == "i18n._" {
			if key := extractStringArg(call, 0); key != "" {
				pos := fset.Position(call.Pos())
				usages = append(usages, KeyUsage{
					Key:      key,
					File:     filename,
					Line:     pos.Line,
					Function: "T",
				})
			}
		} else if funcName == "C" || funcName == "i18n.C" {
			if key := extractStringArg(call, 0); key != "" {
				pos := fset.Position(call.Pos())
				usages = append(usages, KeyUsage{
					Key:      key,
					File:     filename,
					Line:     pos.Line,
					Function: "C",
				})
			}
		} else if funcName == "I" || funcName == "i18n.I" {
			if key := extractStringArg(call, 0); key != "" {
				pos := fset.Position(call.Pos())
				usages = append(usages, KeyUsage{
					Key:      key,
					File:     filename,
					Line:     pos.Line,
					Function: "C", // I() is an intent builder
				})
			}
		}

		return true
	})

	return usages
}

// getFuncName extracts the function name from a call expression.
func getFuncName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		if ident, ok := fn.X.(*ast.Ident); ok {
			return ident.Name + "." + fn.Sel.Name
		}
	}
	return ""
}

// extractStringArg extracts a string literal from a call argument.
func extractStringArg(call *ast.CallExpr, index int) string {
	if index >= len(call.Args) {
		return ""
	}

	arg := call.Args[index]

	// Direct string literal
	if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		// Remove quotes
		s := lit.Value
		if len(s) >= 2 {
			return s[1 : len(s)-1]
		}
	}

	// Identifier (constant reference) - we skip these as they're type-safe
	if _, ok := arg.(*ast.Ident); ok {
		return "" // Skip constants like IntentCoreDelete
	}

	// Selector (like i18n.IntentCoreDelete) - skip these too
	if _, ok := arg.(*ast.SelectorExpr); ok {
		return ""
	}

	return ""
}

// validate validates key usages against valid keys and intents.
func validate(usages []KeyUsage, validKeys, validIntents map[string]bool) ValidationResult {
	result := ValidationResult{
		TotalKeys: len(usages),
	}

	for _, usage := range usages {
		if usage.Function == "C" {
			result.IntentKeys++
			// Check intent keys
			if validIntents[usage.Key] {
				result.ValidKeys++
			} else {
				// Also allow custom intents (non-core.* prefix)
				if !strings.HasPrefix(usage.Key, "core.") {
					result.ValidKeys++ // Assume custom intents are valid
				} else {
					result.MissingKeys = append(result.MissingKeys, usage)
				}
			}
		} else {
			result.MessageKeys++
			// Check message keys
			if validKeys[usage.Key] {
				result.ValidKeys++
			} else if strings.HasPrefix(usage.Key, "core.") {
				// core.* keys used with T() are intent keys
				if validIntents[usage.Key] {
					result.ValidKeys++
				} else {
					result.MissingKeys = append(result.MissingKeys, usage)
				}
			} else {
				result.MissingKeys = append(result.MissingKeys, usage)
			}
		}
	}

	return result
}

// printReport prints the validation report.
func printReport(result ValidationResult) {
	fmt.Printf("i18n Validation Report\n")
	fmt.Printf("======================\n\n")
	fmt.Printf("Total keys scanned:  %d\n", result.TotalKeys)
	fmt.Printf("  Message keys (T):  %d\n", result.MessageKeys)
	fmt.Printf("  Intent keys (C):   %d\n", result.IntentKeys)
	fmt.Printf("Valid keys:          %d\n", result.ValidKeys)
	fmt.Printf("Missing keys:        %d\n", len(result.MissingKeys))

	if len(result.MissingKeys) > 0 {
		fmt.Printf("\nMissing Keys:\n")
		fmt.Printf("-------------\n")

		// Sort by file then line
		sort.Slice(result.MissingKeys, func(i, j int) bool {
			if result.MissingKeys[i].File != result.MissingKeys[j].File {
				return result.MissingKeys[i].File < result.MissingKeys[j].File
			}
			return result.MissingKeys[i].Line < result.MissingKeys[j].Line
		})

		for _, usage := range result.MissingKeys {
			fmt.Printf("  %s:%d: %s(%q)\n", usage.File, usage.Line, usage.Function, usage.Key)
		}

		fmt.Printf("\nAdd these keys to pkg/i18n/locales/en_GB.json or use constants from pkg/i18n/keys.go\n")
	} else {
		fmt.Printf("\nAll keys are valid!\n")
	}
}
