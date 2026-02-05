package mcp

import (
	"context"
	"net"
	"os"

	"github.com/host-uk/core/pkg/log"
)

// ServeUnix starts a Unix domain socket server for the MCP service.
// The socket file is created at the given path and removed on shutdown.
// It accepts connections and spawns a new MCP server session for each connection.
func (s *Service) ServeUnix(ctx context.Context, socketPath string) error {
	// Clean up any stale socket file
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("Failed to remove stale socket", "path", socketPath, "err", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
	}()

	// Close listener when context is cancelled to unblock Accept
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	s.logger.Security("MCP Unix server listening", "path", socketPath, "user", log.Username())

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				s.logger.Error("MCP Unix accept error", "err", err, "user", log.Username())
				continue
			}
		}

		s.logger.Security("MCP Unix connection accepted", "user", log.Username())
		go s.handleConnection(ctx, conn)
	}
}
