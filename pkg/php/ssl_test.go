package php

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSSLDir_Good(t *testing.T) {
	t.Run("uses provided directory", func(t *testing.T) {
		dir := t.TempDir()
		customDir := filepath.Join(dir, "custom-ssl")

		result, err := GetSSLDir(SSLOptions{Dir: customDir})

		assert.NoError(t, err)
		assert.Equal(t, customDir, result)

		// Verify directory was created
		info, err := os.Stat(result)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("uses default directory when not specified", func(t *testing.T) {
		// Skip if we can't get home dir
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("cannot get home directory")
		}

		result, err := GetSSLDir(SSLOptions{})

		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(home, DefaultSSLDir), result)
	})
}

func TestCertPaths_Good(t *testing.T) {
	t.Run("returns correct paths for domain", func(t *testing.T) {
		dir := t.TempDir()

		certFile, keyFile, err := CertPaths("example.test", SSLOptions{Dir: dir})

		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "example.test.pem"), certFile)
		assert.Equal(t, filepath.Join(dir, "example.test-key.pem"), keyFile)
	})

	t.Run("handles domain with subdomain", func(t *testing.T) {
		dir := t.TempDir()

		certFile, keyFile, err := CertPaths("app.example.test", SSLOptions{Dir: dir})

		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "app.example.test.pem"), certFile)
		assert.Equal(t, filepath.Join(dir, "app.example.test-key.pem"), keyFile)
	})
}

func TestCertsExist_Good(t *testing.T) {
	t.Run("returns true when both files exist", func(t *testing.T) {
		dir := t.TempDir()
		domain := "myapp.test"

		// Create cert and key files
		certFile := filepath.Join(dir, domain+".pem")
		keyFile := filepath.Join(dir, domain+"-key.pem")

		err := os.WriteFile(certFile, []byte("cert content"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(keyFile, []byte("key content"), 0644)
		require.NoError(t, err)

		assert.True(t, CertsExist(domain, SSLOptions{Dir: dir}))
	})
}

func TestCertsExist_Bad(t *testing.T) {
	t.Run("returns false when cert missing", func(t *testing.T) {
		dir := t.TempDir()
		domain := "myapp.test"

		// Create only key file
		keyFile := filepath.Join(dir, domain+"-key.pem")
		err := os.WriteFile(keyFile, []byte("key content"), 0644)
		require.NoError(t, err)

		assert.False(t, CertsExist(domain, SSLOptions{Dir: dir}))
	})

	t.Run("returns false when key missing", func(t *testing.T) {
		dir := t.TempDir()
		domain := "myapp.test"

		// Create only cert file
		certFile := filepath.Join(dir, domain+".pem")
		err := os.WriteFile(certFile, []byte("cert content"), 0644)
		require.NoError(t, err)

		assert.False(t, CertsExist(domain, SSLOptions{Dir: dir}))
	})

	t.Run("returns false when neither exists", func(t *testing.T) {
		dir := t.TempDir()
		domain := "myapp.test"

		assert.False(t, CertsExist(domain, SSLOptions{Dir: dir}))
	})

	t.Run("returns false for invalid directory", func(t *testing.T) {
		// Use invalid directory path
		assert.False(t, CertsExist("domain.test", SSLOptions{Dir: "/nonexistent/path/that/does/not/exist"}))
	})
}

func TestSetupSSL_Bad(t *testing.T) {
	t.Run("returns error when mkcert not installed", func(t *testing.T) {
		// This test assumes mkcert might not be installed
		// If it is installed, we skip this test
		if IsMkcertInstalled() {
			t.Skip("mkcert is installed, skipping error test")
		}

		err := SetupSSL("example.test", SSLOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mkcert is not installed")
	})
}

func TestSetupSSLIfNeeded_Good(t *testing.T) {
	t.Run("returns existing certs without regenerating", func(t *testing.T) {
		dir := t.TempDir()
		domain := "existing.test"

		// Create existing cert files
		certFile := filepath.Join(dir, domain+".pem")
		keyFile := filepath.Join(dir, domain+"-key.pem")

		err := os.WriteFile(certFile, []byte("existing cert"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(keyFile, []byte("existing key"), 0644)
		require.NoError(t, err)

		resultCert, resultKey, err := SetupSSLIfNeeded(domain, SSLOptions{Dir: dir})

		assert.NoError(t, err)
		assert.Equal(t, certFile, resultCert)
		assert.Equal(t, keyFile, resultKey)

		// Verify files weren't modified
		data, err := os.ReadFile(certFile)
		require.NoError(t, err)
		assert.Equal(t, "existing cert", string(data))
	})
}

func TestIsMkcertInstalled_Good(t *testing.T) {
	// This test just verifies the function runs without error
	// The actual result depends on whether mkcert is installed
	result := IsMkcertInstalled()
	t.Logf("mkcert installed: %v", result)
}

func TestDefaultSSLDir_Good(t *testing.T) {
	t.Run("constant has expected value", func(t *testing.T) {
		assert.Equal(t, ".core/ssl", DefaultSSLDir)
	})
}
