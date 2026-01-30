// Package i18n provides internationalization for the CLI.
package i18n

import (
	"fmt"
	"time"
)

// TimeAgo returns a localized relative time string.
//
//	TimeAgo(time.Now().Add(-5 * time.Minute)) // "5 minutes ago"
//	TimeAgo(time.Now().Add(-1 * time.Hour))   // "1 hour ago"
func TimeAgo(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return T("time.just_now")
	case duration < time.Hour:
		mins := int(duration.Minutes())
		return FormatAgo(mins, "minute")
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		return FormatAgo(hours, "hour")
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		return FormatAgo(days, "day")
	default:
		weeks := int(duration.Hours() / (24 * 7))
		return FormatAgo(weeks, "week")
	}
}

// FormatAgo formats "N unit ago" with proper pluralization.
// Uses locale-specific patterns from time.ago.{unit}.
//
//	FormatAgo(5, "minute") // "5 minutes ago"
//	FormatAgo(1, "hour")   // "1 hour ago"
func FormatAgo(count int, unit string) string {
	svc := Default()
	if svc == nil {
		return fmt.Sprintf("%d %ss ago", count, unit)
	}

	// Try locale-specific pattern: time.ago.{unit}
	key := "time.ago." + unit
	result := svc.T(key, map[string]any{"Count": count})

	// If key was returned as-is (not found), compose fallback
	if result == key {
		return fmt.Sprintf("%d %s ago", count, Pluralize(unit, count))
	}

	return result
}
