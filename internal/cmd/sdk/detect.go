package sdk

import (
	"fmt"
	"path/filepath"
	"strings"

	coreio "github.com/host-uk/core/pkg/io"
)

// commonSpecPaths are checked in order when no spec is configured.
var commonSpecPaths = []string{
	"api/openapi.yaml",
	"api/openapi.json",
	"openapi.yaml",
	"openapi.json",
	"docs/api.yaml",
	"docs/api.json",
	"swagger.yaml",
	"swagger.json",
}

// DetectSpec finds the OpenAPI spec file.
// Priority: config path -> common paths -> Laravel Scramble.
func (s *SDK) DetectSpec() (string, error) {
	// 1. Check configured path
	if s.config.Spec != "" {
		specPath := filepath.Join(s.projectDir, s.config.Spec)
		if coreio.Local.IsFile(specPath) {
			return specPath, nil
		}
		return "", fmt.Errorf("sdk.DetectSpec: configured spec not found: %s", s.config.Spec)
	}

	// 2. Check common paths
	for _, p := range commonSpecPaths {
		specPath := filepath.Join(s.projectDir, p)
		if coreio.Local.IsFile(specPath) {
			return specPath, nil
		}
	}

	// 3. Try Laravel Scramble detection
	specPath, err := s.detectScramble()
	if err == nil {
		return specPath, nil
	}

	return "", fmt.Errorf("sdk.DetectSpec: no OpenAPI spec found (checked config, common paths, Scramble)")
}

// detectScramble checks for Laravel Scramble and exports the spec.
func (s *SDK) detectScramble() (string, error) {
	composerPath := filepath.Join(s.projectDir, "composer.json")
	if !coreio.Local.IsFile(composerPath) {
		return "", fmt.Errorf("no composer.json")
	}

	// Check for scramble in composer.json
	data, err := coreio.Local.Read(composerPath)
	if err != nil {
		return "", err
	}

	// Simple check for scramble package
	if !containsScramble(data) {
		return "", fmt.Errorf("scramble not found in composer.json")
	}

	// TODO: Run php artisan scramble:export
	return "", fmt.Errorf("scramble export not implemented")
}

// containsScramble checks if composer.json includes scramble.
func containsScramble(content string) bool {
	return strings.Contains(content, "dedoc/scramble") ||
		strings.Contains(content, "\"scramble\"")
}
