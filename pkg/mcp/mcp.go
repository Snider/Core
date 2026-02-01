// Package mcp provides a lightweight MCP (Model Context Protocol) server for CLI use.
// For full GUI integration (display, webview, process management), see core-gui/pkg/mcp.
package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Service provides a lightweight MCP server with file operations only.
// For full GUI features, use the core-gui package.
type Service struct {
	server *mcp.Server
}

// New creates a new MCP service with file operations.
func New() *Service {
	impl := &mcp.Implementation{
		Name:    "core-cli",
		Version: "0.1.0",
	}

	server := mcp.NewServer(impl, nil)
	s := &Service{server: server}
	s.registerTools()
	return s
}

// registerTools adds file operation tools to the MCP server.
func (s *Service) registerTools() {
	// File operations
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "file_read",
		Description: "Read the contents of a file",
	}, s.readFile)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "file_write",
		Description: "Write content to a file",
	}, s.writeFile)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "file_delete",
		Description: "Delete a file or empty directory",
	}, s.deleteFile)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "file_rename",
		Description: "Rename or move a file",
	}, s.renameFile)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "file_exists",
		Description: "Check if a file or directory exists",
	}, s.fileExists)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "file_edit",
		Description: "Edit a file by replacing old_string with new_string. Use replace_all=true to replace all occurrences.",
	}, s.editDiff)

	// Directory operations
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "dir_list",
		Description: "List contents of a directory",
	}, s.listDirectory)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "dir_create",
		Description: "Create a new directory",
	}, s.createDirectory)

	// Language detection
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "lang_detect",
		Description: "Detect the programming language of a file",
	}, s.detectLanguage)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "lang_list",
		Description: "Get list of supported programming languages",
	}, s.getSupportedLanguages)
}

// Tool input/output types for MCP file operations.

// ReadFileInput contains parameters for reading a file.
type ReadFileInput struct {
	Path string `json:"path"`
}

// ReadFileOutput contains the result of reading a file.
type ReadFileOutput struct {
	Content  string `json:"content"`
	Language string `json:"language"`
	Path     string `json:"path"`
}

// WriteFileInput contains parameters for writing a file.
type WriteFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteFileOutput contains the result of writing a file.
type WriteFileOutput struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
}

// ListDirectoryInput contains parameters for listing a directory.
type ListDirectoryInput struct {
	Path string `json:"path"`
}

// ListDirectoryOutput contains the result of listing a directory.
type ListDirectoryOutput struct {
	Entries []DirectoryEntry `json:"entries"`
	Path    string           `json:"path"`
}

// DirectoryEntry represents a single entry in a directory listing.
type DirectoryEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

// CreateDirectoryInput contains parameters for creating a directory.
type CreateDirectoryInput struct {
	Path string `json:"path"`
}

// CreateDirectoryOutput contains the result of creating a directory.
type CreateDirectoryOutput struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
}

// DeleteFileInput contains parameters for deleting a file.
type DeleteFileInput struct {
	Path string `json:"path"`
}

// DeleteFileOutput contains the result of deleting a file.
type DeleteFileOutput struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
}

// RenameFileInput contains parameters for renaming a file.
type RenameFileInput struct {
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
}

// RenameFileOutput contains the result of renaming a file.
type RenameFileOutput struct {
	Success bool   `json:"success"`
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
}

// FileExistsInput contains parameters for checking file existence.
type FileExistsInput struct {
	Path string `json:"path"`
}

// FileExistsOutput contains the result of checking file existence.
type FileExistsOutput struct {
	Exists bool   `json:"exists"`
	IsDir  bool   `json:"isDir"`
	Path   string `json:"path"`
}

// DetectLanguageInput contains parameters for detecting file language.
type DetectLanguageInput struct {
	Path string `json:"path"`
}

// DetectLanguageOutput contains the detected programming language.
type DetectLanguageOutput struct {
	Language string `json:"language"`
	Path     string `json:"path"`
}

// GetSupportedLanguagesInput is an empty struct for the languages query.
type GetSupportedLanguagesInput struct{}

// GetSupportedLanguagesOutput contains the list of supported languages.
type GetSupportedLanguagesOutput struct {
	Languages []LanguageInfo `json:"languages"`
}

