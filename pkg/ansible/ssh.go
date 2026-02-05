package ansible

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/host-uk/core/pkg/log"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHClient handles SSH connections to remote hosts.
type SSHClient struct {
	host       string
	port       int
	user       string
	password   string
	keyFile    string
	client     *ssh.Client
	mu         sync.Mutex
	become     bool
	becomeUser string
	becomePass string
	timeout    time.Duration
}

// SSHConfig holds SSH connection configuration.
type SSHConfig struct {
	Host       string
	Port       int
	User       string
	Password   string
	KeyFile    string
	Become     bool
	BecomeUser string
	BecomePass string
	Timeout    time.Duration
}

// NewSSHClient creates a new SSH client.
func NewSSHClient(cfg SSHConfig) (*SSHClient, error) {
	if cfg.Port == 0 {
		cfg.Port = 22
	}
	if cfg.User == "" {
		cfg.User = "root"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	client := &SSHClient{
		host:       cfg.Host,
		port:       cfg.Port,
		user:       cfg.User,
		password:   cfg.Password,
		keyFile:    cfg.KeyFile,
		become:     cfg.Become,
		becomeUser: cfg.BecomeUser,
		becomePass: cfg.BecomePass,
		timeout:    cfg.Timeout,
	}

	return client, nil
}

// Connect establishes the SSH connection.
func (c *SSHClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		return nil
	}

	var authMethods []ssh.AuthMethod

	// Try key-based auth first
	if c.keyFile != "" {
		keyPath := c.keyFile
		if strings.HasPrefix(keyPath, "~") {
			home, _ := os.UserHomeDir()
			keyPath = filepath.Join(home, keyPath[1:])
		}

		if key, err := os.ReadFile(keyPath); err == nil {
			if signer, err := ssh.ParsePrivateKey(key); err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	// Try default SSH keys
	if len(authMethods) == 0 {
		home, _ := os.UserHomeDir()
		defaultKeys := []string{
			filepath.Join(home, ".ssh", "id_ed25519"),
			filepath.Join(home, ".ssh", "id_rsa"),
		}
		for _, keyPath := range defaultKeys {
			if key, err := os.ReadFile(keyPath); err == nil {
				if signer, err := ssh.ParsePrivateKey(key); err == nil {
					authMethods = append(authMethods, ssh.PublicKeys(signer))
					break
				}
			}
		}
	}

	// Fall back to password auth
	if c.password != "" {
		authMethods = append(authMethods, ssh.Password(c.password))
		authMethods = append(authMethods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i := range questions {
				answers[i] = c.password
			}
			return answers, nil
		}))
	}

	if len(authMethods) == 0 {
		return log.E("ssh.Connect", "no authentication method available", nil)
	}

	// Host key verification
	var hostKeyCallback ssh.HostKeyCallback

	home, err := os.UserHomeDir()
	if err != nil {
		return log.E("ssh.Connect", "failed to get user home dir", err)
	}
	knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")

	// Ensure known_hosts file exists
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(knownHostsPath), 0700); err != nil {
			return log.E("ssh.Connect", "failed to create .ssh dir", err)
		}
		if err := os.WriteFile(knownHostsPath, nil, 0600); err != nil {
			return log.E("ssh.Connect", "failed to create known_hosts file", err)
		}
	}

	cb, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return log.E("ssh.Connect", "failed to load known_hosts", err)
	}
	hostKeyCallback = cb

	config := &ssh.ClientConfig{
		User:            c.user,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         c.timeout,
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)

	// Connect with context timeout
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return log.E("ssh.Connect", fmt.Sprintf("dial %s", addr), err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		// conn is closed by NewClientConn on error
		return log.E("ssh.Connect", fmt.Sprintf("ssh connect %s", addr), err)
	}

	c.client = ssh.NewClient(sshConn, chans, reqs)
	return nil
}

// Close closes the SSH connection.
func (c *SSHClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		err := c.client.Close()
		c.client = nil
		return err
	}
	return nil
}

// Run executes a command on the remote host.
func (c *SSHClient) Run(ctx context.Context, cmd string) (stdout, stderr string, exitCode int, err error) {
	if err := c.Connect(ctx); err != nil {
		return "", "", -1, err
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", "", -1, log.E("ssh.Run", "new session", err)
	}
	defer func() { _ = session.Close() }()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// Apply become if needed
	if c.become {
		becomeUser := c.becomeUser
		if becomeUser == "" {
			becomeUser = "root"
		}
		// Escape single quotes in the command
		escapedCmd := strings.ReplaceAll(cmd, "'", "'\\''")
		if c.becomePass != "" {
			// Use sudo with password via stdin (-S flag)
			// We launch a goroutine to write the password to stdin
			cmd = fmt.Sprintf("sudo -S -u %s bash -c '%s'", becomeUser, escapedCmd)
			stdin, err := session.StdinPipe()
			if err != nil {
				return "", "", -1, log.E("ssh.Run", "stdin pipe", err)
			}
			go func() {
				defer func() { _ = stdin.Close() }()
				_, _ = io.WriteString(stdin, c.becomePass+"\n")
			}()
		} else if c.password != "" {
			// Try using connection password for sudo
			cmd = fmt.Sprintf("sudo -S -u %s bash -c '%s'", becomeUser, escapedCmd)
			stdin, err := session.StdinPipe()
			if err != nil {
				return "", "", -1, log.E("ssh.Run", "stdin pipe", err)
			}
			go func() {
				defer func() { _ = stdin.Close() }()
				_, _ = io.WriteString(stdin, c.password+"\n")
			}()
		} else {
			// Try passwordless sudo
			cmd = fmt.Sprintf("sudo -n -u %s bash -c '%s'", becomeUser, escapedCmd)
		}
	}

	// Run with context
	done := make(chan error, 1)
	go func() {
		done <- session.Run(cmd)
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGKILL)
		return "", "", -1, ctx.Err()
	case err := <-done:
		exitCode = 0
		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				exitCode = exitErr.ExitStatus()
			} else {
				return stdoutBuf.String(), stderrBuf.String(), -1, err
			}
		}
		return stdoutBuf.String(), stderrBuf.String(), exitCode, nil
	}
}

