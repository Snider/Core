package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIO_Read_Good(t *testing.T) {
	medium := NewMockMedium()
	medium.Files["test.txt"] = "hello"

	content, err := Read(medium, "test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello", content)
}

func TestIO_Read_Bad(t *testing.T) {
	medium := NewMockMedium()

	_, err := Read(medium, "nonexistent.txt")
	assert.Error(t, err)
}

func TestIO_Write_Good(t *testing.T) {
	medium := NewMockMedium()

	err := Write(medium, "test.txt", "hello")
	assert.NoError(t, err)

	writtenContent, ok := medium.Files["test.txt"]
	assert.True(t, ok)
	assert.Equal(t, "hello", writtenContent)
}

// TODO: The current MockMedium cannot simulate a write error.
// func TestIO_Write_Bad(t *testing.T) {
// 	medium := NewMockMedium()
// 	// How to make Write fail?
// 	err := Write(medium, "test.txt", "hello")
// 	assert.Error(t, err)
// }

func TestIO_EnsureDir_Good(t *testing.T) {
	medium := NewMockMedium()
	err := EnsureDir(medium, "testdir")
	assert.NoError(t, err)
	exists := medium.Dirs["testdir"]
	assert.True(t, exists)
}

// TODO: The current MockMedium cannot simulate an EnsureDir error.
// func TestIO_EnsureDir_Bad(t *testing.T) {
// 	medium := NewMockMedium()
// 	// How to make EnsureDir fail?
// 	err := EnsureDir(medium, "testdir")
// 	assert.Error(t, err)
// }

func TestIO_IsFile_Good(t *testing.T) {
	medium := NewMockMedium()
	medium.Files["test.txt"] = "content"
	assert.True(t, IsFile(medium, "test.txt"))
	assert.False(t, IsFile(medium, "nonexistent.txt"))
}

func TestIO_Copy_Good(t *testing.T) {
	source := NewMockMedium()
	source.Files["source.txt"] = "hello"

	dest := NewMockMedium()

	err := Copy(source, "source.txt", dest, "dest.txt")
	assert.NoError(t, err)

	copiedContent, ok := dest.Files["dest.txt"]
	assert.True(t, ok)
	assert.Equal(t, "hello", copiedContent)
}

func TestIO_Copy_Bad(t *testing.T) {
	source := NewMockMedium() // No source file
	dest := NewMockMedium()

	err := Copy(source, "source.txt", dest, "dest.txt")
	assert.Error(t, err)
}
