// Package mcp provides an MCP (Model Context Protocol) server for Core.
// This allows Claude Code and other MCP clients to interact with Core's
// IDE, file system, and display services.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/host-uk/core/pkg/core"
	"github.com/host-uk/core/pkg/display"
	"github.com/host-uk/core/pkg/ide"
	"github.com/host-uk/core/pkg/process"
	"github.com/host-uk/core/pkg/webview"
	"github.com/host-uk/core/pkg/ws"
)

// Service provides an MCP server that exposes Core functionality.
type Service struct {
	core      *core.Core
	server    *mcp.Server
	ide       *ide.Service
	display   *display.Service
	process   *process.Service
	webview   *webview.Service
	wsHub     *ws.Hub
	wsPort    int
	wsRunning bool
}

// New creates a new MCP service.
func New(c *core.Core) *Service {
	impl := &mcp.Implementation{
		Name:    "core",
		Version: "0.1.0",
	}

	server := mcp.NewServer(impl, nil)
	s := &Service{
		core:    c,
		server:  server,
		process: process.New(),
	}

	// Try to get the IDE service if available
	if c != nil {
		ideSvc, _ := core.ServiceFor[*ide.Service](c, "github.com/host-uk/core/ide")
		s.ide = ideSvc
	}

	s.registerTools()
	return s
}

// NewStandalone creates an MCP service without a Core instance.
// This allows running the MCP server independently with basic file operations.
func NewStandalone() *Service {
	return NewStandaloneWithPort(9876)
}

// NewStandaloneWithPort creates an MCP service with a specific WebSocket port.
func NewStandaloneWithPort(wsPort int) *Service {
	impl := &mcp.Implementation{
		Name:    "core",
		Version: "0.1.0",
	}

	server := mcp.NewServer(impl, nil)
	hub := ws.NewHub()
	proc := process.New()

	s := &Service{
		server:  server,
		process: proc,
		wsHub:   hub,
		wsPort:  wsPort,
	}

	// Wire process output to WebSocket
	proc.OnOutput(func(processID string, output string) {
		hub.SendProcessOutput(processID, output)
	})

	proc.OnStatusChange(func(processID string, status process.Status, exitCode int) {
		hub.SendProcessStatus(processID, string(status), exitCode)
	})

	s.registerTools()
	return s
}

// registerTools adds all Core tools to the MCP server.
// Naming convention: prefix_action for discoverability
// file_* dir_* lang_* process_*
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

	// Process management
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "process_start",
		Description: "Start a new process with the given command and arguments",
	}, s.processStart)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "process_stop",
		Description: "Stop a running process gracefully",
	}, s.processStop)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "process_kill",
		Description: "Forcefully kill a process",
	}, s.processKill)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "process_list",
		Description: "List all managed processes",
	}, s.processList)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "process_output",
		Description: "Get the output of a process",
	}, s.processOutput)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "process_input",
		Description: "Send input to a running process stdin",
	}, s.processSendInput)

	// WebSocket streaming
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "ws_start",
		Description: "Start WebSocket server for real-time streaming",
	}, s.wsStart)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "ws_info",
		Description: "Get WebSocket server info (port, connected clients)",
	}, s.wsInfo)

	// WebView interaction (only available when embedded in GUI app)
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "webview_list",
		Description: "List all open windows in the application",
	}, s.webviewList)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "webview_eval",
		Description: "Execute JavaScript in a window and return the result",
	}, s.webviewEval)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "webview_console",
		Description: "Get captured console messages (log, warn, error) from the WebView",
	}, s.webviewConsole)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "webview_click",
		Description: "Click an element by CSS selector",
	}, s.webviewClick)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "webview_type",
		Description: "Type text into an element by CSS selector",
	}, s.webviewType)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "webview_query",
		Description: "Query elements by CSS selector and return info about matches",
	}, s.webviewQuery)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "webview_navigate",
		Description: "Navigate to a URL or Angular route",
	}, s.webviewNavigate)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "webview_source",
		Description: "Get the current page HTML source",
	}, s.webviewSource)

	// Window/Display management (the unique value-add for native app control)
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "window_list",
		Description: "List all windows with their positions and sizes",
	}, s.windowList)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "window_get",
		Description: "Get detailed info about a specific window",
	}, s.windowGet)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "window_position",
		Description: "Move a window to a specific position (x, y coordinates)",
	}, s.windowPosition)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "window_size",
		Description: "Resize a window to specific dimensions",
	}, s.windowSize)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "window_bounds",
		Description: "Set both position and size of a window in one call",
	}, s.windowBounds)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "window_maximize",
		Description: "Maximize a window",
	}, s.windowMaximize)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "window_minimize",
		Description: "Minimize a window",
	}, s.windowMinimize)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "window_restore",
		Description: "Restore a window from maximized/minimized state",
	}, s.windowRestore)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "window_focus",
		Description: "Bring a window to the front and focus it",
	}, s.windowFocus)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "screen_list",
		Description: "List all available screens/monitors with their dimensions",
	}, s.screenList)
}