// RunScript runs a script on the remote host.
func (c *SSHClient) RunScript(ctx context.Context, script string) (stdout, stderr string, exitCode int, err error) {
	// Escape the script for heredoc
	cmd := fmt.Sprintf("bash <<'ANSIBLE_SCRIPT_EOF'\n%s\nANSIBLE_SCRIPT_EOF", script)
	return c.Run(ctx, cmd)
}

// Upload copies a file to the remote host.
func (c *SSHClient) Upload(ctx context.Context, local io.Reader, remote string, mode os.FileMode) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	// Read content
	content, err := io.ReadAll(local)
	if err != nil {
		return log.E("ssh.Upload", "read content", err)
	}

	// Create parent directory
	dir := filepath.Dir(remote)
	dirCmd := fmt.Sprintf("mkdir -p %q", dir)
	if c.become {
		dirCmd = fmt.Sprintf("sudo mkdir -p %q", dir)
	}
	if _, _, _, err := c.Run(ctx, dirCmd); err != nil {
		return log.E("ssh.Upload", "create parent dir", err)
	}

	// Use cat to write the file (simpler than SCP)
	writeCmd := fmt.Sprintf("cat > %q && chmod %o %q", remote, mode, remote)

	// If become is needed, we construct a command that reads password then content from stdin
	// But we need to be careful with handling stdin for sudo + cat.
	// We'll use a session with piped stdin.

	session2, err := c.client.NewSession()
	if err != nil {
		return log.E("ssh.Upload", "new session for write", err)
	}
	defer func() { _ = session2.Close() }()

	stdin, err := session2.StdinPipe()
	if err != nil {
		return log.E("ssh.Upload", "stdin pipe", err)
	}

	var stderrBuf bytes.Buffer
	session2.Stderr = &stderrBuf

	if c.become {
		becomeUser := c.becomeUser
		if becomeUser == "" {
			becomeUser = "root"
		}

		pass := c.becomePass
		if pass == "" {
			pass = c.password
		}

		if pass != "" {
			// Use sudo -S with password from stdin
			writeCmd = fmt.Sprintf("sudo -S -u %s bash -c 'cat > %q && chmod %o %q'",
				becomeUser, remote, mode, remote)
		} else {
			// Use passwordless sudo (sudo -n) to avoid consuming file content as password
			writeCmd = fmt.Sprintf("sudo -n -u %s bash -c 'cat > %q && chmod %o %q'",
				becomeUser, remote, mode, remote)
		}

		if err := session2.Start(writeCmd); err != nil {
			return log.E("ssh.Upload", "start write", err)
		}

		go func() {
			defer func() { _ = stdin.Close() }()
			if pass != "" {
				_, _ = io.WriteString(stdin, pass+"\n")
			}
			_, _ = stdin.Write(content)
		}()
	} else {
		// Normal write
		if err := session2.Start(writeCmd); err != nil {
			return log.E("ssh.Upload", "start write", err)
		}

		go func() {
			defer func() { _ = stdin.Close() }()
			_, _ = stdin.Write(content)
		}()
	}

	if err := session2.Wait(); err != nil {
		return log.E("ssh.Upload", fmt.Sprintf("write failed (stderr: %s)", stderrBuf.String()), err)
	}

	return nil
}

// Download copies a file from the remote host.
func (c *SSHClient) Download(ctx context.Context, remote string) ([]byte, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	cmd := fmt.Sprintf("cat %q", remote)

	stdout, stderr, exitCode, err := c.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if exitCode != 0 {
		return nil, log.E("ssh.Download", fmt.Sprintf("cat failed: %s", stderr), nil)
	}

	return []byte(stdout), nil
}

// FileExists checks if a file exists on the remote host.
func (c *SSHClient) FileExists(ctx context.Context, path string) (bool, error) {
	cmd := fmt.Sprintf("test -e %q && echo yes || echo no", path)
	stdout, _, exitCode, err := c.Run(ctx, cmd)
	if err != nil {
		return false, err
	}
	if exitCode != 0 {
		// test command failed but didn't error - file doesn't exist
		return false, nil
	}
	return strings.TrimSpace(stdout) == "yes", nil
}

// Stat returns file info from the remote host.
func (c *SSHClient) Stat(ctx context.Context, path string) (map[string]any, error) {
	// Simple approach - get basic file info
	cmd := fmt.Sprintf(`
if [ -e %q ]; then
  if [ -d %q ]; then
    echo "exists=true isdir=true"
  else
    echo "exists=true isdir=false"
  fi
else
  echo "exists=false"
fi
`, path, path)

	stdout, _, _, err := c.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any)
	parts := strings.Fields(strings.TrimSpace(stdout))
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1] == "true"
		}
	}

	return result, nil
}

// SetBecome enables privilege escalation.
func (c *SSHClient) SetBecome(become bool, user, password string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.become = become
	if user != "" {
		c.becomeUser = user
	}
	if password != "" {
		c.becomePass = password
	}
}
