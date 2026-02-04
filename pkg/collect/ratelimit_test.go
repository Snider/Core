package collect

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Wait_Good(t *testing.T) {
	rl := NewRateLimiter()
	rl.SetDelay("test", 50*time.Millisecond)

	ctx := context.Background()

	// First call should return immediately
	start := time.Now()
	err := rl.Wait(ctx, "test")
	assert.NoError(t, err)
	assert.Less(t, time.Since(start), 50*time.Millisecond)

	// Second call should wait at least the delay
	start = time.Now()
	err = rl.Wait(ctx, "test")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, time.Since(start), 40*time.Millisecond) // allow small timing variance
}

func TestRateLimiter_Wait_Bad_ContextCancelled(t *testing.T) {
	rl := NewRateLimiter()
	rl.SetDelay("test", 5*time.Second)

	ctx := context.Background()

	// First call to set the last time
	err := rl.Wait(ctx, "test")
	assert.NoError(t, err)

	// Cancel context before second call
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = rl.Wait(ctx, "test")
	assert.Error(t, err)
}

func TestRateLimiter_SetDelay_Good(t *testing.T) {
	rl := NewRateLimiter()
	rl.SetDelay("custom", 3*time.Second)
	assert.Equal(t, 3*time.Second, rl.GetDelay("custom"))
}

func TestRateLimiter_GetDelay_Good_Defaults(t *testing.T) {
	rl := NewRateLimiter()

	assert.Equal(t, 500*time.Millisecond, rl.GetDelay("github"))
	assert.Equal(t, 2*time.Second, rl.GetDelay("bitcointalk"))
	assert.Equal(t, 1500*time.Millisecond, rl.GetDelay("coingecko"))
	assert.Equal(t, 1*time.Second, rl.GetDelay("iacr"))
}

func TestRateLimiter_GetDelay_Good_UnknownSource(t *testing.T) {
	rl := NewRateLimiter()
	// Unknown sources should get the default 500ms delay
	assert.Equal(t, 500*time.Millisecond, rl.GetDelay("unknown"))
}

func TestRateLimiter_Wait_Good_UnknownSource(t *testing.T) {
	rl := NewRateLimiter()
	ctx := context.Background()

	// Unknown source should use default delay of 500ms
	err := rl.Wait(ctx, "unknown-source")
	assert.NoError(t, err)
}

func TestNewRateLimiter_Good(t *testing.T) {
	rl := NewRateLimiter()
	assert.NotNil(t, rl)
	assert.NotNil(t, rl.delays)
	assert.NotNil(t, rl.last)
	assert.Len(t, rl.delays, len(defaultDelays))
}
