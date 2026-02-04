package collect

import (
	"context"
	"testing"
	"time"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestGitHubCollector_Name_Good(t *testing.T) {
	g := &GitHubCollector{Org: "host-uk", Repo: "core"}
	assert.Equal(t, "github:host-uk/core", g.Name())
}

func TestGitHubCollector_Name_Good_OrgOnly(t *testing.T) {
	g := &GitHubCollector{Org: "host-uk"}
	assert.Equal(t, "github:host-uk", g.Name())
}

func TestGitHubCollector_Collect_Good_DryRun(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	var progressEmitted bool
	cfg.Dispatcher.On(EventProgress, func(e Event) {
		progressEmitted = true
	})

	g := &GitHubCollector{Org: "host-uk", Repo: "core"}
	result, err := g.Collect(context.Background(), cfg)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.Items)
	assert.True(t, progressEmitted, "Should emit progress event in dry-run mode")
}

func TestGitHubCollector_Collect_Good_DryRun_IssuesOnly(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	g := &GitHubCollector{Org: "test-org", Repo: "test-repo", IssuesOnly: true}
	result, err := g.Collect(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Items)
}

func TestGitHubCollector_Collect_Good_DryRun_PRsOnly(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	g := &GitHubCollector{Org: "test-org", Repo: "test-repo", PRsOnly: true}
	result, err := g.Collect(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Items)
}

func TestFormatIssueMarkdown_Good(t *testing.T) {
	issue := ghIssue{
		Number:    42,
		Title:     "Test Issue",
		State:     "open",
		Author:    ghAuthor{Login: "testuser"},
		Body:      "This is the body.",
		CreatedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		Labels: []ghLabel{
			{Name: "bug"},
			{Name: "priority"},
		},
		URL: "https://github.com/test/repo/issues/42",
	}

	md := formatIssueMarkdown(issue)

	assert.Contains(t, md, "# Test Issue")
	assert.Contains(t, md, "**Number:** #42")
	assert.Contains(t, md, "**State:** open")
	assert.Contains(t, md, "**Author:** testuser")
	assert.Contains(t, md, "**Labels:** bug, priority")
	assert.Contains(t, md, "This is the body.")
	assert.Contains(t, md, "**URL:** https://github.com/test/repo/issues/42")
}

func TestFormatIssueMarkdown_Good_NoLabels(t *testing.T) {
	issue := ghIssue{
		Number: 1,
		Title:  "Simple",
		State:  "closed",
		Author: ghAuthor{Login: "user"},
	}

	md := formatIssueMarkdown(issue)

	assert.Contains(t, md, "# Simple")
	assert.NotContains(t, md, "**Labels:**")
}
