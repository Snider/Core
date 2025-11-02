package cmd

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/leaanthony/clir"
)

func AddDocGenCommand(parent *clir.Command) {
	cmd := parent.NewSubCommand("docgen", "Generates Markdown documentation for the public API of the services.")
	cmd.Action(func() error {
		return runDocGen()
	})
}

func runDocGen() error {
	const pkgDir = "pkg"
	const outDir = "docs/services"

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	dirs, err := os.ReadDir(pkgDir)
	if err != nil {
		return fmt.Errorf("failed to read pkg directory: %w", err)
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		serviceName := dir.Name()
		servicePath := filepath.Join(pkgDir, serviceName)

		if err := generateDocsForService(servicePath, serviceName, outDir); err != nil {
			fmt.Printf("Warning: Could not generate docs for service '%s': %v\n", serviceName, err)
		}
	}

	fmt.Println("Documentation generated successfully in", outDir)
	return nil
}

func generateDocsForService(servicePath, serviceName, outDir string) error {
	fset := token.NewFileSet()
	filter := func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}
	pkgs, err := parser.ParseDir(fset, servicePath, filter, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse directory %s: %w", servicePath, err)
	}

	internalPath := filepath.Join(servicePath, "internal")
	if _, err := os.Stat(internalPath); err == nil {
		pkgs, err = parser.ParseDir(fset, internalPath, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse internal directory %s: %w", internalPath, err)
		}
	}

	var pkg *ast.Package
	for _, p := range pkgs {
		pkg = p
		break
	}
	if pkg == nil {
		return fmt.Errorf("no package found in %s", servicePath)
	}

	docPkg := doc.New(pkg, "./", doc.AllDecls)

	md, err := generateMarkdown(docPkg)
	if err != nil {
		return fmt.Errorf("failed to generate markdown: %w", err)
	}

	outFile := filepath.Join(outDir, serviceName+".md")
	return os.WriteFile(outFile, []byte(md), 0644)
}

const docTemplate = `---
title: {{ .Name }}
---
# Service: ` + "`" + `{{ .Name }}` + "`" + `

{{ .Doc }}

{{if .Consts}}
## Constants
{{range .Consts}}
` + "```go" + `
{{- range .Names }}{{ . }}{{ end }}
` + "```" + `
{{ .Doc }}
{{end}}{{end}}

{{if .Types}}
## Types
{{range .Types}}
### ` + "`" + `type {{ .Name }}` + "`" + `
` + "```go" + `
type {{ .Name }} {{.Decl | formatNode}}
` + "```" + `
{{ .Doc }}

{{if .Methods}}
#### Methods
{{range .Methods}}
- ` + "`" + `{{ .Name }}({{ .Decl.Type.Params | formatParams }}) {{ .Decl.Type.Results | formatParams }}` + "`" + `: {{ .Doc | oneLine }}
{{end}}{{end}}

{{end}}{{end}}

{{if .Funcs}}
## Functions
{{range .Funcs}}
- ` + "`" + `{{ .Name }}({{ .Decl.Type.Params | formatParams }}) {{ .Decl.Type.Results | formatParams }}` + "`" + `: {{ .Doc | oneLine }}
{{end}}{{end}}
`

func generateMarkdown(pkg *doc.Package) (string, error) {
	funcMap := template.FuncMap{
		"oneLine": func(s string) string {
			return strings.TrimSpace(strings.Replace(s, "\n", " ", -1))
		},
		"formatNode": func(decl *ast.GenDecl) string {
			if len(decl.Specs) == 0 {
				return ""
			}
			spec := decl.Specs[0].(*ast.TypeSpec)
			return nodeToString(spec.Type)
		},
		"formatParams": func(fieldList *ast.FieldList) string {
			if fieldList == nil {
				return ""
			}
			var params []string
			for _, p := range fieldList.List {
				var names []string
				for _, name := range p.Names {
					names = append(names, name.Name)
				}
				typeStr := nodeToString(p.Type)
				if len(names) > 0 {
					params = append(params, strings.Join(names, ", ")+" "+typeStr)
				} else {
					params = append(params, typeStr)
				}
			}
			return strings.Join(params, ", ")
		},
	}

	tmpl, err := template.New("doc").Funcs(funcMap).Parse(docTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, pkg); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func nodeToString(node ast.Node) string {
	var buf bytes.Buffer
	err := ast.Fprint(&buf, token.NewFileSet(), node, nil)
	if err != nil {
		return ""
	}
	return buf.String()
}
