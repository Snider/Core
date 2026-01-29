package generators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GoGenerator generates Go SDKs from OpenAPI specs.
type GoGenerator struct{}

// NewGoGenerator creates a new Go generator.
func NewGoGenerator() *GoGenerator {
	return &GoGenerator{}
}

// Language returns the generator's target language identifier.
func (g *GoGenerator) Language() string {
	return "go"
}

// Available checks if generator dependencies are installed.
func (g *GoGenerator) Available() bool {
	_, err := exec.LookPath("oapi-codegen")
	return err == nil
}

// Install returns instructions for installing the generator.
func (g *GoGenerator) Install() string {
	return "go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest"
}

// Generate creates SDK from OpenAPI spec.
func (g *GoGenerator) Generate(ctx context.Context, opts Options) error {
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("go.Generate: failed to create output dir: %w", err)
	}

	if g.Available() {
		return g.generateNative(ctx, opts)
	}
	return g.generateDocker(ctx, opts)
}

func (g *GoGenerator) generateNative(ctx context.Context, opts Options) error {
	outputFile := filepath.Join(opts.OutputDir, "client.go")

	cmd := exec.CommandContext(ctx, "oapi-codegen",
		"-package", opts.PackageName,
		"-generate", "types,client",
		"-o", outputFile,
		opts.SpecPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go.generateNative: %w", err)
	}

	goMod := fmt.Sprintf("module %s\n\ngo 1.21\n", opts.PackageName)
	return os.WriteFile(filepath.Join(opts.OutputDir, "go.mod"), []byte(goMod), 0644)
}

func (g *GoGenerator) generateDocker(ctx context.Context, opts Options) error {
	specDir := filepath.Dir(opts.SpecPath)
	specName := filepath.Base(opts.SpecPath)

	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"-v", specDir+":/spec",
		"-v", opts.OutputDir+":/out",
		"openapitools/openapi-generator-cli", "generate",
		"-i", "/spec/"+specName,
		"-g", "go",
		"-o", "/out",
		"--additional-properties=packageName="+opts.PackageName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
