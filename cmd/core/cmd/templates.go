package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/pkg/container"
	"github.com/leaanthony/clir"
)

var (
	varStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b"))
	defaultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).Italic(true)
)

// AddTemplatesCommand adds the 'templates' command and subcommands.
func AddTemplatesCommand(parent *clir.Cli) {
	templatesCmd := parent.NewSubCommand("templates", "Manage LinuxKit templates")
	templatesCmd.LongDescription("Manage LinuxKit YAML templates for building VMs.\n\n" +
		"Templates provide pre-configured LinuxKit configurations for common use cases.\n" +
		"They support variable substitution with ${VAR} and ${VAR:-default} syntax.\n\n" +
		"Examples:\n" +
		"  core templates                    # List available templates\n" +
		"  core templates show core-dev      # Show template content\n" +
		"  core templates vars server-php    # Show template variables")

	// Default action: list templates
	templatesCmd.Action(func() error {
		return listTemplates()
	})

	// Add subcommands
	addTemplatesShowCommand(templatesCmd)
	addTemplatesVarsCommand(templatesCmd)
}

// addTemplatesShowCommand adds the 'templates show' subcommand.
func addTemplatesShowCommand(parent *clir.Command) {
	showCmd := parent.NewSubCommand("show", "Display template content")
	showCmd.LongDescription("Display the content of a LinuxKit template.\n\n" +
		"Examples:\n" +
		"  core templates show core-dev\n" +
		"  core templates show server-php")

	showCmd.Action(func() error {
		args := showCmd.OtherArgs()
		if len(args) == 0 {
			return fmt.Errorf("template name is required")
		}
		return showTemplate(args[0])
	})
}

// addTemplatesVarsCommand adds the 'templates vars' subcommand.
func addTemplatesVarsCommand(parent *clir.Command) {
	varsCmd := parent.NewSubCommand("vars", "Show template variables")
	varsCmd.LongDescription("Display all variables used in a template.\n\n" +
		"Shows required variables (no default) and optional variables (with defaults).\n\n" +
		"Examples:\n" +
		"  core templates vars core-dev\n" +
		"  core templates vars server-php")

	varsCmd.Action(func() error {
		args := varsCmd.OtherArgs()
		if len(args) == 0 {
			return fmt.Errorf("template name is required")
		}
		return showTemplateVars(args[0])
	})
}

func listTemplates() error {
	templates := container.ListTemplates()

	if len(templates) == 0 {
		fmt.Println("No templates available.")
		return nil
	}

	fmt.Printf("%s\n\n", repoNameStyle.Render("Available LinuxKit Templates"))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-----------")

	for _, tmpl := range templates {
		desc := tmpl.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\n", repoNameStyle.Render(tmpl.Name), desc)
	}
	w.Flush()

	fmt.Println()
	fmt.Printf("Show template: %s\n", dimStyle.Render("core templates show <name>"))
	fmt.Printf("Show variables: %s\n", dimStyle.Render("core templates vars <name>"))
	fmt.Printf("Run from template: %s\n", dimStyle.Render("core run --template <name> --var SSH_KEY=\"...\""))

	return nil
}

func showTemplate(name string) error {
	content, err := container.GetTemplate(name)
	if err != nil {
		return err
	}

	fmt.Printf("%s %s\n\n", dimStyle.Render("Template:"), repoNameStyle.Render(name))
	fmt.Println(content)

	return nil
}

func showTemplateVars(name string) error {
	content, err := container.GetTemplate(name)
	if err != nil {
		return err
	}

	required, optional := container.ExtractVariables(content)

	fmt.Printf("%s %s\n\n", dimStyle.Render("Template:"), repoNameStyle.Render(name))

	if len(required) > 0 {
		fmt.Printf("%s\n", errorStyle.Render("Required Variables (no default):"))
		for _, v := range required {
			fmt.Printf("  %s\n", varStyle.Render("${"+v+"}"))
		}
		fmt.Println()
	}

	if len(optional) > 0 {
		fmt.Printf("%s\n", successStyle.Render("Optional Variables (with defaults):"))
		for v, def := range optional {
			fmt.Printf("  %s = %s\n",
				varStyle.Render("${"+v+"}"),
				defaultStyle.Render(def))
		}
		fmt.Println()
	}

	if len(required) == 0 && len(optional) == 0 {
		fmt.Println("No variables in this template.")
	}

	return nil
}

