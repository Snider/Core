package updater

import (
	"context"
)

// MockGithubClient is a mock implementation of the GithubClient interface for testing.
type MockGithubClient struct {
	GetLatestReleaseFunc        func(ctx context.Context, owner, repo, channel string) (*Release, error)
	GetReleaseByPullRequestFunc func(ctx context.Context, owner, repo string, prNumber int) (*Release, error)
	GetPublicReposFunc          func(ctx context.Context, userOrOrg string) ([]string, error)
}

// GetLatestRelease mocks the GetLatestRelease method of the GithubClient interface.
func (m *MockGithubClient) GetLatestRelease(ctx context.Context, owner, repo, channel string) (*Release, error) {
	if m.GetLatestReleaseFunc != nil {
		return m.GetLatestReleaseFunc(ctx, owner, repo, channel)
	}
	return nil, nil
}

// GetReleaseByPullRequest mocks the GetReleaseByPullRequest method of the GithubClient interface.
func (m *MockGithubClient) GetReleaseByPullRequest(ctx context.Context, owner, repo string, prNumber int) (*Release, error) {
	if m.GetReleaseByPullRequestFunc != nil {
		return m.GetReleaseByPullRequestFunc(ctx, owner, repo, prNumber)
	}
	return nil, nil
}

// GetPublicRepos mocks the GetPublicRepos method of the GithubClient interface.
func (m *MockGithubClient) GetPublicRepos(ctx context.Context, userOrOrg string) ([]string, error) {
	if m.GetPublicReposFunc != nil {
		return m.GetPublicReposFunc(ctx, userOrOrg)
	}
	return []string{"repo1", "repo2"}, nil
}