// LanguageInfo describes a supported programming language.
type LanguageInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
}

// EditDiffInput contains parameters for editing a file via diff.
type EditDiffInput struct {
	Path       string `json:"path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

// EditDiffOutput contains the result of a diff-based edit operation.
type EditDiffOutput struct {
	Path         string `json:"path"`
	Success      bool   `json:"success"`
	Replacements int    `json:"replacements"`
}

// Tool handlers

func (s *Service) readFile(ctx context.Context, req *mcp.CallToolRequest, input ReadFileInput) (*mcp.CallToolResult, ReadFileOutput, error) {
	content, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, ReadFileOutput{}, fmt.Errorf("failed to read file: %w", err)
	}
	return nil, ReadFileOutput{
		Content:  string(content),
		Language: detectLanguageFromPath(input.Path),
		Path:     input.Path,
	}, nil
}

func (s *Service) writeFile(ctx context.Context, req *mcp.CallToolRequest, input WriteFileInput) (*mcp.CallToolResult, WriteFileOutput, error) {
	dir := filepath.Dir(input.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, WriteFileOutput{}, fmt.Errorf("failed to create directory: %w", err)
	}
	err := os.WriteFile(input.Path, []byte(input.Content), 0644)
	if err != nil {
		return nil, WriteFileOutput{}, fmt.Errorf("failed to write file: %w", err)
	}
	return nil, WriteFileOutput{Success: true, Path: input.Path}, nil
}

func (s *Service) listDirectory(ctx context.Context, req *mcp.CallToolRequest, input ListDirectoryInput) (*mcp.CallToolResult, ListDirectoryOutput, error) {
	entries, err := os.ReadDir(input.Path)
	if err != nil {
		return nil, ListDirectoryOutput{}, fmt.Errorf("failed to list directory: %w", err)
	}
	result := make([]DirectoryEntry, 0, len(entries))
	for _, e := range entries {
		info, _ := e.Info()
		var size int64
		if info != nil {
			size = info.Size()
		}
		result = append(result, DirectoryEntry{
			Name:  e.Name(),
			Path:  filepath.Join(input.Path, e.Name()),
			IsDir: e.IsDir(),
			Size:  size,
		})
	}
	return nil, ListDirectoryOutput{Entries: result, Path: input.Path}, nil
}

func (s *Service) createDirectory(ctx context.Context, req *mcp.CallToolRequest, input CreateDirectoryInput) (*mcp.CallToolResult, CreateDirectoryOutput, error) {
	err := os.MkdirAll(input.Path, 0755)
	if err != nil {
		return nil, CreateDirectoryOutput{}, fmt.Errorf("failed to create directory: %w", err)
	}
	return nil, CreateDirectoryOutput{Success: true, Path: input.Path}, nil
}

func (s *Service) deleteFile(ctx context.Context, req *mcp.CallToolRequest, input DeleteFileInput) (*mcp.CallToolResult, DeleteFileOutput, error) {
	err := os.Remove(input.Path)
	if err != nil {
		return nil, DeleteFileOutput{}, fmt.Errorf("failed to delete file: %w", err)
	}
	return nil, DeleteFileOutput{Success: true, Path: input.Path}, nil
}

func (s *Service) renameFile(ctx context.Context, req *mcp.CallToolRequest, input RenameFileInput) (*mcp.CallToolResult, RenameFileOutput, error) {
	err := os.Rename(input.OldPath, input.NewPath)
	if err != nil {
		return nil, RenameFileOutput{}, fmt.Errorf("failed to rename file: %w", err)
	}
	return nil, RenameFileOutput{Success: true, OldPath: input.OldPath, NewPath: input.NewPath}, nil
}

func (s *Service) fileExists(ctx context.Context, req *mcp.CallToolRequest, input FileExistsInput) (*mcp.CallToolResult, FileExistsOutput, error) {
	info, err := os.Stat(input.Path)
	if os.IsNotExist(err) {
		return nil, FileExistsOutput{Exists: false, IsDir: false, Path: input.Path}, nil
	}
	if err != nil {
		return nil, FileExistsOutput{}, fmt.Errorf("failed to check file: %w", err)
	}
	return nil, FileExistsOutput{Exists: true, IsDir: info.IsDir(), Path: input.Path}, nil
}

func (s *Service) detectLanguage(ctx context.Context, req *mcp.CallToolRequest, input DetectLanguageInput) (*mcp.CallToolResult, DetectLanguageOutput, error) {
	lang := detectLanguageFromPath(input.Path)
	return nil, DetectLanguageOutput{Language: lang, Path: input.Path}, nil
}

func (s *Service) getSupportedLanguages(ctx context.Context, req *mcp.CallToolRequest, input GetSupportedLanguagesInput) (*mcp.CallToolResult, GetSupportedLanguagesOutput, error) {
	languages := []LanguageInfo{
		{ID: "typescript", Name: "TypeScript", Extensions: []string{".ts", ".tsx"}},
		{ID: "javascript", Name: "JavaScript", Extensions: []string{".js", ".jsx"}},
		{ID: "go", Name: "Go", Extensions: []string{".go"}},
		{ID: "python", Name: "Python", Extensions: []string{".py"}},
		{ID: "rust", Name: "Rust", Extensions: []string{".rs"}},
		{ID: "java", Name: "Java", Extensions: []string{".java"}},
		{ID: "php", Name: "PHP", Extensions: []string{".php"}},
		{ID: "ruby", Name: "Ruby", Extensions: []string{".rb"}},
		{ID: "html", Name: "HTML", Extensions: []string{".html", ".htm"}},
		{ID: "css", Name: "CSS", Extensions: []string{".css"}},
		{ID: "json", Name: "JSON", Extensions: []string{".json"}},
		{ID: "yaml", Name: "YAML", Extensions: []string{".yaml", ".yml"}},
		{ID: "markdown", Name: "Markdown", Extensions: []string{".md", ".markdown"}},
		{ID: "sql", Name: "SQL", Extensions: []string{".sql"}},
		{ID: "shell", Name: "Shell", Extensions: []string{".sh", ".bash"}},
	}
	return nil, GetSupportedLanguagesOutput{Languages: languages}, nil
}

func (s *Service) editDiff(ctx context.Context, req *mcp.CallToolRequest, input EditDiffInput) (*mcp.CallToolResult, EditDiffOutput, error) {
	if input.OldString == "" {
		return nil, EditDiffOutput{}, fmt.Errorf("old_string cannot be empty")
	}

	content, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, EditDiffOutput{}, fmt.Errorf("failed to read file: %w", err)
	}

	fileContent := string(content)
	count := 0

	if input.ReplaceAll {
		count = strings.Count(fileContent, input.OldString)
		if count == 0 {
			return nil, EditDiffOutput{}, fmt.Errorf("old_string not found in file")
		}
		fileContent = strings.ReplaceAll(fileContent, input.OldString, input.NewString)
	} else {
		if !strings.Contains(fileContent, input.OldString) {
			return nil, EditDiffOutput{}, fmt.Errorf("old_string not found in file")
		}
		fileContent = strings.Replace(fileContent, input.OldString, input.NewString, 1)
		count = 1
	}

	err = os.WriteFile(input.Path, []byte(fileContent), 0644)
	if err != nil {
		return nil, EditDiffOutput{}, fmt.Errorf("failed to write file: %w", err)
	}

	return nil, EditDiffOutput{
		Path:         input.Path,
		Success:      true,
		Replacements: count,
	}, nil
}

// detectLanguageFromPath maps file extensions to language IDs.
func detectLanguageFromPath(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".rb":
		return "ruby"
	case ".java":
		return "java"
	case ".php":
		return "php"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc", ".cxx":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".scss":
		return "scss"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".xml":
		return "xml"
	case ".md", ".markdown":
		return "markdown"
	case ".sql":
		return "sql"
	case ".sh", ".bash":
		return "shell"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	default:
		if filepath.Base(path) == "Dockerfile" {
			return "dockerfile"
		}
		return "plaintext"
	}
}

// Run starts the MCP server on stdio.
func (s *Service) Run(ctx context.Context) error {
	return s.server.Run(ctx, &mcp.StdioTransport{})
}

// Server returns the underlying MCP server for advanced configuration.
func (s *Service) Server() *mcp.Server {
	return s.server
}
