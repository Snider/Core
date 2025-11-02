package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/leaanthony/clir"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// AddTestGenCommand adds the 'test-gen' command to the given parent command.
func AddTestGenCommand(parent *clir.Command) {
	testGenCmd := parent.NewSubCommand("test-gen", "Generates baseline test files for public service APIs.")
	testGenCmd.LongDescription("This command scans for public services and generates a standard set of API contract tests for each one.")
	testGenCmd.Action(func() error {
		if err := runTestGen(); err != nil {
			return fmt.Errorf("Error during test generation: %w", err)
		}
		fmt.Println("API test files generated successfully.")
		return nil
	})
}

const testFileTemplate = `package {{.ServiceName}}_test

import (
	"testing"

	"github.com/Snider/Core/{{.ServiceName}}"
	"github.com/Snider/Core/pkg/core"
)

// TestNew ensures that the public constructor New is available.
func TestNew(t *testing.T) {
	if {{.ServiceName}}.New == nil {
		t.Fatal("{{.ServiceName}}.New constructor is nil")
	}
	// Note: This is a basic check. Some services may require a core instance
	// or other arguments. This test can be expanded as needed.
}

// TestRegister ensures that the public factory Register is available.
func TestRegister(t *testing.T) {
	if {{.ServiceName}}.Register == nil {
		t.Fatal("{{.ServiceName}}.Register factory is nil")
	}
}

// TestInterfaceCompliance ensures that the public Service type correctly
// implements the public {{.InterfaceName}} interface. This is a compile-time check.
func TestInterfaceCompliance(t *testing.T) {
	// This is a compile-time check. If it compiles, the test passes.
	var _ core.{{.InterfaceName}} = (*{{.ServiceName}}.Service)(nil)
}
`

func runTestGen() error {
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
		publicDir := serviceName

		// Check if a corresponding top-level public API directory exists.
		if _, err := os.Stat(publicDir); os.IsNotExist(err) {
			continue // Not a public service, so we skip it.
		}

		testFilePath := filepath.Join(publicDir, serviceName+"_test.go")
		fmt.Printf("Generating test file for service '%s' at %s\n", serviceName, testFilePath)

		if err := generateTestFile(testFilePath, serviceName); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not generate test for service '%s': %v\n", serviceName, err)
		}
	}

	return nil
}

func generateTestFile(path, serviceName string) error {
	tmpl, err := template.New("test").Parse(testFileTemplate)
	if err != nil {
		return err
	}

	tcaser := cases.Title(language.English)
	interfaceName := tcaser.String(serviceName)

	data := struct {
		ServiceName   string
		InterfaceName string
	}{
		ServiceName:   serviceName,
		InterfaceName: interfaceName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}
