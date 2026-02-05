// cmd_project.go implements the main project build logic.
//
// This handles auto-detection of project types (Go, Wails, Docker, LinuxKit, Taskfile)
// and orchestrates the build process including signing, archiving, and checksums.

package buildcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/host-uk/core/pkg/build"
	"github.com/host-uk/core/pkg/build/builders"
	"github.com/host-uk/core/pkg/build/signing"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/io"
)

// runProjectBuild handles the main `core build` command with auto-detection.
func runProjectBuild(ctx context.Context, buildType string, ciMode bool, targetsFlag string, outputDir string, doArchive bool, doChecksum bool, configPath string, format string, push bool, imageName string, noSign bool, notarize bool, verbose bool) error {
	// Use local filesystem as the default medium
	fs := io.Local

	// Get current working directory as project root
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
	}

	// Load configuration from .core/build.yaml (or defaults)
	buildCfg, err := build.LoadConfig(fs, projectDir)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "load config"}), err)
	}

	// Detect project type if not specified
	var projectType build.ProjectType
	if buildType != "" {
		projectType = build.ProjectType(buildType)
	} else {
		projectType, err = build.PrimaryType(fs, projectDir)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "detect project type"}), err)
		}
		if projectType == "" {
			return fmt.Errorf("%s", i18n.T("cmd.build.error.no_project_type", map[string]interface{}{"Dir": projectDir}))
		}
	}

	// Determine targets
	var buildTargets []build.Target
	if targetsFlag != "" {
		// Parse from command line
		buildTargets, err = parseTargets(targetsFlag)
		if err != nil {
			return err
		}
	} else if len(buildCfg.Targets) > 0 {
		// Use config targets
		buildTargets = buildCfg.ToTargets()
	} else {
		// Fall back to current OS/arch
		buildTargets = []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}
	}

	// Determine output directory
	if outputDir == "" {
		outputDir = "dist"
	}
	if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(projectDir, outputDir)
	}
	outputDir = filepath.Clean(outputDir)

	// Ensure config path is absolute if provided
	if configPath != "" && !filepath.IsAbs(configPath) {
		configPath = filepath.Join(projectDir, configPath)
	}

	// Determine binary name
	binaryName := buildCfg.Project.Binary
	if binaryName == "" {
		binaryName = buildCfg.Project.Name
	}
	if binaryName == "" {
		binaryName = filepath.Base(projectDir)
	}

	// Print build info (verbose mode only)
	if verbose && !ciMode {
		fmt.Printf("%s %s\n", buildHeaderStyle.Render(i18n.T("cmd.build.label.build")), i18n.T("cmd.build.building_project"))
		fmt.Printf("  %s %s\n", i18n.T("cmd.build.label.type"), buildTargetStyle.Render(string(projectType)))
		fmt.Printf("  %s %s\n", i18n.T("cmd.build.label.output"), buildTargetStyle.Render(outputDir))
		fmt.Printf("  %s %s\n", i18n.T("cmd.build.label.binary"), buildTargetStyle.Render(binaryName))
		fmt.Printf("  %s %s\n", i18n.T("cmd.build.label.targets"), buildTargetStyle.Render(formatTargets(buildTargets)))
		fmt.Println()
	}

	// Get the appropriate builder
	builder, err := getBuilder(projectType)
	if err != nil {
		return err
	}

	// Create build config for the builder
	cfg := &build.Config{
		FS:         fs,
		ProjectDir: projectDir,
		OutputDir:  outputDir,
		Name:       binaryName,
		Version:    buildCfg.Project.Name, // Could be enhanced with git describe
		LDFlags:    buildCfg.Build.LDFlags,
		// Docker/LinuxKit specific
		Dockerfile:     configPath, // Reuse for Dockerfile path
		LinuxKitConfig: configPath,
		Push:           push,
		Image:          imageName,
	}

	// Parse formats for LinuxKit
	if format != "" {
		cfg.Formats = strings.Split(format, ",")
	}

	// Execute build
	artifacts, err := builder.Build(ctx, cfg, buildTargets)
	if err != nil {
		if !ciMode {
			fmt.Printf("%s %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), err)
		}
		return err
	}

	if verbose && !ciMode {
		fmt.Printf("%s %s\n", buildSuccessStyle.Render(i18n.T("common.label.success")), i18n.T("cmd.build.built_artifacts", map[string]interface{}{"Count": len(artifacts)}))
		fmt.Println()
		for _, artifact := range artifacts {
			relPath, err := filepath.Rel(projectDir, artifact.Path)
			if err != nil {
				relPath = artifact.Path
			}
			fmt.Printf("  %s %s %s\n",
				buildSuccessStyle.Render("*"),
				buildTargetStyle.Render(relPath),
				buildDimStyle.Render(fmt.Sprintf("(%s/%s)", artifact.OS, artifact.Arch)),
			)
		}
	}

	// Sign macOS binaries if enabled
	signCfg := buildCfg.Sign
	if notarize {
		signCfg.MacOS.Notarize = true
	}
	if noSign {
		signCfg.Enabled = false
	}

	if signCfg.Enabled && runtime.GOOS == "darwin" {
		if verbose && !ciMode {
			fmt.Println()
			fmt.Printf("%s %s\n", buildHeaderStyle.Render(i18n.T("cmd.build.label.sign")), i18n.T("cmd.build.signing_binaries"))
		}

		// Convert build.Artifact to signing.Artifact
		signingArtifacts := make([]signing.Artifact, len(artifacts))
		for i, a := range artifacts {
			signingArtifacts[i] = signing.Artifact{Path: a.Path, OS: a.OS, Arch: a.Arch}
		}

		if err := signing.SignBinaries(ctx, fs, signCfg, signingArtifacts); err != nil {
			if !ciMode {
				fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("cmd.build.error.signing_failed"), err)
			}
			return err
		}

		if signCfg.MacOS.Notarize {
			if err := signing.NotarizeBinaries(ctx, fs, signCfg, signingArtifacts); err != nil {
				if !ciMode {
					fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("cmd.build.error.notarization_failed"), err)
				}
				return err
			}
		}
	}

	// Archive artifacts if enabled
	var archivedArtifacts []build.Artifact
	if doArchive && len(artifacts) > 0 {
		if verbose && !ciMode {
			fmt.Println()
			fmt.Printf("%s %s\n", buildHeaderStyle.Render(i18n.T("cmd.build.label.archive")), i18n.T("cmd.build.creating_archives"))
		}

		archivedArtifacts, err = build.ArchiveAll(fs, artifacts)
		if err != nil {
			if !ciMode {
				fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("cmd.build.error.archive_failed"), err)
			}
			return err
		}

		if verbose && !ciMode {
			for _, artifact := range archivedArtifacts {
				relPath, err := filepath.Rel(projectDir, artifact.Path)
				if err != nil {
					relPath = artifact.Path
				}
				fmt.Printf("  %s %s %s\n",
					buildSuccessStyle.Render("*"),
					buildTargetStyle.Render(relPath),
					buildDimStyle.Render(fmt.Sprintf("(%s/%s)", artifact.OS, artifact.Arch)),
				)
			}
		}
	}

	// Compute checksums if enabled
	var checksummedArtifacts []build.Artifact
	if doChecksum && len(archivedArtifacts) > 0 {
		checksummedArtifacts, err = computeAndWriteChecksums(ctx, projectDir, outputDir, archivedArtifacts, signCfg, ciMode, verbose)
		if err != nil {
			return err
		}
	} else if doChecksum && len(artifacts) > 0 && !doArchive {
		// Checksum raw binaries if archiving is disabled
		checksummedArtifacts, err = computeAndWriteChecksums(ctx, projectDir, outputDir, artifacts, signCfg, ciMode, verbose)
		if err != nil {
			return err
		}
	}

	// Output results
	if ciMode {
		// Determine which artifacts to output (prefer checksummed > archived > raw)
		var outputArtifacts []build.Artifact
		if len(checksummedArtifacts) > 0 {
			outputArtifacts = checksummedArtifacts
		} else if len(archivedArtifacts) > 0 {
			outputArtifacts = archivedArtifacts
		} else {
			outputArtifacts = artifacts
		}

		// JSON output for CI
		output, err := json.MarshalIndent(outputArtifacts, "", "  ")
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "marshal artifacts"}), err)
		}
		fmt.Println(string(output))
	} else if !verbose {
		// Minimal output: just success with artifact count
		fmt.Printf("%s %s %s\n",
			buildSuccessStyle.Render(i18n.T("common.label.success")),
			i18n.T("cmd.build.built_artifacts", map[string]interface{}{"Count": len(artifacts)}),
			buildDimStyle.Render(fmt.Sprintf("(%s)", outputDir)),
		)
	}

	return nil
}

