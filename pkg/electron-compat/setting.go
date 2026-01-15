package electroncompat

// SettingService provides application settings operations.
// This corresponds to the Setting IPC service from the Electron app.
type SettingService struct{}

// NewSettingService creates a new SettingService instance.
func NewSettingService() *SettingService {
	return &SettingService{}
}

// GetExplorer returns the configured block explorer URL.
func (s *SettingService) GetExplorer() (string, error) {
	return "", notImplemented("Setting", "getExplorer")
}

// SetExplorer sets the block explorer URL.
func (s *SettingService) SetExplorer(url string) error {
	return notImplemented("Setting", "setExplorer")
}

// GetLocale returns the current locale setting.
func (s *SettingService) GetLocale() (string, error) {
	return "", notImplemented("Setting", "getLocale")
}

// SetLocale sets the locale.
func (s *SettingService) SetLocale(locale string) error {
	return notImplemented("Setting", "setLocale")
}

// GetCustomLocale returns custom locale overrides.
func (s *SettingService) GetCustomLocale() (map[string]string, error) {
	return nil, notImplemented("Setting", "getCustomLocale")
}

// SetCustomLocale sets custom locale overrides.
func (s *SettingService) SetCustomLocale(locale map[string]string) error {
	return notImplemented("Setting", "setCustomLocale")
}

// GetLatestRelease returns the latest release info.
func (s *SettingService) GetLatestRelease() (map[string]any, error) {
	return nil, notImplemented("Setting", "getLatestRelease")
}
