package ide

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Snider/Core/pkg/core"
)

// Options holds configuration for the IDE service.
type Options struct {
	// DefaultLanguage is the default language for new files.
	DefaultLanguage string
}

// Service provides IDE functionality for code editing, file management, and project operations.
type Service struct {
	*core.ServiceRuntime[Options]
	config Options
}

// FileInfo represents information about a file for the editor.
type FileInfo struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	Language string `json:"language"`
	IsNew    bool   `json:"isNew"`
}

// New creates a new IDE service instance.
func New() (*Service, error) {
	return &Service{
		config: Options{
			DefaultLanguage: "typescript",
		},
	}, nil
}

// Register creates and registers a new IDE service with the given Core instance.
func Register(c *core.Core) (any, error) {
	s, err := New()
	if err != nil {
		return nil, err
	}
	s.ServiceRuntime = core.NewServiceRuntime[Options](c, Options{})
	return s, nil
}

// ServiceName returns the canonical name for this service.
func (s *Service) ServiceName() string {
	return "github.com/Snider/Core/ide"
}

// NewFile creates a new untitled file with the specified language.
func (s *Service) NewFile(language string) FileInfo {
	if language == "" {
		language = s.config.DefaultLanguage
	}
	return FileInfo{
		Path:     "",
		Name:     "Untitled",
		Content:  "",
		Language: language,
		IsNew:    true,
	}
}

// OpenFile reads a file from disk and returns its content with language detection.
func (s *Service) OpenFile(path string) (FileInfo, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Path:     path,
		Name:     filepath.Base(path),
		Content:  string(content),
		Language: detectLanguage(path),
		IsNew:    false,
	}, nil
}

// SaveFile saves content to the specified path.
func (s *Service) SaveFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// ReadFile reads content from a file without additional metadata.
func (s *Service) ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ListDirectory returns a list of files and directories in the given path.
func (s *Service) ListDirectory(path string) ([]DirectoryEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]DirectoryEntry, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		result = append(result, DirectoryEntry{
			Name:  entry.Name(),
			Path:  filepath.Join(path, entry.Name()),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		})
	}
	return result, nil
}

// DirectoryEntry represents a file or directory in a listing.
type DirectoryEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

// DetectLanguage returns the Monaco editor language for a given file path.
func (s *Service) DetectLanguage(path string) string {
	return detectLanguage(path)
}

// detectLanguage maps file extensions to Monaco editor languages.
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".ts":
		return "typescript"
	case ".tsx":
		return "typescript"
	case ".js":
		return "javascript"
	case ".jsx":
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
	case ".dockerfile":
		return "dockerfile"
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
		// Check for Dockerfile without extension
		if strings.ToLower(filepath.Base(path)) == "dockerfile" {
			return "dockerfile"
		}
		return "plaintext"
	}
}

// GetSupportedLanguages returns a list of languages supported by the editor.
func (s *Service) GetSupportedLanguages() []LanguageInfo {
	return []LanguageInfo{
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
}

// LanguageInfo describes a supported programming language.
type LanguageInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
}

// FileExists checks if a file exists at the given path.
func (s *Service) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CreateDirectory creates a new directory at the given path.
func (s *Service) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// DeleteFile removes a file at the given path.
func (s *Service) DeleteFile(path string) error {
	return os.Remove(path)
}

// RenameFile renames/moves a file from oldPath to newPath.
func (s *Service) RenameFile(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}