// computeAndWriteChecksums computes checksums for artifacts and writes CHECKSUMS.txt.
func computeAndWriteChecksums(ctx context.Context, projectDir, outputDir string, artifacts []build.Artifact, signCfg signing.SignConfig, ciMode bool, verbose bool) ([]build.Artifact, error) {
	fs := io.Local
	if verbose && !ciMode {
		fmt.Println()
		fmt.Printf("%s %s\n", buildHeaderStyle.Render(i18n.T("cmd.build.label.checksum")), i18n.T("cmd.build.computing_checksums"))
	}

	checksummedArtifacts, err := build.ChecksumAll(fs, artifacts)
	if err != nil {
		if !ciMode {
			fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("cmd.build.error.checksum_failed"), err)
		}
		return nil, err
	}

	// Write CHECKSUMS.txt
	checksumPath := filepath.Join(outputDir, "CHECKSUMS.txt")
	if err := build.WriteChecksumFile(fs, checksummedArtifacts, checksumPath); err != nil {
		if !ciMode {
			fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("common.error.failed", map[string]any{"Action": "write CHECKSUMS.txt"}), err)
		}
		return nil, err
	}

	// Sign checksums with GPG
	if signCfg.Enabled {
		if err := signing.SignChecksums(ctx, fs, signCfg, checksumPath); err != nil {
			if !ciMode {
				fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("cmd.build.error.gpg_signing_failed"), err)
			}
			return nil, err
		}
	}

	if verbose && !ciMode {
		for _, artifact := range checksummedArtifacts {
			relPath, err := filepath.Rel(projectDir, artifact.Path)
			if err != nil {
				relPath = artifact.Path
			}
			fmt.Printf("  %s %s\n",
				buildSuccessStyle.Render("*"),
				buildTargetStyle.Render(relPath),
			)
			fmt.Printf("    %s\n", buildDimStyle.Render(artifact.Checksum))
		}

		relChecksumPath, err := filepath.Rel(projectDir, checksumPath)
		if err != nil {
			relChecksumPath = checksumPath
		}
		fmt.Printf("  %s %s\n",
			buildSuccessStyle.Render("*"),
			buildTargetStyle.Render(relChecksumPath),
		)
	}

	return checksummedArtifacts, nil
}

