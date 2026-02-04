package php

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseService_Name_Good(t *testing.T) {
	t.Run("returns service name", func(t *testing.T) {
		s := &baseService{name: "TestService"}
		assert.Equal(t, "TestService", s.Name())
	})
}

func TestBaseService_Status_Good(t *testing.T) {
	t.Run("returns status when not running", func(t *testing.T) {
		s := &baseService{
			name:    "TestService",
			port:    8080,
			running: false,
		}

		status := s.Status()
		assert.Equal(t, "TestService", status.Name)
		assert.Equal(t, 8080, status.Port)
		assert.False(t, status.Running)
		assert.Equal(t, 0, status.PID)
	})

	t.Run("returns status when running", func(t *testing.T) {
		s := &baseService{
			name:    "TestService",
			port:    8080,
			running: true,
		}

		status := s.Status()
		assert.True(t, status.Running)
	})

	t.Run("returns error in status", func(t *testing.T) {
		testErr := assert.AnError
		s := &baseService{
			name:      "TestService",
			lastError: testErr,
		}

		status := s.Status()
		assert.Equal(t, testErr, status.Error)
	})
}

func TestBaseService_Logs_Good(t *testing.T) {
	t.Run("returns log file content", func(t *testing.T) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "test.log")
		err := os.WriteFile(logPath, []byte("test log content"), 0644)
		require.NoError(t, err)

		s := &baseService{logPath: logPath}
		reader, err := s.Logs(false)

		assert.NoError(t, err)
		assert.NotNil(t, reader)
		_ = reader.Close()
	})

	t.Run("returns tail reader in follow mode", func(t *testing.T) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "test.log")
		err := os.WriteFile(logPath, []byte("test log content"), 0644)
		require.NoError(t, err)

		s := &baseService{logPath: logPath}
		reader, err := s.Logs(true)

		assert.NoError(t, err)
		assert.NotNil(t, reader)
		// Verify it's a tailReader by checking it implements ReadCloser
		_, ok := reader.(*tailReader)
		assert.True(t, ok)
		_ = reader.Close()
	})
}

func TestBaseService_Logs_Bad(t *testing.T) {
	t.Run("returns error when no log path", func(t *testing.T) {
		s := &baseService{name: "TestService"}
		_, err := s.Logs(false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no log file available")
	})

	t.Run("returns error when log file doesn't exist", func(t *testing.T) {
		s := &baseService{logPath: "/nonexistent/path/log.log"}
		_, err := s.Logs(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to open log file")
	})
}

func TestTailReader_Good(t *testing.T) {
	t.Run("creates new tail reader", func(t *testing.T) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "test.log")
		err := os.WriteFile(logPath, []byte("content"), 0644)
		require.NoError(t, err)

		file, err := os.Open(logPath)
		require.NoError(t, err)
		defer func() { _ = file.Close() }()

		reader := newTailReader(file)
		assert.NotNil(t, reader)
		assert.NotNil(t, reader.file)
		assert.NotNil(t, reader.reader)
		assert.False(t, reader.closed)
	})

	t.Run("closes file on Close", func(t *testing.T) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "test.log")
		err := os.WriteFile(logPath, []byte("content"), 0644)
		require.NoError(t, err)

		file, err := os.Open(logPath)
		require.NoError(t, err)

		reader := newTailReader(file)
		err = reader.Close()
		assert.NoError(t, err)
		assert.True(t, reader.closed)
	})

	t.Run("returns EOF when closed", func(t *testing.T) {
		dir := t.TempDir()
		logPath := filepath.Join(dir, "test.log")
		err := os.WriteFile(logPath, []byte("content"), 0644)
		require.NoError(t, err)

		file, err := os.Open(logPath)
		require.NoError(t, err)

		reader := newTailReader(file)
		_ = reader.Close()

		buf := make([]byte, 100)
		n, _ := reader.Read(buf)
		// When closed, should return 0 bytes (the closed flag causes early return)
		assert.Equal(t, 0, n)
	})
}

