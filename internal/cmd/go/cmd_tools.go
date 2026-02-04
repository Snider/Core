package gocmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

var (
	installVerbose bool
	installNoCgo   bool
)

func addGoInstallCommand(parent *cli.Command) {
	installCmd := &cli.Command{
		Use:   "install [path]",
		Short: "Install Go binary",
		Long:  "Install Go binary to $GOPATH/bin",
		RunE: func(cmd *cli.Command, args []string) error {
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

			cli.Print("%s %s\n", dimStyle.Render(i18n.Label("install")), i18n.Progress("install"))
			cli.Print("  %s %s\n", dimStyle.Render(i18n.Label("path")), installPath)
			if installNoCgo {
				cli.Print("  %s %s\n", dimStyle.Render(i18n.Label("cgo")), "disabled")
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
				cli.Print("\n%s\n", errorStyle.Render(i18n.T("i18n.fail.install", "binary")))
				return err
			}

			// Show where it was installed
			gopath := os.Getenv("GOPATH")
			if gopath == "" {
				home, _ := os.UserHomeDir()
				gopath = filepath.Join(home, "go")
			}
			binDir := filepath.Join(gopath, "bin")

			cli.Print("\n%s %s\n", successStyle.Render(i18n.T("i18n.done.install")), binDir)
			return nil
		},
	}

	installCmd.Flags().BoolVarP(&installVerbose, "verbose", "v", false, "Verbose output")
	installCmd.Flags().BoolVar(&installNoCgo, "no-cgo", false, "Disable CGO")

	parent.AddCommand(installCmd)
}

func addGoModCommand(parent *cli.Command) {
	modCmd := &cli.Command{
		Use:   "mod",
		Short: "Module management",
		Long:  "Go module management commands",
	}

	// tidy
	tidyCmd := &cli.Command{
		Use:   "tidy",
		Short: "Run go mod tidy",
		RunE: func(cmd *cli.Command, args []string) error {
			execCmd := exec.Command("go", "mod", "tidy")
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	// download
	downloadCmd := &cli.Command{
		Use:   "download",
		Short: "Download module dependencies",
		RunE: func(cmd *cli.Command, args []string) error {
			execCmd := exec.Command("go", "mod", "download")
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	// verify
	verifyCmd := &cli.Command{
		Use:   "verify",
		Short: "Verify module checksums",
		RunE: func(cmd *cli.Command, args []string) error {
			execCmd := exec.Command("go", "mod", "verify")
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	// graph
	graphCmd := &cli.Command{
		Use:   "graph",
		Short: "Print module dependency graph",
		RunE: func(cmd *cli.Command, args []string) error {
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

func addGoWorkCommand(parent *cli.Command) {
	workCmd := &cli.Command{
		Use:   "work",
		Short: "Workspace management",
		Long:  "Go workspace management commands",
	}

	// sync
	syncCmd := &cli.Command{
		Use:   "sync",
		Short: "Sync workspace modules",
		RunE: func(cmd *cli.Command, args []string) error {
			execCmd := exec.Command("go", "work", "sync")
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	// init
	initCmd := &cli.Command{
		Use:   "init",
		Short: "Initialise a new workspace",
		RunE: func(cmd *cli.Command, args []string) error {
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
	useCmd := &cli.Command{
		Use:   "use [modules...]",
		Short: "Add modules to workspace",
		RunE: func(cmd *cli.Command, args []string) error {
			if len(args) == 0 {
				// Auto-detect modules
				modules := findGoModules(".")
				if len(modules) == 0 {
					return errors.New("no Go modules found")
				}
				for _, mod := range modules {
					execCmd := exec.Command("go", "work", "use", mod)
					execCmd.Stdout = os.Stdout
					execCmd.Stderr = os.Stderr
					if err := execCmd.Run(); err != nil {
						return err
					}
					cli.Print("%s %s\n", successStyle.Render(i18n.T("i18n.done.add")), mod)
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
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
