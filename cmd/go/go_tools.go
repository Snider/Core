package gocmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

var (
	installVerbose bool
	installNoCgo   bool
)

func addGoInstallCommand(parent *cobra.Command) {
	installCmd := &cobra.Command{
		Use:   "install [path]",
		Short: i18n.T("cmd.go.install.short"),
		Long:  i18n.T("cmd.go.install.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get install path from args or default to current dir
			installPath := "./..."
			if len(args) > 0 {
				installPath = args[0]
			}

			// Detect if we're in a module with cmd/ subdirectories or a root main.go
			if installPath == "./..." {
				if _, err := os.Stat("core.go"); err == nil {
					installPath = "."
				} else if entries, err := os.ReadDir("cmd"); err == nil && len(entries) > 0 {
					installPath = "./cmd/..."
				} else if _, err := os.Stat("main.go"); err == nil {
					installPath = "."
				}
			}

			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.install")), i18n.T("common.status.installing"))
			fmt.Printf("  %s %s\n", dimStyle.Render(i18n.T("common.label.path")), installPath)
			if installNoCgo {
				fmt.Printf("  %s %s\n", dimStyle.Render(i18n.T("cmd.go.install.cgo_label")), i18n.T("cmd.go.install.cgo_disabled"))
			}

			cmdArgs := []string{"install"}
			if installVerbose {
				cmdArgs = append(cmdArgs, "-v")
			}
			cmdArgs = append(cmdArgs, installPath)

			execCmd := exec.Command("go", cmdArgs...)
			if installNoCgo {
				execCmd.Env = append(os.Environ(), "CGO_ENABLED=0")
			}
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			if err := execCmd.Run(); err != nil {
				fmt.Printf("\n%s\n", errorStyle.Render(i18n.T("cmd.go.install.failed")))
				return err
			}

			// Show where it was installed
			gopath := os.Getenv("GOPATH")
			if gopath == "" {
				home, _ := os.UserHomeDir()
				gopath = filepath.Join(home, "go")
			}
			binDir := filepath.Join(gopath, "bin")

			fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("cmd.go.install.success")), i18n.T("cmd.go.install.installed_to", map[string]interface{}{"Path": binDir}))
			return nil
		},
	}

	installCmd.Flags().BoolVarP(&installVerbose, "verbose", "v", false, i18n.T("common.flag.verbose"))
	installCmd.Flags().BoolVar(&installNoCgo, "no-cgo", false, i18n.T("cmd.go.install.flag.no_cgo"))

	parent.AddCommand(installCmd)
}

func addGoModCommand(parent *cobra.Command) {
	modCmd := &cobra.Command{
		Use:   "mod",
		Short: i18n.T("cmd.go.mod.short"),
		Long:  i18n.T("cmd.go.mod.long"),
	}

	// tidy
	tidyCmd := &cobra.Command{
		Use:   "tidy",
		Short: i18n.T("cmd.go.mod.tidy.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			execCmd := exec.Command("go", "mod", "tidy")
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	// download
	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: i18n.T("cmd.go.mod.download.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			execCmd := exec.Command("go", "mod", "download")
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	// verify
	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: i18n.T("cmd.go.mod.verify.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			execCmd := exec.Command("go", "mod", "verify")
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	// graph
	graphCmd := &cobra.Command{
		Use:   "graph",
		Short: i18n.T("cmd.go.mod.graph.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			execCmd := exec.Command("go", "mod", "graph")
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	modCmd.AddCommand(tidyCmd)
	modCmd.AddCommand(downloadCmd)
	modCmd.AddCommand(verifyCmd)
	modCmd.AddCommand(graphCmd)
	parent.AddCommand(modCmd)
}

func addGoWorkCommand(parent *cobra.Command) {
	workCmd := &cobra.Command{
		Use:   "work",
		Short: i18n.T("cmd.go.work.short"),
		Long:  i18n.T("cmd.go.work.long"),
	}

	// sync
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: i18n.T("cmd.go.work.sync.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			execCmd := exec.Command("go", "work", "sync")
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	// init
	initCmd := &cobra.Command{
		Use:   "init",
		Short: i18n.T("cmd.go.work.init.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			execCmd := exec.Command("go", "work", "init")
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			if err := execCmd.Run(); err != nil {
				return err
			}
			// Auto-add current module if go.mod exists
			if _, err := os.Stat("go.mod"); err == nil {
				execCmd = exec.Command("go", "work", "use", ".")
				execCmd.Stdout = os.Stdout
				execCmd.Stderr = os.Stderr
				return execCmd.Run()
			}
			return nil
		},
	}

	// use
	useCmd := &cobra.Command{
		Use:   "use [modules...]",
		Short: i18n.T("cmd.go.work.use.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Auto-detect modules
				modules := findGoModules(".")
				if len(modules) == 0 {
					return fmt.Errorf(i18n.T("cmd.go.work.error.no_modules"))
				}
				for _, mod := range modules {
					execCmd := exec.Command("go", "work", "use", mod)
					execCmd.Stdout = os.Stdout
					execCmd.Stderr = os.Stderr
					if err := execCmd.Run(); err != nil {
						return err
					}
					fmt.Println(i18n.T("cmd.go.work.added", map[string]interface{}{"Module": mod}))
				}
				return nil
			}

			cmdArgs := append([]string{"work", "use"}, args...)
			execCmd := exec.Command("go", cmdArgs...)
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	workCmd.AddCommand(syncCmd)
	workCmd.AddCommand(initCmd)
	workCmd.AddCommand(useCmd)
	parent.AddCommand(workCmd)
}

func findGoModules(root string) []string {
	var modules []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "go.mod" && path != "go.mod" {
			modules = append(modules, filepath.Dir(path))
		}
		return nil
	})
	return modules
}
