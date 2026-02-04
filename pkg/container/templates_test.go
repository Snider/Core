package container

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTemplates_Good(t *testing.T) {
	templates := ListTemplates()

	// Should have at least the builtin templates
	assert.GreaterOrEqual(t, len(templates), 2)

	// Find the core-dev template
	var found bool
	for _, tmpl := range templates {
		if tmpl.Name == "core-dev" {
			found = true
			assert.NotEmpty(t, tmpl.Description)
			assert.NotEmpty(t, tmpl.Path)
			break
		}
	}
	assert.True(t, found, "core-dev template should exist")

	// Find the server-php template
	found = false
	for _, tmpl := range templates {
		if tmpl.Name == "server-php" {
			found = true
			assert.NotEmpty(t, tmpl.Description)
			assert.NotEmpty(t, tmpl.Path)
			break
		}
	}
	assert.True(t, found, "server-php template should exist")
}

func TestGetTemplate_Good_CoreDev(t *testing.T) {
	content, err := GetTemplate("core-dev")

	require.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Contains(t, content, "kernel:")
	assert.Contains(t, content, "linuxkit/kernel")
	assert.Contains(t, content, "${SSH_KEY}")
	assert.Contains(t, content, "services:")
}

func TestGetTemplate_Good_ServerPhp(t *testing.T) {
	content, err := GetTemplate("server-php")

	require.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Contains(t, content, "kernel:")
	assert.Contains(t, content, "frankenphp")
	assert.Contains(t, content, "${SSH_KEY}")
	assert.Contains(t, content, "${DOMAIN:-localhost}")
}

func TestGetTemplate_Bad_NotFound(t *testing.T) {
	_, err := GetTemplate("nonexistent-template")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

func TestApplyVariables_Good_SimpleSubstitution(t *testing.T) {
	content := "Hello ${NAME}, welcome to ${PLACE}!"
	vars := map[string]string{
		"NAME":  "World",
		"PLACE": "Core",
	}

	result, err := ApplyVariables(content, vars)

	require.NoError(t, err)
	assert.Equal(t, "Hello World, welcome to Core!", result)
}

func TestApplyVariables_Good_WithDefaults(t *testing.T) {
	content := "Memory: ${MEMORY:-1024}MB, CPUs: ${CPUS:-2}"
	vars := map[string]string{
		"MEMORY": "2048",
		// CPUS not provided, should use default
	}

	result, err := ApplyVariables(content, vars)

	require.NoError(t, err)
	assert.Equal(t, "Memory: 2048MB, CPUs: 2", result)
}

func TestApplyVariables_Good_AllDefaults(t *testing.T) {
	content := "${HOST:-localhost}:${PORT:-8080}"
	vars := map[string]string{} // No vars provided

	result, err := ApplyVariables(content, vars)

	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", result)
}

func TestApplyVariables_Good_MixedSyntax(t *testing.T) {
	content := `
hostname: ${HOSTNAME:-myhost}
ssh_key: ${SSH_KEY}
memory: ${MEMORY:-512}
`
	vars := map[string]string{
		"SSH_KEY":  "ssh-rsa AAAA...",
		"HOSTNAME": "custom-host",
	}

	result, err := ApplyVariables(content, vars)

	require.NoError(t, err)
	assert.Contains(t, result, "hostname: custom-host")
	assert.Contains(t, result, "ssh_key: ssh-rsa AAAA...")
	assert.Contains(t, result, "memory: 512")
}

func TestApplyVariables_Good_EmptyDefault(t *testing.T) {
	content := "value: ${OPT:-}"
	vars := map[string]string{}

	result, err := ApplyVariables(content, vars)

	require.NoError(t, err)
	assert.Equal(t, "value: ", result)
}

func TestApplyVariables_Bad_MissingRequired(t *testing.T) {
	content := "SSH Key: ${SSH_KEY}"
	vars := map[string]string{} // Missing required SSH_KEY

	_, err := ApplyVariables(content, vars)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required variables")
	assert.Contains(t, err.Error(), "SSH_KEY")
}

func TestApplyVariables_Bad_MultipleMissing(t *testing.T) {
	content := "${VAR1} and ${VAR2} and ${VAR3}"
	vars := map[string]string{
		"VAR2": "provided",
	}

	_, err := ApplyVariables(content, vars)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required variables")
	// Should mention both missing vars
	errStr := err.Error()
	assert.True(t, strings.Contains(errStr, "VAR1") || strings.Contains(errStr, "VAR3"))
}

func TestApplyTemplate_Good(t *testing.T) {
	vars := map[string]string{
		"SSH_KEY": "ssh-rsa AAAA... user@host",
	}

	result, err := ApplyTemplate("core-dev", vars)

	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "ssh-rsa AAAA... user@host")
	// Default values should be applied
	assert.Contains(t, result, "core-dev") // HOSTNAME default
}

