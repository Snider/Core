package gitea

import (
	"time"

	"code.gitea.io/sdk/gitea"

	"github.com/host-uk/core/pkg/log"
)

// PRMeta holds structural signals from a pull request,
// used by the pipeline MetaReader for AI-driven workflows.
type PRMeta struct {
	Number       int64
	Title        string
	State        string
	Author       string
	Branch       string
	BaseBranch   string
	Labels       []string
	Assignees    []string
	IsMerged     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CommentCount int
}

// Comment represents a comment with metadata.
type Comment struct {
	ID        int64
	Author    string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const commentPageSize = 50

// GetPRMeta returns structural signals for a pull request.
// This is the Gitea side of the dual MetaReader described in the pipeline design.
func (c *Client) GetPRMeta(owner, repo string, pr int64) (*PRMeta, error) {
	pull, _, err := c.api.GetPullRequest(owner, repo, pr)
	if err != nil {
		return nil, log.E("gitea.GetPRMeta", "failed to get PR metadata", err)
	}

	meta := &PRMeta{
		Number:     pull.Index,
		Title:      pull.Title,
		State:      string(pull.State),
		Branch:     pull.Head.Ref,
		BaseBranch: pull.Base.Ref,
		IsMerged:   pull.HasMerged,
	}

	if pull.Created != nil {
		meta.CreatedAt = *pull.Created
	}
	if pull.Updated != nil {
		meta.UpdatedAt = *pull.Updated
	}

	if pull.Poster != nil {
		meta.Author = pull.Poster.UserName
	}

	for _, label := range pull.Labels {
		meta.Labels = append(meta.Labels, label.Name)
	}

	for _, assignee := range pull.Assignees {
		meta.Assignees = append(meta.Assignees, assignee.UserName)
	}

	// Fetch comment count from the issue side (PRs are issues in Gitea).
	// Paginate to get an accurate count.
	count := 0
	page := 1
	for {
		comments, _, listErr := c.api.ListIssueComments(owner, repo, pr, gitea.ListIssueCommentOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: commentPageSize},
		})
		if listErr != nil {
			break
		}
		count += len(comments)
		if len(comments) < commentPageSize {
			break
		}
		page++
	}
	meta.CommentCount = count

	return meta, nil
}

// GetCommentBodies returns all comment bodies for a pull request.
// This reads full content, which is safe on the home lab Gitea instance.
func (c *Client) GetCommentBodies(owner, repo string, pr int64) ([]Comment, error) {
	var comments []Comment
	page := 1

	for {
		raw, _, err := c.api.ListIssueComments(owner, repo, pr, gitea.ListIssueCommentOptions{
			ListOptions: gitea.ListOptions{Page: page, PageSize: commentPageSize},
		})
		if err != nil {
			return nil, log.E("gitea.GetCommentBodies", "failed to get PR comments", err)
		}

		if len(raw) == 0 {
			break
		}

		for _, rc := range raw {
			comment := Comment{
				ID:        rc.ID,
				Body:      rc.Body,
				CreatedAt: rc.Created,
				UpdatedAt: rc.Updated,
			}
			if rc.Poster != nil {
				comment.Author = rc.Poster.UserName
			}
			comments = append(comments, comment)
		}

		if len(raw) < commentPageSize {
			break
		}
		page++
	}

	return comments, nil
}

// GetIssueBody returns the body text of an issue.
// This reads full content, which is safe on the home lab Gitea instance.
func (c *Client) GetIssueBody(owner, repo string, issue int64) (string, error) {
	iss, _, err := c.api.GetIssue(owner, repo, issue)
	if err != nil {
		return "", log.E("gitea.GetIssueBody", "failed to get issue body", err)
	}

	return iss.Body, nil
}
