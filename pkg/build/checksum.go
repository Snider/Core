// Package build provides project type detection and cross-compilation for the Core build system.
package build

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Checksum computes SHA256 for an artifact and returns the artifact with the Checksum field filled.
func Checksum(artifact Artifact) (Artifact, error) {
	if artifact.Path == "" {
		return Artifact{}, fmt.Errorf("build.Checksum: artifact path is empty")
	}

	// Open the file
	file, err := os.Open(artifact.Path)
	if err != nil {
		return Artifact{}, fmt.Errorf("build.Checksum: failed to open file: %w", err)
	}
	defer file.Close()

	// Compute SHA256 hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return Artifact{}, fmt.Errorf("build.Checksum: failed to hash file: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))

	return Artifact{
		Path:     artifact.Path,
		OS:       artifact.OS,
		Arch:     artifact.Arch,
		Checksum: checksum,
	}, nil
}

// ChecksumAll computes checksums for all artifacts.
// Returns a slice of artifacts with their Checksum fields filled.
func ChecksumAll(artifacts []Artifact) ([]Artifact, error) {
	if len(artifacts) == 0 {
		return nil, nil
	}

	var checksummed []Artifact
	for _, artifact := range artifacts {
		cs, err := Checksum(artifact)
		if err != nil {
			return checksummed, fmt.Errorf("build.ChecksumAll: failed to checksum %s: %w", artifact.Path, err)
		}
		checksummed = append(checksummed, cs)
	}

	return checksummed, nil
}

// WriteChecksumFile writes a CHECKSUMS.txt file with the format:
//
//	sha256hash  filename1
//	sha256hash  filename2
//
// The artifacts should have their Checksum fields filled (call ChecksumAll first).
// Filenames are relative to the output directory (just the basename).
func WriteChecksumFile(artifacts []Artifact, path string) error {
	if len(artifacts) == 0 {
		return nil
	}

	// Build the content
	var lines []string
	for _, artifact := range artifacts {
		if artifact.Checksum == "" {
			return fmt.Errorf("build.WriteChecksumFile: artifact %s has no checksum", artifact.Path)
		}
		filename := filepath.Base(artifact.Path)
		lines = append(lines, fmt.Sprintf("%s  %s", artifact.Checksum, filename))
	}

	// Sort lines for consistent output
	sort.Strings(lines)

	content := strings.Join(lines, "\n") + "\n"

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("build.WriteChecksumFile: failed to create directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("build.WriteChecksumFile: failed to write file: %w", err)
	}

	return nil
}
