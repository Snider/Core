package gocmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	installVerbose bool
	installNoCgo   bool
)

func addGoInstallCommand(parent *cobra.Command) {
	installCmd := &cobra.Command{
		Use:   "install [path]",
		Short: "Install Go binary",
		Long: "Install Go binary to $GOPATH/bin.\n\n" +
			"Examples:\n" +
			"  core go install              # Install current module\n" +
			"  core go install ./cmd/core   # Install specific path\n" +
			"  core go install --no-cgo     # Pure Go (no C dependencies)\n" +
			"  core go install -v           # Verbose output",
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

			fmt.Printf("%s Installing\n", dimStyle.Render("Install:"))
			fmt.Printf("  %s %s\n", dimStyle.Render("Path:"), installPath)
			if installNoCgo {
				fmt.Printf("  %s %s\n", dimStyle.Render("CGO:"), "disabled")
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
				fmt.Printf("\n%s\n", errorStyle.Render("FAIL Install failed"))
				return err
			}

			// Show where it was installed
			gopath := os.Getenv("GOPATH")
			if gopath == "" {
				home, _ := os.UserHomeDir()
				gopath = filepath.Join(home, "go")
			}
			binDir := filepath.Join(gopath, "bin")

			fmt.Printf("\n%s Installed to %s\n", successStyle.Render("OK"), binDir)
			return nil
		},
	}

	installCmd.Flags().BoolVarP(&installVerbose, "verbose", "v", false, "Verbose output")
	installCmd.Flags().BoolVar(&installNoCgo, "no-cgo", false, "Disable CGO (CGO_ENABLED=0)")

	parent.AddCommand(installCmd)
}

func addGoModCommand(parent *cobra.Command) {
	modCmd := &cobra.Command{
		Use:   "mod",
		Short: "Module management",
		Long: "Go module management commands.\n\n" +
			"Commands:\n" +
			"  tidy      Add missing and remove unused modules\n" +
			"  download  Download modules to local cache\n" +
			"  verify    Verify dependencies\n" +
			"  graph     Print module dependency graph",
	}

	// tidy
	tidyCmd := &cobra.Command{
		Use:   "tidy",
		Short: "Tidy go.mod",
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
		Short: "Download modules",
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
		Short: "Verify dependencies",
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
		Short: "Print dependency graph",
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
		Short: "Workspace management",
		Long: "Go workspace management commands.\n\n" +
			"Commands:\n" +
			"  sync    Sync go.work with modules\n" +
			"  init    Initialize go.work\n" +
			"  use     Add module to workspace",
	}

	// sync
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync workspace",
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
		Short: "Initialize workspace",
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
		Short: "Add module to workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Auto-detect modules
				modules := findGoModules(".")
				if len(modules) == 0 {
					return fmt.Errorf("no go.mod files found")
				}
				for _, mod := range modules {
					execCmd := exec.Command("go", "work", "use", mod)
					execCmd.Stdout = os.Stdout
					execCmd.Stderr = os.Stderr
					if err := execCmd.Run(); err != nil {
						return err
					}
					fmt.Printf("Added %s\n", mod)
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
