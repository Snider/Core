package mcp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// maxMCPMessageSize is the maximum size for MCP JSON-RPC messages (10 MB).
const maxMCPMessageSize = 10 * 1024 * 1024

// TCPTransport manages a TCP listener for MCP.
type TCPTransport struct {
	addr     string
	listener net.Listener
}

// DefaultTCPAddr is the default address for the MCP TCP transport.
const DefaultTCPAddr = "127.0.0.1:9100"

// NewTCPTransport creates a new TCP transport listener.
// It listens on the provided address (e.g. "localhost:9100").
// If addr is empty, it defaults to 127.0.0.1:9100.
// A warning is printed to stderr if binding to 0.0.0.0 (all interfaces).
func NewTCPTransport(addr string) (*TCPTransport, error) {
	if addr == "" {
		addr = DefaultTCPAddr
	}

	// Warn if binding to all interfaces
	if strings.HasPrefix(addr, "0.0.0.0:") {
		fmt.Fprintln(os.Stderr, "WARNING: MCP TCP server binding to all interfaces (0.0.0.0). This may expose the service to the network.")
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPTransport{addr: addr, listener: listener}, nil
}

// ServeTCP starts a TCP server for the MCP service.
// It accepts connections and spawns a new MCP server session for each connection.
func (s *Service) ServeTCP(ctx context.Context, addr string) error {
	t, err := NewTCPTransport(addr)
	if err != nil {
		return err
	}
	defer func() { _ = t.listener.Close() }()

	// Close listener when context is cancelled to unblock Accept
	go func() {
		<-ctx.Done()
		_ = t.listener.Close()
	}()

	if addr == "" {
		addr = t.listener.Addr().String()
	}
s.logger.Security("MCP TCP server listening", "addr", addr, "user", log.Username())

	for {
		conn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
s.logger.Error("MCP TCP accept error", "err", err, "user", log.Username())
				continue
			}
		}

		s.logger.Security("MCP TCP connection accepted", "remote", conn.RemoteAddr().String(), "user", log.Username())
		go s.handleConnection(ctx, conn)
	}
}

func (s *Service) handleConnection(ctx context.Context, conn net.Conn) {
	// Note: We don't defer conn.Close() here because it's closed by the Server/Transport

	// Create new server instance for this connection
	impl := &mcp.Implementation{
		Name:    "core-cli",
		Version: "0.1.0",
	}
	server := mcp.NewServer(impl, nil)
	s.registerTools(server)

	// Create transport for this connection
	transport := &connTransport{conn: conn}

	// Run server (blocks until connection closed)
	// Server.Run calls Connect, then Read loop.
	if err := server.Run(ctx, transport); err != nil {
s.logger.Error("MCP TCP connection error", "err", err, "remote", conn.RemoteAddr().String(), "user", log.Username())
	}
}

// connTransport adapts net.Conn to mcp.Transport
type connTransport struct {
	conn net.Conn
}

func (t *connTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	scanner := bufio.NewScanner(t.conn)
	scanner.Buffer(make([]byte, 64*1024), maxMCPMessageSize)
	return &connConnection{
		conn:    t.conn,
		scanner: scanner,
	}, nil
}

// connConnection implements mcp.Connection
type connConnection struct {
	conn    net.Conn
	scanner *bufio.Scanner
}

func (c *connConnection) Read(ctx context.Context) (jsonrpc.Message, error) {
	// Blocks until line is read
	if !c.scanner.Scan() {
		if err := c.scanner.Err(); err != nil {
			return nil, err
		}
		// EOF - connection closed cleanly
		return nil, io.EOF
	}
	line := c.scanner.Bytes()
	return jsonrpc.DecodeMessage(line)
}

func (c *connConnection) Write(ctx context.Context, msg jsonrpc.Message) error {
	data, err := jsonrpc.EncodeMessage(msg)
	if err != nil {
		return err
	}
	// Append newline for line-delimited JSON
	data = append(data, '\n')
	_, err = c.conn.Write(data)
	return err
}

func (c *connConnection) Close() error {
	return c.conn.Close()
}

func (c *connConnection) SessionID() string {
	return "tcp-session" // Unique ID might be better, but optional
}
