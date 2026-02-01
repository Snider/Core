package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath_Good(t *testing.T) {
	// Create a temp directory as workspace root
	tmpDir := t.TempDir()

	s := New(WithWorkspaceRoot(tmpDir))

	// Test valid path within workspace
	validPath := filepath.Join(tmpDir, "test.txt")
	result, err := s.validatePath(validPath)
	if err != nil {
		t.Errorf("Expected no error for valid path, got: %v", err)
	}
	if result != validPath {
		t.Errorf("Expected path %s, got %s", validPath, result)
	}

	// Test nested path within workspace
	nestedPath := filepath.Join(tmpDir, "subdir", "test.txt")
	result, err = s.validatePath(nestedPath)
	if err != nil {
		t.Errorf("Expected no error for nested path, got: %v", err)
	}
	if result != nestedPath {
		t.Errorf("Expected path %s, got %s", nestedPath, result)
	}
}

func TestValidatePath_Bad_DirectoryTraversal(t *testing.T) {
	// Create a temp directory as workspace root
	tmpDir := t.TempDir()

	s := New(WithWorkspaceRoot(tmpDir))

	// Test path traversal attempt
	traversalPath := filepath.Join(tmpDir, "..", "etc", "passwd")
	_, err := s.validatePath(traversalPath)
	if err == nil {
		t.Error("Expected error for directory traversal attempt")
	}

	// Test explicit parent directory
	parentPath := filepath.Join(tmpDir, "..")
	_, err = s.validatePath(parentPath)
	if err == nil {
		t.Error("Expected error for parent directory access")
	}

	// Test absolute path outside workspace
	outsidePath := "/etc/passwd"
	_, err = s.validatePath(outsidePath)
	if err == nil {
		t.Error("Expected error for path outside workspace")
	}
}

func TestValidatePath_Good_NoRestriction(t *testing.T) {
	// Create service with no workspace restriction
	s := New(WithWorkspaceRoot(""))

	// Any path should be allowed
	result, err := s.validatePath("/etc/passwd")
	if err != nil {
		t.Errorf("Expected no error with no restriction, got: %v", err)
	}
	if result != "/etc/passwd" {
		t.Errorf("Expected path /etc/passwd, got %s", result)
	}
}

func TestNew_Good_DefaultWorkspace(t *testing.T) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	s := New()

	// Default should be current working directory
	if s.workspaceRoot != cwd {
		t.Errorf("Expected default workspace root %s, got %s", cwd, s.workspaceRoot)
	}
}

func TestNew_Good_CustomWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	s := New(WithWorkspaceRoot(tmpDir))

	if s.workspaceRoot != tmpDir {
		t.Errorf("Expected workspace root %s, got %s", tmpDir, s.workspaceRoot)
	}
}

func TestValidatePath_Good_RelativePath(t *testing.T) {
	// Create a temp directory as workspace root
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	realTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("Failed to resolve symlinks: %v", err)
	}

	// Change to the temp directory
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(realTmpDir)

	s := New(WithWorkspaceRoot(realTmpDir))

	// Test relative path within workspace
	result, err := s.validatePath("test.txt")
	if err != nil {
		t.Errorf("Expected no error for relative path, got: %v", err)
	}
	expected := filepath.Join(realTmpDir, "test.txt")
	if result != expected {
		t.Errorf("Expected path %s, got %s", expected, result)
	}
}
