// IPC Stub classes mapping old Electron IPC services and methods.
// These stubs let the web build compile and run without native IPC.
// Each method throws a NotImplementedError to highlight what needs
// to be replaced in a native wrapper or future web-compatible API.
//
// WAILS3 INTEGRATION:
// These stubs will be replaced by Wails3 auto-generated bindings.
// See WAILS3_INTEGRATION.md for complete migration guide.
//
// Pattern:
// 1. Create Go service structs with exported methods (e.g., NodeService, WalletService)
// 2. Register services in Wails3 main.go: application.NewService(&NodeService{})
// 3. Run `wails3 generate bindings` to create TypeScript bindings
// 4. Import generated bindings: import { GetInfo } from '../bindings/.../nodeservice'
// 5. Replace stub calls with binding calls: await GetInfo() instead of IPC.Node.getInfo()
//
// Each service below maps 1:1 to a Go service struct that will be created.

export class NotImplementedError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'NotImplementedError';
  }
}

function notImplemented(service: string, method: string): never {
  throw new NotImplementedError(`IPC ${service}.${method} is not implemented in the web build.`);
}

function makeIpcStub<T extends Record<string, any>>(service: string, methods: string[]): T {
  const obj: Record<string, any> = {};
  for (const m of methods) {
    obj[m] = (..._args: any[]) => notImplemented(service, m);
  }
  return obj as T;
}

// Services and their methods as defined in old/app/background/**/client.js
export const Node = makeIpcStub('Node', [
  'start',
  'stop',
  'reset',
  'generateToAddress',
  'getAPIKey',
  'getNoDns',
  'getSpvMode',
  'getInfo',
  'getNameInfo',
  'getTXByAddresses',
  'getNameByHash',
  'getBlockByHeight',
  'getTx',
  'broadcastRawTx',
  'sendRawAirdrop',
  'getFees',
  'getAverageBlockTime',
  'getMTP',
  'getCoin',
  'verifyMessageWithName',
  'setNodeDir',
  'setAPIKey',
  'setNoDns',
  'setSpvMode',
  'getDir',
  'getHNSPrice',
  'testCustomRPCClient',
  'getDNSSECProof',
  'sendRawClaim',
]);

export const Wallet = makeIpcStub('Wallet', [
  'start',
  'getAPIKey',
  'setAPIKey',
  'getWalletInfo',
  'getAccountInfo',
  'getCoin',
  'getTX',
  'getNames',
  'createNewWallet',
  'importSeed',
  'generateReceivingAddress',
  'getAuctionInfo',
  'getTransactionHistory',
  'getPendingTransactions',
  'getBids',
  'getBlind',
  'getMasterHDKey',
  'hasAddress',
  'setPassphrase',
  'revealSeed',
  'estimateTxFee',
  'estimateMaxSend',
  'removeWalletById',
  'updateAccountDepth',
  'findNonce',
  'findNonceCancel',
  'encryptWallet',
  'backup',
  'rescan',
  'deepClean',
  'reset',
  'sendOpen',
  'sendBid',
  'sendRegister',
  'sendUpdate',
  'sendReveal',
  'sendRedeem',
  'sendRenewal',
  'sendRevealAll',
  'sendRedeemAll',
  'sendRegisterAll',
  'signMessageWithName',
  'transferMany',
  'finalizeAll',
  'finalizeMany',
  'renewAll',
  'renewMany',
  'sendTransfer',
  'cancelTransfer',
  'finalizeTransfer',
  'finalizeWithPayment',
  'claimPaidTransfer',
  'revokeName',
  'send',
  'lock',
  'unlock',
  'isLocked',
  'addSharedKey',
  'removeSharedKey',
  'getNonce',
  'importNonce',
  'zap',
  'importName',
  'rpcGetWalletInfo',
  'loadTransaction',
  'listWallets',
  'getStats',
  'isReady',
  'createClaim',
  'sendClaim',
]);

export const Setting = makeIpcStub('Setting', [
  'getExplorer',
  'setExplorer',
  'getLocale',
  'setLocale',
  'getCustomLocale',
  'setCustomLocale',
  'getLatestRelease',
]);

export const Ledger = makeIpcStub('Ledger', [
  'getXPub',
  'getAppVersion',
]);

export const DB = makeIpcStub('DB', [
  'open',
  'close',
  'put',
  'get',
  'del',
  'getUserDir',
]);

export const Analytics = makeIpcStub('Analytics', [
  'setOptIn',
  'getOptIn',
  'track',
  'screenView',
]);

export const Connections = makeIpcStub('Connections', [
  'getConnection',
  'setConnection',
  'setConnectionType',
  'getCustomRPC',
]);

export const Shakedex = makeIpcStub('Shakedex', [
  'fulfillSwap',
  'getFulfillments',
  'finalizeSwap',
  'transferLock',
  'transferCancel',
  'getListings',
  'finalizeLock',
  'finalizeCancel',
  'launchAuction',
  'downloadProofs',
  'restoreOneListing',
  'restoreOneFill',
  'getExchangeAuctions',
  'listAuction',
  'getFeeInfo',
  'getBestBid',
]);

export const Claim = makeIpcStub('Claim', [
  'airdropGenerateProofs',
]);

export const Logger = makeIpcStub('Logger', [
  'info',
  'warn',
  'error',
  'log',
  'download',
]);

export const Hip2 = makeIpcStub('Hip2', [
  'getPort',
  'setPort',
  'fetchAddress',
  'setServers',
]);

// Aggregate facade to import from components/services if needed
export const IPC = {
  Node,
  Wallet,
  Setting,
  Ledger,
  DB,
  Analytics,
  Connections,
  Shakedex,
  Claim,
  Logger,
  Hip2,
};
