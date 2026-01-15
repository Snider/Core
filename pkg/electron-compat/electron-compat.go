// Package electroncompat provides Go implementations for services that were
// previously implemented as Electron IPC handlers. These services bridge
// the frontend Angular application with the Go backend.
//
// Migration from Electron:
// The original application used Electron's IPC for communication between
// the renderer (UI) and main (backend) processes. With the migration to
// Wails v3, these services are now implemented as Go structs with methods
// that Wails automatically exposes to the frontend via generated bindings.
//
// Services in this package:
//   - Node: Blockchain node operations (start/stop, queries, transactions)
//   - Wallet: Wallet management (create, import, send, receive)
//   - Setting: Application settings persistence
//   - Ledger: Hardware wallet (Ledger) integration
//   - DB: Key-value storage for application data
//   - Analytics: Usage tracking and metrics
//   - Connections: RPC connection management
//   - Shakedex: Decentralized exchange operations
//   - Claim: Airdrop and claim proof generation
//   - Logger: Structured logging
//   - Hip2: HIP-2 DNS protocol support
package electroncompat

import "errors"

// ErrNotImplemented is returned when a method stub is called that hasn't
// been fully implemented yet.
var ErrNotImplemented = errors.New("not implemented")

// NotImplementedError wraps an operation name for methods that are stubs.
type NotImplementedError struct {
	Service string
	Method  string
}

func (e *NotImplementedError) Error() string {
	return "IPC " + e.Service + "." + e.Method + " is not implemented"
}

// notImplemented returns a NotImplementedError for the given service and method.
func notImplemented(service, method string) error {
	return &NotImplementedError{Service: service, Method: method}
}
