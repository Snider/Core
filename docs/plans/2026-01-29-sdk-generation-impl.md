# SDK Generation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Generate typed API clients from OpenAPI specs for TypeScript, Python, Go, and PHP with breaking change detection.

**Architecture:** Hybrid generator approach - native tools where available (openapi-typescript-codegen, openapi-python-client, oapi-codegen), Docker fallback for others (openapi-generator). Detection flow: config → common paths → Laravel Scramble. Breaking changes via oasdiff library.

**Tech Stack:** Go, oasdiff, kin-openapi, embedded templates, exec for native generators, Docker for fallback

---

### Task 1: Create SDK Package Structure

**Files:**
- Create: `pkg/sdk/sdk.go`
- Create: `pkg/sdk/go.mod`

**Step 1: Create go.mod for sdk package**

```go
module github.com/host-uk/core/pkg/sdk

go 1.25

require (
	github.com/getkin/kin-openapi v0.128.0
	github.com/tufin/oasdiff v1.10.25
	gopkg.in/yaml.v3 v3.0.1
)
```

**Step 2: Create sdk.go with types and config**

```go
// Package sdk provides OpenAPI SDK generation and diff capabilities.
package sdk

import (
	"context"
	"fmt"
)

// Config holds SDK generation configuration from .core/release.yaml.
type Config struct {
	// Spec is the path to the OpenAPI spec file (auto-detected if empty).
	Spec string `yaml:"spec,omitempty"`
	// Languages to generate SDKs for.
	Languages []string `yaml:"languages,omitempty"`
	// Output directory (default: sdk/).
	Output string `yaml:"output,omitempty"`
	// Package naming configuration.
	Package PackageConfig `yaml:"package,omitempty"`
	// Diff configuration for breaking change detection.
	Diff DiffConfig `yaml:"diff,omitempty"`
	// Publish configuration for monorepo publishing.
	Publish PublishConfig `yaml:"publish,omitempty"`
}

// PackageConfig holds package naming configuration.
type PackageConfig struct {
	// Name is the base package name.
	Name string `yaml:"name,omitempty"`
	// Version is the SDK version (supports templates like {{.Version}}).
	Version string `yaml:"version,omitempty"`
}

// DiffConfig holds breaking change detection configuration.
type DiffConfig struct {
	// Enabled determines whether to run diff checks.
	Enabled bool `yaml:"enabled,omitempty"`
	// FailOnBreaking fails the release if breaking changes are detected.
	FailOnBreaking bool `yaml:"fail_on_breaking,omitempty"`
}

// PublishConfig holds monorepo publishing configuration.
type PublishConfig struct {
	// Repo is the SDK monorepo (e.g., "myorg/sdks").
	Repo string `yaml:"repo,omitempty"`
	// Path is the subdirectory for this SDK (e.g., "packages/myapi").
	Path string `yaml:"path,omitempty"`
}

// SDK orchestrates OpenAPI SDK generation.
type SDK struct {
	config     *Config
	projectDir string
}

// New creates a new SDK instance.
func New(projectDir string, config *Config) *SDK {
	if config == nil {
		config = DefaultConfig()
	}
	return &SDK{
		config:     config,
		projectDir: projectDir,
	}
}

// DefaultConfig returns sensible defaults for SDK configuration.
func DefaultConfig() *Config {
	return &Config{
		Languages: []string{"typescript", "python", "go", "php"},
		Output:    "sdk",
		Diff: DiffConfig{
			Enabled:        true,
			FailOnBreaking: false,
		},
	}
}

// Generate generates SDKs for all configured languages.
func (s *SDK) Generate(ctx context.Context) error {
	return fmt.Errorf("sdk.Generate: not implemented")
}

// GenerateLanguage generates SDK for a specific language.
func (s *SDK) GenerateLanguage(ctx context.Context, lang string) error {
	return fmt.Errorf("sdk.GenerateLanguage: not implemented")
}
```

**Step 3: Add to go.work**

Run: `cd /Users/snider/Code/Core && echo "	./pkg/sdk" >> go.work && go work sync`

**Step 4: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/sdk/...`
Expected: No errors

**Step 5: Commit**

```bash
git add pkg/sdk/
git add go.work go.work.sum
git commit -m "feat(sdk): add SDK package structure with types

Initial pkg/sdk setup with Config types for OpenAPI SDK generation.
Includes language selection, diff config, and publish config.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 2: Implement OpenAPI Spec Detection

**Files:**
- Create: `pkg/sdk/detect.go`
- Create: `pkg/sdk/detect_test.go`

**Step 1: Write the failing test**