func TestFrankenPHPService_Extended(t *testing.T) {
	t.Run("all options set correctly", func(t *testing.T) {
		opts := FrankenPHPOptions{
			Port:      9000,
			HTTPSPort: 9443,
			HTTPS:     true,
			CertFile:  "/path/to/cert.pem",
			KeyFile:   "/path/to/key.pem",
		}

		service := NewFrankenPHPService("/project", opts)

		assert.Equal(t, "FrankenPHP", service.Name())
		assert.Equal(t, 9000, service.port)
		assert.Equal(t, 9443, service.httpsPort)
		assert.True(t, service.https)
		assert.Equal(t, "/path/to/cert.pem", service.certFile)
		assert.Equal(t, "/path/to/key.pem", service.keyFile)
		assert.Equal(t, "/project", service.dir)
	})
}

func TestViteService_Extended(t *testing.T) {
	t.Run("auto-detects package manager", func(t *testing.T) {
		dir := t.TempDir()
		// Create bun.lockb to trigger bun detection
		err := os.WriteFile(filepath.Join(dir, "bun.lockb"), []byte(""), 0644)
		require.NoError(t, err)

		service := NewViteService(dir, ViteOptions{})

		assert.Equal(t, "bun", service.packageManager)
	})

	t.Run("uses provided package manager", func(t *testing.T) {
		dir := t.TempDir()

		service := NewViteService(dir, ViteOptions{PackageManager: "pnpm"})

		assert.Equal(t, "pnpm", service.packageManager)
	})
}

func TestHorizonService_Extended(t *testing.T) {
	t.Run("has zero port", func(t *testing.T) {
		service := NewHorizonService("/project")
		assert.Equal(t, 0, service.port)
	})
}

func TestReverbService_Extended(t *testing.T) {
	t.Run("uses default port 8080", func(t *testing.T) {
		service := NewReverbService("/project", ReverbOptions{})
		assert.Equal(t, 8080, service.port)
	})

	t.Run("uses custom port", func(t *testing.T) {
		service := NewReverbService("/project", ReverbOptions{Port: 9090})
		assert.Equal(t, 9090, service.port)
	})
}

func TestRedisService_Extended(t *testing.T) {
	t.Run("uses default port 6379", func(t *testing.T) {
		service := NewRedisService("/project", RedisOptions{})
		assert.Equal(t, 6379, service.port)
	})

	t.Run("accepts config file", func(t *testing.T) {
		service := NewRedisService("/project", RedisOptions{ConfigFile: "/path/to/redis.conf"})
		assert.Equal(t, "/path/to/redis.conf", service.configFile)
	})
}

func TestServiceStatus_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		testErr := assert.AnError
		status := ServiceStatus{
			Name:    "TestService",
			Running: true,
			PID:     12345,
			Port:    8080,
			Error:   testErr,
		}

		assert.Equal(t, "TestService", status.Name)
		assert.True(t, status.Running)
		assert.Equal(t, 12345, status.PID)
		assert.Equal(t, 8080, status.Port)
		assert.Equal(t, testErr, status.Error)
	})
}

func TestFrankenPHPOptions_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := FrankenPHPOptions{
			Port:      8000,
			HTTPSPort: 443,
			HTTPS:     true,
			CertFile:  "cert.pem",
			KeyFile:   "key.pem",
		}

		assert.Equal(t, 8000, opts.Port)
		assert.Equal(t, 443, opts.HTTPSPort)
		assert.True(t, opts.HTTPS)
		assert.Equal(t, "cert.pem", opts.CertFile)
		assert.Equal(t, "key.pem", opts.KeyFile)
	})
}

func TestViteOptions_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := ViteOptions{
			Port:           5173,
			PackageManager: "bun",
		}

		assert.Equal(t, 5173, opts.Port)
		assert.Equal(t, "bun", opts.PackageManager)
	})
}

func TestReverbOptions_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := ReverbOptions{Port: 8080}
		assert.Equal(t, 8080, opts.Port)
	})
}

func TestRedisOptions_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := RedisOptions{
			Port:       6379,
			ConfigFile: "redis.conf",
		}

		assert.Equal(t, 6379, opts.Port)
		assert.Equal(t, "redis.conf", opts.ConfigFile)
	})
}

func TestBaseService_StopProcess_Good(t *testing.T) {
	t.Run("returns nil when not running", func(t *testing.T) {
		s := &baseService{running: false}
		err := s.stopProcess()
		assert.NoError(t, err)
	})

	t.Run("returns nil when cmd is nil", func(t *testing.T) {
		s := &baseService{running: true, cmd: nil}
		err := s.stopProcess()
		assert.NoError(t, err)
	})
}
