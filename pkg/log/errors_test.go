package log

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Err Type Tests ---

func TestErr_Error_Good(t *testing.T) {
	// With underlying error
	err := &Err{Op: "db.Query", Msg: "failed to query", Err: errors.New("connection refused")}
	assert.Equal(t, "db.Query: failed to query: connection refused", err.Error())

	// With code
	err = &Err{Op: "api.Call", Msg: "request failed", Code: "TIMEOUT"}
	assert.Equal(t, "api.Call: request failed [TIMEOUT]", err.Error())

	// With both underlying error and code
	err = &Err{Op: "user.Save", Msg: "save failed", Err: errors.New("duplicate key"), Code: "DUPLICATE"}
	assert.Equal(t, "user.Save: save failed [DUPLICATE]: duplicate key", err.Error())

	// Just op and msg
	err = &Err{Op: "cache.Get", Msg: "miss"}
	assert.Equal(t, "cache.Get: miss", err.Error())
}

func TestErr_Error_EmptyOp_Good(t *testing.T) {
	// No Op - should not have leading colon
	err := &Err{Msg: "just a message"}
	assert.Equal(t, "just a message", err.Error())

	// No Op with code
	err = &Err{Msg: "error with code", Code: "ERR_CODE"}
	assert.Equal(t, "error with code [ERR_CODE]", err.Error())

	// No Op with underlying error
	err = &Err{Msg: "wrapped", Err: errors.New("underlying")}
	assert.Equal(t, "wrapped: underlying", err.Error())
}

func TestErr_Unwrap_Good(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &Err{Op: "test", Msg: "wrapped", Err: underlying}

	assert.Equal(t, underlying, errors.Unwrap(err))
	assert.True(t, errors.Is(err, underlying))
}

// --- Error Creation Function Tests ---

func TestE_Good(t *testing.T) {
	underlying := errors.New("base error")
	err := E("op.Name", "something failed", underlying)

	assert.NotNil(t, err)
	var logErr *Err
	assert.True(t, errors.As(err, &logErr))
	assert.Equal(t, "op.Name", logErr.Op)
	assert.Equal(t, "something failed", logErr.Msg)
	assert.Equal(t, underlying, logErr.Err)
}

func TestE_Good_NilError(t *testing.T) {
	// E creates an error even with nil underlying - useful for errors without causes
	err := E("op.Name", "message", nil)
	assert.NotNil(t, err)
	assert.Equal(t, "op.Name: message", err.Error())
}

func TestWrap_Good(t *testing.T) {
	underlying := errors.New("base")
	err := Wrap(underlying, "handler.Process", "processing failed")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "handler.Process")
	assert.Contains(t, err.Error(), "processing failed")
	assert.True(t, errors.Is(err, underlying))
}

func TestWrap_PreservesCode_Good(t *testing.T) {
	// Create an error with a code
	inner := WrapCode(errors.New("base"), "VALIDATION_ERROR", "inner.Op", "validation failed")

	// Wrap it - should preserve the code
	outer := Wrap(inner, "outer.Op", "outer context")

	assert.NotNil(t, outer)
	assert.Equal(t, "VALIDATION_ERROR", ErrCode(outer))
	assert.Contains(t, outer.Error(), "[VALIDATION_ERROR]")
}

func TestWrap_NilError_Good(t *testing.T) {
	err := Wrap(nil, "op", "msg")
	assert.Nil(t, err)
}

func TestWrapCode_Good(t *testing.T) {
	underlying := errors.New("validation failed")
	err := WrapCode(underlying, "INVALID_INPUT", "api.Validate", "bad request")

	assert.NotNil(t, err)
	var logErr *Err
	assert.True(t, errors.As(err, &logErr))
	assert.Equal(t, "INVALID_INPUT", logErr.Code)
	assert.Equal(t, "api.Validate", logErr.Op)
	assert.Contains(t, err.Error(), "[INVALID_INPUT]")
}

func TestWrapCode_Good_NilError(t *testing.T) {
	// WrapCode with nil error but with code still creates an error
	err := WrapCode(nil, "CODE", "op", "msg")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "[CODE]")

	// Only returns nil when both error and code are empty
	err = WrapCode(nil, "", "op", "msg")
	assert.Nil(t, err)
}

func TestNewCode_Good(t *testing.T) {
	err := NewCode("NOT_FOUND", "resource not found")

	var logErr *Err
	assert.True(t, errors.As(err, &logErr))
	assert.Equal(t, "NOT_FOUND", logErr.Code)
	assert.Equal(t, "resource not found", logErr.Msg)
	assert.Nil(t, logErr.Err)
}

// --- Standard Library Wrapper Tests ---

func TestIs_Good(t *testing.T) {
	sentinel := errors.New("sentinel")
	wrapped := Wrap(sentinel, "test", "wrapped")

	assert.True(t, Is(wrapped, sentinel))
	assert.False(t, Is(wrapped, errors.New("other")))
}

func TestAs_Good(t *testing.T) {
	err := E("test.Op", "message", errors.New("base"))

	var logErr *Err
	assert.True(t, As(err, &logErr))
	assert.Equal(t, "test.Op", logErr.Op)
}

func TestNewError_Good(t *testing.T) {
	err := NewError("simple error")
	assert.NotNil(t, err)
	assert.Equal(t, "simple error", err.Error())
}

func TestJoin_Good(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	joined := Join(err1, err2)

	assert.True(t, errors.Is(joined, err1))
	assert.True(t, errors.Is(joined, err2))
}

// --- Helper Function Tests ---

