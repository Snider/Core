package generators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("php.Generate: failed to create output dir: %w", err)
	}

	specDir := filepath.Dir(opts.SpecPath)
	specName := filepath.Base(opts.SpecPath)

	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"-v", specDir+":/spec",
		"-v", opts.OutputDir+":/out",
		"openapitools/openapi-generator-cli", "generate",
		"-i", "/spec/"+specName,
		"-g", "php",
		"-o", "/out",
		"--additional-properties=invokerPackage="+opts.PackageName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("php.Generate: %w", err)
	}
	return nil
}
