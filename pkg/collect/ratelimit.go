package collect

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	core "github.com/host-uk/core/pkg/framework/core"
)

// RateLimiter tracks per-source rate limiting to avoid overwhelming APIs.
type RateLimiter struct {
	mu     sync.Mutex
	delays map[string]time.Duration
	last   map[string]time.Time
}

// Default rate limit delays per source.
var defaultDelays = map[string]time.Duration{
	"github":      500 * time.Millisecond,
	"bitcointalk": 2 * time.Second,
	"coingecko":   1500 * time.Millisecond,
	"iacr":        1 * time.Second,
	"arxiv":       1 * time.Second,
}

// NewRateLimiter creates a limiter with default delays.
func NewRateLimiter() *RateLimiter {
	delays := make(map[string]time.Duration, len(defaultDelays))
	for k, v := range defaultDelays {
		delays[k] = v
	}
	return &RateLimiter{
		delays: delays,
		last:   make(map[string]time.Time),
	}
}

// Wait blocks until the rate limit allows the next request for the given source.
// It respects context cancellation.
func (r *RateLimiter) Wait(ctx context.Context, source string) error {
	r.mu.Lock()
	delay, ok := r.delays[source]
	if !ok {
		delay = 500 * time.Millisecond
	}
	lastTime := r.last[source]

	elapsed := time.Since(lastTime)
	if elapsed >= delay {
		// Enough time has passed — claim the slot immediately.
		r.last[source] = time.Now()
		r.mu.Unlock()
		return nil
	}

	remaining := delay - elapsed
	r.mu.Unlock()

	// Wait outside the lock, then reclaim.
	select {
	case <-ctx.Done():
		return core.E("collect.RateLimiter.Wait", "context cancelled", ctx.Err())
	case <-time.After(remaining):
	}

	r.mu.Lock()
	r.last[source] = time.Now()
	r.mu.Unlock()

	return nil
}

// SetDelay sets the delay for a source.
func (r *RateLimiter) SetDelay(source string, d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.delays[source] = d
}

// GetDelay returns the delay configured for a source.
func (r *RateLimiter) GetDelay(source string) time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d, ok := r.delays[source]; ok {
		return d
	}
	return 500 * time.Millisecond
}

// CheckGitHubRateLimit checks GitHub API rate limit status via gh api.
// Returns used and limit counts. Auto-pauses at 75% usage by increasing
// the GitHub rate limit delay.
func (r *RateLimiter) CheckGitHubRateLimit() (used, limit int, err error) {
	cmd := exec.Command("gh", "api", "rate_limit", "--jq", ".rate | \"\\(.used) \\(.limit)\"")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, core.E("collect.RateLimiter.CheckGitHubRateLimit", "failed to check rate limit", err)
	}

	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return 0, 0, core.E("collect.RateLimiter.CheckGitHubRateLimit",
			fmt.Sprintf("unexpected output format: %q", string(out)), nil)
	}

	used, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, core.E("collect.RateLimiter.CheckGitHubRateLimit", "failed to parse used count", err)
	}

	limit, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, core.E("collect.RateLimiter.CheckGitHubRateLimit", "failed to parse limit count", err)
	}

	// Auto-pause at 75% usage
	if limit > 0 {
		usage := float64(used) / float64(limit)
		if usage >= 0.75 {
			r.SetDelay("github", 5*time.Second)
		}
	}

	return used, limit, nil
}
