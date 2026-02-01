package php

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSLOptions_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := SSLOptions{Dir: "/custom/ssl/dir"}
		assert.Equal(t, "/custom/ssl/dir", opts.Dir)
	})
}

func TestGetSSLDir_Bad(t *testing.T) {
	t.Run("fails to create directory in invalid path", func(t *testing.T) {
		// Try to create a directory in a path that can't exist
		opts := SSLOptions{Dir: "/dev/null/cannot/create"}
		_, err := GetSSLDir(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to create SSL directory")
	})
}

func TestCertPaths_Bad(t *testing.T) {
	t.Run("fails when GetSSLDir fails", func(t *testing.T) {
		opts := SSLOptions{Dir: "/dev/null/cannot/create"}
		_, _, err := CertPaths("domain.test", opts)
		assert.Error(t, err)
	})
}

func TestCertsExist_Detailed(t *testing.T) {
	t.Run("returns true when both cert and key exist", func(t *testing.T) {
		dir := t.TempDir()
		domain := "test.local"

		// Create both files
		certPath := filepath.Join(dir, domain+".pem")
		keyPath := filepath.Join(dir, domain+"-key.pem")

		err := os.WriteFile(certPath, []byte("cert"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(keyPath, []byte("key"), 0644)
		require.NoError(t, err)

		result := CertsExist(domain, SSLOptions{Dir: dir})
		assert.True(t, result)
	})

	t.Run("returns false when only cert exists", func(t *testing.T) {
		dir := t.TempDir()
		domain := "test.local"

		certPath := filepath.Join(dir, domain+".pem")
		err := os.WriteFile(certPath, []byte("cert"), 0644)
		require.NoError(t, err)

		result := CertsExist(domain, SSLOptions{Dir: dir})
		assert.False(t, result)
	})

	t.Run("returns false when only key exists", func(t *testing.T) {
		dir := t.TempDir()
		domain := "test.local"

		keyPath := filepath.Join(dir, domain+"-key.pem")
		err := os.WriteFile(keyPath, []byte("key"), 0644)
		require.NoError(t, err)

		result := CertsExist(domain, SSLOptions{Dir: dir})
		assert.False(t, result)
	})

	t.Run("returns false when CertPaths fails", func(t *testing.T) {
		result := CertsExist("domain.test", SSLOptions{Dir: "/dev/null/cannot/create"})
		assert.False(t, result)
	})
}

func TestSetupSSL_RequiresMkcert(t *testing.T) {
	t.Run("fails when mkcert not installed", func(t *testing.T) {
		if IsMkcertInstalled() {
			t.Skip("mkcert is installed, skipping error test")
		}

		err := SetupSSL("example.test", SSLOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mkcert is not installed")
	})
}

func TestSetupSSLIfNeeded_UsesExisting(t *testing.T) {
	t.Run("returns existing certs without regenerating", func(t *testing.T) {
		dir := t.TempDir()
		domain := "existing.test"

		// Create existing certs
		certPath := filepath.Join(dir, domain+".pem")
		keyPath := filepath.Join(dir, domain+"-key.pem")

		err := os.WriteFile(certPath, []byte("existing cert"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(keyPath, []byte("existing key"), 0644)
		require.NoError(t, err)

		resultCert, resultKey, err := SetupSSLIfNeeded(domain, SSLOptions{Dir: dir})

		assert.NoError(t, err)
		assert.Equal(t, certPath, resultCert)
		assert.Equal(t, keyPath, resultKey)

		// Verify original content wasn't changed
		content, _ := os.ReadFile(certPath)
		assert.Equal(t, "existing cert", string(content))
	})
}

func TestSetupSSLIfNeeded_Bad(t *testing.T) {
	t.Run("fails when CertPaths fails", func(t *testing.T) {
		_, _, err := SetupSSLIfNeeded("domain.test", SSLOptions{Dir: "/dev/null/cannot/create"})
		assert.Error(t, err)
	})

	t.Run("fails when SetupSSL fails", func(t *testing.T) {
		if IsMkcertInstalled() {
			t.Skip("mkcert is installed, skipping error test")
		}

		dir := t.TempDir()
		_, _, err := SetupSSLIfNeeded("domain.test", SSLOptions{Dir: dir})
		assert.Error(t, err)
	})
}

func TestInstallMkcertCA_Bad(t *testing.T) {
	t.Run("fails when mkcert not installed", func(t *testing.T) {
		if IsMkcertInstalled() {
			t.Skip("mkcert is installed, skipping error test")
		}

		err := InstallMkcertCA()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mkcert is not installed")
	})
}

func TestGetMkcertCARoot_Bad(t *testing.T) {
	t.Run("fails when mkcert not installed", func(t *testing.T) {
		if IsMkcertInstalled() {
			t.Skip("mkcert is installed, skipping error test")
		}

		_, err := GetMkcertCARoot()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mkcert is not installed")
	})
}

func TestCertPathsNaming(t *testing.T) {
	t.Run("uses correct naming convention", func(t *testing.T) {
		dir := t.TempDir()
		domain := "myapp.example.com"

		certFile, keyFile, err := CertPaths(domain, SSLOptions{Dir: dir})

		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "myapp.example.com.pem"), certFile)
		assert.Equal(t, filepath.Join(dir, "myapp.example.com-key.pem"), keyFile)
	})

	t.Run("handles localhost", func(t *testing.T) {
		dir := t.TempDir()

		certFile, keyFile, err := CertPaths("localhost", SSLOptions{Dir: dir})

		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "localhost.pem"), certFile)
		assert.Equal(t, filepath.Join(dir, "localhost-key.pem"), keyFile)
	})

	t.Run("handles wildcard-like domains", func(t *testing.T) {
		dir := t.TempDir()
		domain := "*.example.com"

		certFile, keyFile, err := CertPaths(domain, SSLOptions{Dir: dir})

		assert.NoError(t, err)
		assert.Contains(t, certFile, "*.example.com.pem")
		assert.Contains(t, keyFile, "*.example.com-key.pem")
	})
}

func TestDefaultSSLDir_Value(t *testing.T) {
	t.Run("has expected default value", func(t *testing.T) {
		assert.Equal(t, ".core/ssl", DefaultSSLDir)
	})
}

func TestGetSSLDir_CreatesDirectory(t *testing.T) {
	t.Run("creates nested directory structure", func(t *testing.T) {
		baseDir := t.TempDir()
		nestedDir := filepath.Join(baseDir, "level1", "level2", "ssl")

		dir, err := GetSSLDir(SSLOptions{Dir: nestedDir})

		assert.NoError(t, err)
		assert.Equal(t, nestedDir, dir)

		// Verify directory exists
		info, err := os.Stat(dir)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}
