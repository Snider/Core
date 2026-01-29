package signing

import (
	"runtime"
	"testing"
)

func TestMacOSSigner_Good_Name(t *testing.T) {
	s := NewMacOSSigner(MacOSConfig{Identity: "Developer ID Application: Test"})
	if s.Name() != "codesign" {
		t.Errorf("expected name 'codesign', got %q", s.Name())
	}
}

func TestMacOSSigner_Good_Available(t *testing.T) {
	s := NewMacOSSigner(MacOSConfig{Identity: "Developer ID Application: Test"})

	// Only available on macOS with identity set
	if runtime.GOOS == "darwin" {
		// May or may not be available depending on Xcode
		_ = s.Available()
	} else {
		if s.Available() {
			t.Error("expected Available() to be false on non-macOS")
		}
	}
}

func TestMacOSSigner_Bad_NoIdentity(t *testing.T) {
	s := NewMacOSSigner(MacOSConfig{})
	if s.Available() {
		t.Error("expected Available() to be false when identity is empty")
	}
}
