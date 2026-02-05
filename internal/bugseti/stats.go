// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// StatsService tracks user contribution statistics.
type StatsService struct {
	config *ConfigService
	stats  *Stats
	mu     sync.RWMutex
}

// Stats contains all tracked statistics.
type Stats struct {
	// Issue stats
	IssuesAttempted int `json:"issuesAttempted"`
	IssuesCompleted int `json:"issuesCompleted"`
	IssuesSkipped   int `json:"issuesSkipped"`

	// PR stats
	PRsSubmitted int `json:"prsSubmitted"`
	PRsMerged    int `json:"prsMerged"`
	PRsRejected  int `json:"prsRejected"`

	// Repository stats
	ReposContributed map[string]*RepoStats `json:"reposContributed"`

	// Streaks
	CurrentStreak int       `json:"currentStreak"`
	LongestStreak int       `json:"longestStreak"`
	LastActivity  time.Time `json:"lastActivity"`

	// Time tracking
	TotalTimeSpent   time.Duration `json:"totalTimeSpent"`
	AverageTimePerPR time.Duration `json:"averageTimePerPR"`

	// Activity history (last 30 days)
	DailyActivity map[string]*DayStats `json:"dailyActivity"`
}

// RepoStats contains statistics for a single repository.
type RepoStats struct {
	Name         string    `json:"name"`
	IssuesFixed  int       `json:"issuesFixed"`
	PRsSubmitted int       `json:"prsSubmitted"`
	PRsMerged    int       `json:"prsMerged"`
	FirstContrib time.Time `json:"firstContrib"`
	LastContrib  time.Time `json:"lastContrib"`
}

// DayStats contains statistics for a single day.
type DayStats struct {
	Date         string `json:"date"`
	IssuesWorked int    `json:"issuesWorked"`
	PRsSubmitted int    `json:"prsSubmitted"`
	TimeSpent    int    `json:"timeSpentMinutes"`
}

// NewStatsService creates a new StatsService.
func NewStatsService(config *ConfigService) *StatsService {
	s := &StatsService{
		config: config,
		stats: &Stats{
			ReposContributed: make(map[string]*RepoStats),
			DailyActivity:    make(map[string]*DayStats),
		},
	}
	s.load()
	return s
}

// ServiceName returns the service name for Wails.
func (s *StatsService) ServiceName() string {
	return "StatsService"
}

// GetStats returns a copy of the current statistics.
func (s *StatsService) GetStats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.stats
}

// RecordIssueAttempted records that an issue was started.
func (s *StatsService) RecordIssueAttempted(repo string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.IssuesAttempted++
	s.ensureRepo(repo)
	s.updateStreak()
	s.updateDailyActivity("issue")
	s.save()
}

// RecordIssueCompleted records that an issue was completed.
func (s *StatsService) RecordIssueCompleted(repo string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.IssuesCompleted++
	if rs, ok := s.stats.ReposContributed[repo]; ok {
		rs.IssuesFixed++
		rs.LastContrib = time.Now()
	}
	s.save()
}

// RecordIssueSkipped records that an issue was skipped.
func (s *StatsService) RecordIssueSkipped() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.IssuesSkipped++
	s.save()
}

// RecordPRSubmitted records that a PR was submitted.
func (s *StatsService) RecordPRSubmitted(repo string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.PRsSubmitted++
	if rs, ok := s.stats.ReposContributed[repo]; ok {
		rs.PRsSubmitted++
		rs.LastContrib = time.Now()
	}
	s.updateDailyActivity("pr")
	s.save()
}

// RecordPRMerged records that a PR was merged.
func (s *StatsService) RecordPRMerged(repo string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.PRsMerged++
	if rs, ok := s.stats.ReposContributed[repo]; ok {
		rs.PRsMerged++
	}
	s.save()
}

// RecordPRRejected records that a PR was rejected.
func (s *StatsService) RecordPRRejected() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.PRsRejected++
	s.save()
}

// RecordTimeSpent adds time spent on an issue.
func (s *StatsService) RecordTimeSpent(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.TotalTimeSpent += duration

	// Recalculate average
	if s.stats.PRsSubmitted > 0 {
		s.stats.AverageTimePerPR = s.stats.TotalTimeSpent / time.Duration(s.stats.PRsSubmitted)
	}

	// Update daily activity
	today := time.Now().Format("2006-01-02")
	if day, ok := s.stats.DailyActivity[today]; ok {
		day.TimeSpent += int(duration.Minutes())
	}

	s.save()
}

