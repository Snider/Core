package signing

import (
	"testing"
)

func TestGPGSigner_Good_Name(t *testing.T) {
	s := NewGPGSigner("ABCD1234")
	if s.Name() != "gpg" {
		t.Errorf("expected name 'gpg', got %q", s.Name())
	}
}

func TestGPGSigner_Good_Available(t *testing.T) {
	s := NewGPGSigner("ABCD1234")
	// Available depends on gpg being installed
	_ = s.Available()
}

func TestGPGSigner_Bad_NoKey(t *testing.T) {
	s := NewGPGSigner("")
	if s.Available() {
		t.Error("expected Available() to be false when key is empty")
	}
}
