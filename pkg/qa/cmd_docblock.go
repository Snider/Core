// cmd_docblock.go implements docblock/docstring coverage checking for Go code.
//
// Usage:
//
//	core qa docblock              # Check current directory
//	core qa docblock ./pkg/...    # Check specific packages
//	core qa docblock --threshold=80  # Require 80% coverage
package qa

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

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

// Docblock command flags
var (
	docblockThreshold float64
	docblockVerbose   bool
	docblockJSON      bool
)

// addDocblockCommand adds the 'docblock' command to qa.
func addDocblockCommand(parent *cli.Command) {
	docblockCmd := &cli.Command{
		Use:   "docblock [packages...]",
		Short: i18n.T("cmd.qa.docblock.short"),
		Long:  i18n.T("cmd.qa.docblock.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			paths := args
			if len(paths) == 0 {
				paths = []string{"./..."}
			}
			return RunDocblockCheck(paths, docblockThreshold, docblockVerbose, docblockJSON)
		},
	}

	docblockCmd.Flags().Float64Var(&docblockThreshold, "threshold", 80, i18n.T("cmd.qa.docblock.flag.threshold"))
	docblockCmd.Flags().BoolVarP(&docblockVerbose, "verbose", "v", false, i18n.T("common.flag.verbose"))
	docblockCmd.Flags().BoolVar(&docblockJSON, "json", false, i18n.T("common.flag.json"))

	parent.AddCommand(docblockCmd)
}

// DocblockResult holds the result of a docblock coverage check.
type DocblockResult struct {
	Coverage   float64           `json:"coverage"`
	Threshold  float64           `json:"threshold"`
	Total      int               `json:"total"`
	Documented int               `json:"documented"`
	Missing    []MissingDocblock `json:"missing,omitempty"`
	Passed     bool              `json:"passed"`
}

// MissingDocblock represents an exported symbol without documentation.
type MissingDocblock struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Name   string `json:"name"`
	Kind   string `json:"kind"` // func, type, const, var
	Reason string `json:"reason,omitempty"`
}

// RunDocblockCheck checks docblock coverage for the given packages.
func RunDocblockCheck(paths []string, threshold float64, verbose, jsonOutput bool) error {
	result, err := CheckDocblockCoverage(paths)
	if err != nil {
		return err
	}
	result.Threshold = threshold
	result.Passed = result.Coverage >= threshold

	if jsonOutput {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		if !result.Passed {
			return cli.Err("docblock coverage %.1f%% below threshold %.1f%%", result.Coverage, threshold)
		}
		return nil
	}

	// Sort missing by file then line
	sort.Slice(result.Missing, func(i, j int) bool {
		if result.Missing[i].File != result.Missing[j].File {
			return result.Missing[i].File < result.Missing[j].File
		}
		return result.Missing[i].Line < result.Missing[j].Line
	})

	// Print result
	if verbose && len(result.Missing) > 0 {
		cli.Print("%s\n\n", i18n.T("cmd.qa.docblock.missing_docs"))
		for _, m := range result.Missing {
			cli.Print("  %s:%d: %s %s\n",
				dimStyle.Render(m.File),
				m.Line,
				dimStyle.Render(m.Kind),
				m.Name,
			)
		}
		cli.Blank()
	}

	// Summary
	coverageStr := fmt.Sprintf("%.1f%%", result.Coverage)
	thresholdStr := fmt.Sprintf("%.1f%%", threshold)

	if result.Passed {
		cli.Print("%s %s %s/%s (%s >= %s)\n",
			successStyle.Render(i18n.T("common.label.success")),
			i18n.T("cmd.qa.docblock.coverage"),
			fmt.Sprintf("%d", result.Documented),
			fmt.Sprintf("%d", result.Total),
			successStyle.Render(coverageStr),
			thresholdStr,
		)
		return nil
	}

	cli.Print("%s %s %s/%s (%s < %s)\n",
		errorStyle.Render(i18n.T("common.label.error")),
		i18n.T("cmd.qa.docblock.coverage"),
		fmt.Sprintf("%d", result.Documented),
		fmt.Sprintf("%d", result.Total),
		errorStyle.Render(coverageStr),
		thresholdStr,
	)

	// Always show compact file:line list when failing (token-efficient for AI agents)
	if len(result.Missing) > 0 {
		cli.Blank()
		for _, m := range result.Missing {
			cli.Print("%s:%d\n", m.File, m.Line)
		}
	}

	return cli.Err("docblock coverage %.1f%% below threshold %.1f%%", result.Coverage, threshold)
}

