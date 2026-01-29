package devops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImageManager_Good_IsInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tmpDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Not installed yet
	if mgr.IsInstalled() {
		t.Error("expected IsInstalled() to be false")
	}

	// Create fake image
	imagePath := filepath.Join(tmpDir, ImageName())
	os.WriteFile(imagePath, []byte("fake"), 0644)

	// Now installed
	if !mgr.IsInstalled() {
		t.Error("expected IsInstalled() to be true")
	}
}
