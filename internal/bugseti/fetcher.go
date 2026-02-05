// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// FetcherService fetches issues from configured OSS repositories.
type FetcherService struct {
	config   *ConfigService
	notify   *NotifyService
	running  bool
	mu       sync.RWMutex
	stopCh   chan struct{}
	issuesCh chan []*Issue
}

// NewFetcherService creates a new FetcherService.
func NewFetcherService(config *ConfigService, notify *NotifyService) *FetcherService {
	return &FetcherService{
		config:   config,
		notify:   notify,
		issuesCh: make(chan []*Issue, 10),
	}
}

// ServiceName returns the service name for Wails.
func (f *FetcherService) ServiceName() string {
	return "FetcherService"
}

// Start begins fetching issues from configured repositories.
func (f *FetcherService) Start() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.running {
		return nil
	}

	f.running = true
	f.stopCh = make(chan struct{})

	go f.fetchLoop()
	log.Println("FetcherService started")
	return nil
}

// Pause stops fetching issues.
func (f *FetcherService) Pause() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.running {
		return
	}

	f.running = false
	close(f.stopCh)
	log.Println("FetcherService paused")
}

// IsRunning returns whether the fetcher is actively running.
func (f *FetcherService) IsRunning() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.running
}

// Issues returns a channel that receives batches of fetched issues.
func (f *FetcherService) Issues() <-chan []*Issue {
	return f.issuesCh
}

// fetchLoop periodically fetches issues from all configured repositories.
func (f *FetcherService) fetchLoop() {
	// Initial fetch
	f.fetchAll()

	// Set up ticker for periodic fetching
	interval := f.config.GetFetchInterval()
	if interval < time.Minute {
		interval = 15 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-f.stopCh:
			return
		case <-ticker.C:
			// Check if within work hours
			if f.config.IsWithinWorkHours() {
				f.fetchAll()
			}
		}
	}
}

// fetchAll fetches issues from all configured repositories.
func (f *FetcherService) fetchAll() {
	repos := f.config.GetWatchedRepos()
	if len(repos) == 0 {
		log.Println("No repositories configured")
		return
	}

	var allIssues []*Issue
	for _, repo := range repos {
		issues, err := f.fetchFromRepo(repo)
		if err != nil {
			log.Printf("Error fetching from %s: %v", repo, err)
			continue
		}
		allIssues = append(allIssues, issues...)
	}

	if len(allIssues) > 0 {
		select {
		case f.issuesCh <- allIssues:
			f.notify.Notify("BugSETI", fmt.Sprintf("Found %d new issues", len(allIssues)))
		default:
			// Channel full, skip
		}
	}
}

// fetchFromRepo fetches issues from a single repository using GitHub CLI.
func (f *FetcherService) fetchFromRepo(repo string) ([]*Issue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build query for good first issues
	labels := f.config.GetLabels()
	if len(labels) == 0 {
		labels = []string{"good first issue", "help wanted", "beginner-friendly"}
	}

	labelQuery := strings.Join(labels, ",")

	// Use gh CLI to fetch issues
	cmd := exec.CommandContext(ctx, "gh", "issue", "list",
		"--repo", repo,
		"--label", labelQuery,
		"--state", "open",
		"--limit", "20",
		"--json", "number,title,body,url,labels,createdAt,author")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh issue list failed: %w", err)
	}

	var ghIssues []struct {
		Number    int       `json:"number"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		URL       string    `json:"url"`
		CreatedAt time.Time `json:"createdAt"`
		Author    struct {
			Login string `json:"login"`
		} `json:"author"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}

	if err := json.Unmarshal(output, &ghIssues); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	issues := make([]*Issue, 0, len(ghIssues))
	for _, gi := range ghIssues {
		labels := make([]string, len(gi.Labels))
		for i, l := range gi.Labels {
			labels[i] = l.Name
		}

		issues = append(issues, &Issue{
			ID:        fmt.Sprintf("%s#%d", repo, gi.Number),
			Number:    gi.Number,
			Repo:      repo,
			Title:     gi.Title,
			Body:      gi.Body,
			URL:       gi.URL,
			Labels:    labels,
			Author:    gi.Author.Login,
			CreatedAt: gi.CreatedAt,
			Priority:  calculatePriority(labels),
		})
	}

	return issues, nil
}

// FetchIssue fetches a single issue by repo and number.
func (f *FetcherService) FetchIssue(repo string, number int) (*Issue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", "issue", "view",
		"--repo", repo,
		fmt.Sprintf("%d", number),
		"--json", "number,title,body,url,labels,createdAt,author,comments")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh issue view failed: %w", err)
	}

	var ghIssue struct {
		Number    int       `json:"number"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		URL       string    `json:"url"`
		CreatedAt time.Time `json:"createdAt"`
		Author    struct {
			Login string `json:"login"`
		} `json:"author"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Comments []struct {
			Body   string `json:"body"`
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
		} `json:"comments"`
	}

	if err := json.Unmarshal(output, &ghIssue); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	labels := make([]string, len(ghIssue.Labels))
	for i, l := range ghIssue.Labels {
		labels[i] = l.Name
	}

	comments := make([]Comment, len(ghIssue.Comments))
	for i, c := range ghIssue.Comments {
		comments[i] = Comment{
			Author: c.Author.Login,
			Body:   c.Body,
		}
	}

	return &Issue{
		ID:        fmt.Sprintf("%s#%d", repo, ghIssue.Number),
		Number:    ghIssue.Number,
		Repo:      repo,
		Title:     ghIssue.Title,
		Body:      ghIssue.Body,
		URL:       ghIssue.URL,
		Labels:    labels,
		Author:    ghIssue.Author.Login,
		CreatedAt: ghIssue.CreatedAt,
		Priority:  calculatePriority(labels),
		Comments:  comments,
	}, nil
}

// calculatePriority assigns a priority score based on labels.
func calculatePriority(labels []string) int {
	priority := 50 // Default priority

	for _, label := range labels {
		lower := strings.ToLower(label)
		switch {
		case strings.Contains(lower, "good first issue"):
			priority += 30
		case strings.Contains(lower, "help wanted"):
			priority += 20
		case strings.Contains(lower, "beginner"):
			priority += 25
		case strings.Contains(lower, "easy"):
			priority += 20
		case strings.Contains(lower, "bug"):
			priority += 10
		case strings.Contains(lower, "documentation"):
			priority += 5
		case strings.Contains(lower, "priority"):
			priority += 15
		}
	}

	return priority
}
