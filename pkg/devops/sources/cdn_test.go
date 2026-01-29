package sources

import (
	"testing"
)

func TestCDNSource_Good_Available(t *testing.T) {
	src := NewCDNSource(SourceConfig{
		CDNURL:    "https://images.example.com",
		ImageName: "core-devops-darwin-arm64.qcow2",
	})

	if src.Name() != "cdn" {
		t.Errorf("expected name 'cdn', got %q", src.Name())
	}

	// CDN is available if URL is configured
	if !src.Available() {
		t.Error("expected Available() to be true when URL is set")
	}
}

func TestCDNSource_Bad_NoURL(t *testing.T) {
	src := NewCDNSource(SourceConfig{
		ImageName: "core-devops-darwin-arm64.qcow2",
	})

	if src.Available() {
		t.Error("expected Available() to be false when URL is empty")
	}
}
