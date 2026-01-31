package dev

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"text/template"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// addSyncCommand adds the 'sync' command to the given parent command.
func addSyncCommand(parent *cli.Command) {
	syncCmd := &cli.Command{
		Use:   "sync",
		Short: i18n.T("cmd.dev.sync.short"),
		Long:  i18n.T("cmd.dev.sync.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			if err := runSync(); err != nil {
				return cli.Wrap(err, i18n.Label("error"))
			}
			cli.Text(i18n.T("i18n.done.sync", "public APIs"))
			return nil
		},
	}

	parent.AddCommand(syncCmd)
}

type symbolInfo struct {
	Name string
	Kind string // "var", "func", "type", "const"
}

func runSync() error {
	pkgDir := "pkg"
	internalDirs, err := os.ReadDir(pkgDir)
	if err != nil {
		return cli.Wrap(err, "failed to read pkg directory")
	}

	for _, dir := range internalDirs {
		if !dir.IsDir() || dir.Name() == "core" {
			continue
		}

		serviceName := dir.Name()
		internalFile := filepath.Join(pkgDir, serviceName, serviceName+".go")
		publicDir := serviceName
		publicFile := filepath.Join(publicDir, serviceName+".go")

		if _, err := os.Stat(internalFile); os.IsNotExist(err) {
			continue
		}

		symbols, err := getExportedSymbols(internalFile)
		if err != nil {
			return cli.Wrap(err, cli.Sprintf("error getting symbols for service '%s'", serviceName))
		}

		if err := generatePublicAPIFile(publicDir, publicFile, serviceName, symbols); err != nil {
			return cli.Wrap(err, cli.Sprintf("error generating public API file for service '%s'", serviceName))
		}
	}

	return nil
}

func getExportedSymbols(path string) ([]symbolInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var symbols []symbolInfo
	for name, obj := range node.Scope.Objects {
		if ast.IsExported(name) {
			kind := "unknown"
			switch obj.Kind {
			case ast.Con:
				kind = "const"
			case ast.Var:
				kind = "var"
			case ast.Fun:
				kind = "func"
			case ast.Typ:
				kind = "type"
			}
			if kind != "unknown" {
				symbols = append(symbols, symbolInfo{Name: name, Kind: kind})
			}
		}
	}
	return symbols, nil
}

const publicAPITemplate = `// package {{.ServiceName}} provides the public API for the {{.ServiceName}} service.
package {{.ServiceName}}

import (
	// Import the internal implementation with an alias.
	impl "github.com/host-uk/core/{{.ServiceName}}"

	// Import the core contracts to re-export the interface.
	"github.com/host-uk/core/core"
)

{{range .Symbols}}
{{- if eq .Kind "type"}}
// {{.Name}} is the public type for the {{.Name}} service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type {{.Name}} = impl.{{.Name}}
{{else if eq .Kind "const"}}
// {{.Name}} is a public constant that points to the real constant in the implementation package.
const {{.Name}} = impl.{{.Name}}
{{else if eq .Kind "var"}}
// {{.Name}} is a public variable that points to the real variable in the implementation package.
var {{.Name}} = impl.{{.Name}}
{{else if eq .Kind "func"}}
// {{.Name}} is a public function that points to the real function in the implementation package.
var {{.Name}} = impl.{{.Name}}
{{end}}
{{end}}

// {{.InterfaceName}} is the public interface for the {{.ServiceName}} service.
type {{.InterfaceName}} = core.{{.InterfaceName}}
`

func generatePublicAPIFile(dir, path, serviceName string, symbols []symbolInfo) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	tmpl, err := template.New("publicAPI").Parse(publicAPITemplate)
	if err != nil {
		return err
	}

	tcaser := cases.Title(language.English)
	interfaceName := tcaser.String(serviceName)

	data := struct {
		ServiceName   string
		Symbols       []symbolInfo
		InterfaceName string
	}{
		ServiceName:   serviceName,
		Symbols:       symbols,
		InterfaceName: interfaceName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}
