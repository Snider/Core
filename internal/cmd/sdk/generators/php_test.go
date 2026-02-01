package generators

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPHPGenerator_Good_Available(t *testing.T) {
	g := NewPHPGenerator()

	// These should not panic
	lang := g.Language()
	if lang != "php" {
		t.Errorf("expected language 'php', got '%s'", lang)
	}

	_ = g.Available()

	install := g.Install()
	if install == "" {
		t.Error("expected non-empty install instructions")
	}
}

func TestPHPGenerator_Good_Generate(t *testing.T) {
	g := NewPHPGenerator()
	if !g.Available() {
		t.Skip("no PHP generator available (docker not installed)")
	}

	// Create temp directories
	tmpDir := t.TempDir()
	specPath := createTestSpec(t, tmpDir)
	outputDir := filepath.Join(tmpDir, "output")

	opts := Options{
		SpecPath:    specPath,
		OutputDir:   outputDir,
		PackageName: "TestClient",
		Version:     "1.0.0",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	err := g.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory was not created")
	}
}
