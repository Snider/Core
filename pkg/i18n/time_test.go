package i18n

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatAgo(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	tests := []struct {
		name     string
		count    int
		unit     string
		expected string
	}{
		{"1 second", 1, "second", "1 second ago"},
		{"5 seconds", 5, "second", "5 seconds ago"},
		{"1 minute", 1, "minute", "1 minute ago"},
		{"30 minutes", 30, "minute", "30 minutes ago"},
		{"1 hour", 1, "hour", "1 hour ago"},
		{"3 hours", 3, "hour", "3 hours ago"},
		{"1 day", 1, "day", "1 day ago"},
		{"7 days", 7, "day", "7 days ago"},
		{"1 week", 1, "week", "1 week ago"},
		{"2 weeks", 2, "week", "2 weeks ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatAgo(tt.count, tt.unit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeAgo(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	tests := []struct {
		name     string
		ago      time.Duration
		expected string
	}{
		{"just now", 30 * time.Second, "just now"},
		{"1 minute", 1 * time.Minute, "1 minute ago"},
		{"5 minutes", 5 * time.Minute, "5 minutes ago"},
		{"1 hour", 1 * time.Hour, "1 hour ago"},
		{"3 hours", 3 * time.Hour, "3 hours ago"},
		{"1 day", 24 * time.Hour, "1 day ago"},
		{"3 days", 3 * 24 * time.Hour, "3 days ago"},
		{"1 week", 7 * 24 * time.Hour, "1 week ago"},
		{"2 weeks", 14 * 24 * time.Hour, "2 weeks ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeAgo(time.Now().Add(-tt.ago))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestI18nAgoNamespace(t *testing.T) {
	svc, err := New()
	require.NoError(t, err)
	SetDefault(svc)

	t.Run("i18n.ago pattern", func(t *testing.T) {
		result := T("i18n.ago", 5, "minute")
		assert.Equal(t, "5 minutes ago", result)
	})

	t.Run("i18n.ago singular", func(t *testing.T) {
		result := T("i18n.ago", 1, "hour")
		assert.Equal(t, "1 hour ago", result)
	})
}
