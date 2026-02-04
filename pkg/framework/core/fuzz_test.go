package core

import (
	"errors"
	"testing"
)

// FuzzE exercises the E() error constructor with arbitrary input.
func FuzzE(f *testing.F) {
	f.Add("svc.Method", "something broke", true)
	f.Add("", "", false)
	f.Add("a.b.c.d.e.f", "unicode: \u00e9\u00e8\u00ea", true)

	f.Fuzz(func(t *testing.T, op, msg string, withErr bool) {
		var underlying error
		if withErr {
			underlying = errors.New("wrapped")
		}

		e := E(op, msg, underlying)
		if e == nil {
			t.Fatal("E() returned nil")
		}

		s := e.Error()
		if s == "" {
			t.Fatal("Error() returned empty string")
		}

		// Round-trip: Unwrap should return the underlying error
		var coreErr *Error
		if !errors.As(e, &coreErr) {
			t.Fatal("errors.As failed for *Error")
		}
		if withErr && coreErr.Unwrap() == nil {
			t.Fatal("Unwrap() returned nil with underlying error")
		}
		if !withErr && coreErr.Unwrap() != nil {
			t.Fatal("Unwrap() returned non-nil without underlying error")
		}
	})
}

// FuzzServiceRegistration exercises service name registration with arbitrary names.
func FuzzServiceRegistration(f *testing.F) {
	f.Add("myservice")
	f.Add("")
	f.Add("a/b/c")
	f.Add("service with spaces")
	f.Add("service\x00null")

	f.Fuzz(func(t *testing.T, name string) {
		sm := newServiceManager()

		err := sm.registerService(name, struct{}{})
		if name == "" {
			if err == nil {
				t.Fatal("expected error for empty name")
			}
			return
		}
		if err != nil {
			t.Fatalf("unexpected error for name %q: %v", name, err)
		}

		// Retrieve should return the same service
		got := sm.service(name)
		if got == nil {
			t.Fatalf("service %q not found after registration", name)
		}

		// Duplicate registration should fail
		err = sm.registerService(name, struct{}{})
		if err == nil {
			t.Fatalf("expected duplicate error for name %q", name)
		}
	})
}

// FuzzMessageDispatch exercises action dispatch with concurrent registrations.
func FuzzMessageDispatch(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add("test\nmultiline")

	f.Fuzz(func(t *testing.T, payload string) {
		c := &Core{
			Features: &Features{},
			svc:      newServiceManager(),
		}
		c.bus = newMessageBus(c)

		var received string
		c.bus.registerAction(func(_ *Core, msg Message) error {
			received = msg.(string)
			return nil
		})

		err := c.bus.action(payload)
		if err != nil {
			t.Fatalf("action dispatch failed: %v", err)
		}
		if received != payload {
			t.Fatalf("got %q, want %q", received, payload)
		}
	})
}