// GetRepoStats returns statistics for a specific repository.
func (s *StatsService) GetRepoStats(repo string) *RepoStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats.ReposContributed[repo]
}

// GetTopRepos returns the top N repositories by contributions.
func (s *StatsService) GetTopRepos(n int) []*RepoStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repos := make([]*RepoStats, 0, len(s.stats.ReposContributed))
	for _, rs := range s.stats.ReposContributed {
		repos = append(repos, rs)
	}

	// Sort by PRs merged (descending)
	for i := 0; i < len(repos)-1; i++ {
		for j := i + 1; j < len(repos); j++ {
			if repos[j].PRsMerged > repos[i].PRsMerged {
				repos[i], repos[j] = repos[j], repos[i]
			}
		}
	}

	if n > len(repos) {
		n = len(repos)
	}
	return repos[:n]
}

// GetActivityHistory returns the activity for the last N days.
func (s *StatsService) GetActivityHistory(days int) []*DayStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*DayStats, 0, days)
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		if day, ok := s.stats.DailyActivity[date]; ok {
			result = append(result, day)
		} else {
			result = append(result, &DayStats{Date: date})
		}
	}

	return result
}

// ensureRepo creates a repo stats entry if it doesn't exist.
func (s *StatsService) ensureRepo(repo string) {
	if _, ok := s.stats.ReposContributed[repo]; !ok {
		s.stats.ReposContributed[repo] = &RepoStats{
			Name:         repo,
			FirstContrib: time.Now(),
			LastContrib:  time.Now(),
		}
	}
}

// updateStreak updates the contribution streak.
func (s *StatsService) updateStreak() {
	now := time.Now()
	lastActivity := s.stats.LastActivity

	if lastActivity.IsZero() {
		s.stats.CurrentStreak = 1
	} else {
		daysSince := int(now.Sub(lastActivity).Hours() / 24)
		if daysSince <= 1 {
			// Same day or next day
			if daysSince == 1 || now.Day() != lastActivity.Day() {
				s.stats.CurrentStreak++
			}
		} else {
			// Streak broken
			s.stats.CurrentStreak = 1
		}
	}

	if s.stats.CurrentStreak > s.stats.LongestStreak {
		s.stats.LongestStreak = s.stats.CurrentStreak
	}

	s.stats.LastActivity = now
}

// updateDailyActivity updates today's activity.
func (s *StatsService) updateDailyActivity(activityType string) {
	today := time.Now().Format("2006-01-02")

	if _, ok := s.stats.DailyActivity[today]; !ok {
		s.stats.DailyActivity[today] = &DayStats{Date: today}
	}

	day := s.stats.DailyActivity[today]
	switch activityType {
	case "issue":
		day.IssuesWorked++
	case "pr":
		day.PRsSubmitted++
	}

	// Clean up old entries (keep last 90 days)
	cutoff := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
	for date := range s.stats.DailyActivity {
		if date < cutoff {
			delete(s.stats.DailyActivity, date)
		}
	}
}

// save persists stats to disk.
func (s *StatsService) save() {
	dataDir := s.config.GetDataDir()
	if dataDir == "" {
		return
	}

	path := filepath.Join(dataDir, "stats.json")
	data, err := json.MarshalIndent(s.stats, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal stats: %v", err)
		return
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("Failed to save stats: %v", err)
	}
}

// load restores stats from disk.
func (s *StatsService) load() {
	dataDir := s.config.GetDataDir()
	if dataDir == "" {
		return
	}

	path := filepath.Join(dataDir, "stats.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to read stats: %v", err)
		}
		return
	}

	var stats Stats
	if err := json.Unmarshal(data, &stats); err != nil {
		log.Printf("Failed to unmarshal stats: %v", err)
		return
	}

	// Ensure maps are initialized
	if stats.ReposContributed == nil {
		stats.ReposContributed = make(map[string]*RepoStats)
	}
	if stats.DailyActivity == nil {
		stats.DailyActivity = make(map[string]*DayStats)
	}

	s.stats = &stats
}

// Reset clears all statistics.
func (s *StatsService) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats = &Stats{
		ReposContributed: make(map[string]*RepoStats),
		DailyActivity:    make(map[string]*DayStats),
	}
	s.save()
	return nil
}
