package pkgcmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/i18n"
	coreio "github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

var (
	installTargetDir string
	installAddToReg  bool
)

// addPkgInstallCommand adds the 'pkg install' command.
func addPkgInstallCommand(parent *cobra.Command) {
	installCmd := &cobra.Command{
		Use:   "install <org/repo>",
		Short: i18n.T("cmd.pkg.install.short"),
		Long:  i18n.T("cmd.pkg.install.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New(i18n.T("cmd.pkg.error.repo_required"))
			}
			return runPkgInstall(args[0], installTargetDir, installAddToReg)
		},
	}

	installCmd.Flags().StringVar(&installTargetDir, "dir", "", i18n.T("cmd.pkg.install.flag.dir"))
	installCmd.Flags().BoolVar(&installAddToReg, "add", false, i18n.T("cmd.pkg.install.flag.add"))

	parent.AddCommand(installCmd)
}

func runPkgInstall(repoArg, targetDir string, addToRegistry bool) error {
	ctx := context.Background()

	// Parse org/repo
	parts := strings.Split(repoArg, "/")
	if len(parts) != 2 {
		return errors.New(i18n.T("cmd.pkg.error.invalid_repo_format"))
	}
	org, repoName := parts[0], parts[1]

	// Determine target directory
	if targetDir == "" {
		if regPath, err := repos.FindRegistry(coreio.Local); err == nil {
			if reg, err := repos.LoadRegistry(coreio.Local, regPath); err == nil {
				targetDir = reg.BasePath
				if targetDir == "" {
					targetDir = "./packages"
				}
				if !filepath.IsAbs(targetDir) {
					targetDir = filepath.Join(filepath.Dir(regPath), targetDir)
				}
			}
		}
		if targetDir == "" {
			targetDir = "."
		}
	}

	if strings.HasPrefix(targetDir, "~/") {
		home, _ := os.UserHomeDir()
		targetDir = filepath.Join(home, targetDir[2:])
	}

	repoPath := filepath.Join(targetDir, repoName)

	if coreio.Local.Exists(filepath.Join(repoPath, ".git")) {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.Label("skip")), i18n.T("cmd.pkg.install.already_exists", map[string]string{"Name": repoName, "Path": repoPath}))
		return nil
	}

	if err := coreio.Local.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("i18n.fail.create", "directory"), err)
	}

	fmt.Printf("%s %s/%s\n", dimStyle.Render(i18n.T("cmd.pkg.install.installing_label")), org, repoName)
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.Label("target")), repoPath)
	fmt.Println()

	fmt.Printf("  %s... ", dimStyle.Render(i18n.T("common.status.cloning")))
	err := gitClone(ctx, org, repoName, repoPath)
	if err != nil {
		fmt.Printf("%s\n", errorStyle.Render("✗ "+err.Error()))
		return err
	}
	fmt.Printf("%s\n", successStyle.Render("✓"))

	if addToRegistry {
		if err := addToRegistryFile(org, repoName); err != nil {
			fmt.Printf("  %s %s: %s\n", errorStyle.Render("✗"), i18n.T("cmd.pkg.install.add_to_registry"), err)
		} else {
			fmt.Printf("  %s %s\n", successStyle.Render("✓"), i18n.T("cmd.pkg.install.added_to_registry"))
		}
	}

	fmt.Println()
	fmt.Printf("%s %s\n", successStyle.Render(i18n.T("i18n.done.install")), i18n.T("cmd.pkg.install.installed", map[string]string{"Name": repoName}))

	return nil
}

func addToRegistryFile(org, repoName string) error {
	regPath, err := repos.FindRegistry(coreio.Local)
	if err != nil {
		return errors.New(i18n.T("cmd.pkg.error.no_repos_yaml"))
	}

	reg, err := repos.LoadRegistry(coreio.Local, regPath)
	if err != nil {
		return err
	}

	if _, exists := reg.Get(repoName); exists {
		return nil
	}

	content, err := coreio.Local.Read(regPath)
	if err != nil {
		return err
	}

	repoType := detectRepoType(repoName)
	entry := fmt.Sprintf("\n  %s:\n    type: %s\n    description: (installed via core pkg install)\n",
		repoName, repoType)

	content += entry
	return coreio.Local.Write(regPath, content)
}

func detectRepoType(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "-mod-") || strings.HasSuffix(lower, "-mod") {
		return "module"
	}
	if strings.Contains(lower, "-plug-") || strings.HasSuffix(lower, "-plug") {
		return "plugin"
	}
	if strings.Contains(lower, "-services-") || strings.HasSuffix(lower, "-services") {
		return "service"
	}
	if strings.Contains(lower, "-website-") || strings.HasSuffix(lower, "-website") {
		return "website"
	}
	if strings.HasPrefix(lower, "core-") {
		return "package"
	}
	return "package"
}