// SetWebView sets the WebView service for GUI interaction.
// This must be called when running embedded in the GUI app.
func (s *Service) SetWebView(wv *webview.Service) {
	s.webview = wv
}

// SetDisplay sets the Display service for window management.
// This must be called when running embedded in the GUI app.
func (s *Service) SetDisplay(d *display.Service) {
	s.display = d
}

// Tool input/output types

// ReadFileInput contains parameters for reading a file.
type ReadFileInput struct {
	// Absolute path to the file to read.
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
	// Absolute path to the file to write.
	Path string `json:"path"`
	// Content to write to the file.
	Content string `json:"content"`
}

// WriteFileOutput contains the result of writing a file.
type WriteFileOutput struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
}

// ListDirectoryInput contains parameters for listing a directory.
type ListDirectoryInput struct {
	// Absolute path to the directory to list.
	Path string `json:"path"`
}

// ListDirectoryOutput contains the result of listing a directory.
type ListDirectoryOutput struct {
	Entries []DirectoryEntry `json:"entries"`
	Path    string           `json:"path"`
}

// DirectoryEntry represents a file or directory entry.
type DirectoryEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

// CreateDirectoryInput contains parameters for creating a directory.
type CreateDirectoryInput struct {
	// Absolute path to the directory to create.
	Path string `json:"path"`
}

// CreateDirectoryOutput contains the result of creating a directory.
type CreateDirectoryOutput struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
}

// DeleteFileInput contains parameters for deleting a file.
type DeleteFileInput struct {
	// Absolute path to the file to delete.
	Path string `json:"path"`
}

// DeleteFileOutput contains the result of deleting a file.
type DeleteFileOutput struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
}

// RenameFileInput contains parameters for renaming a file.
type RenameFileInput struct {
	// Current path of the file.
	OldPath string `json:"oldPath"`
	// New path for the file.
	NewPath string `json:"newPath"`
}

// RenameFileOutput contains the result of renaming a file.
type RenameFileOutput struct {
	Success bool   `json:"success"`
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
}

// FileExistsInput contains parameters for checking if a file exists.
type FileExistsInput struct {
	// Absolute path to check.
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
	// File path to detect language for.
	Path string `json:"path"`
}

type DetectLanguageOutput struct {
	Language string `json:"language"`
	Path     string `json:"path"`
}

type GetSupportedLanguagesInput struct{}

type GetSupportedLanguagesOutput struct {
	Languages []LanguageInfo `json:"languages"`
}

type LanguageInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
}

// EditDiffInput contains parameters for diff-based editing.
type EditDiffInput struct {
	// Absolute path to the file to edit.
	Path string `json:"path"`
	// The text to find and replace.
	OldString string `json:"old_string"`
	// The replacement text.
	NewString string `json:"new_string"`
	// Replace all occurrences if true, otherwise only the first.
	ReplaceAll bool `json:"replace_all,omitempty"`
}

// EditDiffOutput contains the result of the edit.
type EditDiffOutput struct {
	Path         string `json:"path"`
	Success      bool   `json:"success"`
	Replacements int    `json:"replacements"`
}

// Tool handlers

func (s *Service) readFile(ctx context.Context, req *mcp.CallToolRequest, input ReadFileInput) (*mcp.CallToolResult, ReadFileOutput, error) {
	if s.ide != nil {
		info, err := s.ide.OpenFile(input.Path)
		if err != nil {
			return nil, ReadFileOutput{}, fmt.Errorf("failed to read file: %w", err)
		}
		return nil, ReadFileOutput{
			Content:  info.Content,
			Language: info.Language,
			Path:     info.Path,
		}, nil
	}

	// Fallback to direct file read
	content, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, ReadFileOutput{}, fmt.Errorf("failed to read file: %w", err)
	}
	return nil, ReadFileOutput{
		Content:  string(content),
		Language: detectLanguage(input.Path),
		Path:     input.Path,
	}, nil
}

func (s *Service) writeFile(ctx context.Context, req *mcp.CallToolRequest, input WriteFileInput) (*mcp.CallToolResult, WriteFileOutput, error) {
	if s.ide != nil {
		err := s.ide.SaveFile(input.Path, input.Content)
		if err != nil {
			return nil, WriteFileOutput{}, fmt.Errorf("failed to write file: %w", err)
		}
		return nil, WriteFileOutput{Success: true, Path: input.Path}, nil
	}

	// Fallback to direct file write
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
	if s.ide != nil {
		entries, err := s.ide.ListDirectory(input.Path)
		if err != nil {
			return nil, ListDirectoryOutput{}, fmt.Errorf("failed to list directory: %w", err)
		}
		result := make([]DirectoryEntry, 0, len(entries))
		for _, e := range entries {
			result = append(result, DirectoryEntry{
				Name:  e.Name,
				Path:  e.Path,
				IsDir: e.IsDir,
				Size:  e.Size,
			})
		}
		return nil, ListDirectoryOutput{Entries: result, Path: input.Path}, nil
	}

	// Fallback to direct directory listing
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
	if s.ide != nil {
		err := s.ide.CreateDirectory(input.Path)
		if err != nil {
			return nil, CreateDirectoryOutput{}, fmt.Errorf("failed to create directory: %w", err)
		}
		return nil, CreateDirectoryOutput{Success: true, Path: input.Path}, nil
	}

	err := os.MkdirAll(input.Path, 0755)
	if err != nil {
		return nil, CreateDirectoryOutput{}, fmt.Errorf("failed to create directory: %w", err)
	}
	return nil, CreateDirectoryOutput{Success: true, Path: input.Path}, nil
}