```go
package sdk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSpec_Good_ConfigPath(t *testing.T) {
	// Create temp directory with spec at configured path
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "api", "spec.yaml")
	os.MkdirAll(filepath.Dir(specPath), 0755)
	os.WriteFile(specPath, []byte("openapi: 3.0.0"), 0644)

	sdk := New(tmpDir, &Config{Spec: "api/spec.yaml"})
	got, err := sdk.DetectSpec()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != specPath {
		t.Errorf("got %q, want %q", got, specPath)
	}
}

func TestDetectSpec_Good_CommonPath(t *testing.T) {
	// Create temp directory with spec at common path
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.yaml")
	os.WriteFile(specPath, []byte("openapi: 3.0.0"), 0644)

	sdk := New(tmpDir, nil)
	got, err := sdk.DetectSpec()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != specPath {
		t.Errorf("got %q, want %q", got, specPath)
	}
}

func TestDetectSpec_Bad_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	sdk := New(tmpDir, nil)
	_, err := sdk.DetectSpec()
	if err == nil {
		t.Fatal("expected error for missing spec")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/... -run TestDetectSpec -v`
Expected: FAIL (DetectSpec not defined)

**Step 3: Write minimal implementation**

```go
package sdk

import (
	"fmt"
	"os"
	"path/filepath"
)

// commonSpecPaths are checked in order when no spec is configured.
var commonSpecPaths = []string{
	"api/openapi.yaml",
	"api/openapi.json",
	"openapi.yaml",
	"openapi.json",
	"docs/api.yaml",
	"docs/api.json",
	"swagger.yaml",
	"swagger.json",
}

// DetectSpec finds the OpenAPI spec file.
// Priority: config path → common paths → Laravel Scramble.
func (s *SDK) DetectSpec() (string, error) {
	// 1. Check configured path
	if s.config.Spec != "" {
		specPath := filepath.Join(s.projectDir, s.config.Spec)
		if _, err := os.Stat(specPath); err == nil {
			return specPath, nil
		}
		return "", fmt.Errorf("sdk.DetectSpec: configured spec not found: %s", s.config.Spec)
	}

	// 2. Check common paths
	for _, p := range commonSpecPaths {
		specPath := filepath.Join(s.projectDir, p)
		if _, err := os.Stat(specPath); err == nil {
			return specPath, nil
		}
	}

	// 3. Try Laravel Scramble detection
	specPath, err := s.detectScramble()
	if err == nil {
		return specPath, nil
	}

	return "", fmt.Errorf("sdk.DetectSpec: no OpenAPI spec found (checked config, common paths, Scramble)")
}

// detectScramble checks for Laravel Scramble and exports the spec.
func (s *SDK) detectScramble() (string, error) {
	composerPath := filepath.Join(s.projectDir, "composer.json")
	if _, err := os.Stat(composerPath); err != nil {
		return "", fmt.Errorf("no composer.json")
	}

	// Check for scramble in composer.json
	data, err := os.ReadFile(composerPath)
	if err != nil {
		return "", err
	}

	// Simple check for scramble package
	if !containsScramble(data) {
		return "", fmt.Errorf("scramble not found in composer.json")
	}

	// TODO: Run php artisan scramble:export
	return "", fmt.Errorf("scramble export not implemented")
}

// containsScramble checks if composer.json includes scramble.
func containsScramble(data []byte) bool {
	return len(data) > 0 &&
		(contains(data, "dedoc/scramble") || contains(data, "\"scramble\""))
}

// contains is a simple byte slice search.
func contains(data []byte, substr string) bool {
	return len(data) >= len(substr) &&
		string(data) != "" &&
		indexOf(string(data), substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/... -run TestDetectSpec -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/sdk/detect.go pkg/sdk/detect_test.go
git commit -m "feat(sdk): add OpenAPI spec detection

Detects OpenAPI spec via:
1. Configured spec path
2. Common paths (api/openapi.yaml, openapi.yaml, etc.)
3. Laravel Scramble (stub for now)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 3: Define Generator Interface

**Files:**
- Create: `pkg/sdk/generators/generator.go`

**Step 1: Create generator interface**

```go
// Package generators provides SDK code generators for different languages.
package generators

import (
	"context"
)

// Options holds common generation options.
type Options struct {
	// SpecPath is the path to the OpenAPI spec file.
	SpecPath string
	// OutputDir is where to write the generated SDK.
	OutputDir string
	// PackageName is the package/module name.
	PackageName string
	// Version is the SDK version.
	Version string
}

// Generator defines the interface for SDK generators.
type Generator interface {
	// Language returns the generator's target language identifier.
	Language() string

	// Generate creates SDK from OpenAPI spec.
	Generate(ctx context.Context, opts Options) error

	// Available checks if generator dependencies are installed.
	Available() bool

	// Install returns instructions for installing the generator.
	Install() string
}

