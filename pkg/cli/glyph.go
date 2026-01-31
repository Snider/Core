package cli

import (
	"bytes"
	"unicode"
)

// GlyphTheme defines which symbols to use.
type GlyphTheme int

const (
	ThemeUnicode GlyphTheme = iota
	ThemeEmoji
	ThemeASCII
)

var currentTheme = ThemeUnicode

func UseUnicode() { currentTheme = ThemeUnicode }
func UseEmoji()   { currentTheme = ThemeEmoji }
func UseASCII()   { currentTheme = ThemeASCII }

func glyphMap() map[string]string {
	switch currentTheme {
	case ThemeEmoji:
		return glyphMapEmoji
	case ThemeASCII:
		return glyphMapASCII
	default:
		return glyphMapUnicode
	}
}

// Glyph converts a shortcode to its symbol.
func Glyph(code string) string {
	if sym, ok := glyphMap()[code]; ok {
		return sym
	}
	return code
}

func compileGlyphs(x string) string {
	if x == "" {
		return ""
	}
	input := bytes.NewBufferString(x)
	output := bytes.NewBufferString("")

	for {
		r, _, err := input.ReadRune()
		if err != nil {
			break
		}
		if r == ':' {
			output.WriteString(replaceGlyph(input))
		} else {
			output.WriteRune(r)
		}
	}
	return output.String()
}

func replaceGlyph(input *bytes.Buffer) string {
	code := bytes.NewBufferString(":")
	for {
		r, _, err := input.ReadRune()
		if err != nil {
			return code.String()
		}
		if r == ':' && code.Len() == 1 {
			return code.String() + replaceGlyph(input)
		}
		code.WriteRune(r)
		if unicode.IsSpace(r) {
			return code.String()
		}
		if r == ':' {
			return Glyph(code.String())
		}
	}
}
