// Package updater provides auto-update functionality for BugSETI.
package updater

import (
	"fmt"
	"regexp"
	"strings"
)

// Channel represents an update channel.
type Channel string

const (
	// ChannelStable is the production release channel.
	// Tags: bugseti-vX.Y.Z (e.g., bugseti-v1.0.0)
	ChannelStable Channel = "stable"

	// ChannelBeta is the pre-release testing channel.
	// Tags: bugseti-vX.Y.Z-beta.N (e.g., bugseti-v1.0.0-beta.1)
	ChannelBeta Channel = "beta"

	// ChannelNightly is the latest development builds channel.
	// Tags: bugseti-nightly-YYYYMMDD (e.g., bugseti-nightly-20260205)
	ChannelNightly Channel = "nightly"
)

// String returns the string representation of the channel.
func (c Channel) String() string {
	return string(c)
}

// DisplayName returns a human-readable name for the channel.
func (c Channel) DisplayName() string {
	switch c {
	case ChannelStable:
		return "Stable"
	case ChannelBeta:
		return "Beta"
	case ChannelNightly:
		return "Nightly"
	default:
		return "Unknown"
	}
}

// Description returns a description of the channel.
func (c Channel) Description() string {
	switch c {
	case ChannelStable:
		return "Production releases - most stable, recommended for most users"
	case ChannelBeta:
		return "Pre-release builds - new features being tested before stable release"
	case ChannelNightly:
		return "Latest development builds - bleeding edge, may be unstable"
	default:
		return "Unknown channel"
	}
}

// TagPrefix returns the tag prefix used for this channel.
func (c Channel) TagPrefix() string {
	switch c {
	case ChannelStable:
		return "bugseti-v"
	case ChannelBeta:
		return "bugseti-v"
	case ChannelNightly:
		return "bugseti-nightly-"
	default:
		return ""
	}
}

// TagPattern returns a regex pattern to match tags for this channel.
func (c Channel) TagPattern() *regexp.Regexp {
	switch c {
	case ChannelStable:
		// Match bugseti-vX.Y.Z but NOT bugseti-vX.Y.Z-beta.N
		return regexp.MustCompile(`^bugseti-v(\d+\.\d+\.\d+)$`)
	case ChannelBeta:
		// Match bugseti-vX.Y.Z-beta.N
		return regexp.MustCompile(`^bugseti-v(\d+\.\d+\.\d+-beta\.\d+)$`)
	case ChannelNightly:
		// Match bugseti-nightly-YYYYMMDD
		return regexp.MustCompile(`^bugseti-nightly-(\d{8})$`)
	default:
		return nil
	}
}

// MatchesTag returns true if the given tag matches this channel's pattern.
func (c Channel) MatchesTag(tag string) bool {
	pattern := c.TagPattern()
	if pattern == nil {
		return false
	}
	return pattern.MatchString(tag)
}

// ExtractVersion extracts the version from a tag for this channel.
func (c Channel) ExtractVersion(tag string) string {
	pattern := c.TagPattern()
	if pattern == nil {
		return ""
	}
	matches := pattern.FindStringSubmatch(tag)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

// AllChannels returns all available channels.
func AllChannels() []Channel {
	return []Channel{ChannelStable, ChannelBeta, ChannelNightly}
}

// ParseChannel parses a string into a Channel.
func ParseChannel(s string) (Channel, error) {
	switch strings.ToLower(s) {
	case "stable":
		return ChannelStable, nil
	case "beta":
		return ChannelBeta, nil
	case "nightly":
		return ChannelNightly, nil
	default:
		return "", fmt.Errorf("unknown channel: %s", s)
	}
}

// ChannelInfo contains information about an update channel.
type ChannelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetChannelInfo returns information about a channel.
func GetChannelInfo(c Channel) ChannelInfo {
	return ChannelInfo{
		ID:          c.String(),
		Name:        c.DisplayName(),
		Description: c.Description(),
	}
}

// GetAllChannelInfo returns information about all channels.
func GetAllChannelInfo() []ChannelInfo {
	channels := AllChannels()
	info := make([]ChannelInfo, len(channels))
	for i, c := range channels {
		info[i] = GetChannelInfo(c)
	}
	return info
}

// IncludesPrerelease returns true if the channel includes pre-release versions.
func (c Channel) IncludesPrerelease() bool {
	return c == ChannelBeta || c == ChannelNightly
}

// IncludesChannel returns true if this channel should include releases from the given channel.
// For example, beta channel includes stable releases, nightly includes both.
func (c Channel) IncludesChannel(other Channel) bool {
	switch c {
	case ChannelStable:
		return other == ChannelStable
	case ChannelBeta:
		return other == ChannelStable || other == ChannelBeta
	case ChannelNightly:
		return true // Nightly users can see all releases
	default:
		return false
	}
}