// CheckDocblockCoverage analyzes Go packages for docblock coverage.
func CheckDocblockCoverage(patterns []string) (*DocblockResult, error) {
	result := &DocblockResult{}

	// Expand patterns to actual directories
	dirs, err := expandPatterns(patterns)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()

	for _, dir := range dirs {
		pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go")
		}, parser.ParseComments)
		if err != nil {
			// Log parse errors but continue to check other directories
			fmt.Fprintf(os.Stderr, "warning: failed to parse %s: %v\n", dir, err)
			continue
		}

		for _, pkg := range pkgs {
			for filename, file := range pkg.Files {
				checkFile(fset, filename, file, result)
			}
		}
	}

	if result.Total > 0 {
		result.Coverage = float64(result.Documented) / float64(result.Total) * 100
	}

	return result, nil
}

// expandPatterns expands Go package patterns like ./... to actual directories.
func expandPatterns(patterns []string) ([]string, error) {
	var dirs []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		if strings.HasSuffix(pattern, "/...") {
			// Recursive pattern
			base := strings.TrimSuffix(pattern, "/...")
			if base == "." {
				base = "."
			}
			err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Skip errors
				}
				if !info.IsDir() {
					return nil
				}
				// Skip vendor, testdata, and hidden directories (but not "." itself)
				name := info.Name()
				if name == "vendor" || name == "testdata" || (strings.HasPrefix(name, ".") && name != ".") {
					return filepath.SkipDir
				}
				// Check if directory has Go files
				if hasGoFiles(path) && !seen[path] {
					dirs = append(dirs, path)
					seen[path] = true
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			// Single directory
			path := pattern
			if !seen[path] && hasGoFiles(path) {
				dirs = append(dirs, path)
				seen[path] = true
			}
		}
	}

	return dirs, nil
}

// hasGoFiles checks if a directory contains Go files.
func hasGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			return true
		}
	}
	return false
}

// checkFile analyzes a single file for docblock coverage.
func checkFile(fset *token.FileSet, filename string, file *ast.File, result *DocblockResult) {
	// Make filename relative if possible
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, filename); err == nil {
			filename = rel
		}
	}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Skip unexported functions
			if !ast.IsExported(d.Name.Name) {
				continue
			}
			// Skip methods on unexported types
			if d.Recv != nil && len(d.Recv.List) > 0 {
				if recvType := getReceiverTypeName(d.Recv.List[0].Type); recvType != "" && !ast.IsExported(recvType) {
					continue
				}
			}

			result.Total++
			if d.Doc != nil && len(d.Doc.List) > 0 {
				result.Documented++
			} else {
				pos := fset.Position(d.Pos())
				result.Missing = append(result.Missing, MissingDocblock{
					File: filename,
					Line: pos.Line,
					Name: d.Name.Name,
					Kind: "func",
				})
			}

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if !ast.IsExported(s.Name.Name) {
						continue
					}
					result.Total++
					// Type can have doc on GenDecl or TypeSpec
					if (d.Doc != nil && len(d.Doc.List) > 0) || (s.Doc != nil && len(s.Doc.List) > 0) {
						result.Documented++
					} else {
						pos := fset.Position(s.Pos())
						result.Missing = append(result.Missing, MissingDocblock{
							File: filename,
							Line: pos.Line,
							Name: s.Name.Name,
							Kind: "type",
						})
					}

				case *ast.ValueSpec:
					// Check exported consts and vars
					for _, name := range s.Names {
						if !ast.IsExported(name.Name) {
							continue
						}
						result.Total++
						// Value can have doc on GenDecl or ValueSpec
						if (d.Doc != nil && len(d.Doc.List) > 0) || (s.Doc != nil && len(s.Doc.List) > 0) {
							result.Documented++
						} else {
							pos := fset.Position(name.Pos())
							result.Missing = append(result.Missing, MissingDocblock{
								File: filename,
								Line: pos.Line,
								Name: name.Name,
								Kind: kindFromToken(d.Tok),
							})
						}
					}
				}
			}
		}
	}
}

// getReceiverTypeName extracts the type name from a method receiver.
func getReceiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return getReceiverTypeName(t.X)
	}
	return ""
}

// kindFromToken returns a string representation of the token kind.
func kindFromToken(tok token.Token) string {
	switch tok {
	case token.CONST:
		return "const"
	case token.VAR:
		return "var"
	default:
		return "value"
	}
}
