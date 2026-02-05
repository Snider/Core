// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"container/heap"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// IssueStatus represents the status of an issue in the queue.
type IssueStatus string

const (
	StatusPending    IssueStatus = "pending"
	StatusClaimed    IssueStatus = "claimed"
	StatusInProgress IssueStatus = "in_progress"
	StatusCompleted  IssueStatus = "completed"
	StatusSkipped    IssueStatus = "skipped"
)

// Issue represents a GitHub issue in the queue.
type Issue struct {
	ID        string        `json:"id"`
	Number    int           `json:"number"`
	Repo      string        `json:"repo"`
	Title     string        `json:"title"`
	Body      string        `json:"body"`
	URL       string        `json:"url"`
	Labels    []string      `json:"labels"`
	Author    string        `json:"author"`
	CreatedAt time.Time     `json:"createdAt"`
	Priority  int           `json:"priority"`
	Status    IssueStatus   `json:"status"`
	ClaimedAt time.Time     `json:"claimedAt,omitempty"`
	Context   *IssueContext `json:"context,omitempty"`
	Comments  []Comment     `json:"comments,omitempty"`
	index     int           // For heap interface
}

// Comment represents a comment on an issue.
type Comment struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}

// IssueContext contains AI-prepared context for an issue.
type IssueContext struct {
	Summary       string    `json:"summary"`
	RelevantFiles []string  `json:"relevantFiles"`
	SuggestedFix  string    `json:"suggestedFix"`
	RelatedIssues []string  `json:"relatedIssues"`
	Complexity    string    `json:"complexity"`
	EstimatedTime string    `json:"estimatedTime"`
	PreparedAt    time.Time `json:"preparedAt"`
}

// QueueService manages the priority queue of issues.
type QueueService struct {
	config  *ConfigService
	issues  issueHeap
	seen    map[string]bool
	current *Issue
	mu      sync.RWMutex
}

// issueHeap implements heap.Interface for priority queue.
type issueHeap []*Issue

func (h issueHeap) Len() int           { return len(h) }
func (h issueHeap) Less(i, j int) bool { return h[i].Priority > h[j].Priority } // Higher priority first
func (h issueHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *issueHeap) Push(x any) {
	n := len(*h)
	item := x.(*Issue)
	item.index = n
	*h = append(*h, item)
}

func (h *issueHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

// NewQueueService creates a new QueueService.
func NewQueueService(config *ConfigService) *QueueService {
	q := &QueueService{
		config: config,
		issues: make(issueHeap, 0),
		seen:   make(map[string]bool),
	}
	heap.Init(&q.issues)
	q.load() // Load persisted queue
	return q
}

// ServiceName returns the service name for Wails.
func (q *QueueService) ServiceName() string {
	return "QueueService"
}

// Add adds issues to the queue, deduplicating by ID.
func (q *QueueService) Add(issues []*Issue) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	added := 0
	for _, issue := range issues {
		if q.seen[issue.ID] {
			continue
		}
		q.seen[issue.ID] = true
		issue.Status = StatusPending
		heap.Push(&q.issues, issue)
		added++
	}

	if added > 0 {
		q.save()
	}
	return added
}

// Size returns the number of issues in the queue.
func (q *QueueService) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.issues)
}

// CurrentIssue returns the issue currently being worked on.
func (q *QueueService) CurrentIssue() *Issue {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.current
}

// Next claims and returns the next issue from the queue.
func (q *QueueService) Next() *Issue {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.issues) == 0 {
		return nil
	}

	// Pop the highest priority issue
	issue := heap.Pop(&q.issues).(*Issue)
	issue.Status = StatusClaimed
	issue.ClaimedAt = time.Now()
	q.current = issue
	q.save()
	return issue
}

// Skip marks the current issue as skipped and moves to the next.
func (q *QueueService) Skip() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.current != nil {
		q.current.Status = StatusSkipped
		q.current = nil
		q.save()
	}
}

// Complete marks the current issue as completed.
func (q *QueueService) Complete() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.current != nil {
		q.current.Status = StatusCompleted
		q.current = nil
		q.save()
	}
}

// SetInProgress marks the current issue as in progress.
func (q *QueueService) SetInProgress() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.current != nil {
		q.current.Status = StatusInProgress
		q.save()
	}
}

// SetContext sets the AI-prepared context for the current issue.
func (q *QueueService) SetContext(ctx *IssueContext) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.current != nil {
		q.current.Context = ctx
		q.save()
	}
}

// GetPending returns all pending issues.
func (q *QueueService) GetPending() []*Issue {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]*Issue, 0, len(q.issues))
	for _, issue := range q.issues {
		if issue.Status == StatusPending {
			result = append(result, issue)
		}
	}
	return result
}

// Clear removes all issues from the queue.
func (q *QueueService) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.issues = make(issueHeap, 0)
	q.seen = make(map[string]bool)
	q.current = nil
	heap.Init(&q.issues)
	q.save()
}

// queueState represents the persisted queue state.
type queueState struct {
	Issues  []*Issue `json:"issues"`
	Current *Issue   `json:"current"`
	Seen    []string `json:"seen"`
}

// save persists the queue to disk.
func (q *QueueService) save() {
	dataDir := q.config.GetDataDir()
	if dataDir == "" {
		return
	}

	path := filepath.Join(dataDir, "queue.json")

	seen := make([]string, 0, len(q.seen))
	for id := range q.seen {
		seen = append(seen, id)
	}

	state := queueState{
		Issues:  []*Issue(q.issues),
		Current: q.current,
		Seen:    seen,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal queue: %v", err)
		return
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("Failed to save queue: %v", err)
	}
}

// load restores the queue from disk.
func (q *QueueService) load() {
	dataDir := q.config.GetDataDir()
	if dataDir == "" {
		return
	}

	path := filepath.Join(dataDir, "queue.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to read queue: %v", err)
		}
		return
	}

	var state queueState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("Failed to unmarshal queue: %v", err)
		return
	}

	q.issues = state.Issues
	heap.Init(&q.issues)
	q.current = state.Current
	q.seen = make(map[string]bool)
	for _, id := range state.Seen {
		q.seen[id] = true
	}
}
