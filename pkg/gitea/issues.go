package gitea

import (
	"code.gitea.io/sdk/gitea"

	"github.com/host-uk/core/pkg/log"
)

// ListIssuesOpts configures issue listing.
type ListIssuesOpts struct {
	State string // "open", "closed", "all"
	Page  int
	Limit int
}

// ListIssues returns issues for the given repository.
func (c *Client) ListIssues(owner, repo string, opts ListIssuesOpts) ([]*gitea.Issue, error) {
	state := gitea.StateOpen
	switch opts.State {
	case "closed":
		state = gitea.StateClosed
	case "all":
		state = gitea.StateAll
	}

	limit := opts.Limit
	if limit == 0 {
		limit = 50
	}

	page := opts.Page
	if page == 0 {
		page = 1
	}

	issues, _, err := c.api.ListRepoIssues(owner, repo, gitea.ListIssueOption{
		ListOptions: gitea.ListOptions{Page: page, PageSize: limit},
		State:       state,
		Type:        gitea.IssueTypeIssue,
	})
	if err != nil {
		return nil, log.E("gitea.ListIssues", "failed to list issues", err)
	}

	return issues, nil
}

// GetIssue returns a single issue by number.
func (c *Client) GetIssue(owner, repo string, number int64) (*gitea.Issue, error) {
	issue, _, err := c.api.GetIssue(owner, repo, number)
	if err != nil {
		return nil, log.E("gitea.GetIssue", "failed to get issue", err)
	}

	return issue, nil
}

// CreateIssue creates a new issue in the given repository.
func (c *Client) CreateIssue(owner, repo string, opts gitea.CreateIssueOption) (*gitea.Issue, error) {
	issue, _, err := c.api.CreateIssue(owner, repo, opts)
	if err != nil {
		return nil, log.E("gitea.CreateIssue", "failed to create issue", err)
	}

	return issue, nil
}

// ListPullRequests returns pull requests for the given repository.
func (c *Client) ListPullRequests(owner, repo string, state string) ([]*gitea.PullRequest, error) {
	st := gitea.StateOpen
	switch state {
	case "closed":
		st = gitea.StateClosed
	case "all":
		st = gitea.StateAll
	}

	var all []*gitea.PullRequest
	page := 1

	for {
		prs, resp, err := c.api.ListRepoPullRequests(owner, repo, gitea.ListPullRequestsOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: 50},
			State:       st,
		})
		if err != nil {
			return nil, log.E("gitea.ListPullRequests", "failed to list pull requests", err)
		}

		all = append(all, prs...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// GetPullRequest returns a single pull request by number.
func (c *Client) GetPullRequest(owner, repo string, number int64) (*gitea.PullRequest, error) {
	pr, _, err := c.api.GetPullRequest(owner, repo, number)
	if err != nil {
		return nil, log.E("gitea.GetPullRequest", "failed to get pull request", err)
	}

	return pr, nil
}