func (s *Service) deleteFile(ctx context.Context, req *mcp.CallToolRequest, input DeleteFileInput) (*mcp.CallToolResult, DeleteFileOutput, error) {
	if s.ide != nil {
		err := s.ide.DeleteFile(input.Path)
		if err != nil {
			return nil, DeleteFileOutput{}, fmt.Errorf("failed to delete file: %w", err)
		}
		return nil, DeleteFileOutput{Success: true, Path: input.Path}, nil
	}

	err := os.Remove(input.Path)
	if err != nil {
		return nil, DeleteFileOutput{}, fmt.Errorf("failed to delete file: %w", err)
	}
	return nil, DeleteFileOutput{Success: true, Path: input.Path}, nil
}

func (s *Service) renameFile(ctx context.Context, req *mcp.CallToolRequest, input RenameFileInput) (*mcp.CallToolResult, RenameFileOutput, error) {
	if s.ide != nil {
		err := s.ide.RenameFile(input.OldPath, input.NewPath)
		if err != nil {
			return nil, RenameFileOutput{}, fmt.Errorf("failed to rename file: %w", err)
		}
		return nil, RenameFileOutput{Success: true, OldPath: input.OldPath, NewPath: input.NewPath}, nil
	}

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
	lang := detectLanguage(input.Path)
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
		{ID: "csharp", Name: "C#", Extensions: []string{".cs"}},
		{ID: "cpp", Name: "C++", Extensions: []string{".cpp", ".hpp", ".cc", ".cxx"}},
		{ID: "c", Name: "C", Extensions: []string{".c", ".h"}},
		{ID: "html", Name: "HTML", Extensions: []string{".html", ".htm"}},
		{ID: "css", Name: "CSS", Extensions: []string{".css"}},
		{ID: "scss", Name: "SCSS", Extensions: []string{".scss"}},
		{ID: "json", Name: "JSON", Extensions: []string{".json"}},
		{ID: "yaml", Name: "YAML", Extensions: []string{".yaml", ".yml"}},
		{ID: "markdown", Name: "Markdown", Extensions: []string{".md", ".markdown"}},
		{ID: "sql", Name: "SQL", Extensions: []string{".sql"}},
		{ID: "shell", Name: "Shell", Extensions: []string{".sh", ".bash"}},
		{ID: "xml", Name: "XML", Extensions: []string{".xml"}},
		{ID: "swift", Name: "Swift", Extensions: []string{".swift"}},
		{ID: "kotlin", Name: "Kotlin", Extensions: []string{".kt", ".kts"}},
		{ID: "php", Name: "PHP", Extensions: []string{".php"}},
		{ID: "ruby", Name: "Ruby", Extensions: []string{".rb"}},
	}
	return nil, GetSupportedLanguagesOutput{Languages: languages}, nil
}

func (s *Service) editDiff(ctx context.Context, req *mcp.CallToolRequest, input EditDiffInput) (*mcp.CallToolResult, EditDiffOutput, error) {
	// Read the file
	content, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, EditDiffOutput{}, fmt.Errorf("failed to read file: %w", err)
	}

	fileContent := string(content)
	count := 0

	if input.ReplaceAll {
		// Count occurrences
		count = strings.Count(fileContent, input.OldString)
		if count == 0 {
			return nil, EditDiffOutput{}, fmt.Errorf("old_string not found in file")
		}
		fileContent = strings.ReplaceAll(fileContent, input.OldString, input.NewString)
	} else {
		// Replace only first occurrence
		if !strings.Contains(fileContent, input.OldString) {
			return nil, EditDiffOutput{}, fmt.Errorf("old_string not found in file")
		}
		fileContent = strings.Replace(fileContent, input.OldString, input.NewString, 1)
		count = 1
	}

	// Write the file back
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

// detectLanguage maps file extensions to Monaco editor languages.
func detectLanguage(path string) string {
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
	case ".less":
		return "less"
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
	case ".ps1":
		return "powershell"
	case ".toml":
		return "toml"
	case ".ini", ".cfg":
		return "ini"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".php":
		return "php"
	case ".r":
		return "r"
	case ".lua":
		return "lua"
	case ".pl", ".pm":
		return "perl"
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

// Process management types

// ProcessStartInput contains parameters for starting a process.
type ProcessStartInput struct {
	// Command to execute.
	Command string `json:"command"`
	// Arguments for the command.
	Args []string `json:"args,omitempty"`
	// Working directory for the process.
	Dir string `json:"dir,omitempty"`
}

// ProcessStartOutput contains the result of starting a process.
type ProcessStartOutput struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	Dir       string    `json:"dir"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"startedAt"`
}

