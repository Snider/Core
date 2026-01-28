package php

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	// DefaultSSLDir is the default directory for SSL certificates.
	DefaultSSLDir = ".core/ssl"
)

// SSLOptions configures SSL certificate generation.
type SSLOptions struct {
	// Dir is the directory to store certificates.
	// Defaults to ~/.core/ssl/
	Dir string
}

// GetSSLDir returns the SSL directory, creating it if necessary.
func GetSSLDir(opts SSLOptions) (string, error) {
	dir := opts.Dir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		dir = filepath.Join(home, DefaultSSLDir)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create SSL directory: %w", err)
	}

	return dir, nil
}

// CertPaths returns the paths to the certificate and key files for a domain.
func CertPaths(domain string, opts SSLOptions) (certFile, keyFile string, err error) {
	dir, err := GetSSLDir(opts)
	if err != nil {
		return "", "", err
	}

	certFile = filepath.Join(dir, fmt.Sprintf("%s.pem", domain))
	keyFile = filepath.Join(dir, fmt.Sprintf("%s-key.pem", domain))

	return certFile, keyFile, nil
}

// CertsExist checks if SSL certificates exist for the given domain.
func CertsExist(domain string, opts SSLOptions) bool {
	certFile, keyFile, err := CertPaths(domain, opts)
	if err != nil {
		return false
	}

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return false
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// SetupSSL creates local SSL certificates using mkcert.
// It installs the local CA if not already installed and generates
// certificates for the given domain.
func SetupSSL(domain string, opts SSLOptions) error {
	// Check if mkcert is installed
	if _, err := exec.LookPath("mkcert"); err != nil {
		return fmt.Errorf("mkcert is not installed. Install it with: brew install mkcert (macOS) or see https://github.com/FiloSottile/mkcert")
	}

	dir, err := GetSSLDir(opts)
	if err != nil {
		return err
	}

	// Install local CA (idempotent operation)
	installCmd := exec.Command("mkcert", "-install")
	if output, err := installCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install mkcert CA: %w\n%s", err, output)
	}

	// Generate certificates
	certFile := filepath.Join(dir, fmt.Sprintf("%s.pem", domain))
	keyFile := filepath.Join(dir, fmt.Sprintf("%s-key.pem", domain))

	// mkcert generates cert and key with specific naming
	genCmd := exec.Command("mkcert",
		"-cert-file", certFile,
		"-key-file", keyFile,
		domain,
		"localhost",
		"127.0.0.1",
		"::1",
	)

	if output, err := genCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to generate certificates: %w\n%s", err, output)
	}

	return nil
}

// SetupSSLIfNeeded checks if certificates exist and creates them if not.
func SetupSSLIfNeeded(domain string, opts SSLOptions) (certFile, keyFile string, err error) {
	certFile, keyFile, err = CertPaths(domain, opts)
	if err != nil {
		return "", "", err
	}

	if !CertsExist(domain, opts) {
		if err := SetupSSL(domain, opts); err != nil {
			return "", "", err
		}
	}

	return certFile, keyFile, nil
}

// IsMkcertInstalled checks if mkcert is available in PATH.
func IsMkcertInstalled() bool {
	_, err := exec.LookPath("mkcert")
	return err == nil
}

// InstallMkcertCA installs the local CA for mkcert.
func InstallMkcertCA() error {
	if !IsMkcertInstalled() {
		return fmt.Errorf("mkcert is not installed")
	}

	cmd := exec.Command("mkcert", "-install")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install mkcert CA: %w\n%s", err, output)
	}

	return nil
}

// GetMkcertCARoot returns the path to the mkcert CA root directory.
func GetMkcertCARoot() (string, error) {
	if !IsMkcertInstalled() {
		return "", fmt.Errorf("mkcert is not installed")
	}

	cmd := exec.Command("mkcert", "-CAROOT")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get mkcert CA root: %w", err)
	}

	return filepath.Clean(string(output)), nil
}
