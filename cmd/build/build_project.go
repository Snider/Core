// build_project.go implements the main project build logic.
//
// This handles auto-detection of project types (Go, Wails, Docker, LinuxKit, Taskfile)
// and orchestrates the build process including signing, archiving, and checksums.

package build

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	buildpkg "github.com/host-uk/core/pkg/build"
	"github.com/host-uk/core/pkg/build/builders"
	"github.com/host-uk/core/pkg/build/signing"
	"github.com/host-uk/core/pkg/i18n"
)

// runProjectBuild handles the main `core build` command with auto-detection.
func runProjectBuild(buildType string, ciMode bool, targetsFlag string, outputDir string, doArchive bool, doChecksum bool, configPath string, format string, push bool, imageName string, noSign bool, notarize bool) error {
	// Get current working directory as project root
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
	}

	// Load configuration from .core/build.yaml (or defaults)
	buildCfg, err := buildpkg.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "load config"}), err)
	}

	// Detect project type if not specified
	var projectType buildpkg.ProjectType
	if buildType != "" {
		projectType = buildpkg.ProjectType(buildType)
	} else {
		projectType, err = buildpkg.PrimaryType(projectDir)
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "detect project type"}), err)
		}
		if projectType == "" {
			return fmt.Errorf("%s", i18n.T("cmd.build.error.no_project_type", map[string]interface{}{"Dir": projectDir}))
		}
	}

	// Determine targets
	var buildTargets []buildpkg.Target
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
		buildTargets = []buildpkg.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}
	}

	// Determine output directory
	if outputDir == "" {
		outputDir = "dist"
	}

	// Determine binary name
	binaryName := buildCfg.Project.Binary
	if binaryName == "" {
		binaryName = buildCfg.Project.Name
	}
	if binaryName == "" {
		binaryName = filepath.Base(projectDir)
	}

	// Print build info (unless CI mode)
	if !ciMode {
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
	cfg := &buildpkg.Config{
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
	ctx := context.Background()
	artifacts, err := builder.Build(ctx, cfg, buildTargets)
	if err != nil {
		if !ciMode {
			fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("common.error.failed", map[string]any{"Action": "build"}), err)
		}
		return err
	}

	if !ciMode {
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
		if !ciMode {
			fmt.Println()
			fmt.Printf("%s %s\n", buildHeaderStyle.Render(i18n.T("cmd.build.label.sign")), i18n.T("cmd.build.signing_binaries"))
		}

		// Convert buildpkg.Artifact to signing.Artifact
		signingArtifacts := make([]signing.Artifact, len(artifacts))
		for i, a := range artifacts {
			signingArtifacts[i] = signing.Artifact{Path: a.Path, OS: a.OS, Arch: a.Arch}
		}

		if err := signing.SignBinaries(ctx, signCfg, signingArtifacts); err != nil {
			if !ciMode {
				fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("cmd.build.error.signing_failed"), err)
			}
			return err
		}

		if signCfg.MacOS.Notarize {
			if err := signing.NotarizeBinaries(ctx, signCfg, signingArtifacts); err != nil {
				if !ciMode {
					fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("cmd.build.error.notarization_failed"), err)
				}
				return err
			}
		}
	}

	// Archive artifacts if enabled
	var archivedArtifacts []buildpkg.Artifact
	if doArchive && len(artifacts) > 0 {
		if !ciMode {
			fmt.Println()
			fmt.Printf("%s %s\n", buildHeaderStyle.Render(i18n.T("cmd.build.label.archive")), i18n.T("cmd.build.creating_archives"))
		}

		archivedArtifacts, err = buildpkg.ArchiveAll(artifacts)
		if err != nil {
			if !ciMode {
				fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("cmd.build.error.archive_failed"), err)
			}
			return err
		}

		if !ciMode {
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
	var checksummedArtifacts []buildpkg.Artifact
	if doChecksum && len(archivedArtifacts) > 0 {
		checksummedArtifacts, err = computeAndWriteChecksums(ctx, projectDir, outputDir, archivedArtifacts, signCfg, ciMode)
		if err != nil {
			return err
		}
	} else if doChecksum && len(artifacts) > 0 && !doArchive {
		// Checksum raw binaries if archiving is disabled
		checksummedArtifacts, err = computeAndWriteChecksums(ctx, projectDir, outputDir, artifacts, signCfg, ciMode)
		if err != nil {
			return err
		}
	}

	// Output results for CI mode
	if ciMode {
		// Determine which artifacts to output (prefer checksummed > archived > raw)
		var outputArtifacts []buildpkg.Artifact
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
	}

	return nil
}

