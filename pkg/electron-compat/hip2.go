package electroncompat

// Hip2Service provides HIP-2 DNS protocol operations.
// This corresponds to the Hip2 IPC service from the Electron app.
// HIP-2 defines how to resolve Handshake names to payment addresses.
type Hip2Service struct{}

// NewHip2Service creates a new Hip2Service instance.
func NewHip2Service() *Hip2Service {
	return &Hip2Service{}
}

// GetPort returns the HIP-2 server port.
func (s *Hip2Service) GetPort() (int, error) {
	return 0, notImplemented("Hip2", "getPort")
}

// SetPort sets the HIP-2 server port.
func (s *Hip2Service) SetPort(port int) error {
	return notImplemented("Hip2", "setPort")
}

// FetchAddress fetches an address for a name using HIP-2.
func (s *Hip2Service) FetchAddress(name string) (string, error) {
	return "", notImplemented("Hip2", "fetchAddress")
}

// SetServers sets the HIP-2 resolver servers.
func (s *Hip2Service) SetServers(servers []string) error {
	return notImplemented("Hip2", "setServers")
}
