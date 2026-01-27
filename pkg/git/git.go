// Package git provides utilities for git operations across multiple repositories.
package git

import (
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// RepoStatus represents the git status of a single repository.
type RepoStatus struct {
	Name      string
	Path      string
	Modified  int
	Untracked int
	Staged    int
	Ahead     int
	Behind    int
	Branch    string
	Error     error
}

// IsDirty returns true if there are uncommitted changes.
func (s *RepoStatus) IsDirty() bool {
	return s.Modified > 0 || s.Untracked > 0 || s.Staged > 0
}

// HasUnpushed returns true if there are commits to push.
func (s *RepoStatus) HasUnpushed() bool {
	return s.Ahead > 0
}

// HasUnpulled returns true if there are commits to pull.
func (s *RepoStatus) HasUnpulled() bool {
	return s.Behind > 0
}

// StatusOptions configures the status check.
type StatusOptions struct {
	// Paths is a list of repo paths to check
	Paths []string
	// Names maps paths to display names
	Names map[string]string
}

// Status checks git status for multiple repositories in parallel.
func Status(ctx context.Context, opts StatusOptions) []RepoStatus {
	var wg sync.WaitGroup
	results := make([]RepoStatus, len(opts.Paths))

	for i, path := range opts.Paths {
		wg.Add(1)
		go func(idx int, repoPath string) {
			defer wg.Done()
			name := opts.Names[repoPath]
			if name == "" {
				name = repoPath
			}
			results[idx] = getStatus(ctx, repoPath, name)
		}(i, path)
	}

	wg.Wait()
	return results
}

// getStatus gets the git status for a single repository.
func getStatus(ctx context.Context, path, name string) RepoStatus {
	status := RepoStatus{
		Name: name,
		Path: path,
	}

	// Get current branch
	branch, err := gitCommand(ctx, path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		status.Error = err
		return status
	}
	status.Branch = strings.TrimSpace(branch)

	// Get porcelain status
	porcelain, err := gitCommand(ctx, path, "status", "--porcelain")
	if err != nil {
		status.Error = err
		return status
	}

	// Parse status output
	for _, line := range strings.Split(porcelain, "\n") {
		if len(line) < 2 {
			continue
		}
		x, y := line[0], line[1]

		// Untracked
		if x == '?' && y == '?' {
			status.Untracked++
			continue
		}

		// Staged (index has changes)
		if x == 'A' || x == 'D' || x == 'R' || x == 'M' {
			status.Staged++
		}

		// Modified in working tree
		if y == 'M' || y == 'D' {
			status.Modified++
		}
	}

	// Get ahead/behind counts
	ahead, behind := getAheadBehind(ctx, path)
	status.Ahead = ahead
	status.Behind = behind

	return status
}

// getAheadBehind returns the number of commits ahead and behind upstream.
func getAheadBehind(ctx context.Context, path string) (ahead, behind int) {
	// Try to get ahead count
	aheadStr, err := gitCommand(ctx, path, "rev-list", "--count", "@{u}..HEAD")
	if err == nil {
		ahead, _ = strconv.Atoi(strings.TrimSpace(aheadStr))
	}

	// Try to get behind count
	behindStr, err := gitCommand(ctx, path, "rev-list", "--count", "HEAD..@{u}")
	if err == nil {
		behind, _ = strconv.Atoi(strings.TrimSpace(behindStr))
	}

	return ahead, behind
}

// Push pushes commits for a single repository.
func Push(ctx context.Context, path string) error {
	_, err := gitCommand(ctx, path, "push")
	return err
}

// PushResult represents the result of a push operation.
type PushResult struct {
	Name    string
	Path    string
	Success bool
	Error   error
}

// PushMultiple pushes multiple repositories sequentially.
// Sequential because SSH passphrase prompts need user interaction.
func PushMultiple(ctx context.Context, paths []string, names map[string]string) []PushResult {
	results := make([]PushResult, len(paths))

	for i, path := range paths {
		name := names[path]
		if name == "" {
			name = path
		}

		result := PushResult{
			Name: name,
			Path: path,
		}

		err := Push(ctx, path)
		if err != nil {
			result.Error = err
		} else {
			result.Success = true
		}

		results[i] = result
	}

	return results
}

// gitCommand runs a git command and returns stdout.
func gitCommand(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return stdout.String(), nil
}
