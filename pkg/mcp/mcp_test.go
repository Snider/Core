package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_Good_DefaultWorkspace(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if s.workspaceRoot != cwd {
		t.Errorf("Expected default workspace root %s, got %s", cwd, s.workspaceRoot)
	}
	if s.medium == nil {
		t.Error("Expected medium to be set")
	}
}

func TestNew_Good_CustomWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	s, err := New(WithWorkspaceRoot(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if s.workspaceRoot != tmpDir {
		t.Errorf("Expected workspace root %s, got %s", tmpDir, s.workspaceRoot)
	}
	if s.medium == nil {
		t.Error("Expected medium to be set")
	}
}

func TestNew_Good_NoRestriction(t *testing.T) {
	s, err := New(WithWorkspaceRoot(""))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if s.workspaceRoot != "" {
		t.Errorf("Expected empty workspace root, got %s", s.workspaceRoot)
	}
	if s.medium == nil {
		t.Error("Expected medium to be set (unsandboxed)")
	}
}

func TestMedium_Good_ReadWrite(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := New(WithWorkspaceRoot(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Write a file
	testContent := "hello world"
	err = s.medium.Write("test.txt", testContent)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Read it back
	content, err := s.medium.Read("test.txt")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if content != testContent {
		t.Errorf("Expected content %q, got %q", testContent, content)
	}

	// Verify file exists on disk
	diskPath := filepath.Join(tmpDir, "test.txt")
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		t.Error("File should exist on disk")
	}
}

func TestMedium_Good_EnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := New(WithWorkspaceRoot(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	err = s.medium.EnsureDir("subdir/nested")
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Verify directory exists
	diskPath := filepath.Join(tmpDir, "subdir", "nested")
	info, err := os.Stat(diskPath)
	if os.IsNotExist(err) {
		t.Error("Directory should exist on disk")
	}
	if err == nil && !info.IsDir() {
		t.Error("Path should be a directory")
	}
}

func TestMedium_Good_IsFile(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := New(WithWorkspaceRoot(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// File doesn't exist yet
	if s.medium.IsFile("test.txt") {
		t.Error("File should not exist yet")
	}

	// Create the file
	_ = s.medium.Write("test.txt", "content")

	// Now it should exist
	if !s.medium.IsFile("test.txt") {
		t.Error("File should exist after write")
	}
}

func TestSandboxing_Traversal_Sanitized(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := New(WithWorkspaceRoot(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Path traversal is sanitized (.. becomes .), so ../secret.txt becomes
	// ./secret.txt in the workspace. Since that file doesn't exist, we get
	// a file not found error (not a traversal error).
	_, err = s.medium.Read("../secret.txt")
	if err == nil {
		t.Error("Expected error (file not found)")
	}

	// Absolute paths are allowed through - they access the real filesystem.
	// This is intentional for full filesystem access. Callers wanting sandboxing
	// should validate inputs before calling Medium.
}

func TestSandboxing_Symlinks_Followed(t *testing.T) {
	tmpDir := t.TempDir()
	outsideDir := t.TempDir()

	// Create a target file outside workspace
	targetFile := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(targetFile, []byte("secret"), 0644); err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create symlink inside workspace pointing outside
	symlinkPath := filepath.Join(tmpDir, "link")
	if err := os.Symlink(targetFile, symlinkPath); err != nil {
		t.Skipf("Symlinks not supported: %v", err)
	}

	s, err := New(WithWorkspaceRoot(tmpDir))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Symlinks are followed - no traversal blocking at Medium level.
	// This is intentional for simplicity. Callers wanting to block symlinks
	// should validate inputs before calling Medium.
	content, err := s.medium.Read("link")
	if err != nil {
		t.Errorf("Expected symlink to be followed, got error: %v", err)
	}
	if content != "secret" {
		t.Errorf("Expected 'secret', got '%s'", content)
	}
}
