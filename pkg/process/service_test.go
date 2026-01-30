package process

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/host-uk/core/pkg/framework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(t *testing.T) (*Service, *framework.Core) {
	t.Helper()

	core, err := framework.New(
		framework.WithName("process", NewService(Options{BufferSize: 1024})),
	)
	require.NoError(t, err)

	svc, err := framework.ServiceFor[*Service](core, "process")
	require.NoError(t, err)

	return svc, core
}

func TestService_Start(t *testing.T) {
	t.Run("echo command", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "echo", "hello")
		require.NoError(t, err)
		require.NotNil(t, proc)

		assert.NotEmpty(t, proc.ID)
		assert.Equal(t, "echo", proc.Command)
		assert.Equal(t, []string{"hello"}, proc.Args)

		// Wait for completion
		<-proc.Done()

		assert.Equal(t, StatusExited, proc.Status)
		assert.Equal(t, 0, proc.ExitCode)
		assert.Contains(t, proc.Output(), "hello")
	})

	t.Run("failing command", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.Start(context.Background(), "sh", "-c", "exit 42")
		require.NoError(t, err)

		<-proc.Done()

		assert.Equal(t, StatusExited, proc.Status)
		assert.Equal(t, 42, proc.ExitCode)
	})

	t.Run("non-existent command", func(t *testing.T) {
		svc, _ := newTestService(t)

		_, err := svc.Start(context.Background(), "nonexistent_command_xyz")
		assert.Error(t, err)
	})

	t.Run("with working directory", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, err := svc.StartWithOptions(context.Background(), RunOptions{
			Command: "pwd",
			Dir:     "/tmp",
		})
		require.NoError(t, err)

		<-proc.Done()

		// On macOS /tmp is a symlink to /private/tmp
		output := strings.TrimSpace(proc.Output())
		assert.True(t, output == "/tmp" || output == "/private/tmp", "got: %s", output)
	})

	t.Run("context cancellation", func(t *testing.T) {
		svc, _ := newTestService(t)

		ctx, cancel := context.WithCancel(context.Background())
		proc, err := svc.Start(ctx, "sleep", "10")
		require.NoError(t, err)

		// Cancel immediately
		cancel()

		select {
		case <-proc.Done():
			// Good - process was killed
		case <-time.After(2 * time.Second):
			t.Fatal("process should have been killed")
		}
	})
}

func TestService_Run(t *testing.T) {
	t.Run("returns output", func(t *testing.T) {
		svc, _ := newTestService(t)

		output, err := svc.Run(context.Background(), "echo", "hello world")
		require.NoError(t, err)
		assert.Contains(t, output, "hello world")
	})

	t.Run("returns error on failure", func(t *testing.T) {
		svc, _ := newTestService(t)

		_, err := svc.Run(context.Background(), "sh", "-c", "exit 1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exited with code 1")
	})
}

func TestService_Actions(t *testing.T) {
	t.Run("broadcasts events", func(t *testing.T) {
		core, err := framework.New(
			framework.WithName("process", NewService(Options{})),
		)
		require.NoError(t, err)

		var started []ActionProcessStarted
		var outputs []ActionProcessOutput
		var exited []ActionProcessExited
		var mu sync.Mutex

		core.RegisterAction(func(c *framework.Core, msg framework.Message) error {
			mu.Lock()
			defer mu.Unlock()
			switch m := msg.(type) {
			case ActionProcessStarted:
				started = append(started, m)
			case ActionProcessOutput:
				outputs = append(outputs, m)
			case ActionProcessExited:
				exited = append(exited, m)
			}
			return nil
		})

		svc, _ := framework.ServiceFor[*Service](core, "process")
		proc, err := svc.Start(context.Background(), "echo", "test")
		require.NoError(t, err)

		<-proc.Done()

		// Give time for events to propagate
		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()

		assert.Len(t, started, 1)
		assert.Equal(t, "echo", started[0].Command)
		assert.Equal(t, []string{"test"}, started[0].Args)

		assert.NotEmpty(t, outputs)
		foundTest := false
		for _, o := range outputs {
			if strings.Contains(o.Line, "test") {
				foundTest = true
				break
			}
		}
		assert.True(t, foundTest, "should have output containing 'test'")

		assert.Len(t, exited, 1)
		assert.Equal(t, 0, exited[0].ExitCode)
	})
}

func TestService_List(t *testing.T) {
	t.Run("tracks processes", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc1, _ := svc.Start(context.Background(), "echo", "1")
		proc2, _ := svc.Start(context.Background(), "echo", "2")

		<-proc1.Done()
		<-proc2.Done()

		list := svc.List()
		assert.Len(t, list, 2)
	})

	t.Run("get by id", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, _ := svc.Start(context.Background(), "echo", "test")
		<-proc.Done()

		got, err := svc.Get(proc.ID)
		require.NoError(t, err)
		assert.Equal(t, proc.ID, got.ID)
	})

	t.Run("get not found", func(t *testing.T) {
		svc, _ := newTestService(t)

		_, err := svc.Get("nonexistent")
		assert.ErrorIs(t, err, ErrProcessNotFound)
	})
}

func TestService_Remove(t *testing.T) {
	t.Run("removes completed process", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc, _ := svc.Start(context.Background(), "echo", "test")
		<-proc.Done()

		err := svc.Remove(proc.ID)
		require.NoError(t, err)

		_, err = svc.Get(proc.ID)
		assert.ErrorIs(t, err, ErrProcessNotFound)
	})

	t.Run("cannot remove running process", func(t *testing.T) {
		svc, _ := newTestService(t)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		proc, _ := svc.Start(ctx, "sleep", "10")

		err := svc.Remove(proc.ID)
		assert.Error(t, err)

		cancel()
		<-proc.Done()
	})
}

func TestService_Clear(t *testing.T) {
	t.Run("clears completed processes", func(t *testing.T) {
		svc, _ := newTestService(t)

		proc1, _ := svc.Start(context.Background(), "echo", "1")
		proc2, _ := svc.Start(context.Background(), "echo", "2")

		<-proc1.Done()
		<-proc2.Done()

		assert.Len(t, svc.List(), 2)

		svc.Clear()

		assert.Len(t, svc.List(), 0)
	})
}

func TestProcess_Info(t *testing.T) {
	svc, _ := newTestService(t)

	proc, _ := svc.Start(context.Background(), "echo", "hello")
	<-proc.Done()

	info := proc.Info()
	assert.Equal(t, proc.ID, info.ID)
	assert.Equal(t, "echo", info.Command)
	assert.Equal(t, []string{"hello"}, info.Args)
	assert.Equal(t, StatusExited, info.Status)
	assert.Equal(t, 0, info.ExitCode)
}
