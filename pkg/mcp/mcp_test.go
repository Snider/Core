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

	s := New()

	if s.workspaceRoot != cwd {
		t.Errorf("Expected default workspace root %s, got %s", cwd, s.workspaceRoot)
	}
	if s.medium == nil {
		t.Error("Expected medium to be set")
	}
}

func TestNew_Good_CustomWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	s := New(WithWorkspaceRoot(tmpDir))

	if s.workspaceRoot != tmpDir {
		t.Errorf("Expected workspace root %s, got %s", tmpDir, s.workspaceRoot)
	}
	if s.medium == nil {
		t.Error("Expected medium to be set")
	}
}

func TestNew_Good_NoRestriction(t *testing.T) {
	s := New(WithWorkspaceRoot(""))

	if s.workspaceRoot != "" {
		t.Errorf("Expected empty workspace root, got %s", s.workspaceRoot)
	}
	if s.medium == nil {
		t.Error("Expected medium to be set (unsandboxed)")
	}
}

func TestMedium_Good_ReadWrite(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(WithWorkspaceRoot(tmpDir))

	// Write a file
	testContent := "hello world"
	err := s.medium.Write("test.txt", testContent)
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
	s := New(WithWorkspaceRoot(tmpDir))

	err := s.medium.EnsureDir("subdir/nested")
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
	s := New(WithWorkspaceRoot(tmpDir))

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

func TestResolvePath_Good(t *testing.T) {
	tmpDir := t.TempDir()
	s := New(WithWorkspaceRoot(tmpDir))

	// Relative path should resolve to workspace
	resolved := s.resolvePath("test.txt")
	expected := filepath.Join(tmpDir, "test.txt")
	if resolved != expected {
		t.Errorf("Expected %s, got %s", expected, resolved)
	}

	// Absolute path should stay absolute
	absPath := "/etc/passwd"
	resolved = s.resolvePath(absPath)
	if resolved != absPath {
		t.Errorf("Expected %s, got %s", absPath, resolved)
	}
}

func TestResolvePath_Good_NoWorkspace(t *testing.T) {
	s := New(WithWorkspaceRoot(""))

	// With no workspace, relative paths resolve to cwd
	cwd, _ := os.Getwd()
	resolved := s.resolvePath("test.txt")
	expected := filepath.Join(cwd, "test.txt")
	if resolved != expected {
		t.Errorf("Expected %s, got %s", expected, resolved)
	}
}
