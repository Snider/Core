package electroncompat

// AnalyticsService provides usage tracking operations.
// This corresponds to the Analytics IPC service from the Electron app.
type AnalyticsService struct{}

// NewAnalyticsService creates a new AnalyticsService instance.
func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{}
}

// SetOptIn sets whether analytics are enabled.
func (s *AnalyticsService) SetOptIn(optIn bool) error {
	return notImplemented("Analytics", "setOptIn")
}

// GetOptIn returns whether analytics are enabled.
func (s *AnalyticsService) GetOptIn() (bool, error) {
	return false, notImplemented("Analytics", "getOptIn")
}

// Track tracks an event.
func (s *AnalyticsService) Track(event string, properties map[string]any) error {
	return notImplemented("Analytics", "track")
}

// ScreenView tracks a screen view.
func (s *AnalyticsService) ScreenView(screen string) error {
	return notImplemented("Analytics", "screenView")
}
