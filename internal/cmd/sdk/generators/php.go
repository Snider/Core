package generators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	coreio "github.com/host-uk/core/pkg/io"
)

// PHPGenerator generates PHP SDKs from OpenAPI specs.
type PHPGenerator struct{}

// NewPHPGenerator creates a new PHP generator.
func NewPHPGenerator() *PHPGenerator {
	return &PHPGenerator{}
}

// Language returns the generator's target language identifier.
func (g *PHPGenerator) Language() string {
	return "php"
}

// Available checks if generator dependencies are installed.
func (g *PHPGenerator) Available() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// Install returns instructions for installing the generator.
func (g *PHPGenerator) Install() string {
	return "Docker is required for PHP SDK generation"
}

// Generate creates SDK from OpenAPI spec.
func (g *PHPGenerator) Generate(ctx context.Context, opts Options) error {
	if !g.Available() {
		return fmt.Errorf("php.Generate: Docker is required but not available")
	}

	if err := coreio.Local.EnsureDir(opts.OutputDir); err != nil {
		return fmt.Errorf("php.Generate: failed to create output dir: %w", err)
	}

	specDir := filepath.Dir(opts.SpecPath)
	specName := filepath.Base(opts.SpecPath)

	args := []string{"run", "--rm"}
	args = append(args, dockerUserArgs()...)
	args = append(args,
		"-v", specDir+":/spec",
		"-v", opts.OutputDir+":/out",
		"openapitools/openapi-generator-cli", "generate",
		"-i", "/spec/"+specName,
		"-g", "php",
		"-o", "/out",
		"--additional-properties=invokerPackage="+opts.PackageName,
	)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("php.Generate: %w", err)
	}
	return nil
}
