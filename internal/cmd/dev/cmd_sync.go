package dev

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"text/template"

	"github.com/host-uk/core/pkg/cli"  // Added
	"github.com/host-uk/core/pkg/i18n" // Added
	coreio "github.com/host-uk/core/pkg/io"
	// Added
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// syncInternalToPublic handles the synchronization of internal packages to public-facing directories.
// This function is a placeholder for future implementation.
func syncInternalToPublic(ctx context.Context, publicDir string) error {
	// 1. Clean public/internal
	// 2. Copy relevant files from internal/ to public/internal/
	//    Usually just shared logic, not private stuff.

	// For now, let's assume we copy specific safe packages
	// Logic to be refined.

	// Example migration of os calls:
	// internalDirs, err := os.ReadDir(pkgDir) -> coreio.Local.List(pkgDir)
	// os.Stat -> coreio.Local.IsFile (returns bool) or List for existence check
	// os.MkdirAll -> coreio.Local.EnsureDir
	// os.WriteFile -> coreio.Local.Write

	return nil
}

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
	internalDirs, err := coreio.Local.List(pkgDir)
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

		if !coreio.Local.IsFile(internalFile) {
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
	// ParseFile expects a filename/path and reads it using os.Open by default if content is nil.
	// Since we want to use our Medium abstraction, we should read the file content first.
	content, err := coreio.Local.Read(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	// ParseFile can take content as string (src argument).
	node, err := parser.ParseFile(fset, path, content, parser.ParseComments)
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
	if err := coreio.Local.EnsureDir(dir); err != nil {
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

	return coreio.Local.Write(path, buf.String())
}