// Registry holds available generators.
type Registry struct {
	generators map[string]Generator
}

// NewRegistry creates a registry with all available generators.
func NewRegistry() *Registry {
	r := &Registry{
		generators: make(map[string]Generator),
	}
	// Generators will be registered in subsequent tasks
	return r
}

// Get returns a generator by language.
func (r *Registry) Get(lang string) (Generator, bool) {
	g, ok := r.generators[lang]
	return g, ok
}

// Register adds a generator to the registry.
func (r *Registry) Register(g Generator) {
	r.generators[g.Language()] = g
}

// Languages returns all registered language identifiers.
func (r *Registry) Languages() []string {
	langs := make([]string, 0, len(r.generators))
	for lang := range r.generators {
		langs = append(langs, lang)
	}
	return langs
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/sdk/...`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/sdk/generators/generator.go
git commit -m "feat(sdk): add Generator interface and Registry

Defines the common interface for SDK generators with:
- Generate(), Available(), Install() methods
- Registry for managing multiple generators

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 4: Implement TypeScript Generator

**Files:**
- Create: `pkg/sdk/generators/typescript.go`
- Create: `pkg/sdk/generators/typescript_test.go`

**Step 1: Write the failing test**

```go
package generators

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestTypeScriptGenerator_Good_Available(t *testing.T) {
	g := NewTypeScriptGenerator()
	// Just check it doesn't panic
	_ = g.Available()
	_ = g.Language()
	_ = g.Install()
}

func TestTypeScriptGenerator_Good_Generate(t *testing.T) {
	// Skip if no generator available
	g := NewTypeScriptGenerator()
	if !g.Available() && !dockerAvailable() {
		t.Skip("no TypeScript generator available (need openapi-typescript-codegen or Docker)")
	}

	// Create temp spec
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.yaml")
	spec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
`
	os.WriteFile(specPath, []byte(spec), 0644)

	outputDir := filepath.Join(tmpDir, "sdk", "typescript")
	err := g.Generate(context.Background(), Options{
		SpecPath:    specPath,
		OutputDir:   outputDir,
		PackageName: "test-api",
		Version:     "1.0.0",
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check output exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created")
	}
}

func dockerAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/generators/... -run TestTypeScriptGenerator -v`
Expected: FAIL (NewTypeScriptGenerator not defined)

**Step 3: Write implementation**

```go
package generators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// TypeScriptGenerator generates TypeScript SDKs using openapi-typescript-codegen.
type TypeScriptGenerator struct{}

// NewTypeScriptGenerator creates a new TypeScript generator.
func NewTypeScriptGenerator() *TypeScriptGenerator {
	return &TypeScriptGenerator{}
}

// Language returns "typescript".
func (g *TypeScriptGenerator) Language() string {
	return "typescript"
}

// Available checks if openapi-typescript-codegen is installed.
func (g *TypeScriptGenerator) Available() bool {
	_, err := exec.LookPath("openapi-typescript-codegen")
	if err == nil {
		return true
	}
	// Also check npx availability
	_, err = exec.LookPath("npx")
	return err == nil
}

// Install returns installation instructions.
func (g *TypeScriptGenerator) Install() string {
	return "npm install -g openapi-typescript-codegen"
}

// Generate creates TypeScript SDK from OpenAPI spec.
func (g *TypeScriptGenerator) Generate(ctx context.Context, opts Options) error {
	// Ensure output directory exists
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("typescript.Generate: failed to create output dir: %w", err)
	}

	// Try native generator first
	if g.nativeAvailable() {
		return g.generateNative(ctx, opts)
	}

	// Try npx
	if g.npxAvailable() {
		return g.generateNpx(ctx, opts)
	}

	// Fall back to Docker
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
	// Use openapi-generator via Docker
	specDir := filepath.Dir(opts.SpecPath)
	specName := filepath.Base(opts.SpecPath)

	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"-v", specDir+":/spec",
		"-v", opts.OutputDir+":/out",
		"openapitools/openapi-generator-cli", "generate",
		"-i", "/spec/"+specName,
		"-g", "typescript-fetch",
		"-o", "/out",
		"--additional-properties=npmName="+opts.PackageName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("typescript.generateDocker: %w", err)
	}
	return nil
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/generators/... -run TestTypeScriptGenerator -v`
Expected: PASS (or skip if no generator available)

**Step 5: Commit**

```bash
git add pkg/sdk/generators/typescript.go pkg/sdk/generators/typescript_test.go
git commit -m "feat(sdk): add TypeScript generator

Uses openapi-typescript-codegen (native or npx) with Docker fallback.
Generates TypeScript-fetch client from OpenAPI spec.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 5: Implement Python Generator

**Files:**
- Create: `pkg/sdk/generators/python.go`
- Create: `pkg/sdk/generators/python_test.go`

**Step 1: Write the failing test**

```go
package generators

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPythonGenerator_Good_Available(t *testing.T) {
	g := NewPythonGenerator()
	_ = g.Available()
	_ = g.Language()
	_ = g.Install()
}

func TestPythonGenerator_Good_Generate(t *testing.T) {
	g := NewPythonGenerator()
	if !g.Available() && !dockerAvailable() {
		t.Skip("no Python generator available")
	}

	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.yaml")
	spec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
`
	os.WriteFile(specPath, []byte(spec), 0644)

	outputDir := filepath.Join(tmpDir, "sdk", "python")
	err := g.Generate(context.Background(), Options{
		SpecPath:    specPath,
		OutputDir:   outputDir,
		PackageName: "test_api",
		Version:     "1.0.0",
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/generators/... -run TestPythonGenerator -v`
Expected: FAIL

**Step 3: Write implementation**

```go
package generators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// PythonGenerator generates Python SDKs using openapi-python-client.
type PythonGenerator struct{}

// NewPythonGenerator creates a new Python generator.
func NewPythonGenerator() *PythonGenerator {
	return &PythonGenerator{}
}

// Language returns "python".
func (g *PythonGenerator) Language() string {
	return "python"
}

// Available checks if openapi-python-client is installed.
func (g *PythonGenerator) Available() bool {
	_, err := exec.LookPath("openapi-python-client")
	return err == nil
}

// Install returns installation instructions.
func (g *PythonGenerator) Install() string {
	return "pip install openapi-python-client"
}

// Generate creates Python SDK from OpenAPI spec.
func (g *PythonGenerator) Generate(ctx context.Context, opts Options) error {
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("python.Generate: failed to create output dir: %w", err)
	}

	if g.Available() {
		return g.generateNative(ctx, opts)
	}
	return g.generateDocker(ctx, opts)
}

func (g *PythonGenerator) generateNative(ctx context.Context, opts Options) error {
	// openapi-python-client creates a directory named after the package
	// We need to generate into a temp location then move
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

	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"-v", specDir+":/spec",
		"-v", opts.OutputDir+":/out",
		"openapitools/openapi-generator-cli", "generate",
		"-i", "/spec/"+specName,
		"-g", "python",
		"-o", "/out",
		"--additional-properties=packageName="+opts.PackageName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/generators/... -run TestPythonGenerator -v`
Expected: PASS (or skip)

**Step 5: Commit**

```bash
git add pkg/sdk/generators/python.go pkg/sdk/generators/python_test.go
git commit -m "feat(sdk): add Python generator

Uses openapi-python-client with Docker fallback.
Generates Python client from OpenAPI spec.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 6: Implement Go Generator

**Files:**
- Create: `pkg/sdk/generators/go.go`
- Create: `pkg/sdk/generators/go_test.go`

**Step 1: Write the failing test**

```go
package generators

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGoGenerator_Good_Available(t *testing.T) {
	g := NewGoGenerator()
	_ = g.Available()
	_ = g.Language()
	_ = g.Install()
}

func TestGoGenerator_Good_Generate(t *testing.T) {
	g := NewGoGenerator()
	if !g.Available() && !dockerAvailable() {
		t.Skip("no Go generator available")
	}

	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.yaml")
	spec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
`
	os.WriteFile(specPath, []byte(spec), 0644)

	outputDir := filepath.Join(tmpDir, "sdk", "go")
	err := g.Generate(context.Background(), Options{
		SpecPath:    specPath,
		OutputDir:   outputDir,
		PackageName: "testapi",
		Version:     "1.0.0",
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/generators/... -run TestGoGenerator -v`
Expected: FAIL

**Step 3: Write implementation**

```go
package generators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GoGenerator generates Go SDKs using oapi-codegen.
type GoGenerator struct{}

// NewGoGenerator creates a new Go generator.
func NewGoGenerator() *GoGenerator {
	return &GoGenerator{}
}

// Language returns "go".
func (g *GoGenerator) Language() string {
	return "go"
}

// Available checks if oapi-codegen is installed.
func (g *GoGenerator) Available() bool {
	_, err := exec.LookPath("oapi-codegen")
	return err == nil
}

// Install returns installation instructions.
func (g *GoGenerator) Install() string {
	return "go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest"
}

// Generate creates Go SDK from OpenAPI spec.
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

	// Create go.mod
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
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/generators/... -run TestGoGenerator -v`
Expected: PASS (or skip)

**Step 5: Commit**

```bash
git add pkg/sdk/generators/go.go pkg/sdk/generators/go_test.go
git commit -m "feat(sdk): add Go generator

Uses oapi-codegen with Docker fallback.
Generates Go client and types from OpenAPI spec.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 7: Implement PHP Generator

**Files:**
- Create: `pkg/sdk/generators/php.go`
- Create: `pkg/sdk/generators/php_test.go`

**Step 1: Write the failing test**

```go
package generators

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPHPGenerator_Good_Available(t *testing.T) {
	g := NewPHPGenerator()
	_ = g.Available()
	_ = g.Language()
	_ = g.Install()
}

func TestPHPGenerator_Good_Generate(t *testing.T) {
	g := NewPHPGenerator()
	if !g.Available() {
		t.Skip("Docker not available for PHP generator")
	}

	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "spec.yaml")
	spec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
`
	os.WriteFile(specPath, []byte(spec), 0644)

	outputDir := filepath.Join(tmpDir, "sdk", "php")
	err := g.Generate(context.Background(), Options{
		SpecPath:    specPath,
		OutputDir:   outputDir,
		PackageName: "TestApi",
		Version:     "1.0.0",
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/generators/... -run TestPHPGenerator -v`
Expected: FAIL

**Step 3: Write implementation**

```go
package generators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// PHPGenerator generates PHP SDKs using openapi-generator (Docker).
type PHPGenerator struct{}

// NewPHPGenerator creates a new PHP generator.
func NewPHPGenerator() *PHPGenerator {
	return &PHPGenerator{}
}

// Language returns "php".
func (g *PHPGenerator) Language() string {
	return "php"
}

// Available checks if Docker is available.
func (g *PHPGenerator) Available() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// Install returns installation instructions.
func (g *PHPGenerator) Install() string {
	return "Docker is required for PHP SDK generation"
}

// Generate creates PHP SDK from OpenAPI spec using Docker.
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
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/generators/... -run TestPHPGenerator -v`
Expected: PASS (or skip)

**Step 5: Commit**

```bash
git add pkg/sdk/generators/php.go pkg/sdk/generators/php_test.go
git commit -m "feat(sdk): add PHP generator

Uses openapi-generator via Docker.
Generates PHP client from OpenAPI spec.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 8: Implement Breaking Change Detection

**Files:**
- Create: `pkg/sdk/diff.go`
- Create: `pkg/sdk/diff_test.go`

**Step 1: Write the failing test**

```go
package sdk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiff_Good_NoBreaking(t *testing.T) {
	tmpDir := t.TempDir()

	baseSpec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
`
	revSpec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.1.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
  /status:
    get:
      operationId: getStatus
      responses:
        "200":
          description: OK
`
	basePath := filepath.Join(tmpDir, "base.yaml")
	revPath := filepath.Join(tmpDir, "rev.yaml")
	os.WriteFile(basePath, []byte(baseSpec), 0644)
	os.WriteFile(revPath, []byte(revSpec), 0644)

	result, err := Diff(basePath, revPath)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if result.Breaking {
		t.Error("expected no breaking changes for adding endpoint")
	}
}

func TestDiff_Good_Breaking(t *testing.T) {
	tmpDir := t.TempDir()

	baseSpec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
  /users:
    get:
      operationId: getUsers
      responses:
        "200":
          description: OK
`
	revSpec := `openapi: "3.0.0"
info:
  title: Test API
  version: "2.0.0"
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200":
          description: OK
`
	basePath := filepath.Join(tmpDir, "base.yaml")
	revPath := filepath.Join(tmpDir, "rev.yaml")
	os.WriteFile(basePath, []byte(baseSpec), 0644)
	os.WriteFile(revPath, []byte(revSpec), 0644)

	result, err := Diff(basePath, revPath)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !result.Breaking {
		t.Error("expected breaking change for removed endpoint")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/... -run TestDiff -v`
Expected: FAIL (Diff not defined)

**Step 3: Add oasdiff dependency**

Run: `cd /Users/snider/Code/Core/pkg/sdk && go get github.com/tufin/oasdiff@latest github.com/getkin/kin-openapi@latest`

**Step 4: Write implementation**

```go
package sdk

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/checker"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/load"
)

// DiffResult holds the result of comparing two OpenAPI specs.
type DiffResult struct {
	// Breaking is true if breaking changes were detected.
	Breaking bool
	// Changes is the list of breaking changes.
	Changes []string
	// Summary is a human-readable summary.
	Summary string
}

// Diff compares two OpenAPI specs and detects breaking changes.
func Diff(basePath, revisionPath string) (*DiffResult, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Load specs
	baseSpec, err := load.NewSpecInfo(loader, load.NewSource(basePath))
	if err != nil {
		return nil, fmt.Errorf("sdk.Diff: failed to load base spec: %w", err)
	}

	revSpec, err := load.NewSpecInfo(loader, load.NewSource(revisionPath))
	if err != nil {
		return nil, fmt.Errorf("sdk.Diff: failed to load revision spec: %w", err)
	}

	// Compute diff
	diffResult, err := diff.Get(diff.NewConfig(), baseSpec.Spec, revSpec.Spec)
	if err != nil {
		return nil, fmt.Errorf("sdk.Diff: failed to compute diff: %w", err)
	}

	// Check for breaking changes
	config := checker.GetAllChecks()
	breaks := checker.CheckBackwardCompatibilityUntilLevel(
		config,
		diffResult,
		baseSpec.Spec,
		revSpec.Spec,
		checker.ERR, // Only errors (breaking changes)
	)

	// Build result
	result := &DiffResult{
		Breaking: len(breaks) > 0,
		Changes:  make([]string, 0, len(breaks)),
	}

	for _, b := range breaks {
		result.Changes = append(result.Changes, b.GetUncolorizedText(checker.NewDefaultLocalizer()))
	}

	if result.Breaking {
		result.Summary = fmt.Sprintf("%d breaking change(s) detected", len(breaks))
	} else {
		result.Summary = "No breaking changes"
	}

	return result, nil
}

// DiffExitCode returns the exit code for CI integration.
// 0 = no breaking changes, 1 = breaking changes, 2 = error
func DiffExitCode(result *DiffResult, err error) int {
	if err != nil {
		return 2
	}
	if result.Breaking {
		return 1
	}
	return 0
}
```

**Step 5: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/... -run TestDiff -v`
Expected: PASS

**Step 6: Commit**

```bash
git add pkg/sdk/diff.go pkg/sdk/diff_test.go pkg/sdk/go.mod pkg/sdk/go.sum
git commit -m "feat(sdk): add breaking change detection with oasdiff

Compares OpenAPI specs to detect breaking changes:
- Removed endpoints
- Changed required parameters
- Modified response schemas

Returns CI-friendly exit codes (0=ok, 1=breaking, 2=error).

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 9: Wire Up Generate Command

**Files:**
- Modify: `pkg/sdk/sdk.go`

**Step 1: Update SDK.Generate to use generators**

```go
// Add to sdk.go, replacing the stub Generate method

import (
	"github.com/host-uk/core/pkg/sdk/generators"
)

// Generate generates SDKs for all configured languages.
func (s *SDK) Generate(ctx context.Context) error {
	// Detect spec
	specPath, err := s.DetectSpec()
	if err != nil {
		return err
	}

	// Create registry with all generators
	registry := generators.NewRegistry()
	registry.Register(generators.NewTypeScriptGenerator())
	registry.Register(generators.NewPythonGenerator())
	registry.Register(generators.NewGoGenerator())
	registry.Register(generators.NewPHPGenerator())

	// Generate for each language
	for _, lang := range s.config.Languages {
		if err := s.GenerateLanguage(ctx, lang); err != nil {
			return err
		}
	}

	return nil
}

// GenerateLanguage generates SDK for a specific language.
func (s *SDK) GenerateLanguage(ctx context.Context, lang string) error {
	specPath, err := s.DetectSpec()
	if err != nil {
		return err
	}

	registry := generators.NewRegistry()
	registry.Register(generators.NewTypeScriptGenerator())
	registry.Register(generators.NewPythonGenerator())
	registry.Register(generators.NewGoGenerator())
	registry.Register(generators.NewPHPGenerator())

	gen, ok := registry.Get(lang)
	if !ok {
		return fmt.Errorf("sdk.GenerateLanguage: unknown language: %s", lang)
	}

	if !gen.Available() {
		fmt.Printf("Warning: %s generator not available. Install with: %s\n", lang, gen.Install())
		fmt.Printf("Falling back to Docker...\n")
	}

	outputDir := filepath.Join(s.projectDir, s.config.Output, lang)
	opts := generators.Options{
		SpecPath:    specPath,
		OutputDir:   outputDir,
		PackageName: s.config.Package.Name,
		Version:     s.config.Package.Version,
	}

	fmt.Printf("Generating %s SDK...\n", lang)
	if err := gen.Generate(ctx, opts); err != nil {
		return fmt.Errorf("sdk.GenerateLanguage: %s generation failed: %w", lang, err)
	}
	fmt.Printf("Generated %s SDK at %s\n", lang, outputDir)

	return nil
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/sdk/...`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/sdk/sdk.go
git commit -m "feat(sdk): wire up Generate to use all generators

SDK.Generate() and SDK.GenerateLanguage() now use the
generator registry to generate SDKs for configured languages.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 10: Add CLI Commands

**Files:**
- Create: `cmd/core/cmd/sdk.go`

**Step 1: Create SDK command file**

```go
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/pkg/sdk"
	"github.com/leaanthony/clir"
)

var (
	sdkHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#3b82f6"))

	sdkSuccessStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#22c55e"))

	sdkErrorStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ef4444"))

	sdkDimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280"))
)

// AddSDKCommand adds the sdk command and its subcommands.
func AddSDKCommand(app *clir.Cli) {
	sdkCmd := app.NewSubCommand("sdk", "Generate and manage API SDKs")
	sdkCmd.LongDescription("Generate typed API clients from OpenAPI specs.\n" +
		"Supports TypeScript, Python, Go, and PHP.")

	// sdk generate
	genCmd := sdkCmd.NewSubCommand("generate", "Generate SDKs from OpenAPI spec")
	var specPath, lang string
	genCmd.StringFlag("spec", "Path to OpenAPI spec file", &specPath)
	genCmd.StringFlag("lang", "Generate only this language", &lang)
	genCmd.Action(func() error {
		return runSDKGenerate(specPath, lang)
	})

	// sdk diff
	diffCmd := sdkCmd.NewSubCommand("diff", "Check for breaking API changes")
	var basePath string
	diffCmd.StringFlag("base", "Base spec (version tag or file)", &basePath)
	diffCmd.StringFlag("spec", "Current spec file", &specPath)
	diffCmd.Action(func() error {
		return runSDKDiff(basePath, specPath)
	})

	// sdk validate
	validateCmd := sdkCmd.NewSubCommand("validate", "Validate OpenAPI spec")
	validateCmd.StringFlag("spec", "Path to OpenAPI spec file", &specPath)
	validateCmd.Action(func() error {
		return runSDKValidate(specPath)
	})
}

func runSDKGenerate(specPath, lang string) error {
	ctx := context.Background()

	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load config
	config := sdk.DefaultConfig()
	if specPath != "" {
		config.Spec = specPath
	}

	s := sdk.New(projectDir, config)

	fmt.Printf("%s Generating SDKs\n", sdkHeaderStyle.Render("SDK:"))

	if lang != "" {
		// Generate single language
		if err := s.GenerateLanguage(ctx, lang); err != nil {
			fmt.Printf("%s %v\n", sdkErrorStyle.Render("Error:"), err)
			return err
		}
	} else {
		// Generate all
		if err := s.Generate(ctx); err != nil {
			fmt.Printf("%s %v\n", sdkErrorStyle.Render("Error:"), err)
			return err
		}
	}

	fmt.Printf("%s SDK generation complete\n", sdkSuccessStyle.Render("Success:"))
	return nil
}

func runSDKDiff(basePath, specPath string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Detect current spec if not provided
	if specPath == "" {
		s := sdk.New(projectDir, nil)
		specPath, err = s.DetectSpec()
		if err != nil {
			return err
		}
	}

	if basePath == "" {
		return fmt.Errorf("--base is required (version tag or file path)")
	}

	fmt.Printf("%s Checking for breaking changes\n", sdkHeaderStyle.Render("SDK Diff:"))
	fmt.Printf("  Base:     %s\n", sdkDimStyle.Render(basePath))
	fmt.Printf("  Current:  %s\n", sdkDimStyle.Render(specPath))
	fmt.Println()

	result, err := sdk.Diff(basePath, specPath)
	if err != nil {
		fmt.Printf("%s %v\n", sdkErrorStyle.Render("Error:"), err)
		os.Exit(2)
	}

	if result.Breaking {
		fmt.Printf("%s %s\n", sdkErrorStyle.Render("Breaking:"), result.Summary)
		for _, change := range result.Changes {
			fmt.Printf("  - %s\n", change)
		}
		os.Exit(1)
	}

	fmt.Printf("%s %s\n", sdkSuccessStyle.Render("OK:"), result.Summary)
	return nil
}

func runSDKValidate(specPath string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	s := sdk.New(projectDir, &sdk.Config{Spec: specPath})

	fmt.Printf("%s Validating OpenAPI spec\n", sdkHeaderStyle.Render("SDK:"))

	detectedPath, err := s.DetectSpec()
	if err != nil {
		fmt.Printf("%s %v\n", sdkErrorStyle.Render("Error:"), err)
		return err
	}

	fmt.Printf("  Spec: %s\n", sdkDimStyle.Render(detectedPath))
	fmt.Printf("%s Spec is valid\n", sdkSuccessStyle.Render("OK:"))
	return nil
}
```

**Step 2: Register command in root.go**

Add to root.go after other command registrations:
```go
AddSDKCommand(app)
```

**Step 3: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./cmd/core/...`
Expected: No errors

**Step 4: Commit**

```bash
git add cmd/core/cmd/sdk.go cmd/core/cmd/root.go
git commit -m "feat(cli): add sdk command with generate, diff, validate

Commands:
- core sdk generate [--spec FILE] [--lang LANG]
- core sdk diff --base VERSION [--spec FILE]
- core sdk validate [--spec FILE]

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 11: Add SDK Config to Release Config

**Files:**
- Modify: `pkg/release/config.go`

**Step 1: Add SDK field to Config**

Add to Config struct in config.go:
```go
// SDK configures SDK generation.
SDK *SDKConfig `yaml:"sdk,omitempty"`
```

Add SDKConfig type:
```go
// SDKConfig holds SDK generation configuration.
type SDKConfig struct {
	// Spec is the path to the OpenAPI spec file.
	Spec string `yaml:"spec,omitempty"`
	// Languages to generate.
	Languages []string `yaml:"languages,omitempty"`
	// Output directory (default: sdk/).
	Output string `yaml:"output,omitempty"`
	// Package naming.
	Package SDKPackageConfig `yaml:"package,omitempty"`
	// Diff configuration.
	Diff SDKDiffConfig `yaml:"diff,omitempty"`
	// Publish configuration.
	Publish SDKPublishConfig `yaml:"publish,omitempty"`
}

// SDKPackageConfig holds package naming configuration.
type SDKPackageConfig struct {
	Name    string `yaml:"name,omitempty"`
	Version string `yaml:"version,omitempty"`
}

// SDKDiffConfig holds diff configuration.
type SDKDiffConfig struct {
	Enabled        bool `yaml:"enabled,omitempty"`
	FailOnBreaking bool `yaml:"fail_on_breaking,omitempty"`
}

// SDKPublishConfig holds monorepo publish configuration.
type SDKPublishConfig struct {
	Repo string `yaml:"repo,omitempty"`
	Path string `yaml:"path,omitempty"`
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/release/...`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/release/config.go
git commit -m "feat(release): add SDK configuration to release.yaml

Adds sdk: section to .core/release.yaml for configuring
OpenAPI SDK generation during releases.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 12: Add SDK Example to Docs

**Files:**
- Create: `docs/examples/sdk-full.yaml`

**Step 1: Create example file**

```yaml
# Example: Full SDK Configuration
# Generate typed API clients from OpenAPI specs

sdk:
  # OpenAPI spec source (auto-detected if omitted)
  spec: api/openapi.yaml

  # Languages to generate
  languages:
    - typescript
    - python
    - go
    - php

  # Output directory (default: sdk/)
  output: sdk/

  # Package naming
  package:
    name: myapi
    version: "{{.Version}}"

  # Breaking change detection
  diff:
    enabled: true
    fail_on_breaking: true  # CI fails on breaking changes

  # Optional: publish to monorepo
  publish:
    repo: myorg/sdks
    path: packages/myapi

# Required tools (install one per language):
#   TypeScript: npm i -g openapi-typescript-codegen (or Docker)
#   Python:     pip install openapi-python-client (or Docker)
#   Go:         go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
#   PHP:        Docker required
#
# Usage:
#   core sdk generate              # Generate all configured languages
#   core sdk generate --lang go    # Generate single language
#   core sdk diff --base v1.0.0    # Check for breaking changes
#   core sdk validate              # Validate spec
```

**Step 2: Commit**

```bash
git add docs/examples/sdk-full.yaml
git commit -m "docs: add SDK configuration example

Shows full SDK config with all options:
- Language selection
- Breaking change detection
- Monorepo publishing

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 13: Final Integration Test

**Step 1: Build and verify CLI**

Run: `cd /Users/snider/Code/Core && go build -o bin/core ./cmd/core && ./bin/core sdk --help`
Expected: Shows sdk command help

**Step 2: Run all tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/sdk/... -v`
Expected: All tests pass

**Step 3: Final commit if needed**

```bash
git add -A
git commit -m "chore(sdk): finalize S3.4 SDK generation

All SDK generation features complete:
- OpenAPI spec detection
- TypeScript, Python, Go, PHP generators
- Breaking change detection with oasdiff
- CLI commands (generate, diff, validate)
- Integration with release config

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Summary

13 tasks covering:
1. Package structure
2. Spec detection
3. Generator interface
4. TypeScript generator
5. Python generator
6. Go generator
7. PHP generator
8. Breaking change detection
9. Wire up Generate
10. CLI commands
11. Release config integration
12. Documentation example
13. Integration test
