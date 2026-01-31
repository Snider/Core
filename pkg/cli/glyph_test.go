package cli

import "testing"

func TestGlyph(t *testing.T) {
	UseUnicode()
	if Glyph(":check:") != "✓" {
		t.Errorf("Expected ✓, got %s", Glyph(":check:"))
	}

	UseASCII()
	if Glyph(":check:") != "[OK]" {
		t.Errorf("Expected [OK], got %s", Glyph(":check:"))
	}
}

func TestCompileGlyphs(t *testing.T) {
	UseUnicode()
	got := compileGlyphs("Status: :check:")
	if got != "Status: ✓" {
		t.Errorf("Expected Status: ✓, got %s", got)
	}
}