// RunFromTemplate builds and runs a LinuxKit image from a template.
func RunFromTemplate(templateName string, vars map[string]string, runOpts container.RunOptions) error {
	// Apply template with variables
	content, err := container.ApplyTemplate(templateName, vars)
	if err != nil {
		return fmt.Errorf("failed to apply template: %w", err)
	}

	// Create a temporary directory for the build
	tmpDir, err := os.MkdirTemp("", "core-linuxkit-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write the YAML file
	yamlPath := filepath.Join(tmpDir, templateName+".yml")
	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Template:"), repoNameStyle.Render(templateName))
	fmt.Printf("%s %s\n", dimStyle.Render("Building:"), yamlPath)

	// Build the image using linuxkit
	outputPath := filepath.Join(tmpDir, templateName)
	if err := buildLinuxKitImage(yamlPath, outputPath); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	// Find the built image (linuxkit creates .iso or other format)
	imagePath := findBuiltImage(outputPath)
	if imagePath == "" {
		return fmt.Errorf("no image found after build")
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Image:"), imagePath)
	fmt.Println()

	// Run the image
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf("failed to initialize container manager: %w", err)
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Hypervisor:"), manager.Hypervisor().Name())
	fmt.Println()

	ctx := context.Background()
	c, err := manager.Run(ctx, imagePath, runOpts)
	if err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	if runOpts.Detach {
		fmt.Printf("%s %s\n", successStyle.Render("Started:"), c.ID)
		fmt.Printf("%s %d\n", dimStyle.Render("PID:"), c.PID)
		fmt.Println()
		fmt.Printf("Use 'core logs %s' to view output\n", c.ID[:8])
		fmt.Printf("Use 'core stop %s' to stop\n", c.ID[:8])
	} else {
		fmt.Printf("\n%s %s\n", dimStyle.Render("Container stopped:"), c.ID)
	}

	return nil
}

// buildLinuxKitImage builds a LinuxKit image from a YAML file.
func buildLinuxKitImage(yamlPath, outputPath string) error {
	// Check if linuxkit is available
	lkPath, err := lookupLinuxKit()
	if err != nil {
		return err
	}

	// Build the image
	// linuxkit build -format iso-bios -name <output> <yaml>
	cmd := exec.Command(lkPath, "build",
		"-format", "iso-bios",
		"-name", outputPath,
		yamlPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// findBuiltImage finds the built image file.
func findBuiltImage(basePath string) string {
	// LinuxKit can create different formats
	extensions := []string{".iso", "-bios.iso", ".qcow2", ".raw", ".vmdk"}

	for _, ext := range extensions {
		path := basePath + ext
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check directory for any image file
	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, base) {
			for _, ext := range []string{".iso", ".qcow2", ".raw", ".vmdk"} {
				if strings.HasSuffix(name, ext) {
					return filepath.Join(dir, name)
				}
			}
		}
	}

	return ""
}

// lookupLinuxKit finds the linuxkit binary.
func lookupLinuxKit() (string, error) {
	// Check PATH first
	if path, err := exec.LookPath("linuxkit"); err == nil {
		return path, nil
	}

	// Check common locations
	paths := []string{
		"/usr/local/bin/linuxkit",
		"/opt/homebrew/bin/linuxkit",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("linuxkit not found. Install with: brew install linuxkit (macOS) or see https://github.com/linuxkit/linuxkit")
}

// ParseVarFlags parses --var flags into a map.
// Format: --var KEY=VALUE or --var KEY="VALUE"
func ParseVarFlags(varFlags []string) map[string]string {
	vars := make(map[string]string)

	for _, v := range varFlags {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove surrounding quotes if present
			value = strings.Trim(value, "\"'")
			vars[key] = value
		}
	}

	return vars
}