func TestApplyTemplate_Bad_TemplateNotFound(t *testing.T) {
	vars := map[string]string{
		"SSH_KEY": "test",
	}

	_, err := ApplyTemplate("nonexistent", vars)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

func TestApplyTemplate_Bad_MissingVariable(t *testing.T) {
	// server-php requires SSH_KEY
	vars := map[string]string{} // Missing required SSH_KEY

	_, err := ApplyTemplate("server-php", vars)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required variables")
}

func TestExtractVariables_Good(t *testing.T) {
	content := `
hostname: ${HOSTNAME:-myhost}
ssh_key: ${SSH_KEY}
memory: ${MEMORY:-1024}
cpus: ${CPUS:-2}
api_key: ${API_KEY}
`
	required, optional := ExtractVariables(content)

	// Required variables (no default)
	assert.Contains(t, required, "SSH_KEY")
	assert.Contains(t, required, "API_KEY")
	assert.Len(t, required, 2)

	// Optional variables (with defaults)
	assert.Equal(t, "myhost", optional["HOSTNAME"])
	assert.Equal(t, "1024", optional["MEMORY"])
	assert.Equal(t, "2", optional["CPUS"])
	assert.Len(t, optional, 3)
}

func TestExtractVariables_Good_NoVariables(t *testing.T) {
	content := "This has no variables at all"

	required, optional := ExtractVariables(content)

	assert.Empty(t, required)
	assert.Empty(t, optional)
}

func TestExtractVariables_Good_OnlyDefaults(t *testing.T) {
	content := "${A:-default1} ${B:-default2}"

	required, optional := ExtractVariables(content)

	assert.Empty(t, required)
	assert.Len(t, optional, 2)
	assert.Equal(t, "default1", optional["A"])
	assert.Equal(t, "default2", optional["B"])
}

func TestScanUserTemplates_Good(t *testing.T) {
	// Create a temporary directory with template files
	tmpDir := t.TempDir()

	// Create a valid template file
	templateContent := `# My Custom Template
# A custom template for testing
kernel:
  image: linuxkit/kernel:6.6
`
	err := os.WriteFile(filepath.Join(tmpDir, "custom.yml"), []byte(templateContent), 0644)
	require.NoError(t, err)

	// Create a non-template file (should be ignored)
	err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("Not a template"), 0644)
	require.NoError(t, err)

	templates := scanUserTemplates(tmpDir)

	assert.Len(t, templates, 1)
	assert.Equal(t, "custom", templates[0].Name)
	assert.Equal(t, "My Custom Template", templates[0].Description)
}

func TestScanUserTemplates_Good_MultipleTemplates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple template files
	err := os.WriteFile(filepath.Join(tmpDir, "web.yml"), []byte("# Web Server\nkernel:"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "db.yaml"), []byte("# Database Server\nkernel:"), 0644)
	require.NoError(t, err)

	templates := scanUserTemplates(tmpDir)

	assert.Len(t, templates, 2)

	// Check names are extracted correctly
	names := make(map[string]bool)
	for _, tmpl := range templates {
		names[tmpl.Name] = true
	}
	assert.True(t, names["web"])
	assert.True(t, names["db"])
}

func TestScanUserTemplates_Good_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	templates := scanUserTemplates(tmpDir)

	assert.Empty(t, templates)
}

func TestScanUserTemplates_Bad_NonexistentDirectory(t *testing.T) {
	templates := scanUserTemplates("/nonexistent/path/to/templates")

	assert.Empty(t, templates)
}

func TestExtractTemplateDescription_Good(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yml")

	content := `# My Template Description
# More details here
kernel:
  image: test
`
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	desc := extractTemplateDescription(path)

	assert.Equal(t, "My Template Description", desc)
}

func TestExtractTemplateDescription_Good_NoComments(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yml")

	content := `kernel:
  image: test
`
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	desc := extractTemplateDescription(path)

	assert.Empty(t, desc)
}

func TestExtractTemplateDescription_Bad_FileNotFound(t *testing.T) {
	desc := extractTemplateDescription("/nonexistent/file.yml")

	assert.Empty(t, desc)
}

