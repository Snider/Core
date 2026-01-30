package process

import (
	"context"
	"sync"
	"testing"

	"github.com/host-uk/core/pkg/framework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobal_DefaultNotInitialized(t *testing.T) {
	// Reset global state for this test
	old := defaultService.Swap(nil)
	defer func() {
		if old != nil {
			defaultService.Store(old)
		}
	}()

	assert.Nil(t, Default())

	_, err := Start(context.Background(), "echo", "test")
	assert.ErrorIs(t, err, ErrServiceNotInitialized)

	_, err = Run(context.Background(), "echo", "test")
	assert.ErrorIs(t, err, ErrServiceNotInitialized)

	_, err = Get("proc-1")
	assert.ErrorIs(t, err, ErrServiceNotInitialized)

	assert.Nil(t, List())
	assert.Nil(t, Running())

	err = Kill("proc-1")
	assert.ErrorIs(t, err, ErrServiceNotInitialized)

	_, err = StartWithOptions(context.Background(), RunOptions{Command: "echo"})
	assert.ErrorIs(t, err, ErrServiceNotInitialized)

	_, err = RunWithOptions(context.Background(), RunOptions{Command: "echo"})
	assert.ErrorIs(t, err, ErrServiceNotInitialized)
}

func TestGlobal_SetDefault(t *testing.T) {
	t.Run("sets and retrieves service", func(t *testing.T) {
		// Reset global state
		old := defaultService.Swap(nil)
		defer func() {
			if old != nil {
				defaultService.Store(old)
			}
		}()

		core, err := framework.New(
			framework.WithName("process", NewService(Options{})),
		)
		require.NoError(t, err)

		svc, err := framework.ServiceFor[*Service](core, "process")
		require.NoError(t, err)

		SetDefault(svc)
		assert.Equal(t, svc, Default())
	})

	t.Run("panics on nil", func(t *testing.T) {
		assert.Panics(t, func() {
			SetDefault(nil)
		})
	})
}

func TestGlobal_ConcurrentDefault(t *testing.T) {
	// Reset global state
	old := defaultService.Swap(nil)
	defer func() {
		if old != nil {
			defaultService.Store(old)
		}
	}()

	core, err := framework.New(
		framework.WithName("process", NewService(Options{})),
	)
	require.NoError(t, err)

	svc, err := framework.ServiceFor[*Service](core, "process")
	require.NoError(t, err)

	SetDefault(svc)

	// Concurrent reads of Default()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := Default()
			assert.NotNil(t, s)
			assert.Equal(t, svc, s)
		}()
	}
	wg.Wait()
}

func TestGlobal_ConcurrentSetDefault(t *testing.T) {
	// Reset global state
	old := defaultService.Swap(nil)
	defer func() {
		if old != nil {
			defaultService.Store(old)
		}
	}()

	// Create multiple services
	var services []*Service
	for i := 0; i < 10; i++ {
		core, err := framework.New(
			framework.WithName("process", NewService(Options{})),
		)
		require.NoError(t, err)

		svc, err := framework.ServiceFor[*Service](core, "process")
		require.NoError(t, err)
		services = append(services, svc)
	}

	// Concurrent SetDefault calls - should not panic or race
	var wg sync.WaitGroup
	for _, svc := range services {
		wg.Add(1)
		go func(s *Service) {
			defer wg.Done()
			SetDefault(s)
		}(svc)
	}
	wg.Wait()

	// Final state should be one of the services
	final := Default()
	assert.NotNil(t, final)

	found := false
	for _, svc := range services {
		if svc == final {
			found = true
			break
		}
	}
	assert.True(t, found, "Default should be one of the set services")
}

func TestGlobal_ConcurrentOperations(t *testing.T) {
	// Reset global state
	old := defaultService.Swap(nil)
	defer func() {
		if old != nil {
			defaultService.Store(old)
		}
	}()

	core, err := framework.New(
		framework.WithName("process", NewService(Options{})),
	)
	require.NoError(t, err)

	svc, err := framework.ServiceFor[*Service](core, "process")
	require.NoError(t, err)

	SetDefault(svc)

	// Concurrent Start, List, Get operations
	var wg sync.WaitGroup
	var processes []*Process
	var procMu sync.Mutex

	// Start 20 processes concurrently
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			proc, err := Start(context.Background(), "echo", "concurrent")
			if err == nil {
				procMu.Lock()
				processes = append(processes, proc)
				procMu.Unlock()
			}
		}()
	}

	// Concurrent List calls while starting
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = List()
			_ = Running()
		}()
	}

	wg.Wait()

	// Wait for all processes to complete
	procMu.Lock()
	for _, p := range processes {
		<-p.Done()
	}
	procMu.Unlock()

	// All should have succeeded
	assert.Len(t, processes, 20)

	// Concurrent Get calls
	var wg2 sync.WaitGroup
	for _, p := range processes {
		wg2.Add(1)
		go func(id string) {
			defer wg2.Done()
			got, err := Get(id)
			assert.NoError(t, err)
			assert.NotNil(t, got)
		}(p.ID)
	}
	wg2.Wait()
}

func TestGlobal_StartWithOptions(t *testing.T) {
	svc, _ := newTestService(t)

	// Set as default
	old := defaultService.Swap(svc)
	defer func() {
		if old != nil {
			defaultService.Store(old)
		}
	}()

	proc, err := StartWithOptions(context.Background(), RunOptions{
		Command: "echo",
		Args:    []string{"with", "options"},
	})
	require.NoError(t, err)

	<-proc.Done()

	assert.Equal(t, 0, proc.ExitCode)
	assert.Contains(t, proc.Output(), "with options")
}

func TestGlobal_RunWithOptions(t *testing.T) {
	svc, _ := newTestService(t)

	// Set as default
	old := defaultService.Swap(svc)
	defer func() {
		if old != nil {
			defaultService.Store(old)
		}
	}()

	output, err := RunWithOptions(context.Background(), RunOptions{
		Command: "echo",
		Args:    []string{"run", "options"},
	})
	require.NoError(t, err)
	assert.Contains(t, output, "run options")
}

func TestGlobal_Running(t *testing.T) {
	svc, _ := newTestService(t)

	// Set as default
	old := defaultService.Swap(svc)
	defer func() {
		if old != nil {
			defaultService.Store(old)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start a long-running process
	proc, err := Start(ctx, "sleep", "60")
	require.NoError(t, err)

	running := Running()
	assert.Len(t, running, 1)
	assert.Equal(t, proc.ID, running[0].ID)

	cancel()
	<-proc.Done()

	running = Running()
	assert.Len(t, running, 0)
}