// ProcessIDInput contains a process ID parameter.
type ProcessIDInput struct {
	// Process ID to operate on.
	ID string `json:"id"`
}

// ProcessStopOutput contains the result of stopping a process.
type ProcessStopOutput struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
}

// ProcessListInput is empty but required for the handler signature.
type ProcessListInput struct{}

// ProcessInfo represents process information.
type ProcessInfo struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	Dir       string    `json:"dir"`
	Status    string    `json:"status"`
	ExitCode  int       `json:"exitCode"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"startedAt"`
}

// ProcessListOutput contains the list of processes.
type ProcessListOutput struct {
	Processes []ProcessInfo `json:"processes"`
}

// ProcessOutputOutput contains the captured output of a process.
type ProcessOutputOutput struct {
	ID     string `json:"id"`
	Output string `json:"output"`
	Length int    `json:"length"`
}

// ProcessSendInputInput contains input to send to a process.
type ProcessSendInputInput struct {
	// Process ID to send input to.
	ID string `json:"id"`
	// Input text to send.
	Input string `json:"input"`
}

// ProcessSendInputOutput contains the result of sending input.
type ProcessSendInputOutput struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
}

// Process management handlers

func (s *Service) processStart(ctx context.Context, req *mcp.CallToolRequest, input ProcessStartInput) (*mcp.CallToolResult, ProcessStartOutput, error) {
	dir := input.Dir
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			dir = "."
		}
	}

	proc, err := s.process.Start(input.Command, input.Args, dir)
	if err != nil {
		return nil, ProcessStartOutput{}, fmt.Errorf("failed to start process: %w", err)
	}

	info := proc.Info()
	return nil, ProcessStartOutput{
		ID:        info.ID,
		Command:   info.Command,
		Args:      info.Args,
		Dir:       info.Dir,
		PID:       info.PID,
		StartedAt: info.StartedAt,
	}, nil
}

func (s *Service) processStop(ctx context.Context, req *mcp.CallToolRequest, input ProcessIDInput) (*mcp.CallToolResult, ProcessStopOutput, error) {
	err := s.process.Stop(input.ID)
	if err != nil {
		return nil, ProcessStopOutput{}, fmt.Errorf("failed to stop process: %w", err)
	}
	return nil, ProcessStopOutput{ID: input.ID, Success: true}, nil
}

func (s *Service) processKill(ctx context.Context, req *mcp.CallToolRequest, input ProcessIDInput) (*mcp.CallToolResult, ProcessStopOutput, error) {
	err := s.process.Kill(input.ID)
	if err != nil {
		return nil, ProcessStopOutput{}, fmt.Errorf("failed to kill process: %w", err)
	}
	return nil, ProcessStopOutput{ID: input.ID, Success: true}, nil
}

func (s *Service) processList(ctx context.Context, req *mcp.CallToolRequest, input ProcessListInput) (*mcp.CallToolResult, ProcessListOutput, error) {
	procs := s.process.List()
	result := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		info := p.Info()
		result = append(result, ProcessInfo{
			ID:        info.ID,
			Command:   info.Command,
			Args:      info.Args,
			Dir:       info.Dir,
			Status:    string(info.Status),
			ExitCode:  info.ExitCode,
			PID:       info.PID,
			StartedAt: info.StartedAt,
		})
	}
	return nil, ProcessListOutput{Processes: result}, nil
}

func (s *Service) processOutput(ctx context.Context, req *mcp.CallToolRequest, input ProcessIDInput) (*mcp.CallToolResult, ProcessOutputOutput, error) {
	output, err := s.process.Output(input.ID)
	if err != nil {
		return nil, ProcessOutputOutput{}, fmt.Errorf("failed to get process output: %w", err)
	}
	return nil, ProcessOutputOutput{
		ID:     input.ID,
		Output: output,
		Length: len(output),
	}, nil
}

func (s *Service) processSendInput(ctx context.Context, req *mcp.CallToolRequest, input ProcessSendInputInput) (*mcp.CallToolResult, ProcessSendInputOutput, error) {
	err := s.process.SendInput(input.ID, input.Input)
	if err != nil {
		return nil, ProcessSendInputOutput{}, fmt.Errorf("failed to send input: %w", err)
	}
	return nil, ProcessSendInputOutput{ID: input.ID, Success: true}, nil
}

// WebSocket types

// WsStartInput contains parameters for starting the WebSocket server.
type WsStartInput struct {
	// Port to run WebSocket server on. Defaults to 9876.
	Port int `json:"port,omitempty"`
}

// WsStartOutput contains the result of starting the WebSocket server.
type WsStartOutput struct {
	Port    int    `json:"port"`
	URL     string `json:"url"`
	Started bool   `json:"started"`
}

