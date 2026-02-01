package cli

import "testing"

func TestCheckBuilder(t *testing.T) {
	UseASCII() // Deterministic output

	// Pass
	c := Check("foo").Pass()
	got := c.String()
	if got == "" {
		t.Error("Empty output for Pass")
	}

	// Fail
	c = Check("foo").Fail()
	got = c.String()
	if got == "" {
		t.Error("Empty output for Fail")
	}

	// Skip
	c = Check("foo").Skip()
	got = c.String()
	if got == "" {
		t.Error("Empty output for Skip")
	}

	// Warn
	c = Check("foo").Warn()
	got = c.String()
	if got == "" {
		t.Error("Empty output for Warn")
	}

	// Duration
	c = Check("foo").Pass().Duration("1s")
	got = c.String()
	if got == "" {
		t.Error("Empty output for Duration")
	}

	// Message
	c = Check("foo").Message("status")
	got = c.String()
	if got == "" {
		t.Error("Empty output for Message")
	}
}
