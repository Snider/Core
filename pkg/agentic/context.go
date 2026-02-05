// Package agentic provides AI collaboration features for task management.
package agentic

import (
	"bytes"
	goio "io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/host-uk/core/pkg/ai"
	"github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/log"
)

const maxContextBytes = 5000

// FileContent represents the content of a file for AI context.
type FileContent struct {
	// Path is the relative path to the file.
	Path string `json:"path"`
	// Content is the file content.
	Content string `json:"content"`
	// Language is the detected programming language.
	Language string `json:"language"`
}

// TaskContext contains gathered context for AI collaboration.
type TaskContext struct {
	// Task is the task being worked on.
	Task *Task `json:"task"`
	// Files is a list of relevant file contents.
	Files []FileContent `json:"files"`
	// GitStatus is the current git status output.
	GitStatus string `json:"git_status"`
	// RecentCommits is the recent commit log.
	RecentCommits string `json:"recent_commits"`
	// RelatedCode contains code snippets related to the task.
	RelatedCode []FileContent `json:"related_code"`
	// RAGContext contains relevant documentation from the vector database.
	RAGContext string `json:"rag_context,omitempty"`
}

// BuildTaskContext gathers context for AI collaboration on a task.
func BuildTaskContext(task *Task, dir string) (*TaskContext, error) {
	const op = "agentic.BuildTaskContext"

	if task == nil {
		return nil, log.E(op, "task is required", nil)
	}

	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, log.E(op, "failed to get working directory", err)
		}
		dir = cwd
	}

	ctx := &TaskContext{
		Task: task,
	}

	// Gather files mentioned in the task
	files, err := GatherRelatedFiles(task, dir)
	if err != nil {
		// Non-fatal: continue without files
		files = nil
	}
	ctx.Files = files

	// Get git status
	gitStatus, _ := runGitCommand(dir, "status", "--porcelain")
	ctx.GitStatus = gitStatus

	// Get recent commits
	recentCommits, _ := runGitCommand(dir, "log", "--oneline", "-10")
	ctx.RecentCommits = recentCommits

	// Find related code by searching for keywords
	relatedCode, err := findRelatedCode(task, dir)
	if err != nil {
		relatedCode = nil
	}
	ctx.RelatedCode = relatedCode

	// Query RAG for relevant documentation (graceful degradation)
	ragCtx := ai.QueryRAGForTask(ai.TaskInfo{
		Title:       task.Title,
		Description: task.Description,
	})
	ctx.RAGContext = ragCtx

	return ctx, nil
}

// GatherRelatedFiles reads files mentioned in the task.
func GatherRelatedFiles(task *Task, dir string) ([]FileContent, error) {
	const op = "agentic.GatherRelatedFiles"

	if task == nil {
		return nil, log.E(op, "task is required", nil)
	}

	var files []FileContent

	// Read files explicitly mentioned in the task
	for _, relPath := range task.Files {
		fullPath := relPath
		if !filepath.IsAbs(relPath) {
			fullPath = filepath.Join(dir, relPath)
		}

		content, truncated, err := readAndTruncate(fullPath)
		if err != nil {
			continue
		}

		contentStr := string(content)
		if truncated {
			contentStr += "\n... (truncated)"
		}

		files = append(files, FileContent{
			Path:     relPath,
			Content:  contentStr,
			Language: detectLanguage(relPath),
		})
	}

	return files, nil
}

// findRelatedCode searches for code related to the task by keywords.
func findRelatedCode(task *Task, dir string) ([]FileContent, error) {
	const op = "agentic.findRelatedCode"

	if task == nil {
		return nil, log.E(op, "task is required", nil)
	}

	// Extract keywords from title and description
	keywords := extractKeywords(task.Title + " " + task.Description)
	if len(keywords) == 0 {
		return nil, nil
	}

	var files []FileContent
	seen := make(map[string]bool)

	// Search for each keyword using git grep
	for _, keyword := range keywords {
		if len(keyword) < 3 {
			continue
		}

		output, err := runGitCommand(dir, "grep", "-l", "-i", keyword, "--", "*.go", "*.ts", "*.js", "*.py")
		if err != nil {
			continue
		}

		// Parse matched files
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || seen[line] {
				continue
			}
			seen[line] = true

			// Limit to 10 related files
			if len(files) >= 10 {
				break
			}

			fullPath := line
			if !filepath.IsAbs(line) {
				fullPath = filepath.Join(dir, line)
			}

			content, truncated, err := readAndTruncate(fullPath)
			if err != nil {
				continue
			}

			contentStr := string(content)
			if truncated {
				contentStr += "\n... (truncated)"
			}

			files = append(files, FileContent{
				Path:     line,
				Content:  contentStr,
				Language: detectLanguage(line),
			})
		}

		if len(files) >= 10 {
			break
		}
	}

	return files, nil
}

