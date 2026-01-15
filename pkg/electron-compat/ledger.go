package electroncompat

// LedgerService provides Ledger hardware wallet integration.
// This corresponds to the Ledger IPC service from the Electron app.
type LedgerService struct{}

// NewLedgerService creates a new LedgerService instance.
func NewLedgerService() *LedgerService {
	return &LedgerService{}
}

// GetXPub returns the extended public key from the Ledger device.
func (s *LedgerService) GetXPub(path string) (string, error) {
	return "", notImplemented("Ledger", "getXPub")
}

// GetAppVersion returns the Handshake app version on the Ledger.
func (s *LedgerService) GetAppVersion() (string, error) {
	return "", notImplemented("Ledger", "getAppVersion")
}
