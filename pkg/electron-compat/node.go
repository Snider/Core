package electroncompat

// NodeService provides blockchain node operations.
// This corresponds to the Node IPC service from the Electron app.
type NodeService struct{}

// NewNodeService creates a new NodeService instance.
func NewNodeService() *NodeService {
	return &NodeService{}
}

// Start starts the blockchain node.
func (s *NodeService) Start() error {
	return notImplemented("Node", "start")
}

// Stop stops the blockchain node.
func (s *NodeService) Stop() error {
	return notImplemented("Node", "stop")
}

// Reset resets the blockchain node data.
func (s *NodeService) Reset() error {
	return notImplemented("Node", "reset")
}

// GenerateToAddress generates blocks to an address (regtest mode).
func (s *NodeService) GenerateToAddress(address string, blocks int) error {
	return notImplemented("Node", "generateToAddress")
}

// GetAPIKey returns the node API key.
func (s *NodeService) GetAPIKey() (string, error) {
	return "", notImplemented("Node", "getAPIKey")
}

// GetNoDns returns whether DNS is disabled.
func (s *NodeService) GetNoDns() (bool, error) {
	return false, notImplemented("Node", "getNoDns")
}

// GetSpvMode returns whether SPV mode is enabled.
func (s *NodeService) GetSpvMode() (bool, error) {
	return false, notImplemented("Node", "getSpvMode")
}

// GetInfo returns node information.
func (s *NodeService) GetInfo() (map[string]any, error) {
	return nil, notImplemented("Node", "getInfo")
}

// GetNameInfo returns information about a name.
func (s *NodeService) GetNameInfo(name string) (map[string]any, error) {
	return nil, notImplemented("Node", "getNameInfo")
}

// GetTXByAddresses returns transactions for addresses.
func (s *NodeService) GetTXByAddresses(addresses []string) ([]map[string]any, error) {
	return nil, notImplemented("Node", "getTXByAddresses")
}

// GetNameByHash returns a name by its hash.
func (s *NodeService) GetNameByHash(hash string) (string, error) {
	return "", notImplemented("Node", "getNameByHash")
}

// GetBlockByHeight returns a block by height.
func (s *NodeService) GetBlockByHeight(height int) (map[string]any, error) {
	return nil, notImplemented("Node", "getBlockByHeight")
}

// GetTx returns a transaction by hash.
func (s *NodeService) GetTx(hash string) (map[string]any, error) {
	return nil, notImplemented("Node", "getTx")
}

// BroadcastRawTx broadcasts a raw transaction.
func (s *NodeService) BroadcastRawTx(tx string) (string, error) {
	return "", notImplemented("Node", "broadcastRawTx")
}

// SendRawAirdrop sends a raw airdrop proof.
func (s *NodeService) SendRawAirdrop(proof string) error {
	return notImplemented("Node", "sendRawAirdrop")
}

// GetFees returns current fee estimates.
func (s *NodeService) GetFees() (map[string]any, error) {
	return nil, notImplemented("Node", "getFees")
}

// GetAverageBlockTime returns the average block time.
func (s *NodeService) GetAverageBlockTime() (float64, error) {
	return 0, notImplemented("Node", "getAverageBlockTime")
}

// GetMTP returns the median time past.
func (s *NodeService) GetMTP() (int64, error) {
	return 0, notImplemented("Node", "getMTP")
}

// GetCoin returns a coin by outpoint.
func (s *NodeService) GetCoin(hash string, index int) (map[string]any, error) {
	return nil, notImplemented("Node", "getCoin")
}

// VerifyMessageWithName verifies a signed message.
func (s *NodeService) VerifyMessageWithName(name, signature, message string) (bool, error) {
	return false, notImplemented("Node", "verifyMessageWithName")
}

// SetNodeDir sets the node data directory.
func (s *NodeService) SetNodeDir(dir string) error {
	return notImplemented("Node", "setNodeDir")
}

// SetAPIKey sets the node API key.
func (s *NodeService) SetAPIKey(key string) error {
	return notImplemented("Node", "setAPIKey")
}

// SetNoDns sets whether DNS is disabled.
func (s *NodeService) SetNoDns(noDns bool) error {
	return notImplemented("Node", "setNoDns")
}

// SetSpvMode sets whether SPV mode is enabled.
func (s *NodeService) SetSpvMode(spv bool) error {
	return notImplemented("Node", "setSpvMode")
}

// GetDir returns the node data directory.
func (s *NodeService) GetDir() (string, error) {
	return "", notImplemented("Node", "getDir")
}

// GetHNSPrice returns the current HNS price.
func (s *NodeService) GetHNSPrice() (float64, error) {
	return 0, notImplemented("Node", "getHNSPrice")
}

// TestCustomRPCClient tests a custom RPC connection.
func (s *NodeService) TestCustomRPCClient(host string, port int, apiKey string) error {
	return notImplemented("Node", "testCustomRPCClient")
}

// GetDNSSECProof returns a DNSSEC proof for a name.
func (s *NodeService) GetDNSSECProof(name string) (string, error) {
	return "", notImplemented("Node", "getDNSSECProof")
}

// SendRawClaim sends a raw claim.
func (s *NodeService) SendRawClaim(claim string) error {
	return notImplemented("Node", "sendRawClaim")
}
