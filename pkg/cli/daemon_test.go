package cli

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectMode(t *testing.T) {
	t.Run("daemon mode from env", func(t *testing.T) {
		t.Setenv("CORE_DAEMON", "1")
		assert.Equal(t, ModeDaemon, DetectMode())
	})

	t.Run("mode string", func(t *testing.T) {
		assert.Equal(t, "interactive", ModeInteractive.String())
		assert.Equal(t, "pipe", ModePipe.String())
		assert.Equal(t, "daemon", ModeDaemon.String())
		assert.Equal(t, "unknown", Mode(99).String())
	})
}

func TestPIDFile(t *testing.T) {
	t.Run("acquire and release", func(t *testing.T) {
		m := io.NewMockMedium()
		pidPath := "/tmp/test.pid"

		pid := NewPIDFile(m, pidPath)

		// Acquire should succeed
		err := pid.Acquire()
		require.NoError(t, err)

		// File should exist with our PID
		data, err := m.Read(pidPath)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Release should remove file
		err = pid.Release()
		require.NoError(t, err)

		assert.False(t, m.Exists(pidPath))
	})

	t.Run("stale pid file", func(t *testing.T) {
		m := io.NewMockMedium()
		pidPath := "/tmp/stale.pid"

		// Write a stale PID (non-existent process)
		err := m.Write(pidPath, "999999999")
		require.NoError(t, err)

		pid := NewPIDFile(m, pidPath)

		// Should acquire successfully (stale PID removed)
		err = pid.Acquire()
		require.NoError(t, err)

		err = pid.Release()
		require.NoError(t, err)
	})

	t.Run("creates parent directory", func(t *testing.T) {
		m := io.NewMockMedium()
		pidPath := "/tmp/subdir/nested/test.pid"

		pid := NewPIDFile(m, pidPath)

		err := pid.Acquire()
		require.NoError(t, err)

		assert.True(t, m.Exists(pidPath))

		err = pid.Release()
		require.NoError(t, err)
	})

	t.Run("path getter", func(t *testing.T) {
		m := io.NewMockMedium()
		pid := NewPIDFile(m, "/tmp/test.pid")
		assert.Equal(t, "/tmp/test.pid", pid.Path())
	})
}

func TestHealthServer(t *testing.T) {
	t.Run("health and ready endpoints", func(t *testing.T) {
		hs := NewHealthServer("127.0.0.1:0") // Random port

		err := hs.Start()
		require.NoError(t, err)
		defer func() { _ = hs.Stop(context.Background()) }()

		addr := hs.Addr()
		require.NotEmpty(t, addr)

		// Health should be OK
		resp, err := http.Get("http://" + addr + "/health")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()

		// Ready should be OK by default
		resp, err = http.Get("http://" + addr + "/ready")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()

		// Set not ready
		hs.SetReady(false)

		resp, err = http.Get("http://" + addr + "/ready")
		require.NoError(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		_ = resp.Body.Close()
	})

	t.Run("with health checks", func(t *testing.T) {
		hs := NewHealthServer("127.0.0.1:0")

		healthy := true
		hs.AddCheck(func() error {
			if !healthy {
				return assert.AnError
			}
			return nil
		})

		err := hs.Start()
		require.NoError(t, err)
		defer func() { _ = hs.Stop(context.Background()) }()

		addr := hs.Addr()

		// Should be healthy
		resp, err := http.Get("http://" + addr + "/health")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()

		// Make unhealthy
		healthy = false

		resp, err = http.Get("http://" + addr + "/health")
		require.NoError(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		_ = resp.Body.Close()
	})
}

func TestDaemon(t *testing.T) {
	t.Run("start and stop", func(t *testing.T) {
		m := io.NewMockMedium()
		pidPath := "/tmp/test.pid"

		d := NewDaemon(DaemonOptions{
			Medium:          m,
			PIDFile:         pidPath,
			HealthAddr:      "127.0.0.1:0",
			ShutdownTimeout: 5 * time.Second,
		})

		err := d.Start()
		require.NoError(t, err)

		// Health server should be running
		addr := d.HealthAddr()
		require.NotEmpty(t, addr)

		resp, err := http.Get("http://" + addr + "/health")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()

		// Stop should succeed
		err = d.Stop()
		require.NoError(t, err)

		// PID file should be removed
		assert.False(t, m.Exists(pidPath))
	})

	t.Run("double start fails", func(t *testing.T) {
		d := NewDaemon(DaemonOptions{
			HealthAddr: "127.0.0.1:0",
		})

		err := d.Start()
		require.NoError(t, err)
		defer func() { _ = d.Stop() }()

		err = d.Start()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already running")
	})

	t.Run("run without start fails", func(t *testing.T) {
		d := NewDaemon(DaemonOptions{})

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := d.Run(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not started")
	})

	t.Run("set ready", func(t *testing.T) {
		d := NewDaemon(DaemonOptions{
			HealthAddr: "127.0.0.1:0",
		})

		err := d.Start()
		require.NoError(t, err)
		defer func() { _ = d.Stop() }()

		addr := d.HealthAddr()

		// Initially ready
		resp, _ := http.Get("http://" + addr + "/ready")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()

		// Set not ready
		d.SetReady(false)

		resp, _ = http.Get("http://" + addr + "/ready")
		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		_ = resp.Body.Close()
	})

	t.Run("no health addr returns empty", func(t *testing.T) {
		d := NewDaemon(DaemonOptions{})
		assert.Empty(t, d.HealthAddr())
	})

	t.Run("default shutdown timeout", func(t *testing.T) {
		d := NewDaemon(DaemonOptions{})
		assert.Equal(t, 30*time.Second, d.opts.ShutdownTimeout)
	})
}

func TestRunWithTimeout(t *testing.T) {
	t.Run("creates shutdown function", func(t *testing.T) {
		// Just test that it returns a function
		shutdown := RunWithTimeout(100 * time.Millisecond)
		assert.NotNil(t, shutdown)
	})
}
