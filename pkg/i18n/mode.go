// Package i18n provides internationalization for the CLI.
package i18n

// String returns the string representation of the mode.
func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "normal"
	case ModeStrict:
		return "strict"
	case ModeCollect:
		return "collect"
	default:
		return "unknown"
	}
}
