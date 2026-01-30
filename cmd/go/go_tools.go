package gocmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/leaanthony/clir"
)

func addGoInstallCommand(parent *clir.Command) {
	var verbose bool
	var noCgo bool

	installCmd := parent.NewSubCommand("install", "Install Go binary")
	installCmd.LongDescription("Install Go binary to $GOPATH/bin.\n\n" +
		"Examples:\n" +
		"  core go install              # Install current module\n" +
		"  core go install ./cmd/core   # Install specific path\n" +
		"  core go install --no-cgo     # Pure Go (no C dependencies)\n" +
		"  core go install -v           # Verbose output")

	installCmd.BoolFlag("v", "Verbose output", &verbose)
	installCmd.BoolFlag("no-cgo", "Disable CGO (CGO_ENABLED=0)", &noCgo)

	installCmd.Action(func() error {
		// Get install path from args or default to current dir
		args := installCmd.OtherArgs()
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
		if noCgo {
			fmt.Printf("  %s %s\n", dimStyle.Render("CGO:"), "disabled")
		}

		cmdArgs := []string{"install"}
		if verbose {
			cmdArgs = append(cmdArgs, "-v")
		}
		cmdArgs = append(cmdArgs, installPath)

		cmd := exec.Command("go", cmdArgs...)
		if noCgo {
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
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
	})
}

func addGoModCommand(parent *clir.Command) {
	modCmd := parent.NewSubCommand("mod", "Module management")
	modCmd.LongDescription("Go module management commands.\n\n" +
		"Commands:\n" +
		"  tidy      Add missing and remove unused modules\n" +
		"  download  Download modules to local cache\n" +
		"  verify    Verify dependencies\n" +
		"  graph     Print module dependency graph")

	// tidy
	tidyCmd := modCmd.NewSubCommand("tidy", "Tidy go.mod")
	tidyCmd.Action(func() error {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	// download
	downloadCmd := modCmd.NewSubCommand("download", "Download modules")
	downloadCmd.Action(func() error {
		cmd := exec.Command("go", "mod", "download")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	// verify
	verifyCmd := modCmd.NewSubCommand("verify", "Verify dependencies")
	verifyCmd.Action(func() error {
		cmd := exec.Command("go", "mod", "verify")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	// graph
	graphCmd := modCmd.NewSubCommand("graph", "Print dependency graph")
	graphCmd.Action(func() error {
		cmd := exec.Command("go", "mod", "graph")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
}

func addGoWorkCommand(parent *clir.Command) {
	workCmd := parent.NewSubCommand("work", "Workspace management")
	workCmd.LongDescription("Go workspace management commands.\n\n" +
		"Commands:\n" +
		"  sync    Sync go.work with modules\n" +
		"  init    Initialize go.work\n" +
		"  use     Add module to workspace")

	// sync
	syncCmd := workCmd.NewSubCommand("sync", "Sync workspace")
	syncCmd.Action(func() error {
		cmd := exec.Command("go", "work", "sync")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	// init
	initCmd := workCmd.NewSubCommand("init", "Initialize workspace")
	initCmd.Action(func() error {
		cmd := exec.Command("go", "work", "init")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		// Auto-add current module if go.mod exists
		if _, err := os.Stat("go.mod"); err == nil {
			cmd = exec.Command("go", "work", "use", ".")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
		return nil
	})

	// use
	useCmd := workCmd.NewSubCommand("use", "Add module to workspace")
	useCmd.Action(func() error {
		args := useCmd.OtherArgs()
		if len(args) == 0 {
			// Auto-detect modules
			modules := findGoModules(".")
			if len(modules) == 0 {
				return fmt.Errorf("no go.mod files found")
			}
			for _, mod := range modules {
				cmd := exec.Command("go", "work", "use", mod)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return err
				}
				fmt.Printf("Added %s\n", mod)
			}
			return nil
		}

		cmdArgs := append([]string{"work", "use"}, args...)
		cmd := exec.Command("go", cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
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
