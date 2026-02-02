package help

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

var (
	// frontmatterRegex matches YAML frontmatter delimited by ---
	// Supports both LF and CRLF line endings, and empty frontmatter blocks
	frontmatterRegex = regexp.MustCompile(`(?s)^---\r?\n(.*?)(?:\r?\n)?---\r?\n?`)

	// headingRegex matches markdown headings (# to ######)
	headingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
)

// ParseTopic parses a markdown file into a Topic.
func ParseTopic(path string, content []byte) (*Topic, error) {
	contentStr := string(content)

	topic := &Topic{
		Path:     path,
		ID:       GenerateID(pathToTitle(path)),
		Sections: []Section{},
		Tags:     []string{},
		Related:  []string{},
	}

	// Extract YAML frontmatter if present
	fm, body := ExtractFrontmatter(contentStr)
	if fm != nil {
		topic.Title = fm.Title
		topic.Tags = fm.Tags
		topic.Related = fm.Related
		topic.Order = fm.Order
		if topic.Title != "" {
			topic.ID = GenerateID(topic.Title)
		}
	}

	topic.Content = body

	// Extract sections from headings
	topic.Sections = ExtractSections(body)

	// If no title from frontmatter, try first H1
	if topic.Title == "" && len(topic.Sections) > 0 {
		for _, s := range topic.Sections {
			if s.Level == 1 {
				topic.Title = s.Title
				topic.ID = GenerateID(s.Title)
				break
			}
		}
	}

	return topic, nil
}

// ExtractFrontmatter extracts YAML frontmatter from markdown content.
// Returns the parsed frontmatter and the remaining content.
func ExtractFrontmatter(content string) (*Frontmatter, string) {
	match := frontmatterRegex.FindStringSubmatch(content)
	if match == nil {
		return nil, content
	}

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(match[1]), &fm); err != nil {
		// Invalid YAML, return content as-is
		return nil, content
	}

	// Return content without frontmatter
	body := content[len(match[0]):]
	return &fm, body
}

// ExtractSections parses markdown and returns sections.
func ExtractSections(content string) []Section {
	lines := strings.Split(content, "\n")
	sections := []Section{}

	var currentSection *Section
	var contentLines []string

	for i, line := range lines {
		lineNum := i + 1 // 1-indexed

		match := headingRegex.FindStringSubmatch(line)
		if match != nil {
			// Save previous section's content
			if currentSection != nil {
				currentSection.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
			}

			// Start new section
			level := len(match[1])
			title := strings.TrimSpace(match[2])

			section := Section{
				ID:    GenerateID(title),
				Title: title,
				Level: level,
				Line:  lineNum,
			}
			sections = append(sections, section)
			currentSection = &sections[len(sections)-1]
			contentLines = []string{}
		} else if currentSection != nil {
			contentLines = append(contentLines, line)
		}
	}

	// Save last section's content
	if currentSection != nil {
		currentSection.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
	}

	return sections
}

// GenerateID creates a URL-safe ID from a title.
// "Getting Started" -> "getting-started"
func GenerateID(title string) string {
	var result strings.Builder

	for _, r := range strings.ToLower(title) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			// Only add hyphen if last char isn't already a hyphen
			str := result.String()
			if len(str) > 0 && str[len(str)-1] != '-' {
				result.WriteRune('-')
			}
		}
		// Skip other characters
	}

	// Trim trailing hyphens
	str := result.String()
	return strings.Trim(str, "-")
}

// pathToTitle converts a file path to a title.
// "getting-started.md" -> "Getting Started"
func pathToTitle(path string) string {
	// Get filename without directory (cross-platform)
	filename := filepath.Base(path)

	// Remove extension
	if ext := filepath.Ext(filename); ext != "" {
		filename = strings.TrimSuffix(filename, ext)
	}

	// Replace hyphens/underscores with spaces
	filename = strings.ReplaceAll(filename, "-", " ")
	filename = strings.ReplaceAll(filename, "_", " ")

	// Title case
	words := strings.Fields(filename)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}