// WsInfoInput is empty but required for handler signature.
type WsInfoInput struct{}

// WsInfoOutput contains WebSocket server status.
type WsInfoOutput struct {
	Running  bool   `json:"running"`
	Port     int    `json:"port"`
	URL      string `json:"url"`
	Clients  int    `json:"clients"`
	Channels int    `json:"channels"`
}

// WebSocket handlers

func (s *Service) wsStart(ctx context.Context, req *mcp.CallToolRequest, input WsStartInput) (*mcp.CallToolResult, WsStartOutput, error) {
	if s.wsHub == nil {
		return nil, WsStartOutput{}, fmt.Errorf("WebSocket not available in this configuration")
	}

	// Already running?
	if s.wsRunning {
		url := fmt.Sprintf("ws://localhost:%d/ws", s.wsPort)
		return nil, WsStartOutput{
			Port:    s.wsPort,
			URL:     url,
			Started: true,
		}, nil
	}

	port := input.Port
	if port == 0 {
		port = s.wsPort
	}
	if port == 0 {
		port = 9876
	}

	// Start the hub event loop
	hubCtx := context.Background()
	go s.wsHub.Run(hubCtx)

	// Start HTTP server for WebSocket
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/ws", s.wsHub.HandleWebSocket)
		addr := fmt.Sprintf(":%d", port)
		http.ListenAndServe(addr, mux)
	}()

	s.wsPort = port
	s.wsRunning = true
	url := fmt.Sprintf("ws://localhost:%d/ws", port)

	return nil, WsStartOutput{
		Port:    port,
		URL:     url,
		Started: true,
	}, nil
}

func (s *Service) wsInfo(ctx context.Context, req *mcp.CallToolRequest, input WsInfoInput) (*mcp.CallToolResult, WsInfoOutput, error) {
	if s.wsHub == nil {
		return nil, WsInfoOutput{Running: false}, nil
	}

	stats := s.wsHub.Stats()
	url := ""
	if s.wsRunning {
		url = fmt.Sprintf("ws://localhost:%d/ws", s.wsPort)
	}

	return nil, WsInfoOutput{
		Running:  s.wsRunning,
		Port:     s.wsPort,
		URL:      url,
		Clients:  stats.Clients,
		Channels: stats.Channels,
	}, nil
}

// WebView types

// WebviewListInput is empty.
type WebviewListInput struct{}

// WebviewListOutput contains the list of windows.
type WebviewListOutput struct {
	Windows []WebviewWindowInfo `json:"windows"`
}

// WebviewWindowInfo contains window information.
type WebviewWindowInfo struct {
	Name string `json:"name"`
}

// WebviewEvalInput contains parameters for JS evaluation.
type WebviewEvalInput struct {
	// Window name (empty for first window).
	Window string `json:"window,omitempty"`
	// JavaScript code to execute.
	Code string `json:"code"`
}

// WebviewEvalOutput contains the evaluation result.
type WebviewEvalOutput struct {
	Result string `json:"result"`
}

// WebviewConsoleInput contains parameters for console retrieval.
type WebviewConsoleInput struct {
	// Filter by level: log, warn, error, info, debug (empty for all).
	Level string `json:"level,omitempty"`
	// Maximum messages to return.
	Limit int `json:"limit,omitempty"`
	// Clear buffer after reading.
	Clear bool `json:"clear,omitempty"`
}

// WebviewConsoleOutput contains console messages.
type WebviewConsoleOutput struct {
	Messages []webview.ConsoleMessage `json:"messages"`
	Count    int                      `json:"count"`
}

// WebviewClickInput contains parameters for clicking.
type WebviewClickInput struct {
	// Window name (empty for first window).
	Window string `json:"window,omitempty"`
	// CSS selector for the element to click.
	Selector string `json:"selector"`
}

// WebviewClickOutput contains the click result.
type WebviewClickOutput struct {
	Success bool `json:"success"`
}

// WebviewTypeInput contains parameters for typing.
type WebviewTypeInput struct {
	// Window name (empty for first window).
	Window string `json:"window,omitempty"`
	// CSS selector for the input element.
	Selector string `json:"selector"`
	// Text to type.
	Text string `json:"text"`
}

// WebviewTypeOutput contains the type result.
type WebviewTypeOutput struct {
	Success bool `json:"success"`
}

// WebviewQueryInput contains parameters for querying.
type WebviewQueryInput struct {
	// Window name (empty for first window).
	Window string `json:"window,omitempty"`
	// CSS selector to query.
	Selector string `json:"selector"`
}

// WebviewQueryOutput contains query results.
type WebviewQueryOutput struct {
	Elements []map[string]any `json:"elements"`
	Count    int              `json:"count"`
}

// WebviewNavigateInput contains parameters for navigation.
type WebviewNavigateInput struct {
	// Window name (empty for first window).
	Window string `json:"window,omitempty"`
	// URL or route to navigate to.
	URL string `json:"url"`
}

// WebviewNavigateOutput contains navigation result.
type WebviewNavigateOutput struct {
	Success bool `json:"success"`
}

