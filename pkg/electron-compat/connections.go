package electroncompat

// ConnectionsService provides RPC connection management.
// This corresponds to the Connections IPC service from the Electron app.
type ConnectionsService struct{}

// NewConnectionsService creates a new ConnectionsService instance.
func NewConnectionsService() *ConnectionsService {
	return &ConnectionsService{}
}

// ConnectionType represents the type of connection.
type ConnectionType string

const (
	// ConnectionLocal uses a local node.
	ConnectionLocal ConnectionType = "local"
	// ConnectionP2P uses P2P networking.
	ConnectionP2P ConnectionType = "p2p"
	// ConnectionCustom uses a custom RPC endpoint.
	ConnectionCustom ConnectionType = "custom"
)

// GetConnection returns the current connection settings.
func (s *ConnectionsService) GetConnection() (map[string]any, error) {
	return nil, notImplemented("Connections", "getConnection")
}

// SetConnection sets the connection settings.
func (s *ConnectionsService) SetConnection(settings map[string]any) error {
	return notImplemented("Connections", "setConnection")
}

// SetConnectionType sets the connection type.
func (s *ConnectionsService) SetConnectionType(connType ConnectionType) error {
	return notImplemented("Connections", "setConnectionType")
}

// GetCustomRPC returns custom RPC settings.
func (s *ConnectionsService) GetCustomRPC() (map[string]any, error) {
	return nil, notImplemented("Connections", "getCustomRPC")
}
