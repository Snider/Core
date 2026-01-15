package electroncompat

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotImplementedError(t *testing.T) {
	t.Run("creates error with service and method", func(t *testing.T) {
		err := notImplemented("TestService", "testMethod")
		assert.Error(t, err)
		assert.Equal(t, "IPC TestService.testMethod is not implemented", err.Error())
	})

	t.Run("implements error interface", func(t *testing.T) {
		err := notImplemented("Service", "method")
		var notImplErr *NotImplementedError
		assert.True(t, errors.As(err, &notImplErr))
		assert.Equal(t, "Service", notImplErr.Service)
		assert.Equal(t, "method", notImplErr.Method)
	})
}

func TestErrNotImplemented(t *testing.T) {
	assert.Equal(t, "not implemented", ErrNotImplemented.Error())
}

// Helper to check if error is NotImplementedError
func assertNotImplemented(t *testing.T, err error) {
	t.Helper()
	var notImplErr *NotImplementedError
	assert.True(t, errors.As(err, &notImplErr), "expected NotImplementedError")
}

func TestNodeService(t *testing.T) {
	svc := NewNodeService()
	assert.NotNil(t, svc)

	t.Run("Start", func(t *testing.T) {
		assertNotImplemented(t, svc.Start())
	})
	t.Run("Stop", func(t *testing.T) {
		assertNotImplemented(t, svc.Stop())
	})
	t.Run("Reset", func(t *testing.T) {
		assertNotImplemented(t, svc.Reset())
	})
	t.Run("GenerateToAddress", func(t *testing.T) {
		assertNotImplemented(t, svc.GenerateToAddress("addr", 1))
	})
	t.Run("GetAPIKey", func(t *testing.T) {
		_, err := svc.GetAPIKey()
		assertNotImplemented(t, err)
	})
	t.Run("GetNoDns", func(t *testing.T) {
		_, err := svc.GetNoDns()
		assertNotImplemented(t, err)
	})
	t.Run("GetSpvMode", func(t *testing.T) {
		_, err := svc.GetSpvMode()
		assertNotImplemented(t, err)
	})
	t.Run("GetInfo", func(t *testing.T) {
		_, err := svc.GetInfo()
		assertNotImplemented(t, err)
	})
	t.Run("GetNameInfo", func(t *testing.T) {
		_, err := svc.GetNameInfo("name")
		assertNotImplemented(t, err)
	})
	t.Run("GetTXByAddresses", func(t *testing.T) {
		_, err := svc.GetTXByAddresses([]string{"addr"})
		assertNotImplemented(t, err)
	})
	t.Run("GetNameByHash", func(t *testing.T) {
		_, err := svc.GetNameByHash("hash")
		assertNotImplemented(t, err)
	})
	t.Run("GetBlockByHeight", func(t *testing.T) {
		_, err := svc.GetBlockByHeight(1)
		assertNotImplemented(t, err)
	})
	t.Run("GetTx", func(t *testing.T) {
		_, err := svc.GetTx("hash")
		assertNotImplemented(t, err)
	})
	t.Run("BroadcastRawTx", func(t *testing.T) {
		_, err := svc.BroadcastRawTx("tx")
		assertNotImplemented(t, err)
	})
	t.Run("SendRawAirdrop", func(t *testing.T) {
		assertNotImplemented(t, svc.SendRawAirdrop("proof"))
	})
	t.Run("GetFees", func(t *testing.T) {
		_, err := svc.GetFees()
		assertNotImplemented(t, err)
	})
	t.Run("GetAverageBlockTime", func(t *testing.T) {
		_, err := svc.GetAverageBlockTime()
		assertNotImplemented(t, err)
	})
	t.Run("GetMTP", func(t *testing.T) {
		_, err := svc.GetMTP()
		assertNotImplemented(t, err)
	})
	t.Run("GetCoin", func(t *testing.T) {
		_, err := svc.GetCoin("hash", 0)
		assertNotImplemented(t, err)
	})
	t.Run("VerifyMessageWithName", func(t *testing.T) {
		_, err := svc.VerifyMessageWithName("name", "sig", "msg")
		assertNotImplemented(t, err)
	})
	t.Run("SetNodeDir", func(t *testing.T) {
		assertNotImplemented(t, svc.SetNodeDir("dir"))
	})
	t.Run("SetAPIKey", func(t *testing.T) {
		assertNotImplemented(t, svc.SetAPIKey("key"))
	})
	t.Run("SetNoDns", func(t *testing.T) {
		assertNotImplemented(t, svc.SetNoDns(true))
	})
	t.Run("SetSpvMode", func(t *testing.T) {
		assertNotImplemented(t, svc.SetSpvMode(true))
	})
	t.Run("GetDir", func(t *testing.T) {
		_, err := svc.GetDir()
		assertNotImplemented(t, err)
	})
	t.Run("GetHNSPrice", func(t *testing.T) {
		_, err := svc.GetHNSPrice()
		assertNotImplemented(t, err)
	})
	t.Run("TestCustomRPCClient", func(t *testing.T) {
		assertNotImplemented(t, svc.TestCustomRPCClient("host", 8080, "key"))
	})
	t.Run("GetDNSSECProof", func(t *testing.T) {
		_, err := svc.GetDNSSECProof("name")
		assertNotImplemented(t, err)
	})
	t.Run("SendRawClaim", func(t *testing.T) {
		assertNotImplemented(t, svc.SendRawClaim("claim"))
	})
}

