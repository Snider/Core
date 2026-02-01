package generators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// TypeScriptGenerator generates TypeScript SDKs from OpenAPI specs.
type TypeScriptGenerator struct{}

// NewTypeScriptGenerator creates a new TypeScript generator.
func NewTypeScriptGenerator() *TypeScriptGenerator {
	return &TypeScriptGenerator{}
}

// Language returns the generator's target language identifier.
func (g *TypeScriptGenerator) Language() string {
	return "typescript"
}

// Available checks if generator dependencies are installed.
func (g *TypeScriptGenerator) Available() bool {
	_, err := exec.LookPath("openapi-typescript-codegen")
	if err == nil {
		return true
	}
	_, err = exec.LookPath("npx")
	return err == nil
}

// Install returns instructions for installing the generator.
func (g *TypeScriptGenerator) Install() string {
	return "npm install -g openapi-typescript-codegen"
}

// Generate creates SDK from OpenAPI spec.
func (g *TypeScriptGenerator) Generate(ctx context.Context, opts Options) error {
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("typescript.Generate: failed to create output dir: %w", err)
	}

	if g.nativeAvailable() {
		return g.generateNative(ctx, opts)
	}
	if g.npxAvailable() {
		return g.generateNpx(ctx, opts)
	}
	return g.generateDocker(ctx, opts)
}

func (g *TypeScriptGenerator) nativeAvailable() bool {
	_, err := exec.LookPath("openapi-typescript-codegen")
	return err == nil
}

func (g *TypeScriptGenerator) npxAvailable() bool {
	_, err := exec.LookPath("npx")
	return err == nil
}

func (g *TypeScriptGenerator) generateNative(ctx context.Context, opts Options) error {
	cmd := exec.CommandContext(ctx, "openapi-typescript-codegen",
		"--input", opts.SpecPath,
		"--output", opts.OutputDir,
		"--name", opts.PackageName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *TypeScriptGenerator) generateNpx(ctx context.Context, opts Options) error {
	cmd := exec.CommandContext(ctx, "npx", "openapi-typescript-codegen",
		"--input", opts.SpecPath,
		"--output", opts.OutputDir,
		"--name", opts.PackageName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *TypeScriptGenerator) generateDocker(ctx context.Context, opts Options) error {
	specDir := filepath.Dir(opts.SpecPath)
	specName := filepath.Base(opts.SpecPath)

	args := []string{"run", "--rm"}
	args = append(args, dockerUserArgs()...)
	args = append(args,
		"-v", specDir+":/spec",
		"-v", opts.OutputDir+":/out",
		"openapitools/openapi-generator-cli", "generate",
		"-i", "/spec/"+specName,
		"-g", "typescript-fetch",
		"-o", "/out",
		"--additional-properties=npmName="+opts.PackageName,
	)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("typescript.generateDocker: %w", err)
	}
	return nil
}