// parseTargets parses a comma-separated list of OS/arch pairs.
func parseTargets(targetsFlag string) ([]build.Target, error) {
	parts := strings.Split(targetsFlag, ",")
	var targets []build.Target

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		osArch := strings.Split(part, "/")
		if len(osArch) != 2 {
			return nil, fmt.Errorf("%s", i18n.T("cmd.build.error.invalid_target", map[string]interface{}{"Target": part}))
		}

		targets = append(targets, build.Target{
			OS:   strings.TrimSpace(osArch[0]),
			Arch: strings.TrimSpace(osArch[1]),
		})
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("%s", i18n.T("cmd.build.error.no_targets"))
	}

	return targets, nil
}

// formatTargets returns a human-readable string of targets.
func formatTargets(targets []build.Target) string {
	var parts []string
	for _, t := range targets {
		parts = append(parts, t.String())
	}
	return strings.Join(parts, ", ")
}

// getBuilder returns the appropriate builder for the project type.
func getBuilder(projectType build.ProjectType) (build.Builder, error) {
	switch projectType {
	case build.ProjectTypeWails:
		return builders.NewWailsBuilder(), nil
	case build.ProjectTypeGo:
		return builders.NewGoBuilder(), nil
	case build.ProjectTypeDocker:
		return builders.NewDockerBuilder(), nil
	case build.ProjectTypeLinuxKit:
		return builders.NewLinuxKitBuilder(), nil
	case build.ProjectTypeTaskfile:
		return builders.NewTaskfileBuilder(), nil
	case build.ProjectTypeNode:
		return nil, fmt.Errorf("%s", i18n.T("cmd.build.error.node_not_implemented"))
	case build.ProjectTypePHP:
		return nil, fmt.Errorf("%s", i18n.T("cmd.build.error.php_not_implemented"))
	default:
		return nil, fmt.Errorf("%s: %s", i18n.T("cmd.build.error.unsupported_type"), projectType)
	}
}
