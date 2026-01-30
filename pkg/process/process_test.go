package process

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcess_Info(t *testing.T) {
	svc, _ := newTestService(t)

	proc, err := svc.Start(context.Background(), "echo", "hello")
	require.NoError(t, err)

	<-proc.Done()

	info := proc.Info()
	assert.Equal(t, proc.ID, info.ID)
	assert.Equal(t, "echo", info.Command)
	assert.Equal(t, []string{"hello"}, info.Args)
	assert.Equal(t, StatusExited, info.Status)
	assert.Equal(t, 0, info.ExitCode)
	assert.Greater(t, info.Duration, time.Duration(0))
}

func TestProcess_Output(t *testing.T) {
	t.Run("captures stdout", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "echo", "hello world")
		require.NoError(t, err)

		<-proc.Done()

		output := proc.Output()
		assert.Contains(t, output, "hello world")
	})

	t.Run("OutputBytes returns copy", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "echo", "test")
		require.NoError(t, err)

		<-proc.Done()

		bytes := proc.OutputBytes()
		assert.NotNil(t, bytes)
		assert.Contains(t, string(bytes), "test")
	})
}

func TestProcess_IsRunning(t *testing.T) {
	t.Run("true while running", func(t *testing.T) {
		svc, _ := newTestService(t)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		proc, err := svc.Start(ctx, "sleep", "10")
		require.NoError(t, err)

		assert.True(t, proc.IsRunning())

		cancel()
		<-proc.Done()

		assert.False(t, proc.IsRunning())
	})

	t.Run("false after completion", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "echo", "done")
		require.NoError(t, err)

		<-proc.Done()

		assert.False(t, proc.IsRunning())
	})
}

func TestProcess_Wait(t *testing.T) {
	t.Run("returns nil on success", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "echo", "ok")
		require.NoError(t, err)

		err = proc.Wait()
		assert.NoError(t, err)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "sh", "-c", "exit 1")
		require.NoError(t, err)

		err = proc.Wait()
		assert.Error(t, err)
	})
}

func TestProcess_Done(t *testing.T) {
	t.Run("channel closes on completion", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "echo", "test")
		require.NoError(t, err)

		select {
		case <-proc.Done():
			// Success - channel closed
		case <-time.After(5 * time.Second):
			t.Fatal("Done channel should have closed")
		}
	})
}

func TestProcess_Kill(t *testing.T) {
	t.Run("terminates running process", func(t *testing.T) {
		svc, _ := newTestService(t)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		proc, err := svc.Start(ctx, "sleep", "60")
		require.NoError(t, err)

		assert.True(t, proc.IsRunning())

		err = proc.Kill()
		assert.NoError(t, err)

		select {
		case <-proc.Done():
			// Good - process terminated
		case <-time.After(2 * time.Second):
			t.Fatal("process should have been killed")
		}
	})

	t.Run("noop on completed process", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "echo", "done")
		require.NoError(t, err)

		<-proc.Done()

		err = proc.Kill()
		assert.NoError(t, err)
	})
}

func TestProcess_SendInput(t *testing.T) {
	t.Run("writes to stdin", func(t *testing.T) {
		svc, _ := newTestService(t)

		// Use cat to echo back stdin
		proc, err := svc.Start(context.Background(), "cat")
		require.NoError(t, err)

		err = proc.SendInput("hello\n")
		assert.NoError(t, err)

		err = proc.CloseStdin()
		assert.NoError(t, err)

		<-proc.Done()

		assert.Contains(t, proc.Output(), "hello")
	})

	t.Run("error on completed process", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "echo", "done")
		require.NoError(t, err)

		<-proc.Done()

		err = proc.SendInput("test")
		assert.ErrorIs(t, err, ErrProcessNotRunning)
	})
}

func TestProcess_CloseStdin(t *testing.T) {
	t.Run("closes stdin pipe", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "cat")
		require.NoError(t, err)

		err = proc.CloseStdin()
		assert.NoError(t, err)

		// Process should exit now that stdin is closed
		select {
		case <-proc.Done():
			// Good
		case <-time.After(2 * time.Second):
			t.Fatal("cat should exit when stdin is closed")
		}
	})

	t.Run("double close is safe", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "cat")
		require.NoError(t, err)

		// First close
		err = proc.CloseStdin()
		assert.NoError(t, err)

		<-proc.Done()

		// Second close should be safe (stdin already nil)
		err = proc.CloseStdin()
		assert.NoError(t, err)
	})
}