// WebviewSourceInput contains parameters for getting source.
type WebviewSourceInput struct {
	// Window name (empty for first window).
	Window string `json:"window,omitempty"`
}

// WebviewSourceOutput contains the page source.
type WebviewSourceOutput struct {
	HTML   string `json:"html"`
	Length int    `json:"length"`
}

// WebView handlers

func (s *Service) webviewList(ctx context.Context, req *mcp.CallToolRequest, input WebviewListInput) (*mcp.CallToolResult, WebviewListOutput, error) {
	if s.webview == nil {
		return nil, WebviewListOutput{}, fmt.Errorf("WebView not available (MCP server running standalone)")
	}

	windows := s.webview.ListWindows()
	result := make([]WebviewWindowInfo, len(windows))
	for i, w := range windows {
		result[i] = WebviewWindowInfo{Name: w.Name}
	}

	return nil, WebviewListOutput{Windows: result}, nil
}

func (s *Service) webviewEval(ctx context.Context, req *mcp.CallToolRequest, input WebviewEvalInput) (*mcp.CallToolResult, WebviewEvalOutput, error) {
	if s.webview == nil {
		return nil, WebviewEvalOutput{}, fmt.Errorf("WebView not available (MCP server running standalone)")
	}

	result, err := s.webview.ExecJS(input.Window, input.Code)
	if err != nil {
		return nil, WebviewEvalOutput{}, fmt.Errorf("failed to execute JS: %w", err)
	}

	return nil, WebviewEvalOutput{Result: result}, nil
}

func (s *Service) webviewConsole(ctx context.Context, req *mcp.CallToolRequest, input WebviewConsoleInput) (*mcp.CallToolResult, WebviewConsoleOutput, error) {
	if s.webview == nil {
		return nil, WebviewConsoleOutput{}, fmt.Errorf("WebView not available (MCP server running standalone)")
	}

	messages := s.webview.GetConsoleMessages(input.Level, input.Limit)

	if input.Clear {
		s.webview.ClearConsole()
	}

	return nil, WebviewConsoleOutput{
		Messages: messages,
		Count:    len(messages),
	}, nil
}

func (s *Service) webviewClick(ctx context.Context, req *mcp.CallToolRequest, input WebviewClickInput) (*mcp.CallToolResult, WebviewClickOutput, error) {
	if s.webview == nil {
		return nil, WebviewClickOutput{}, fmt.Errorf("WebView not available (MCP server running standalone)")
	}

	err := s.webview.Click(input.Window, input.Selector)
	if err != nil {
		return nil, WebviewClickOutput{}, fmt.Errorf("failed to click: %w", err)
	}

	return nil, WebviewClickOutput{Success: true}, nil
}

func (s *Service) webviewType(ctx context.Context, req *mcp.CallToolRequest, input WebviewTypeInput) (*mcp.CallToolResult, WebviewTypeOutput, error) {
	if s.webview == nil {
		return nil, WebviewTypeOutput{}, fmt.Errorf("WebView not available (MCP server running standalone)")
	}

	err := s.webview.Type(input.Window, input.Selector, input.Text)
	if err != nil {
		return nil, WebviewTypeOutput{}, fmt.Errorf("failed to type: %w", err)
	}

	return nil, WebviewTypeOutput{Success: true}, nil
}

func (s *Service) webviewQuery(ctx context.Context, req *mcp.CallToolRequest, input WebviewQueryInput) (*mcp.CallToolResult, WebviewQueryOutput, error) {
	if s.webview == nil {
		return nil, WebviewQueryOutput{}, fmt.Errorf("WebView not available (MCP server running standalone)")
	}

	result, err := s.webview.QuerySelector(input.Window, input.Selector)
	if err != nil {
		return nil, WebviewQueryOutput{}, fmt.Errorf("failed to query: %w", err)
	}

	// Parse result as JSON array
	var elements []map[string]any
	if err := json.Unmarshal([]byte(result), &elements); err != nil {
		// Return raw result if not valid JSON
		return nil, WebviewQueryOutput{
			Elements: []map[string]any{{"raw": result}},
			Count:    1,
		}, nil
	}

	return nil, WebviewQueryOutput{
		Elements: elements,
		Count:    len(elements),
	}, nil
}

func (s *Service) webviewNavigate(ctx context.Context, req *mcp.CallToolRequest, input WebviewNavigateInput) (*mcp.CallToolResult, WebviewNavigateOutput, error) {
	if s.webview == nil {
		return nil, WebviewNavigateOutput{}, fmt.Errorf("WebView not available (MCP server running standalone)")
	}

	err := s.webview.Navigate(input.Window, input.URL)
	if err != nil {
		return nil, WebviewNavigateOutput{}, fmt.Errorf("failed to navigate: %w", err)
	}

	return nil, WebviewNavigateOutput{Success: true}, nil
}