// computeAndWriteChecksums computes checksums for artifacts and writes CHECKSUMS.txt.
func computeAndWriteChecksums(ctx context.Context, projectDir, outputDir string, artifacts []buildpkg.Artifact, signCfg signing.SignConfig, ciMode bool) ([]buildpkg.Artifact, error) {
	if !ciMode {
		fmt.Println()
		fmt.Printf("%s %s\n", buildHeaderStyle.Render(i18n.T("cmd.build.label.checksum")), i18n.T("cmd.build.computing_checksums"))
	}

	checksummedArtifacts, err := buildpkg.ChecksumAll(artifacts)
	if err != nil {
		if !ciMode {
			fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("cmd.build.error.checksum_failed"), err)
		}
		return nil, err
	}

	// Write CHECKSUMS.txt
	checksumPath := filepath.Join(outputDir, "CHECKSUMS.txt")
	if err := buildpkg.WriteChecksumFile(checksummedArtifacts, checksumPath); err != nil {
		if !ciMode {
			fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("common.error.failed", map[string]any{"Action": "write CHECKSUMS.txt"}), err)
		}
		return nil, err
	}

	// Sign checksums with GPG
	if signCfg.Enabled {
		if err := signing.SignChecksums(ctx, signCfg, checksumPath); err != nil {
			if !ciMode {
				fmt.Printf("%s %s: %v\n", buildErrorStyle.Render(i18n.T("common.label.error")), i18n.T("cmd.build.error.gpg_signing_failed"), err)
			}
			return nil, err
		}
	}

	if !ciMode {
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
func parseTargets(targetsFlag string) ([]buildpkg.Target, error) {
	parts := strings.Split(targetsFlag, ",")
	var targets []buildpkg.Target

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		osArch := strings.Split(part, "/")
		if len(osArch) != 2 {
			return nil, fmt.Errorf("%s", i18n.T("cmd.build.error.invalid_target", map[string]interface{}{"Target": part}))
		}

		targets = append(targets, buildpkg.Target{
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
func formatTargets(targets []buildpkg.Target) string {
	var parts []string
	for _, t := range targets {
		parts = append(parts, t.String())
	}
	return strings.Join(parts, ", ")
}

// getBuilder returns the appropriate builder for the project type.
func getBuilder(projectType buildpkg.ProjectType) (buildpkg.Builder, error) {
	switch projectType {
	case buildpkg.ProjectTypeWails:
		return builders.NewWailsBuilder(), nil
	case buildpkg.ProjectTypeGo:
		return builders.NewGoBuilder(), nil
	case buildpkg.ProjectTypeDocker:
		return builders.NewDockerBuilder(), nil
	case buildpkg.ProjectTypeLinuxKit:
		return builders.NewLinuxKitBuilder(), nil
	case buildpkg.ProjectTypeTaskfile:
		return builders.NewTaskfileBuilder(), nil
	case buildpkg.ProjectTypeNode:
		return nil, fmt.Errorf("%s", i18n.T("cmd.build.error.node_not_implemented"))
	case buildpkg.ProjectTypePHP:
		return nil, fmt.Errorf("%s", i18n.T("cmd.build.error.php_not_implemented"))
	default:
		return nil, fmt.Errorf("%s: %s", i18n.T("cmd.build.error.unsupported_type"), projectType)
	}
}