func TestVariablePatternEdgeCases_Good(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		vars     map[string]string
		expected string
	}{
		{
			name:     "underscore in name",
			content:  "${MY_VAR:-default}",
			vars:     map[string]string{"MY_VAR": "value"},
			expected: "value",
		},
		{
			name:     "numbers in name",
			content:  "${VAR123:-default}",
			vars:     map[string]string{},
			expected: "default",
		},
		{
			name:     "default with special chars",
			content:  "${URL:-http://localhost:8080}",
			vars:     map[string]string{},
			expected: "http://localhost:8080",
		},
		{
			name:     "default with path",
			content:  "${PATH:-/usr/local/bin}",
			vars:     map[string]string{},
			expected: "/usr/local/bin",
		},
		{
			name:     "adjacent variables",
			content:  "${A:-a}${B:-b}${C:-c}",
			vars:     map[string]string{"B": "X"},
			expected: "aXc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ApplyVariables(tt.content, tt.vars)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestListTemplates_Good_WithUserTemplates(t *testing.T) {
	// Create a workspace directory with user templates
	tmpDir := t.TempDir()
	coreDir := filepath.Join(tmpDir, ".core", "linuxkit")
	err := os.MkdirAll(coreDir, 0755)
	require.NoError(t, err)

	// Create a user template
	templateContent := `# Custom user template
kernel:
  image: linuxkit/kernel:6.6
`
	err = os.WriteFile(filepath.Join(coreDir, "user-custom.yml"), []byte(templateContent), 0644)
	require.NoError(t, err)

	// Change to the temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()

	templates := ListTemplates()

	// Should have at least the builtin templates plus the user template
	assert.GreaterOrEqual(t, len(templates), 3)

	// Check that user template is included
	found := false
	for _, tmpl := range templates {
		if tmpl.Name == "user-custom" {
			found = true
			assert.Equal(t, "Custom user template", tmpl.Description)
			break
		}
	}
	assert.True(t, found, "user-custom template should exist")
}

func TestGetTemplate_Good_UserTemplate(t *testing.T) {
	// Create a workspace directory with user templates
	tmpDir := t.TempDir()
	coreDir := filepath.Join(tmpDir, ".core", "linuxkit")
	err := os.MkdirAll(coreDir, 0755)
	require.NoError(t, err)

	// Create a user template
	templateContent := `# My user template
kernel:
  image: linuxkit/kernel:6.6
services:
  - name: test
`
	err = os.WriteFile(filepath.Join(coreDir, "my-user-template.yml"), []byte(templateContent), 0644)
	require.NoError(t, err)

	// Change to the temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()

	content, err := GetTemplate("my-user-template")

	require.NoError(t, err)
	assert.Contains(t, content, "kernel:")
	assert.Contains(t, content, "My user template")
}

func TestScanUserTemplates_Good_SkipsBuiltinNames(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a template with a builtin name (should be skipped)
	err := os.WriteFile(filepath.Join(tmpDir, "core-dev.yml"), []byte("# Duplicate\nkernel:"), 0644)
	require.NoError(t, err)

	// Create a unique template
	err = os.WriteFile(filepath.Join(tmpDir, "unique.yml"), []byte("# Unique\nkernel:"), 0644)
	require.NoError(t, err)

	templates := scanUserTemplates(tmpDir)

	// Should only have the unique template, not the builtin name
	assert.Len(t, templates, 1)
	assert.Equal(t, "unique", templates[0].Name)
}

func TestScanUserTemplates_Good_SkipsDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a subdirectory (should be skipped)
	err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	require.NoError(t, err)

	// Create a valid template
	err = os.WriteFile(filepath.Join(tmpDir, "valid.yml"), []byte("# Valid\nkernel:"), 0644)
	require.NoError(t, err)

	templates := scanUserTemplates(tmpDir)

	assert.Len(t, templates, 1)
	assert.Equal(t, "valid", templates[0].Name)
}

func TestScanUserTemplates_Good_YamlExtension(t *testing.T) {
	tmpDir := t.TempDir()

	// Create templates with both extensions
	err := os.WriteFile(filepath.Join(tmpDir, "template1.yml"), []byte("# Template 1\nkernel:"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "template2.yaml"), []byte("# Template 2\nkernel:"), 0644)
	require.NoError(t, err)

	templates := scanUserTemplates(tmpDir)

	assert.Len(t, templates, 2)

	names := make(map[string]bool)
	for _, tmpl := range templates {
		names[tmpl.Name] = true
	}
	assert.True(t, names["template1"])
	assert.True(t, names["template2"])
}

func TestExtractTemplateDescription_Good_EmptyComment(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yml")

	// First comment is empty, second has content
	content := `#
# Actual description here
kernel:
  image: test
`
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	desc := extractTemplateDescription(path)

	assert.Equal(t, "Actual description here", desc)
}

func TestExtractTemplateDescription_Good_MultipleEmptyComments(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yml")

	// Multiple empty comments before actual content
	content := `#
#
#
# Real description
kernel:
  image: test
`
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	desc := extractTemplateDescription(path)

	assert.Equal(t, "Real description", desc)
}

func TestGetUserTemplatesDir_Good_NoDirectory(t *testing.T) {
	// Save current working directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)

	// Create a temp directory without .core/linuxkit
	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()

	dir := getUserTemplatesDir()

	// Should return empty string since no templates dir exists
	// (unless home dir has one)
	assert.True(t, dir == "" || strings.Contains(dir, "linuxkit"))
}

func TestScanUserTemplates_Good_DefaultDescription(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a template without comments
	content := `kernel:
  image: test
`
	err := os.WriteFile(filepath.Join(tmpDir, "nocomment.yml"), []byte(content), 0644)
	require.NoError(t, err)

	templates := scanUserTemplates(tmpDir)

	assert.Len(t, templates, 1)
	assert.Equal(t, "User-defined template", templates[0].Description)
}
