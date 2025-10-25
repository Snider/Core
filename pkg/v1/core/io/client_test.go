package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	m := NewMockMedium()
	m.Files["test.txt"] = "hello"
	content, err := Read(m, "test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello", content)
}

func TestWrite(t *testing.T) {
	m := NewMockMedium()
	err := Write(m, "test.txt", "hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello", m.Files["test.txt"])
}

func TestCopy(t *testing.T) {
	source := NewMockMedium()
	dest := NewMockMedium()
	source.Files["test.txt"] = "hello"
	err := Copy(source, "test.txt", dest, "test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello", dest.Files["test.txt"])
}
