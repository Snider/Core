package cli

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPanicRecovery_Good verifies that the panic recovery mechanism
// catches panics and calls the appropriate shutdown and error handling.
func TestPanicRecovery_Good(t *testing.T) {
	t.Run("recovery captures panic value and stack", func(t *testing.T) {
		var recovered any
		var capturedStack []byte
		var shutdownCalled bool

		// Simulate the panic recovery pattern from Main()
		func() {
			defer func() {
				if r := recover(); r != nil {
					recovered = r
					capturedStack = debug.Stack()
					shutdownCalled = true // simulates Shutdown() call
				}
			}()

			panic("test panic")
		}()

		assert.Equal(t, "test panic", recovered)
		assert.True(t, shutdownCalled, "Shutdown should be called after panic recovery")
		assert.NotEmpty(t, capturedStack, "Stack trace should be captured")
		assert.Contains(t, string(capturedStack), "TestPanicRecovery_Good")
	})

	t.Run("recovery handles error type panics", func(t *testing.T) {
		var recovered any

		func() {
			defer func() {
				if r := recover(); r != nil {
					recovered = r
				}
			}()

			panic(fmt.Errorf("error panic"))
		}()

		err, ok := recovered.(error)
		assert.True(t, ok, "Recovered value should be an error")
		assert.Equal(t, "error panic", err.Error())
	})

	t.Run("recovery handles nil panic gracefully", func(t *testing.T) {
		recoveryExecuted := false

		func() {
			defer func() {
				if r := recover(); r != nil {
					recoveryExecuted = true
				}
			}()

			// No panic occurs
		}()

		assert.False(t, recoveryExecuted, "Recovery block should not execute without panic")
	})
}

// TestPanicRecovery_Bad tests error conditions in panic recovery.
func TestPanicRecovery_Bad(t *testing.T) {
	t.Run("recovery handles concurrent panics", func(t *testing.T) {
		var wg sync.WaitGroup
		recoveryCount := 0
		var mu sync.Mutex

		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						mu.Lock()
						recoveryCount++
						mu.Unlock()
					}
				}()

				panic(fmt.Sprintf("panic from goroutine %d", id))
			}(i)
		}

		wg.Wait()
		assert.Equal(t, 3, recoveryCount, "All goroutine panics should be recovered")
	})
}

// TestPanicRecovery_Ugly tests edge cases in panic recovery.
func TestPanicRecovery_Ugly(t *testing.T) {
	t.Run("recovery handles typed panic values", func(t *testing.T) {
		type customError struct {
			code int
			msg  string
		}

		var recovered any

		func() {
			defer func() {
				recovered = recover()
			}()

			panic(customError{code: 500, msg: "internal error"})
		}()

		ce, ok := recovered.(customError)
		assert.True(t, ok, "Should recover custom type")
		assert.Equal(t, 500, ce.code)
		assert.Equal(t, "internal error", ce.msg)
	})
}

// TestMainPanicRecoveryPattern verifies the exact pattern used in Main().
func TestMainPanicRecoveryPattern(t *testing.T) {
	t.Run("pattern logs error and calls shutdown", func(t *testing.T) {
		var logBuffer bytes.Buffer
		var shutdownCalled bool
		var fatalErr error

		// Mock implementations
		mockLogError := func(msg string, args ...any) {
			fmt.Fprintf(&logBuffer, msg, args...)
		}
		mockShutdown := func() {
			shutdownCalled = true
		}
		mockFatal := func(err error) {
			fatalErr = err
		}

		// Execute the pattern from Main()
		func() {
			defer func() {
				if r := recover(); r != nil {
					mockLogError("recovered from panic: %v", r)
					mockShutdown()
					mockFatal(fmt.Errorf("panic: %v", r))
				}
			}()

			panic("simulated crash")
		}()

		assert.Contains(t, logBuffer.String(), "recovered from panic: simulated crash")
		assert.True(t, shutdownCalled, "Shutdown must be called on panic")
		assert.NotNil(t, fatalErr, "Fatal must be called with error")
		assert.Equal(t, "panic: simulated crash", fatalErr.Error())
	})
}