func TestWalletService(t *testing.T) {
	svc := NewWalletService()
	assert.NotNil(t, svc)

	t.Run("Start", func(t *testing.T) {
		assertNotImplemented(t, svc.Start())
	})
	t.Run("GetAPIKey", func(t *testing.T) {
		_, err := svc.GetAPIKey()
		assertNotImplemented(t, err)
	})
	t.Run("SetAPIKey", func(t *testing.T) {
		assertNotImplemented(t, svc.SetAPIKey("key"))
	})
	t.Run("GetWalletInfo", func(t *testing.T) {
		_, err := svc.GetWalletInfo("id")
		assertNotImplemented(t, err)
	})
	t.Run("GetAccountInfo", func(t *testing.T) {
		_, err := svc.GetAccountInfo("wallet", "account")
		assertNotImplemented(t, err)
	})
	t.Run("GetCoin", func(t *testing.T) {
		_, err := svc.GetCoin("hash", 0)
		assertNotImplemented(t, err)
	})
	t.Run("GetTX", func(t *testing.T) {
		_, err := svc.GetTX("hash")
		assertNotImplemented(t, err)
	})
	t.Run("GetNames", func(t *testing.T) {
		_, err := svc.GetNames("wallet")
		assertNotImplemented(t, err)
	})
	t.Run("CreateNewWallet", func(t *testing.T) {
		_, err := svc.CreateNewWallet("id", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("ImportSeed", func(t *testing.T) {
		assertNotImplemented(t, svc.ImportSeed("id", "pass", "mnemonic"))
	})
	t.Run("GenerateReceivingAddress", func(t *testing.T) {
		_, err := svc.GenerateReceivingAddress("wallet", "account")
		assertNotImplemented(t, err)
	})
	t.Run("GetAuctionInfo", func(t *testing.T) {
		_, err := svc.GetAuctionInfo("name")
		assertNotImplemented(t, err)
	})
	t.Run("GetTransactionHistory", func(t *testing.T) {
		_, err := svc.GetTransactionHistory("wallet")
		assertNotImplemented(t, err)
	})
	t.Run("GetPendingTransactions", func(t *testing.T) {
		_, err := svc.GetPendingTransactions("wallet")
		assertNotImplemented(t, err)
	})
	t.Run("GetBids", func(t *testing.T) {
		_, err := svc.GetBids("wallet", true)
		assertNotImplemented(t, err)
	})
	t.Run("GetBlind", func(t *testing.T) {
		_, err := svc.GetBlind(100, "nonce")
		assertNotImplemented(t, err)
	})
	t.Run("GetMasterHDKey", func(t *testing.T) {
		_, err := svc.GetMasterHDKey("wallet", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("HasAddress", func(t *testing.T) {
		_, err := svc.HasAddress("wallet", "addr")
		assertNotImplemented(t, err)
	})
	t.Run("SetPassphrase", func(t *testing.T) {
		assertNotImplemented(t, svc.SetPassphrase("wallet", "old", "new"))
	})
	t.Run("RevealSeed", func(t *testing.T) {
		_, err := svc.RevealSeed("wallet", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("EstimateTxFee", func(t *testing.T) {
		_, err := svc.EstimateTxFee(1000)
		assertNotImplemented(t, err)
	})
	t.Run("EstimateMaxSend", func(t *testing.T) {
		_, err := svc.EstimateMaxSend("wallet", "account", 1000)
		assertNotImplemented(t, err)
	})
	t.Run("RemoveWalletById", func(t *testing.T) {
		assertNotImplemented(t, svc.RemoveWalletById("id"))
	})
	t.Run("UpdateAccountDepth", func(t *testing.T) {
		assertNotImplemented(t, svc.UpdateAccountDepth("wallet", "account", 100))
	})
	t.Run("FindNonce", func(t *testing.T) {
		_, err := svc.FindNonce("name", "addr", 100)
		assertNotImplemented(t, err)
	})
	t.Run("FindNonceCancel", func(t *testing.T) {
		assertNotImplemented(t, svc.FindNonceCancel())
	})
	t.Run("EncryptWallet", func(t *testing.T) {
		assertNotImplemented(t, svc.EncryptWallet("wallet", "pass"))
	})
	t.Run("Backup", func(t *testing.T) {
		assertNotImplemented(t, svc.Backup("wallet", "path"))
	})
	t.Run("Rescan", func(t *testing.T) {
		assertNotImplemented(t, svc.Rescan(0))
	})
	t.Run("DeepClean", func(t *testing.T) {
		assertNotImplemented(t, svc.DeepClean("wallet"))
	})
	t.Run("Reset", func(t *testing.T) {
		assertNotImplemented(t, svc.Reset())
	})
	t.Run("SendOpen", func(t *testing.T) {
		_, err := svc.SendOpen("wallet", "pass", "name")
		assertNotImplemented(t, err)
	})
	t.Run("SendBid", func(t *testing.T) {
		_, err := svc.SendBid("wallet", "pass", "name", 100, 200)
		assertNotImplemented(t, err)
	})
	t.Run("SendRegister", func(t *testing.T) {
		_, err := svc.SendRegister("wallet", "pass", "name")
		assertNotImplemented(t, err)
	})
	t.Run("SendUpdate", func(t *testing.T) {
		_, err := svc.SendUpdate("wallet", "pass", "name", nil)
		assertNotImplemented(t, err)
	})
	t.Run("SendReveal", func(t *testing.T) {
		_, err := svc.SendReveal("wallet", "pass", "name")
		assertNotImplemented(t, err)
	})
	t.Run("SendRedeem", func(t *testing.T) {
		_, err := svc.SendRedeem("wallet", "pass", "name")
		assertNotImplemented(t, err)
	})
	t.Run("SendRenewal", func(t *testing.T) {
		_, err := svc.SendRenewal("wallet", "pass", "name")
		assertNotImplemented(t, err)
	})
	t.Run("SendRevealAll", func(t *testing.T) {
		_, err := svc.SendRevealAll("wallet", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("SendRedeemAll", func(t *testing.T) {
		_, err := svc.SendRedeemAll("wallet", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("SendRegisterAll", func(t *testing.T) {
		_, err := svc.SendRegisterAll("wallet", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("SignMessageWithName", func(t *testing.T) {
		_, err := svc.SignMessageWithName("wallet", "pass", "name", "msg")
		assertNotImplemented(t, err)
	})
	t.Run("TransferMany", func(t *testing.T) {
		_, err := svc.TransferMany("wallet", "pass", []string{"name"}, "addr")
		assertNotImplemented(t, err)
	})
	t.Run("FinalizeAll", func(t *testing.T) {
		_, err := svc.FinalizeAll("wallet", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("FinalizeMany", func(t *testing.T) {
		_, err := svc.FinalizeMany("wallet", "pass", []string{"name"})
		assertNotImplemented(t, err)
	})
	t.Run("RenewAll", func(t *testing.T) {
		_, err := svc.RenewAll("wallet", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("RenewMany", func(t *testing.T) {
		_, err := svc.RenewMany("wallet", "pass", []string{"name"})
		assertNotImplemented(t, err)
	})
	t.Run("SendTransfer", func(t *testing.T) {
		_, err := svc.SendTransfer("wallet", "pass", "name", "addr")
		assertNotImplemented(t, err)
	})
	t.Run("CancelTransfer", func(t *testing.T) {
		_, err := svc.CancelTransfer("wallet", "pass", "name")
		assertNotImplemented(t, err)
	})
	t.Run("FinalizeTransfer", func(t *testing.T) {
		_, err := svc.FinalizeTransfer("wallet", "pass", "name")
		assertNotImplemented(t, err)
	})
	t.Run("FinalizeWithPayment", func(t *testing.T) {
		_, err := svc.FinalizeWithPayment("wallet", "pass", "name", "fund", "recv", 100)
		assertNotImplemented(t, err)
	})
	t.Run("ClaimPaidTransfer", func(t *testing.T) {
		_, err := svc.ClaimPaidTransfer("wallet", "pass", "hex")
		assertNotImplemented(t, err)
	})
	t.Run("RevokeName", func(t *testing.T) {
		_, err := svc.RevokeName("wallet", "pass", "name")
		assertNotImplemented(t, err)
	})
	t.Run("Send", func(t *testing.T) {
		_, err := svc.Send("wallet", "pass", "addr", 100)
		assertNotImplemented(t, err)
	})
	t.Run("Lock", func(t *testing.T) {
		assertNotImplemented(t, svc.Lock("wallet"))
	})
	t.Run("Unlock", func(t *testing.T) {
		assertNotImplemented(t, svc.Unlock("wallet", "pass", 60))
	})
	t.Run("IsLocked", func(t *testing.T) {
		_, err := svc.IsLocked("wallet")
		assertNotImplemented(t, err)
	})
	t.Run("AddSharedKey", func(t *testing.T) {
		assertNotImplemented(t, svc.AddSharedKey("wallet", "account", "key"))
	})
	t.Run("RemoveSharedKey", func(t *testing.T) {
		assertNotImplemented(t, svc.RemoveSharedKey("wallet", "account", "key"))
	})
	t.Run("GetNonce", func(t *testing.T) {
		_, err := svc.GetNonce("wallet", "name", "addr", 100)
		assertNotImplemented(t, err)
	})
	t.Run("ImportNonce", func(t *testing.T) {
		assertNotImplemented(t, svc.ImportNonce("wallet", "name", "addr", 100))
	})
	t.Run("Zap", func(t *testing.T) {
		assertNotImplemented(t, svc.Zap("wallet", "account", 3600))
	})
	t.Run("ImportName", func(t *testing.T) {
		assertNotImplemented(t, svc.ImportName("wallet", "name", 0))
	})
	t.Run("RPCGetWalletInfo", func(t *testing.T) {
		_, err := svc.RPCGetWalletInfo("wallet")
		assertNotImplemented(t, err)
	})
	t.Run("LoadTransaction", func(t *testing.T) {
		_, err := svc.LoadTransaction("wallet", "hex")
		assertNotImplemented(t, err)
	})
	t.Run("ListWallets", func(t *testing.T) {
		_, err := svc.ListWallets()
		assertNotImplemented(t, err)
	})
	t.Run("GetStats", func(t *testing.T) {
		_, err := svc.GetStats("wallet")
		assertNotImplemented(t, err)
	})
	t.Run("IsReady", func(t *testing.T) {
		_, err := svc.IsReady()
		assertNotImplemented(t, err)
	})
	t.Run("CreateClaim", func(t *testing.T) {
		_, err := svc.CreateClaim("wallet", "pass", "name")
		assertNotImplemented(t, err)
	})
	t.Run("SendClaim", func(t *testing.T) {
		_, err := svc.SendClaim("wallet", "pass", "name")
		assertNotImplemented(t, err)
	})
}

func TestSettingService(t *testing.T) {
	svc := NewSettingService()
	assert.NotNil(t, svc)

	t.Run("GetExplorer", func(t *testing.T) {
		_, err := svc.GetExplorer()
		assertNotImplemented(t, err)
	})
	t.Run("SetExplorer", func(t *testing.T) {
		assertNotImplemented(t, svc.SetExplorer("url"))
	})
	t.Run("GetLocale", func(t *testing.T) {
		_, err := svc.GetLocale()
		assertNotImplemented(t, err)
	})
	t.Run("SetLocale", func(t *testing.T) {
		assertNotImplemented(t, svc.SetLocale("en"))
	})
	t.Run("GetCustomLocale", func(t *testing.T) {
		_, err := svc.GetCustomLocale()
		assertNotImplemented(t, err)
	})
	t.Run("SetCustomLocale", func(t *testing.T) {
		assertNotImplemented(t, svc.SetCustomLocale(nil))
	})
	t.Run("GetLatestRelease", func(t *testing.T) {
		_, err := svc.GetLatestRelease()
		assertNotImplemented(t, err)
	})
}

func TestLedgerService(t *testing.T) {
	svc := NewLedgerService()
	assert.NotNil(t, svc)

	t.Run("GetAppVersion", func(t *testing.T) {
		_, err := svc.GetAppVersion()
		assertNotImplemented(t, err)
	})
	t.Run("GetXPub", func(t *testing.T) {
		_, err := svc.GetXPub("m/44'/5353'/0'")
		assertNotImplemented(t, err)
	})
}

func TestDBService(t *testing.T) {
	svc := NewDBService()
	assert.NotNil(t, svc)

	t.Run("Open", func(t *testing.T) {
		assertNotImplemented(t, svc.Open("test"))
	})
	t.Run("Close", func(t *testing.T) {
		assertNotImplemented(t, svc.Close())
	})
	t.Run("Get", func(t *testing.T) {
		_, err := svc.Get("key")
		assertNotImplemented(t, err)
	})
	t.Run("Put", func(t *testing.T) {
		assertNotImplemented(t, svc.Put("key", "value"))
	})
	t.Run("Del", func(t *testing.T) {
		assertNotImplemented(t, svc.Del("key"))
	})
	t.Run("GetUserDir", func(t *testing.T) {
		_, err := svc.GetUserDir()
		assertNotImplemented(t, err)
	})
}

func TestAnalyticsService(t *testing.T) {
	svc := NewAnalyticsService()
	assert.NotNil(t, svc)

	t.Run("GetOptIn", func(t *testing.T) {
		_, err := svc.GetOptIn()
		assertNotImplemented(t, err)
	})
	t.Run("SetOptIn", func(t *testing.T) {
		assertNotImplemented(t, svc.SetOptIn(true))
	})
	t.Run("Track", func(t *testing.T) {
		assertNotImplemented(t, svc.Track("event", nil))
	})
}

func TestConnectionsService(t *testing.T) {
	svc := NewConnectionsService()
	assert.NotNil(t, svc)

	t.Run("connection types are defined", func(t *testing.T) {
		assert.Equal(t, ConnectionType("local"), ConnectionLocal)
		assert.Equal(t, ConnectionType("p2p"), ConnectionP2P)
		assert.Equal(t, ConnectionType("custom"), ConnectionCustom)
	})

	t.Run("GetConnection", func(t *testing.T) {
		_, err := svc.GetConnection()
		assertNotImplemented(t, err)
	})
	t.Run("SetConnection", func(t *testing.T) {
		assertNotImplemented(t, svc.SetConnection(nil))
	})
	t.Run("SetConnectionType", func(t *testing.T) {
		assertNotImplemented(t, svc.SetConnectionType(ConnectionLocal))
	})
	t.Run("GetCustomRPC", func(t *testing.T) {
		_, err := svc.GetCustomRPC()
		assertNotImplemented(t, err)
	})
}

func TestShakedexService(t *testing.T) {
	svc := NewShakedexService()
	assert.NotNil(t, svc)

	t.Run("GetListings", func(t *testing.T) {
		_, err := svc.GetListings()
		assertNotImplemented(t, err)
	})
	t.Run("FulfillSwap", func(t *testing.T) {
		_, err := svc.FulfillSwap("offer", "addr")
		assertNotImplemented(t, err)
	})
	t.Run("GetFulfillments", func(t *testing.T) {
		_, err := svc.GetFulfillments()
		assertNotImplemented(t, err)
	})
	t.Run("FinalizeSwap", func(t *testing.T) {
		_, err := svc.FinalizeSwap("id")
		assertNotImplemented(t, err)
	})
	t.Run("TransferLock", func(t *testing.T) {
		_, err := svc.TransferLock("name", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("TransferCancel", func(t *testing.T) {
		_, err := svc.TransferCancel("name", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("FinalizeLock", func(t *testing.T) {
		_, err := svc.FinalizeLock("name", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("FinalizeCancel", func(t *testing.T) {
		_, err := svc.FinalizeCancel("name", "pass")
		assertNotImplemented(t, err)
	})
	t.Run("LaunchAuction", func(t *testing.T) {
		_, err := svc.LaunchAuction("name", nil)
		assertNotImplemented(t, err)
	})
	t.Run("DownloadProofs", func(t *testing.T) {
		_, err := svc.DownloadProofs("id")
		assertNotImplemented(t, err)
	})
	t.Run("RestoreOneListing", func(t *testing.T) {
		assertNotImplemented(t, svc.RestoreOneListing("id"))
	})
	t.Run("RestoreOneFill", func(t *testing.T) {
		assertNotImplemented(t, svc.RestoreOneFill("id"))
	})
	t.Run("GetExchangeAuctions", func(t *testing.T) {
		_, err := svc.GetExchangeAuctions()
		assertNotImplemented(t, err)
	})
	t.Run("ListAuction", func(t *testing.T) {
		_, err := svc.ListAuction("name", 100, 200, 3600)
		assertNotImplemented(t, err)
	})
	t.Run("GetFeeInfo", func(t *testing.T) {
		_, err := svc.GetFeeInfo()
		assertNotImplemented(t, err)
	})
	t.Run("GetBestBid", func(t *testing.T) {
		_, err := svc.GetBestBid("id")
		assertNotImplemented(t, err)
	})
}

func TestClaimService(t *testing.T) {
	svc := NewClaimService()
	assert.NotNil(t, svc)

	t.Run("AirdropGenerateProofs", func(t *testing.T) {
		_, err := svc.AirdropGenerateProofs("address")
		assertNotImplemented(t, err)
	})
}

func TestLoggerService(t *testing.T) {
	svc := NewLoggerService()
	assert.NotNil(t, svc)

	t.Run("Info", func(t *testing.T) {
		assertNotImplemented(t, svc.Info("message"))
	})
	t.Run("Warn", func(t *testing.T) {
		assertNotImplemented(t, svc.Warn("message"))
	})
	t.Run("Error", func(t *testing.T) {
		assertNotImplemented(t, svc.Error("message"))
	})
	t.Run("Log", func(t *testing.T) {
		assertNotImplemented(t, svc.Log("info", "message"))
	})
	t.Run("Download", func(t *testing.T) {
		_, err := svc.Download()
		assertNotImplemented(t, err)
	})
}

func TestHip2Service(t *testing.T) {
	svc := NewHip2Service()
	assert.NotNil(t, svc)

	t.Run("FetchAddress", func(t *testing.T) {
		_, err := svc.FetchAddress("test.hns")
		assertNotImplemented(t, err)
	})
}
