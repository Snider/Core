package dev

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/leaanthony/clir"
)

// AddDocCommand adds the 'doc' command to the provided clir instance.
func AddDocCommand(parent *clir.Command) {
	docCmd := parent.NewSubCommand("doc", "Generate missing docstrings for the public API.")
	docCmd.LongDescription(`This command scans the public API of your project and generates placeholder docstrings for any exported functions, types, or methods that are missing them.

It will also report any values that cannot be determined through reflection, providing a list of file names and line numbers that require manual attention.

If a comment already exists, it will be preserved.`)

	var basePath string
	docCmd.StringFlag("path", "Base path to scan for Go files (defaults to current directory)", &basePath)

	docCmd.Action(func() error {
		if basePath == "" {
			basePath = "."
		}

		reports := []string{}
		fset := token.NewFileSet()

		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && (info.Name() == "vendor" || strings.HasPrefix(info.Name(), ".")) {
				return filepath.SkipDir
			}
			if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
				fileContent, readErr := os.ReadFile(path)
				if readErr != nil {
					return fmt.Errorf("failed to read file %s: %w", path, readErr)
				}

				node, parseErr := parser.ParseFile(fset, path, fileContent, parser.ParseComments)
				if parseErr != nil {
					return fmt.Errorf("failed to parse file %s: %w", path, parseErr)
				}

				for _, decl := range node.Decls {
					switch d := decl.(type) {
					case *ast.GenDecl:
						if d.Doc == nil || len(d.Doc.List) == 0 {
							for _, spec := range d.Specs {
								switch s := spec.(type) {
								case *ast.TypeSpec:
									if ast.IsExported(s.Name.Name) {
										reports = append(reports, fmt.Sprintf("%s:%d: Missing docstring for type %s. Suggested: // %s ...", path, fset.Position(s.Pos()).Line, s.Name.Name, s.Name.Name))
									}
								case *ast.ValueSpec: // For var and const
									for _, name := range s.Names {
										if ast.IsExported(name.Name) {
											reports = append(reports, fmt.Sprintf("%s:%d: Missing docstring for %s %s. Suggested: // %s ...", path, fset.Position(name.Pos()).Line, d.Tok.String(), name.Name, name.Name))
										}
									}
								}
							}
						}
					case *ast.FuncDecl:
						if d.Doc == nil || len(d.Doc.List) == 0 {
							if d.Name != nil && ast.IsExported(d.Name.Name) {
								funcName := d.Name.Name
								if d.Recv != nil && len(d.Recv.List) > 0 {
									// It's a method
									recvType := ""
									switch rt := d.Recv.List[0].Type.(type) {
									case *ast.Ident:
										recvType = rt.Name
									case *ast.StarExpr:
										if id, ok := rt.X.(*ast.Ident); ok {
											recvType = "*" + id.Name
										}
									}
									reports = append(reports, fmt.Sprintf("%s:%d: Missing docstring for method (%s) %s. Suggested: // (%s) %s ...", path, fset.Position(d.Pos()).Line, recvType, funcName, recvType, funcName))
								} else {
									// It's a function
									reports = append(reports, fmt.Sprintf("%s:%d: Missing docstring for function %s. Suggested: // %s ...", path, fset.Position(d.Pos()).Line, funcName, funcName))
								}
							}
						}
					}
				}
			}
			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking through files: %v\n", err)
			return err
		}

		if len(reports) == 0 {
			fmt.Println("No missing docstrings found for exported declarations.")
		} else {
			fmt.Println("Missing docstrings found:")
			for _, r := range reports {
				fmt.Println(r)
			}
		}
		return nil
	})
}