// extractKeywords extracts meaningful words from text for searching.
func extractKeywords(text string) []string {
	// Remove common words and extract identifiers
	text = strings.ToLower(text)

	// Split by non-alphanumeric characters
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	words := re.Split(text, -1)

	// Filter stop words and short words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "is": true, "are": true, "was": true,
		"be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true, "could": true,
		"should": true, "may": true, "might": true, "must": true, "shall": true,
		"this": true, "that": true, "these": true, "those": true, "it": true,
		"add": true, "create": true, "update": true, "fix": true, "remove": true,
		"implement": true, "new": true, "file": true, "code": true,
	}

	var keywords []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) >= 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	// Limit to first 5 keywords
	if len(keywords) > 5 {
		keywords = keywords[:5]
	}

	return keywords
}

// detectLanguage detects the programming language from a file extension.
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	languages := map[string]string{
		".go":    "go",
		".ts":    "typescript",
		".tsx":   "typescript",
		".js":    "javascript",
		".jsx":   "javascript",
		".py":    "python",
		".rs":    "rust",
		".java":  "java",
		".kt":    "kotlin",
		".swift": "swift",
		".c":     "c",
		".cpp":   "cpp",
		".h":     "c",
		".hpp":   "cpp",
		".rb":    "ruby",
		".php":   "php",
		".cs":    "csharp",
		".fs":    "fsharp",
		".scala": "scala",
		".sh":    "bash",
		".bash":  "bash",
		".zsh":   "zsh",
		".yaml":  "yaml",
		".yml":   "yaml",
		".json":  "json",
		".xml":   "xml",
		".html":  "html",
		".css":   "css",
		".scss":  "scss",
		".sql":   "sql",
		".md":    "markdown",
	}

	if lang, ok := languages[ext]; ok {
		return lang
	}
	return "text"
}

// readAndTruncate reads up to maxContextBytes from a file.
func readAndTruncate(path string) ([]byte, bool, error) {
	f, err := io.Local.ReadStream(path)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = f.Close() }()

	// Read up to maxContextBytes + 1 to detect truncation
	reader := goio.LimitReader(f, maxContextBytes+1)
	content, err := goio.ReadAll(reader)
	if err != nil {
		return nil, false, err
	}

	truncated := false
	if len(content) > maxContextBytes {
		content = content[:maxContextBytes]
		truncated = true
	}

	return content, truncated, nil
}

// runGitCommand runs a git command and returns the output.
func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return stdout.String(), nil
}

// FormatContext formats the TaskContext for AI consumption.
func (tc *TaskContext) FormatContext() string {
	var sb strings.Builder

	sb.WriteString("# Task Context\n\n")

	// Task info
	sb.WriteString("## Task\n")
	sb.WriteString("ID: " + tc.Task.ID + "\n")
	sb.WriteString("Title: " + tc.Task.Title + "\n")
	sb.WriteString("Priority: " + string(tc.Task.Priority) + "\n")
	sb.WriteString("Status: " + string(tc.Task.Status) + "\n")
	sb.WriteString("\n### Description\n")
	sb.WriteString(tc.Task.Description + "\n\n")

	// Files
	if len(tc.Files) > 0 {
		sb.WriteString("## Task Files\n")
		for _, f := range tc.Files {
			sb.WriteString("### " + f.Path + " (" + f.Language + ")\n")
			sb.WriteString("```" + f.Language + "\n")
			sb.WriteString(f.Content)
			sb.WriteString("\n```\n\n")
		}
	}

	// Git status
	if tc.GitStatus != "" {
		sb.WriteString("## Git Status\n")
		sb.WriteString("```\n")
		sb.WriteString(tc.GitStatus)
		sb.WriteString("\n```\n\n")
	}

	// Recent commits
	if tc.RecentCommits != "" {
		sb.WriteString("## Recent Commits\n")
		sb.WriteString("```\n")
		sb.WriteString(tc.RecentCommits)
		sb.WriteString("\n```\n\n")
	}

	// Related code
	if len(tc.RelatedCode) > 0 {
		sb.WriteString("## Related Code\n")
		for _, f := range tc.RelatedCode {
			sb.WriteString("### " + f.Path + " (" + f.Language + ")\n")
			sb.WriteString("```" + f.Language + "\n")
			sb.WriteString(f.Content)
			sb.WriteString("\n```\n\n")
		}
	}

	// Relevant documentation from RAG
	if tc.RAGContext != "" {
		sb.WriteString("## Relevant Documentation\n")
		sb.WriteString(tc.RAGContext)
		sb.WriteString("\n\n")
	}

	return sb.String()
}
