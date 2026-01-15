package electroncompat

// ShakedexService provides decentralized exchange operations.
// This corresponds to the Shakedex IPC service from the Electron app.
type ShakedexService struct{}

// NewShakedexService creates a new ShakedexService instance.
func NewShakedexService() *ShakedexService {
	return &ShakedexService{}
}

// FulfillSwap fulfills a swap offer.
func (s *ShakedexService) FulfillSwap(offerID string, fundingAddr string) (map[string]any, error) {
	return nil, notImplemented("Shakedex", "fulfillSwap")
}

// GetFulfillments returns swap fulfillments.
func (s *ShakedexService) GetFulfillments() ([]map[string]any, error) {
	return nil, notImplemented("Shakedex", "getFulfillments")
}

// FinalizeSwap finalizes a swap.
func (s *ShakedexService) FinalizeSwap(fulfillmentID string) (map[string]any, error) {
	return nil, notImplemented("Shakedex", "finalizeSwap")
}

// TransferLock creates a transfer lock.
func (s *ShakedexService) TransferLock(name, passphrase string) (map[string]any, error) {
	return nil, notImplemented("Shakedex", "transferLock")
}

// TransferCancel cancels a transfer lock.
func (s *ShakedexService) TransferCancel(name, passphrase string) (map[string]any, error) {
	return nil, notImplemented("Shakedex", "transferCancel")
}

// GetListings returns active listings.
func (s *ShakedexService) GetListings() ([]map[string]any, error) {
	return nil, notImplemented("Shakedex", "getListings")
}

// FinalizeLock finalizes a transfer lock.
func (s *ShakedexService) FinalizeLock(name, passphrase string) (map[string]any, error) {
	return nil, notImplemented("Shakedex", "finalizeLock")
}

// FinalizeCancel finalizes a cancellation.
func (s *ShakedexService) FinalizeCancel(name, passphrase string) (map[string]any, error) {
	return nil, notImplemented("Shakedex", "finalizeCancel")
}

// LaunchAuction launches a name auction.
func (s *ShakedexService) LaunchAuction(name string, params map[string]any) (map[string]any, error) {
	return nil, notImplemented("Shakedex", "launchAuction")
}

// DownloadProofs downloads auction proofs.
func (s *ShakedexService) DownloadProofs(auctionID string) ([]byte, error) {
	return nil, notImplemented("Shakedex", "downloadProofs")
}

// RestoreOneListing restores a listing.
func (s *ShakedexService) RestoreOneListing(listingID string) error {
	return notImplemented("Shakedex", "restoreOneListing")
}

// RestoreOneFill restores a fill.
func (s *ShakedexService) RestoreOneFill(fillID string) error {
	return notImplemented("Shakedex", "restoreOneFill")
}

// GetExchangeAuctions returns exchange auctions.
func (s *ShakedexService) GetExchangeAuctions() ([]map[string]any, error) {
	return nil, notImplemented("Shakedex", "getExchangeAuctions")
}

// ListAuction lists a name for auction.
func (s *ShakedexService) ListAuction(name string, startPrice, endPrice int64, duration int) (map[string]any, error) {
	return nil, notImplemented("Shakedex", "listAuction")
}

// GetFeeInfo returns fee information.
func (s *ShakedexService) GetFeeInfo() (map[string]any, error) {
	return nil, notImplemented("Shakedex", "getFeeInfo")
}

// GetBestBid returns the best bid for an auction.
func (s *ShakedexService) GetBestBid(auctionID string) (map[string]any, error) {
	return nil, notImplemented("Shakedex", "getBestBid")
}