func (s *Service) webviewSource(ctx context.Context, req *mcp.CallToolRequest, input WebviewSourceInput) (*mcp.CallToolResult, WebviewSourceOutput, error) {
	if s.webview == nil {
		return nil, WebviewSourceOutput{}, fmt.Errorf("WebView not available (MCP server running standalone)")
	}

	html, err := s.webview.GetPageSource(input.Window)
	if err != nil {
		return nil, WebviewSourceOutput{}, fmt.Errorf("failed to get source: %w", err)
	}

	return nil, WebviewSourceOutput{
		HTML:   html,
		Length: len(html),
	}, nil
}

// Window/Display management types

// WindowListInput is empty.
type WindowListInput struct{}

// WindowListOutput contains the list of windows with positions.
type WindowListOutput struct {
	Windows []WindowInfo `json:"windows"`
}

// WindowInfo contains detailed window information.
type WindowInfo struct {
	Name      string `json:"name"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Maximized bool   `json:"maximized"`
}

// WindowGetInput contains the window name to get.
type WindowGetInput struct {
	// Window name to get info for.
	Name string `json:"name"`
}

// WindowGetOutput contains the window information.
type WindowGetOutput struct {
	Window *WindowInfo `json:"window"`
}

// WindowPositionInput contains parameters for moving a window.
type WindowPositionInput struct {
	// Window name to move.
	Name string `json:"name"`
	// X coordinate (pixels from left edge of screen).
	X int `json:"x"`
	// Y coordinate (pixels from top edge of screen).
	Y int `json:"y"`
}

// WindowPositionOutput contains the result of moving a window.
type WindowPositionOutput struct {
	Success bool   `json:"success"`
	Name    string `json:"name"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
}

// WindowSizeInput contains parameters for resizing a window.
type WindowSizeInput struct {
	// Window name to resize.
	Name string `json:"name"`
	// Width in pixels.
	Width int `json:"width"`
	// Height in pixels.
	Height int `json:"height"`
}

// WindowSizeOutput contains the result of resizing a window.
type WindowSizeOutput struct {
	Success bool   `json:"success"`
	Name    string `json:"name"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

// WindowBoundsInput contains parameters for setting window bounds.
type WindowBoundsInput struct {
	// Window name to modify.
	Name string `json:"name"`
	// X coordinate.
	X int `json:"x"`
	// Y coordinate.
	Y int `json:"y"`
	// Width in pixels.
	Width int `json:"width"`
	// Height in pixels.
	Height int `json:"height"`
}

// WindowBoundsOutput contains the result of setting window bounds.
type WindowBoundsOutput struct {
	Success bool   `json:"success"`
	Name    string `json:"name"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

// WindowNameInput contains just a window name.
type WindowNameInput struct {
	// Window name to operate on.
	Name string `json:"name"`
}

// WindowActionOutput contains the result of a window action.
type WindowActionOutput struct {
	Success bool   `json:"success"`
	Name    string `json:"name"`
	Action  string `json:"action"`
}

// ScreenListInput is empty.
type ScreenListInput struct{}

// ScreenListOutput contains the list of screens.
type ScreenListOutput struct {
	Screens []ScreenInfo `json:"screens"`
}

// ScreenInfo contains screen/monitor information.
type ScreenInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Primary bool   `json:"primary"`
}

// Window/Display handlers

func (s *Service) windowList(ctx context.Context, req *mcp.CallToolRequest, input WindowListInput) (*mcp.CallToolResult, WindowListOutput, error) {
	if s.display == nil {
		return nil, WindowListOutput{}, fmt.Errorf("display service not available (MCP server running standalone)")
	}

	windows := s.display.ListWindowInfos()
	result := make([]WindowInfo, len(windows))
	for i, w := range windows {
		result[i] = WindowInfo{
			Name:      w.Name,
			X:         w.X,
			Y:         w.Y,
			Width:     w.Width,
			Height:    w.Height,
			Maximized: w.Maximized,
		}
	}

	return nil, WindowListOutput{Windows: result}, nil
}

func (s *Service) windowGet(ctx context.Context, req *mcp.CallToolRequest, input WindowGetInput) (*mcp.CallToolResult, WindowGetOutput, error) {
	if s.display == nil {
		return nil, WindowGetOutput{}, fmt.Errorf("display service not available (MCP server running standalone)")
	}

	info, err := s.display.GetWindowInfo(input.Name)
	if err != nil {
		return nil, WindowGetOutput{}, fmt.Errorf("failed to get window info: %w", err)
	}

	return nil, WindowGetOutput{
		Window: &WindowInfo{
			Name:      info.Name,
			X:         info.X,
			Y:         info.Y,
			Width:     info.Width,
			Height:    info.Height,
			Maximized: info.Maximized,
		},
	}, nil
}

func (s *Service) windowPosition(ctx context.Context, req *mcp.CallToolRequest, input WindowPositionInput) (*mcp.CallToolResult, WindowPositionOutput, error) {
	if s.display == nil {
		return nil, WindowPositionOutput{}, fmt.Errorf("display service not available (MCP server running standalone)")
	}

	err := s.display.SetWindowPosition(input.Name, input.X, input.Y)
	if err != nil {
		return nil, WindowPositionOutput{}, fmt.Errorf("failed to move window: %w", err)
	}

	return nil, WindowPositionOutput{
		Success: true,
		Name:    input.Name,
		X:       input.X,
		Y:       input.Y,
	}, nil
}

