package vm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/host-uk/core/pkg/container"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// addVMTemplatesCommand adds the 'templates' command under vm.
func addVMTemplatesCommand(parent *cobra.Command) {
	templatesCmd := &cobra.Command{
		Use:   "templates",
		Short: i18n.T("cmd.vm.templates.short"),
		Long:  i18n.T("cmd.vm.templates.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTemplates()
		},
	}

	// Add subcommands
	addTemplatesShowCommand(templatesCmd)
	addTemplatesVarsCommand(templatesCmd)

	parent.AddCommand(templatesCmd)
}

// addTemplatesShowCommand adds the 'templates show' subcommand.
func addTemplatesShowCommand(parent *cobra.Command) {
	showCmd := &cobra.Command{
		Use:   "show <template-name>",
		Short: i18n.T("cmd.vm.templates.show.short"),
		Long:  i18n.T("cmd.vm.templates.show.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New(i18n.T("cmd.vm.error.template_required"))
			}
			return showTemplate(args[0])
		},
	}

	parent.AddCommand(showCmd)
}

// addTemplatesVarsCommand adds the 'templates vars' subcommand.
func addTemplatesVarsCommand(parent *cobra.Command) {
	varsCmd := &cobra.Command{
		Use:   "vars <template-name>",
		Short: i18n.T("cmd.vm.templates.vars.short"),
		Long:  i18n.T("cmd.vm.templates.vars.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New(i18n.T("cmd.vm.error.template_required"))
			}
			return showTemplateVars(args[0])
		},
	}

	parent.AddCommand(varsCmd)
}

func listTemplates() error {
	templates := container.ListTemplates()

	if len(templates) == 0 {
		fmt.Println(i18n.T("cmd.vm.templates.no_templates"))
		return nil
	}

	fmt.Printf("%s\n\n", repoNameStyle.Render(i18n.T("cmd.vm.templates.title")))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, i18n.T("cmd.vm.templates.header"))
	_, _ = fmt.Fprintln(w, "----\t-----------")

	for _, tmpl := range templates {
		desc := tmpl.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\n", repoNameStyle.Render(tmpl.Name), desc)
	}
	_ = w.Flush()

	fmt.Println()
	fmt.Printf("%s %s\n", i18n.T("cmd.vm.templates.hint.show"), dimStyle.Render("core vm templates show <name>"))
	fmt.Printf("%s %s\n", i18n.T("cmd.vm.templates.hint.vars"), dimStyle.Render("core vm templates vars <name>"))
	fmt.Printf("%s %s\n", i18n.T("cmd.vm.templates.hint.run"), dimStyle.Render("core vm run --template <name> --var SSH_KEY=\"...\""))

	return nil
}

func showTemplate(name string) error {
	content, err := container.GetTemplate(name)
	if err != nil {
		return err
	}

	fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("common.label.template")), repoNameStyle.Render(name))
	fmt.Println(content)

	return nil
}

func showTemplateVars(name string) error {
	content, err := container.GetTemplate(name)
	if err != nil {
		return err
	}

	required, optional := container.ExtractVariables(content)

	fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("common.label.template")), repoNameStyle.Render(name))

	if len(required) > 0 {
		fmt.Printf("%s\n", errorStyle.Render(i18n.T("cmd.vm.templates.vars.required")))
		for _, v := range required {
			fmt.Printf("  %s\n", varStyle.Render("${"+v+"}"))
		}
		fmt.Println()
	}

	if len(optional) > 0 {
		fmt.Printf("%s\n", successStyle.Render(i18n.T("cmd.vm.templates.vars.optional")))
		for v, def := range optional {
			fmt.Printf("  %s = %s\n",
				varStyle.Render("${"+v+"}"),
				defaultStyle.Render(def))
		}
		fmt.Println()
	}

	if len(required) == 0 && len(optional) == 0 {
		fmt.Println(i18n.T("cmd.vm.templates.vars.none"))
	}

	return nil
}

// RunFromTemplate builds and runs a LinuxKit image from a template.
func RunFromTemplate(templateName string, vars map[string]string, runOpts container.RunOptions) error {
	// Apply template with variables
	content, err := container.ApplyTemplate(templateName, vars)
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "apply template"})+": %w", err)
	}

	// Create a temporary directory for the build
	tmpDir, err := os.MkdirTemp("", "core-linuxkit-*")
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "create temp directory"})+": %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Write the YAML file
	yamlPath := filepath.Join(tmpDir, templateName+".yml")
	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "write template"})+": %w", err)
	}

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.template")), repoNameStyle.Render(templateName))
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.vm.label.building")), yamlPath)

	// Build the image using linuxkit
	outputPath := filepath.Join(tmpDir, templateName)
	if err := buildLinuxKitImage(yamlPath, outputPath); err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "build image"})+": %w", err)
	}

	// Find the built image (linuxkit creates .iso or other format)
	imagePath := findBuiltImage(outputPath)
	if imagePath == "" {
		return errors.New(i18n.T("cmd.vm.error.no_image_found"))
	}

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.image")), imagePath)
	fmt.Println()

	// Run the image
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "initialize container manager"})+": %w", err)
	}

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.vm.label.hypervisor")), manager.Hypervisor().Name())
	fmt.Println()

	ctx := context.Background()
	c, err := manager.Run(ctx, imagePath, runOpts)
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "run container"})+": %w", err)
	}

	if runOpts.Detach {
		fmt.Printf("%s %s\n", successStyle.Render(i18n.T("common.label.started")), c.ID)
		fmt.Printf("%s %d\n", dimStyle.Render(i18n.T("cmd.vm.label.pid")), c.PID)
		fmt.Println()
		fmt.Println(i18n.T("cmd.vm.hint.view_logs", map[string]interface{}{"ID": c.ID[:8]}))
		fmt.Println(i18n.T("cmd.vm.hint.stop", map[string]interface{}{"ID": c.ID[:8]}))
	} else {
		fmt.Printf("\n%s %s\n", dimStyle.Render(i18n.T("cmd.vm.label.container_stopped")), c.ID)
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
	// linuxkit build --format iso-bios --name <output> <yaml>
	cmd := exec.Command(lkPath, "build",
		"--format", "iso-bios",
		"--name", outputPath,
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

	return "", errors.New(i18n.T("cmd.vm.error.linuxkit_not_found"))
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
