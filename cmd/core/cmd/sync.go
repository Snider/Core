package cmd

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"text/template"

	"github.com/leaanthony/clir"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// AddSyncCommand adds the 'sync' command to the given parent command.
func AddSyncCommand(parent *clir.Command) {
	syncCmd := parent.NewSubCommand("sync", "Synchronizes the public service APIs with their internal implementations.")
	syncCmd.LongDescription("This command scans the 'pkg' directory for services and ensures that the\ntop-level public API for each service is in sync with its internal implementation.\nIt automatically generates the necessary Go files with type aliases.")
	syncCmd.Action(func() error {
		if err := runSync(); err != nil {
			return fmt.Errorf("Error: %w", err)
		}
		fmt.Println("Public APIs synchronized successfully.")
		return nil
	})
}

type symbolInfo struct {
	Name string
	Kind string // "var", "func", "type", "const"
}

func runSync() error {
	pkgDir := "pkg"
	internalDirs, err := os.ReadDir(pkgDir)
	if err != nil {
		return fmt.Errorf("failed to read pkg directory: %w", err)
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
			return fmt.Errorf("error getting symbols for service '%s': %w", serviceName, err)
		}

		if err := generatePublicAPIFile(publicDir, publicFile, serviceName, symbols); err != nil {
			return fmt.Errorf("error generating public API file for service '%s': %w", serviceName, err)
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
	impl "github.com/Snider/Core/{{.ServiceName}}"

	// Import the core contracts to re-export the interface.
	"github.com/Snider/Core/core"
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
