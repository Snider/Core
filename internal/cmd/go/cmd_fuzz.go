package gocmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

var (
	fuzzDuration time.Duration
	fuzzPkg      string
	fuzzRun      string
	fuzzVerbose  bool
)

func addGoFuzzCommand(parent *cli.Command) {
	fuzzCmd := &cli.Command{
		Use:   "fuzz",
		Short: "Run Go fuzz tests",
		Long: `Run Go fuzz tests with configurable duration.

Discovers Fuzz* functions across the project and runs each with go test -fuzz.

Examples:
  core go fuzz                      # Run all fuzz targets for 10s each
  core go fuzz --duration=30s       # Run each target for 30s
  core go fuzz --pkg=./pkg/...      # Fuzz specific package
  core go fuzz --run=FuzzE          # Run only matching fuzz targets`,
		RunE: func(cmd *cli.Command, args []string) error {
			return runGoFuzz(fuzzDuration, fuzzPkg, fuzzRun, fuzzVerbose)
		},
	}

	fuzzCmd.Flags().DurationVar(&fuzzDuration, "duration", 10*time.Second, "Duration per fuzz target")
	fuzzCmd.Flags().StringVar(&fuzzPkg, "pkg", "", "Package to fuzz (default: auto-discover)")
	fuzzCmd.Flags().StringVar(&fuzzRun, "run", "", "Only run fuzz targets matching pattern")
	fuzzCmd.Flags().BoolVarP(&fuzzVerbose, "verbose", "v", false, "Verbose output")

	parent.AddCommand(fuzzCmd)
}

// fuzzTarget represents a discovered fuzz function and its package.
type fuzzTarget struct {
	Pkg  string
	Name string
}

func runGoFuzz(duration time.Duration, pkg, run string, verbose bool) error {
	cli.Print("%s %s\n", dimStyle.Render(i18n.Label("fuzz")), i18n.ProgressSubject("run", "fuzz tests"))
	cli.Blank()

	targets, err := discoverFuzzTargets(pkg, run)
	if err != nil {
		return cli.Wrap(err, "discover fuzz targets")
	}

	if len(targets) == 0 {
		cli.Print("  %s no fuzz targets found\n", dimStyle.Render("—"))
		return nil
	}

	cli.Print("  %s %d target(s), %s each\n", dimStyle.Render(i18n.Label("targets")), len(targets), duration)
	cli.Blank()

	passed := 0
	failed := 0

	for _, t := range targets {
		cli.Print("  %s %s in %s\n", dimStyle.Render("→"), t.Name, t.Pkg)

		args := []string{
			"test",
			fmt.Sprintf("-fuzz=^%s$", t.Name),
			fmt.Sprintf("-fuzztime=%s", duration),
			"-run=^$", // Don't run unit tests
		}
		if verbose {
			args = append(args, "-v")
		}
		args = append(args, t.Pkg)

		cmd := exec.Command("go", args...)
		cmd.Env = append(os.Environ(), "MACOSX_DEPLOYMENT_TARGET=26.0", "CGO_ENABLED=0")
		cmd.Dir, _ = os.Getwd()

		output, runErr := cmd.CombinedOutput()
		outputStr := string(output)

		if runErr != nil {
			failed++
			cli.Print("    %s %s\n", errorStyle.Render(cli.Glyph(":cross:")), runErr.Error())
			if outputStr != "" {
				cli.Text(outputStr)
			}
		} else {
			passed++
			cli.Print("    %s %s\n", successStyle.Render(cli.Glyph(":check:")), i18n.T("i18n.done.pass"))
			if verbose && outputStr != "" {
				cli.Text(outputStr)
			}
		}
	}

	cli.Blank()
	if failed > 0 {
		cli.Print("%s %d passed, %d failed\n", errorStyle.Render(cli.Glyph(":cross:")), passed, failed)
		return cli.Err("fuzz: %d target(s) failed", failed)
	}

	cli.Print("%s %d passed\n", successStyle.Render(cli.Glyph(":check:")), passed)
	return nil
}

// discoverFuzzTargets scans for Fuzz* functions in test files.
func discoverFuzzTargets(pkg, pattern string) ([]fuzzTarget, error) {
	root := "."
	if pkg != "" {
		// Convert Go package pattern to filesystem path
		root = strings.TrimPrefix(pkg, "./")
		root = strings.TrimSuffix(root, "/...")
	}

	fuzzRe := regexp.MustCompile(`^func\s+(Fuzz\w+)\s*\(\s*\w+\s+\*testing\.F\s*\)`)
	var matchRe *regexp.Regexp
	if pattern != "" {
		var err error
		matchRe, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid --run pattern: %w", err)
		}
	}

	var targets []fuzzTarget
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		dir := "./" + filepath.Dir(path)
		for line := range strings.SplitSeq(string(data), "\n") {
			m := fuzzRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			name := m[1]
			if matchRe != nil && !matchRe.MatchString(name) {
				continue
			}
			targets = append(targets, fuzzTarget{Pkg: dir, Name: name})
		}
		return nil
	})
	return targets, err
}
