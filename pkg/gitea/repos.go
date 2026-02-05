package gitea

import (
	"code.gitea.io/sdk/gitea"

	"github.com/host-uk/core/pkg/log"
)

// ListOrgRepos returns all repositories for the given organisation.
func (c *Client) ListOrgRepos(org string) ([]*gitea.Repository, error) {
	var all []*gitea.Repository
	page := 1

	for {
		repos, resp, err := c.api.ListOrgRepos(org, gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("gitea.ListOrgRepos", "failed to list org repos", err)
		}

		all = append(all, repos...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// ListUserRepos returns all repositories for the authenticated user.
func (c *Client) ListUserRepos() ([]*gitea.Repository, error) {
	var all []*gitea.Repository
	page := 1

	for {
		repos, resp, err := c.api.ListMyRepos(gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("gitea.ListUserRepos", "failed to list user repos", err)
		}

		all = append(all, repos...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// GetRepo returns a single repository by owner and name.
func (c *Client) GetRepo(owner, name string) (*gitea.Repository, error) {
	repo, _, err := c.api.GetRepo(owner, name)
	if err != nil {
		return nil, log.E("gitea.GetRepo", "failed to get repo", err)
	}

	return repo, nil
}

// CreateMirror creates a mirror repository on Gitea from a GitHub clone URL.
// This uses the Gitea migration API to set up a pull mirror.
// If authToken is provided, it is used to authenticate against the source (e.g. for private GitHub repos).
func (c *Client) CreateMirror(owner, name, cloneURL, authToken string) (*gitea.Repository, error) {
	opts := gitea.MigrateRepoOption{
		RepoName:    name,
		RepoOwner:   owner,
		CloneAddr:   cloneURL,
		Service:     gitea.GitServiceGithub,
		Mirror:      true,
		Description: "Mirror of " + cloneURL,
	}

	if authToken != "" {
		opts.AuthToken = authToken
	}

	repo, _, err := c.api.MigrateRepo(opts)
	if err != nil {
		return nil, log.E("gitea.CreateMirror", "failed to create mirror", err)
	}

	return repo, nil
}

// DeleteRepo deletes a repository from Gitea.
func (c *Client) DeleteRepo(owner, name string) error {
	_, err := c.api.DeleteRepo(owner, name)
	if err != nil {
		return log.E("gitea.DeleteRepo", "failed to delete repo", err)
	}

	return nil
}

// CreateOrgRepo creates a new empty repository under an organisation.
func (c *Client) CreateOrgRepo(org string, opts gitea.CreateRepoOption) (*gitea.Repository, error) {
	repo, _, err := c.api.CreateOrgRepo(org, opts)
	if err != nil {
		return nil, log.E("gitea.CreateOrgRepo", "failed to create org repo", err)
	}

	return repo, nil
}
