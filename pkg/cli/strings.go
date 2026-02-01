package cli

import "fmt"

// Sprintf formats a string (fmt.Sprintf wrapper).
func Sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// Sprint formats using default formats (fmt.Sprint wrapper).
func Sprint(args ...any) string {
	return fmt.Sprint(args...)
}

// Styled returns text with a style applied.
func Styled(style *AnsiStyle, text string) string {
	return style.Render(text)
}

// Styledf returns formatted text with a style applied.
func Styledf(style *AnsiStyle, format string, args ...any) string {
	return style.Render(fmt.Sprintf(format, args...))
}

// SuccessStr returns success-styled string.
func SuccessStr(msg string) string {
	return SuccessStyle.Render(Glyph(":check:") + " " + msg)
}

// ErrorStr returns error-styled string.
func ErrorStr(msg string) string {
	return ErrorStyle.Render(Glyph(":cross:") + " " + msg)
}

// WarnStr returns warning-styled string.
func WarnStr(msg string) string {
	return WarningStyle.Render(Glyph(":warn:") + " " + msg)
}

// InfoStr returns info-styled string.
func InfoStr(msg string) string {
	return InfoStyle.Render(Glyph(":info:") + " " + msg)
}

// DimStr returns dim-styled string.
func DimStr(msg string) string {
	return DimStyle.Render(msg)
}
