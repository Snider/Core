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
)

// runProjectBuild handles the main `core build` command with auto-detection.
func runProjectBuild(buildType string, ciMode bool, targetsFlag string, outputDir string, doArchive bool, doChecksum bool, configPath string, format string, push bool, imageName string, noSign bool, notarize bool) error {
	// Get current working directory as project root
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load configuration from .core/build.yaml (or defaults)
	buildCfg, err := buildpkg.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Detect project type if not specified
	var projectType buildpkg.ProjectType
	if buildType != "" {
		projectType = buildpkg.ProjectType(buildType)
	} else {
		projectType, err = buildpkg.PrimaryType(projectDir)
		if err != nil {
			return fmt.Errorf("failed to detect project type: %w", err)
		}
		if projectType == "" {
			return fmt.Errorf("no supported project type detected in %s\n"+
				"Supported types: go (go.mod), wails (wails.json), node (package.json), php (composer.json)", projectDir)
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
		fmt.Printf("%s Building project\n", buildHeaderStyle.Render("Build:"))
		fmt.Printf("  Type:    %s\n", buildTargetStyle.Render(string(projectType)))
		fmt.Printf("  Output:  %s\n", buildTargetStyle.Render(outputDir))
		fmt.Printf("  Binary:  %s\n", buildTargetStyle.Render(binaryName))
		fmt.Printf("  Targets: %s\n", buildTargetStyle.Render(formatTargets(buildTargets)))
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
			fmt.Printf("%s Build failed: %v\n", buildErrorStyle.Render("Error:"), err)
		}
		return err
	}

	if !ciMode {
		fmt.Printf("%s Built %d artifact(s)\n", buildSuccessStyle.Render("Success:"), len(artifacts))
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
			fmt.Printf("%s Signing binaries...\n", buildHeaderStyle.Render("Sign:"))
		}

		// Convert buildpkg.Artifact to signing.Artifact
		signingArtifacts := make([]signing.Artifact, len(artifacts))
		for i, a := range artifacts {
			signingArtifacts[i] = signing.Artifact{Path: a.Path, OS: a.OS, Arch: a.Arch}
		}

		if err := signing.SignBinaries(ctx, signCfg, signingArtifacts); err != nil {
			if !ciMode {
				fmt.Printf("%s Signing failed: %v\n", buildErrorStyle.Render("Error:"), err)
			}
			return err
		}

		if signCfg.MacOS.Notarize {
			if err := signing.NotarizeBinaries(ctx, signCfg, signingArtifacts); err != nil {
				if !ciMode {
					fmt.Printf("%s Notarization failed: %v\n", buildErrorStyle.Render("Error:"), err)
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
			fmt.Printf("%s Creating archives...\n", buildHeaderStyle.Render("Archive:"))
		}

		archivedArtifacts, err = buildpkg.ArchiveAll(artifacts)
		if err != nil {
			if !ciMode {
				fmt.Printf("%s Archive failed: %v\n", buildErrorStyle.Render("Error:"), err)
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
			return fmt.Errorf("failed to marshal artifacts: %w", err)
		}
		fmt.Println(string(output))
	}

	return nil
}

// computeAndWriteChecksums computes checksums for artifacts and writes CHECKSUMS.txt.
func computeAndWriteChecksums(ctx context.Context, projectDir, outputDir string, artifacts []buildpkg.Artifact, signCfg signing.SignConfig, ciMode bool) ([]buildpkg.Artifact, error) {
	if !ciMode {
		fmt.Println()
		fmt.Printf("%s Computing checksums...\n", buildHeaderStyle.Render("Checksum:"))
	}

	checksummedArtifacts, err := buildpkg.ChecksumAll(artifacts)
	if err != nil {
		if !ciMode {
			fmt.Printf("%s Checksum failed: %v\n", buildErrorStyle.Render("Error:"), err)
		}
		return nil, err
	}

	// Write CHECKSUMS.txt
	checksumPath := filepath.Join(outputDir, "CHECKSUMS.txt")
	if err := buildpkg.WriteChecksumFile(checksummedArtifacts, checksumPath); err != nil {
		if !ciMode {
			fmt.Printf("%s Failed to write CHECKSUMS.txt: %v\n", buildErrorStyle.Render("Error:"), err)
		}
		return nil, err
	}

	// Sign checksums with GPG
	if signCfg.Enabled {
		if err := signing.SignChecksums(ctx, signCfg, checksumPath); err != nil {
			if !ciMode {
				fmt.Printf("%s GPG signing failed: %v\n", buildErrorStyle.Render("Error:"), err)
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
			return nil, fmt.Errorf("invalid target format %q, expected OS/arch (e.g., linux/amd64)", part)
		}

		targets = append(targets, buildpkg.Target{
			OS:   strings.TrimSpace(osArch[0]),
			Arch: strings.TrimSpace(osArch[1]),
		})
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no valid targets specified")
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
		return nil, fmt.Errorf("Node.js builder not yet implemented")
	case buildpkg.ProjectTypePHP:
		return nil, fmt.Errorf("PHP builder not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported project type: %s", projectType)
	}
}
