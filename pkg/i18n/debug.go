// Package i18n provides internationalization for the CLI.
package i18n

// Debug mode provides visibility into i18n key resolution for development.
// When enabled, translations are prefixed with their key: [cli.success] Success
//
// Usage:
//
//	i18n.SetDebug(true)
//	fmt.Println(i18n.T("cli.success")) // "[cli.success] Success"
//
// This helps identify which keys are being used in the UI, making it easier
// to find and update translations during development.

// SetDebug enables or disables debug mode on the default service.
// In debug mode, translations show their keys: [key] translation
//
//	SetDebug(true)
//	T("cli.success") // "[cli.success] Success"
func SetDebug(enabled bool) {
	if svc := Default(); svc != nil {
		svc.SetDebug(enabled)
	}
}

// SetDebug enables or disables debug mode.
// In debug mode, translations are prefixed with their key:
//
//	[cli.success] Success
//	[core.delete] Delete config.yaml?
func (s *Service) SetDebug(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debug = enabled
}

// Debug returns whether debug mode is enabled.
func (s *Service) Debug() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.debug
}

// debugFormat formats a translation with its key prefix for debug mode.
// Returns "[key] text" format.
func debugFormat(key, text string) string {
	return "[" + key + "] " + text
}
