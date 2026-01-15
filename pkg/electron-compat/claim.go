package electroncompat

// ClaimService provides airdrop and claim proof operations.
// This corresponds to the Claim IPC service from the Electron app.
type ClaimService struct{}

// NewClaimService creates a new ClaimService instance.
func NewClaimService() *ClaimService {
	return &ClaimService{}
}

// AirdropGenerateProofs generates airdrop proofs for an address.
func (s *ClaimService) AirdropGenerateProofs(address string) ([]map[string]any, error) {
	return nil, notImplemented("Claim", "airdropGenerateProofs")
}