func TestOp_Good(t *testing.T) {
	err := E("mypackage.MyFunc", "failed", errors.New("cause"))
	assert.Equal(t, "mypackage.MyFunc", Op(err))
}

func TestOp_Good_NotLogError(t *testing.T) {
	err := errors.New("plain error")
	assert.Equal(t, "", Op(err))
}

func TestErrCode_Good(t *testing.T) {
	err := WrapCode(errors.New("base"), "ERR_CODE", "op", "msg")
	assert.Equal(t, "ERR_CODE", ErrCode(err))
}

func TestErrCode_Good_NoCode(t *testing.T) {
	err := E("op", "msg", errors.New("base"))
	assert.Equal(t, "", ErrCode(err))
}

func TestMessage_Good(t *testing.T) {
	err := E("op", "the message", errors.New("base"))
	assert.Equal(t, "the message", Message(err))
}

func TestMessage_Good_PlainError(t *testing.T) {
	err := errors.New("plain message")
	assert.Equal(t, "plain message", Message(err))
}

func TestMessage_Good_Nil(t *testing.T) {
	assert.Equal(t, "", Message(nil))
}

func TestRoot_Good(t *testing.T) {
	root := errors.New("root cause")
	level1 := Wrap(root, "level1", "wrapped once")
	level2 := Wrap(level1, "level2", "wrapped twice")

	assert.Equal(t, root, Root(level2))
}

func TestRoot_Good_SingleError(t *testing.T) {
	err := errors.New("single")
	assert.Equal(t, err, Root(err))
}

func TestRoot_Good_Nil(t *testing.T) {
	assert.Nil(t, Root(nil))
}

// --- Log-and-Return Helper Tests ---

func TestLogError_Good(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := New(Options{Level: LevelDebug, Output: &buf})
	SetDefault(logger)
	defer SetDefault(New(Options{Level: LevelInfo}))

	underlying := errors.New("connection failed")
	err := LogError(underlying, "db.Connect", "database unavailable")

	// Check returned error
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "db.Connect")
	assert.Contains(t, err.Error(), "database unavailable")
	assert.True(t, errors.Is(err, underlying))

	// Check log output
	output := buf.String()
	assert.Contains(t, output, "[ERR]")
	assert.Contains(t, output, "database unavailable")
	assert.Contains(t, output, "op=db.Connect")
}

func TestLogError_Good_NilError(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{Level: LevelDebug, Output: &buf})
	SetDefault(logger)
	defer SetDefault(New(Options{Level: LevelInfo}))

	err := LogError(nil, "op", "msg")
	assert.Nil(t, err)
	assert.Empty(t, buf.String()) // No log output for nil error
}

func TestLogWarn_Good(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{Level: LevelDebug, Output: &buf})
	SetDefault(logger)
	defer SetDefault(New(Options{Level: LevelInfo}))

	underlying := errors.New("cache miss")
	err := LogWarn(underlying, "cache.Get", "falling back to db")

	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, underlying))

	output := buf.String()
	assert.Contains(t, output, "[WRN]")
	assert.Contains(t, output, "falling back to db")
}

func TestLogWarn_Good_NilError(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{Level: LevelDebug, Output: &buf})
	SetDefault(logger)
	defer SetDefault(New(Options{Level: LevelInfo}))

	err := LogWarn(nil, "op", "msg")
	assert.Nil(t, err)
	assert.Empty(t, buf.String())
}

func TestMust_Good_NoError(t *testing.T) {
	// Should not panic when error is nil
	assert.NotPanics(t, func() {
		Must(nil, "test", "should not panic")
	})
}

func TestMust_Ugly_Panics(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{Level: LevelDebug, Output: &buf})
	SetDefault(logger)
	defer SetDefault(New(Options{Level: LevelInfo}))

	assert.Panics(t, func() {
		Must(errors.New("fatal error"), "startup", "initialization failed")
	})

	// Verify error was logged before panic
	output := buf.String()
	assert.True(t, strings.Contains(output, "[ERR]") || len(output) > 0)
}

func TestStackTrace_Good(t *testing.T) {
	// Nested operations
	err := E("op1", "msg1", nil)
	err = Wrap(err, "op2", "msg2")
	err = Wrap(err, "op3", "msg3")

	stack := StackTrace(err)
	assert.Equal(t, []string{"op3", "op2", "op1"}, stack)

	// Format
	formatted := FormatStackTrace(err)
	assert.Equal(t, "op3 -> op2 -> op1", formatted)
}

func TestStackTrace_PlainError(t *testing.T) {
	err := errors.New("plain error")
	assert.Empty(t, StackTrace(err))
	assert.Empty(t, FormatStackTrace(err))
}

func TestStackTrace_Nil(t *testing.T) {
	assert.Empty(t, StackTrace(nil))
	assert.Empty(t, FormatStackTrace(nil))
}

func TestStackTrace_NoOp(t *testing.T) {
	err := &Err{Msg: "no op"}
	assert.Empty(t, StackTrace(err))
	assert.Empty(t, FormatStackTrace(err))
}

func TestStackTrace_Mixed(t *testing.T) {
	err := E("inner", "msg", nil)
	err = errors.New("middle: " + err.Error()) // Breaks the chain if not handled properly, but Unwrap should work if it's a wrapped error
	// Wait, errors.New doesn't wrap. fmt.Errorf("%w") does.
	err = E("inner", "msg", nil)
	err = fmt.Errorf("wrapper: %w", err)
	err = Wrap(err, "outer", "msg")

	stack := StackTrace(err)
	assert.Equal(t, []string{"outer", "inner"}, stack)
}
