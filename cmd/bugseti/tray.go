// Package main provides the BugSETI system tray application.
package main

import (
	"context"
	"log"

	"github.com/host-uk/core/internal/bugseti"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// TrayService provides system tray bindings for the frontend.
type TrayService struct {
	app     *application.App
	fetcher *bugseti.FetcherService
	queue   *bugseti.QueueService
	config  *bugseti.ConfigService
	stats   *bugseti.StatsService
}

// NewTrayService creates a new TrayService instance.
func NewTrayService(app *application.App) *TrayService {
	return &TrayService{
		app: app,
	}
}

// SetServices sets the service references after initialization.
func (t *TrayService) SetServices(fetcher *bugseti.FetcherService, queue *bugseti.QueueService, config *bugseti.ConfigService, stats *bugseti.StatsService) {
	t.fetcher = fetcher
	t.queue = queue
	t.config = config
	t.stats = stats
}

// ServiceName returns the service name for Wails.
func (t *TrayService) ServiceName() string {
	return "TrayService"
}

// ServiceStartup is called when the Wails application starts.
func (t *TrayService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	log.Println("TrayService started")
	return nil
}

// ServiceShutdown is called when the Wails application shuts down.
func (t *TrayService) ServiceShutdown() error {
	log.Println("TrayService shutdown")
	return nil
}

// TrayStatus represents the current status of the tray.
type TrayStatus struct {
	Running      bool   `json:"running"`
	CurrentIssue string `json:"currentIssue"`
	QueueSize    int    `json:"queueSize"`
	IssuesFixed  int    `json:"issuesFixed"`
	PRsMerged    int    `json:"prsMerged"`
}

// GetStatus returns the current tray status.
func (t *TrayService) GetStatus() TrayStatus {
	var currentIssue string
	if t.queue != nil {
		if issue := t.queue.CurrentIssue(); issue != nil {
			currentIssue = issue.Title
		}
	}

	var queueSize int
	if t.queue != nil {
		queueSize = t.queue.Size()
	}

	var running bool
	if t.fetcher != nil {
		running = t.fetcher.IsRunning()
	}

	var issuesFixed, prsMerged int
	if t.stats != nil {
		stats := t.stats.GetStats()
		issuesFixed = stats.IssuesAttempted
		prsMerged = stats.PRsMerged
	}

	return TrayStatus{
		Running:      running,
		CurrentIssue: currentIssue,
		QueueSize:    queueSize,
		IssuesFixed:  issuesFixed,
		PRsMerged:    prsMerged,
	}
}

// StartFetching starts the issue fetcher.
func (t *TrayService) StartFetching() error {
	if t.fetcher == nil {
		return nil
	}
	return t.fetcher.Start()
}

// PauseFetching pauses the issue fetcher.
func (t *TrayService) PauseFetching() {
	if t.fetcher != nil {
		t.fetcher.Pause()
	}
}

// GetCurrentIssue returns the current issue being worked on.
func (t *TrayService) GetCurrentIssue() *bugseti.Issue {
	if t.queue == nil {
		return nil
	}
	return t.queue.CurrentIssue()
}

// NextIssue moves to the next issue in the queue.
func (t *TrayService) NextIssue() *bugseti.Issue {
	if t.queue == nil {
		return nil
	}
	return t.queue.Next()
}

// SkipIssue skips the current issue.
func (t *TrayService) SkipIssue() {
	if t.queue == nil {
		return
	}
	t.queue.Skip()
}

// ShowWindow shows a specific window by name.
func (t *TrayService) ShowWindow(name string) {
	if t.app == nil {
		return
	}
	// Window will be shown by the frontend via Wails runtime
}

// IsOnboarded returns whether the user has completed onboarding.
func (t *TrayService) IsOnboarded() bool {
	if t.config == nil {
		return false
	}
	return t.config.IsOnboarded()
}

// CompleteOnboarding marks onboarding as complete.
func (t *TrayService) CompleteOnboarding() error {
	if t.config == nil {
		return nil
	}
	return t.config.CompleteOnboarding()
}
