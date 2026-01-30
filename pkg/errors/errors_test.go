package errors

import (
	"io"
	"testing"
)

func TestE(t *testing.T) {
	err := E("user.Create", "validation failed", nil)

	if err.Error() != "user.Create: validation failed" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestE_WithUnderlying(t *testing.T) {
	underlying := New("database connection failed")
	err := E("user.Create", "failed to save", underlying)

	if err.Error() != "user.Create: failed to save: database connection failed" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestWrap(t *testing.T) {
	// Wrap nil returns nil
	if Wrap(nil, "op", "msg") != nil {
		t.Error("expected Wrap(nil) to return nil")
	}

	// Wrap error
	underlying := New("original")
	err := Wrap(underlying, "user.Get", "failed")

	if !Is(err, underlying) {
		t.Error("expected wrapped error to match underlying")
	}
}

func TestWrapCode(t *testing.T) {
	underlying := New("not found")
	err := WrapCode(underlying, "ERR_NOT_FOUND", "user.Get", "user not found")

	var e *Error
	if !As(err, &e) {
		t.Fatal("expected error to be *Error")
	}

	if e.Code != "ERR_NOT_FOUND" {
		t.Errorf("expected code ERR_NOT_FOUND, got %s", e.Code)
	}
}

func TestCode(t *testing.T) {
	err := Code("ERR_VALIDATION", "invalid email")

	var e *Error
	if !As(err, &e) {
		t.Fatal("expected error to be *Error")
	}

	if e.Code != "ERR_VALIDATION" {
		t.Errorf("expected code ERR_VALIDATION, got %s", e.Code)
	}
	if e.Msg != "invalid email" {
		t.Errorf("expected msg 'invalid email', got %s", e.Msg)
	}
}

func TestIs(t *testing.T) {
	err := Wrap(io.EOF, "read", "failed")

	if !Is(err, io.EOF) {
		t.Error("expected Is to find io.EOF in chain")
	}

	if Is(err, io.ErrClosedPipe) {
		t.Error("expected Is to not find io.ErrClosedPipe")
	}
}

func TestAs(t *testing.T) {
	err := E("test.Op", "test message", nil)

	var e *Error
	if !As(err, &e) {
		t.Fatal("expected As to find *Error")
	}

	if e.Op != "test.Op" {
		t.Errorf("expected Op 'test.Op', got %s", e.Op)
	}
}

func TestOp(t *testing.T) {
	err := E("user.Create", "failed", nil)

	if Op(err) != "user.Create" {
		t.Errorf("expected Op 'user.Create', got %s", Op(err))
	}

	// Non-Error returns empty string
	if Op(New("plain error")) != "" {
		t.Error("expected empty Op for non-Error")
	}
}

func TestErrCode(t *testing.T) {
	err := Code("ERR_TEST", "test")

	if ErrCode(err) != "ERR_TEST" {
		t.Errorf("expected code ERR_TEST, got %s", ErrCode(err))
	}

	// Non-Error returns empty string
	if ErrCode(New("plain error")) != "" {
		t.Error("expected empty code for non-Error")
	}
}

func TestMessage(t *testing.T) {
	err := E("op", "the message", nil)

	if Message(err) != "the message" {
		t.Errorf("expected 'the message', got %s", Message(err))
	}

	// Plain error returns full error string
	plain := New("plain error")
	if Message(plain) != "plain error" {
		t.Errorf("expected 'plain error', got %s", Message(plain))
	}

	// Nil returns empty string
	if Message(nil) != "" {
		t.Error("expected empty string for nil")
	}
}

func TestRoot(t *testing.T) {
	root := New("root cause")
	mid := Wrap(root, "mid", "middle")
	top := Wrap(mid, "top", "top level")

	if Root(top) != root {
		t.Error("expected Root to return deepest error")
	}

	// Single error returns itself
	single := New("single")
	if Root(single) != single {
		t.Error("expected Root of single error to return itself")
	}
}

func TestError_Unwrap(t *testing.T) {
	underlying := New("underlying")
	err := E("op", "msg", underlying)

	var e *Error
	if !As(err, &e) {
		t.Fatal("expected *Error")
	}

	if e.Unwrap() != underlying {
		t.Error("expected Unwrap to return underlying error")
	}
}

func TestJoin(t *testing.T) {
	err1 := New("error 1")
	err2 := New("error 2")

	joined := Join(err1, err2)

	if !Is(joined, err1) {
		t.Error("expected joined error to contain err1")
	}
	if !Is(joined, err2) {
		t.Error("expected joined error to contain err2")
	}
}
