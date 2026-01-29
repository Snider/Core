package sdk

import (
	"testing"
)

func TestSDK_Good_SetVersion(t *testing.T) {
	s := New("/tmp", nil)
	s.SetVersion("v1.2.3")

	if s.version != "v1.2.3" {
		t.Errorf("expected version v1.2.3, got %s", s.version)
	}
}

func TestSDK_Good_VersionPassedToGenerator(t *testing.T) {
	config := &Config{
		Languages: []string{"typescript"},
		Output:    "sdk",
		Package: PackageConfig{
			Name: "test-sdk",
		},
	}
	s := New("/tmp", config)
	s.SetVersion("v2.0.0")

	if s.config.Package.Version != "v2.0.0" {
		t.Errorf("expected config version v2.0.0, got %s", s.config.Package.Version)
	}
}