func (s *Service) windowSize(ctx context.Context, req *mcp.CallToolRequest, input WindowSizeInput) (*mcp.CallToolResult, WindowSizeOutput, error) {
	if s.display == nil {
		return nil, WindowSizeOutput{}, fmt.Errorf("display service not available (MCP server running standalone)")
	}

	err := s.display.SetWindowSize(input.Name, input.Width, input.Height)
	if err != nil {
		return nil, WindowSizeOutput{}, fmt.Errorf("failed to resize window: %w", err)
	}

	return nil, WindowSizeOutput{
		Success: true,
		Name:    input.Name,
		Width:   input.Width,
		Height:  input.Height,
	}, nil
}

func (s *Service) windowBounds(ctx context.Context, req *mcp.CallToolRequest, input WindowBoundsInput) (*mcp.CallToolResult, WindowBoundsOutput, error) {
	if s.display == nil {
		return nil, WindowBoundsOutput{}, fmt.Errorf("display service not available (MCP server running standalone)")
	}

	err := s.display.SetWindowBounds(input.Name, input.X, input.Y, input.Width, input.Height)
	if err != nil {
		return nil, WindowBoundsOutput{}, fmt.Errorf("failed to set window bounds: %w", err)
	}

	return nil, WindowBoundsOutput{
		Success: true,
		Name:    input.Name,
		X:       input.X,
		Y:       input.Y,
		Width:   input.Width,
		Height:  input.Height,
	}, nil
}

func (s *Service) windowMaximize(ctx context.Context, req *mcp.CallToolRequest, input WindowNameInput) (*mcp.CallToolResult, WindowActionOutput, error) {
	if s.display == nil {
		return nil, WindowActionOutput{}, fmt.Errorf("display service not available (MCP server running standalone)")
	}

	err := s.display.MaximizeWindow(input.Name)
	if err != nil {
		return nil, WindowActionOutput{}, fmt.Errorf("failed to maximize window: %w", err)
	}

	return nil, WindowActionOutput{
		Success: true,
		Name:    input.Name,
		Action:  "maximize",
	}, nil
}

func (s *Service) windowMinimize(ctx context.Context, req *mcp.CallToolRequest, input WindowNameInput) (*mcp.CallToolResult, WindowActionOutput, error) {
	if s.display == nil {
		return nil, WindowActionOutput{}, fmt.Errorf("display service not available (MCP server running standalone)")
	}

	err := s.display.MinimizeWindow(input.Name)
	if err != nil {
		return nil, WindowActionOutput{}, fmt.Errorf("failed to minimize window: %w", err)
	}

	return nil, WindowActionOutput{
		Success: true,
		Name:    input.Name,
		Action:  "minimize",
	}, nil
}

func (s *Service) windowRestore(ctx context.Context, req *mcp.CallToolRequest, input WindowNameInput) (*mcp.CallToolResult, WindowActionOutput, error) {
	if s.display == nil {
		return nil, WindowActionOutput{}, fmt.Errorf("display service not available (MCP server running standalone)")
	}

	err := s.display.RestoreWindow(input.Name)
	if err != nil {
		return nil, WindowActionOutput{}, fmt.Errorf("failed to restore window: %w", err)
	}

	return nil, WindowActionOutput{
		Success: true,
		Name:    input.Name,
		Action:  "restore",
	}, nil
}

func (s *Service) windowFocus(ctx context.Context, req *mcp.CallToolRequest, input WindowNameInput) (*mcp.CallToolResult, WindowActionOutput, error) {
	if s.display == nil {
		return nil, WindowActionOutput{}, fmt.Errorf("display service not available (MCP server running standalone)")
	}

	err := s.display.FocusWindow(input.Name)
	if err != nil {
		return nil, WindowActionOutput{}, fmt.Errorf("failed to focus window: %w", err)
	}

	return nil, WindowActionOutput{
		Success: true,
		Name:    input.Name,
		Action:  "focus",
	}, nil
}

func (s *Service) screenList(ctx context.Context, req *mcp.CallToolRequest, input ScreenListInput) (*mcp.CallToolResult, ScreenListOutput, error) {
	if s.display == nil {
		return nil, ScreenListOutput{}, fmt.Errorf("display service not available (MCP server running standalone)")
	}

	screens := s.display.GetScreens()
	result := make([]ScreenInfo, len(screens))
	for i, sc := range screens {
		result[i] = ScreenInfo{
			ID:      sc.ID,
			Name:    sc.Name,
			X:       sc.X,
			Y:       sc.Y,
			Width:   sc.Width,
			Height:  sc.Height,
			Primary: sc.Primary,
		}
	}

	return nil, ScreenListOutput{Screens: result}, nil
}
