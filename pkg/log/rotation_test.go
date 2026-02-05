package log

import (
	"strings"
	"testing"
	"time"

	"github.com/host-uk/core/pkg/io"
)

func TestRotatingWriter_Basic(t *testing.T) {
	m := io.NewMockMedium()
	opts := RotationOptions{
		Filename:   "test.log",
		MaxSize:    1, // 1 MB
		MaxBackups: 3,
	}

	w := NewRotatingWriter(opts, m)
	defer w.Close()

	msg := "test message\n"
	_, err := w.Write([]byte(msg))
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	w.Close()

	content, err := m.Read("test.log")
	if err != nil {
		t.Fatalf("failed to read from medium: %v", err)
	}
	if content != msg {
		t.Errorf("expected %q, got %q", msg, content)
	}
}

func TestRotatingWriter_Rotation(t *testing.T) {
	m := io.NewMockMedium()
	opts := RotationOptions{
		Filename:   "test.log",
		MaxSize:    1, // 1 MB
		MaxBackups: 2,
	}

	w := NewRotatingWriter(opts, m)
	defer w.Close()

	// 1. Write almost 1MB
	largeMsg := strings.Repeat("a", 1024*1024-10)
	_, _ = w.Write([]byte(largeMsg))

	// 2. Write more to trigger rotation
	_, _ = w.Write([]byte("trigger rotation\n"))
	w.Close()

	// Check if test.log.1 exists and contains the large message
	if !m.Exists("test.log.1") {
		t.Error("expected test.log.1 to exist")
	}

	// Check if test.log exists and contains the new message
	content, _ := m.Read("test.log")
	if !strings.Contains(content, "trigger rotation") {
		t.Errorf("expected test.log to contain new message, got %q", content)
	}
}

func TestRotatingWriter_Retention(t *testing.T) {
	m := io.NewMockMedium()
	opts := RotationOptions{
		Filename:   "test.log",
		MaxSize:    1,
		MaxBackups: 2,
	}

	w := NewRotatingWriter(opts, m)
	defer w.Close()

	// Trigger rotation 4 times to test retention of only the latest backups
	for i := 1; i <= 4; i++ {
		_, _ = w.Write([]byte(strings.Repeat("a", 1024*1024+1)))
	}
	w.Close()

	// Should have test.log, test.log.1, test.log.2
	// test.log.3 should have been deleted because MaxBackups is 2
	if !m.Exists("test.log") {
		t.Error("expected test.log to exist")
	}
	if !m.Exists("test.log.1") {
		t.Error("expected test.log.1 to exist")
	}
	if !m.Exists("test.log.2") {
		t.Error("expected test.log.2 to exist")
	}
	if m.Exists("test.log.3") {
		t.Error("expected test.log.3 NOT to exist")
	}
}

func TestRotatingWriter_Append(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.Write("test.log", "existing content\n")

	opts := RotationOptions{
		Filename: "test.log",
	}

	w := NewRotatingWriter(opts, m)
	_, _ = w.Write([]byte("new content\n"))
	_ = w.Close()

	content, _ := m.Read("test.log")
	expected := "existing content\nnew content\n"
	if content != expected {
		t.Errorf("expected %q, got %q", expected, content)
	}
}

func TestRotatingWriter_AgeRetention(t *testing.T) {
	m := io.NewMockMedium()
	opts := RotationOptions{
		Filename:   "test.log",
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     7, // 7 days
	}

	w := NewRotatingWriter(opts, m)

	// Create some backup files
	m.Write("test.log.1", "recent")
	m.ModTimes["test.log.1"] = time.Now()

	m.Write("test.log.2", "old")
	m.ModTimes["test.log.2"] = time.Now().AddDate(0, 0, -10) // 10 days old

	// Trigger rotation to run cleanup
	_, _ = w.Write([]byte(strings.Repeat("a", 1024*1024+1)))
	w.Close()

	if !m.Exists("test.log.1") {
		t.Error("expected test.log.1 (now test.log.2) to exist as it's recent")
	}
	// Note: test.log.1 becomes test.log.2 after rotation, etc.
	// But wait, my cleanup runs AFTER rotation.
	// Initial state:
	// test.log.1 (now)
	// test.log.2 (-10d)
	// Write triggers rotation:
	// test.log -> test.log.1
	// test.log.1 -> test.log.2
	// test.log.2 -> test.log.3
	// Then cleanup runs:
	// test.log.1 (now) - keep
	// test.log.2 (now) - keep
	// test.log.3 (-10d) - delete (since MaxAge is 7)

	if m.Exists("test.log.3") {
		t.Error("expected test.log.3 to be deleted as it's too old")
	}
}
