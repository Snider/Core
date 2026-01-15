package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf bytes.Buffer
	done := make(chan bool)

	go func() {
		io.Copy(&buf, r)
		done <- true
	}()

	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	f()

	w.Close()
	<-done
	return buf.String()
}

func TestRootCmd(t *testing.T) {
	output := captureOutput(func() {
		rootCmd.SetArgs([]string{})
		rootCmd.Execute()
	})

	assert.Contains(t, output, "Hello from the i18n CLI!")
}

// Note: Testing Execute() directly is hard because it calls os.Exit(1) on failure.
// But we can test that rootCmd exists and has correct fields.
func TestRootCmdConfig(t *testing.T) {
	assert.Equal(t, "i18n", rootCmd.Use)
	assert.NotEmpty(t, rootCmd.Short)
	assert.NotEmpty(t, rootCmd.Long)
}
