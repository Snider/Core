package generators

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGoGenerator_Good_Available(t *testing.T) {
	g := NewGoGenerator()

	// These should not panic
	lang := g.Language()
	if lang != "go" {
		t.Errorf("expected language 'go', got '%s'", lang)
	}

	_ = g.Available()

	install := g.Install()
	if install == "" {
		t.Error("expected non-empty install instructions")
	}
}

func TestGoGenerator_Good_Generate(t *testing.T) {
	g := NewGoGenerator()
	if !g.Available() && !dockerAvailable() {
		t.Skip("no Go generator available (neither native nor docker)")
	}

	// Create temp directories
	tmpDir := t.TempDir()
	specPath := createTestSpec(t, tmpDir)
	outputDir := filepath.Join(tmpDir, "output")

	opts := Options{
		SpecPath:    specPath,
		OutputDir:   outputDir,
		PackageName: "testclient",
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
