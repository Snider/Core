package generators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	coreio "github.com/host-uk/core/pkg/io"
)

// PythonGenerator generates Python SDKs from OpenAPI specs.
type PythonGenerator struct{}

// NewPythonGenerator creates a new Python generator.
func NewPythonGenerator() *PythonGenerator {
	return &PythonGenerator{}
}

// Language returns the generator's target language identifier.
func (g *PythonGenerator) Language() string {
	return "python"
}

// Available checks if generator dependencies are installed.
func (g *PythonGenerator) Available() bool {
	_, err := exec.LookPath("openapi-python-client")
	return err == nil
}

// Install returns instructions for installing the generator.
func (g *PythonGenerator) Install() string {
	return "pip install openapi-python-client"
}

// Generate creates SDK from OpenAPI spec.
func (g *PythonGenerator) Generate(ctx context.Context, opts Options) error {
	if err := coreio.Local.EnsureDir(opts.OutputDir); err != nil {
		return fmt.Errorf("python.Generate: failed to create output dir: %w", err)
	}

	if g.Available() {
		return g.generateNative(ctx, opts)
	}
	return g.generateDocker(ctx, opts)
}

func (g *PythonGenerator) generateNative(ctx context.Context, opts Options) error {
	parentDir := filepath.Dir(opts.OutputDir)

	cmd := exec.CommandContext(ctx, "openapi-python-client", "generate",
		"--path", opts.SpecPath,
		"--output-path", opts.OutputDir,
	)
	cmd.Dir = parentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *PythonGenerator) generateDocker(ctx context.Context, opts Options) error {
	specDir := filepath.Dir(opts.SpecPath)
	specName := filepath.Base(opts.SpecPath)

	args := []string{"run", "--rm"}
	args = append(args, dockerUserArgs()...)
	args = append(args,
		"-v", specDir+":/spec",
		"-v", opts.OutputDir+":/out",
		"openapitools/openapi-generator-cli", "generate",
		"-i", "/spec/"+specName,
		"-g", "python",
		"-o", "/out",
		"--additional-properties=packageName="+opts.PackageName,
	)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
