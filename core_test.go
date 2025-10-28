package core_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPublicAPICompleteness dynamically discovers all public services and ensures
// their top-level API packages are in sync with their internal implementations.
func TestPublicAPICompleteness(t *testing.T) {
	pkgDir := "pkg"

	// 1. Discover all potential service packages in the pkg/ directory.
	internalDirs, err := os.ReadDir(pkgDir)
	if err != nil {
		t.Fatalf("Failed to read pkg directory: %v", err)
	}

	var allMissingSymbols []string

	for _, dir := range internalDirs {
		if !dir.IsDir() || dir.Name() == "core" {
			continue // Skip files and the core package itself
		}

		serviceName := dir.Name()
		topLevelDir := serviceName

		// 2. Check if a corresponding top-level public API directory exists.
		if _, err := os.Stat(topLevelDir); os.IsNotExist(err) {
			continue // Not a public service, so we skip it.
		}

		// 3. Define paths for public and internal Go files.
		publicFile := filepath.Join(topLevelDir, serviceName+".go")
		internalFile := filepath.Join(pkgDir, serviceName, serviceName+".go")

		// Ensure both files exist before trying to parse them.
		if _, err := os.Stat(publicFile); os.IsNotExist(err) {
			t.Logf("Skipping service '%s': public API file not found at %s", serviceName, publicFile)
			continue
		}
		if _, err := os.Stat(internalFile); os.IsNotExist(err) {
			t.Logf("Skipping service '%s': internal implementation file not found at %s", serviceName, internalFile)
			continue
		}

		// 4. Compare the exported symbols.
		missing, err := compareExports(publicFile, internalFile)
		if err != nil {
			t.Errorf("Error comparing exports for service '%s': %v", serviceName, err)
			continue
		}

		if len(missing) > 0 {
			msg := "- Service: " + serviceName + "\n  - Missing: " + strings.Join(missing, ", ")
			allMissingSymbols = append(allMissingSymbols, msg)
		}
	}

	// 5. Report all discrepancies at the end.
	if len(allMissingSymbols) > 0 {
		t.Errorf("Public APIs are out of sync with internal implementations:\n\n%s",
			strings.Join(allMissingSymbols, "\n"))
	}
}

// compareExports takes two file paths, parses them, and returns a list of
// symbols that are exported in the internal file but not the public one.
func compareExports(publicFile, internalFile string) ([]string, error) {
	publicAPI, err := getExportedSymbols(publicFile)
	if err != nil {
		return nil, err
	}

	internalImpl, err := getExportedSymbols(internalFile)
	if err != nil {
		return nil, err
	}

	publicSymbols := make(map[string]bool)
	for _, sym := range publicAPI {
		publicSymbols[sym] = true
	}

	var missingSymbols []string
	for _, internalSym := range internalImpl {
		// The public API re-exports the interface from core, so we don't expect it here.
		if internalSym == "Config" {
			continue
		}
		if !publicSymbols[internalSym] {
			missingSymbols = append(missingSymbols, internalSym)
		}
	}

	return missingSymbols, nil
}

// getExportedSymbols parses a Go file and returns a slice of its exported symbol names.
func getExportedSymbols(path string) ([]string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, absPath, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var symbols []string
	for name, obj := range node.Scope.Objects {
		if token.IsExported(name) {
			symbols = append(symbols, obj.Name)
		}
	}

	return symbols, nil
}
