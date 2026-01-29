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
