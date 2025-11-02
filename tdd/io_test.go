package tdd

import (
	"errors"
	"testing"

	"github.com/Snider/Core/pkg/io"
	"github.com/stretchr/testify/assert"
)

type MockMedium struct {
	ReadFileFunc  func(path string) (string, error)
	WriteFileFunc func(path, content string) error
	EnsureDirFunc func(path string) error
	IsFileFunc    func(path string) bool
}

func (m *MockMedium) Read(path string) (string, error) {
	if m.ReadFileFunc != nil {
		return m.ReadFileFunc(path)
	}
	return "", errors.New("not implemented")
}

func (m *MockMedium) Write(path, content string) error {
	if m.WriteFileFunc != nil {
		return m.WriteFileFunc(path, content)
	}
	return errors.New("not implemented")
}

func (m *MockMedium) EnsureDir(path string) error {
	if m.EnsureDirFunc != nil {
		return m.EnsureDirFunc(path)
	}
	return errors.New("not implemented")
}

func (m *MockMedium) IsFile(path string) bool {
	if m.IsFileFunc != nil {
		return m.IsFileFunc(path)
	}
	return false
}

func (m *MockMedium) FileGet(path string) (string, error) {
	return m.Read(path)
}

func (m *MockMedium) FileSet(path, content string) error {
	return m.Write(path, content)
}

func TestIO_Read_Good(t *testing.T) {
	medium := &MockMedium{
		ReadFileFunc: func(path string) (string, error) {
			assert.Equal(t, "test.txt", path)
			return "hello", nil
		},
	}
	content, err := io.Read(medium, "test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello", content)
}

func TestIO_Read_Bad(t *testing.T) {
	medium := &MockMedium{
		ReadFileFunc: func(path string) (string, error) {
			return "", assert.AnError
		},
	}
	_, err := io.Read(medium, "test.txt")
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestIO_Write_Good(t *testing.T) {
	medium := &MockMedium{
		WriteFileFunc: func(path, content string) error {
			assert.Equal(t, "test.txt", path)
			assert.Equal(t, "hello", content)
			return nil
		},
	}
	err := io.Write(medium, "test.txt", "hello")
	assert.NoError(t, err)
}

func TestIO_Write_Bad(t *testing.T) {
	medium := &MockMedium{
		WriteFileFunc: func(path, content string) error {
			return assert.AnError
		},
	}
	err := io.Write(medium, "test.txt", "hello")
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestIO_EnsureDir_Good(t *testing.T) {
	medium := &MockMedium{
		EnsureDirFunc: func(path string) error {
			assert.Equal(t, "testdir", path)
			return nil
		},
	}
	err := io.EnsureDir(medium, "testdir")
	assert.NoError(t, err)
}

func TestIO_EnsureDir_Bad(t *testing.T) {
	medium := &MockMedium{
		EnsureDirFunc: func(path string) error {
			return assert.AnError
		},
	}
	err := io.EnsureDir(medium, "testdir")
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestIO_IsFile_Good(t *testing.T) {
	medium := &MockMedium{
		IsFileFunc: func(path string) bool {
			assert.Equal(t, "test.txt", path)
			return true
		},
	}
	assert.True(t, io.IsFile(medium, "test.txt"))
}

func TestIO_Copy_Good(t *testing.T) {
	source := &MockMedium{
		ReadFileFunc: func(path string) (string, error) {
			assert.Equal(t, "source.txt", path)
			return "hello", nil
		},
	}
	dest := &MockMedium{
		WriteFileFunc: func(path, content string) error {
			assert.Equal(t, "dest.txt", path)
			assert.Equal(t, "hello", content)
			return nil
		},
	}
	err := io.Copy(source, "source.txt", dest, "dest.txt")
	assert.NoError(t, err)
}

func TestIO_Copy_Bad(t *testing.T) {
	source := &MockMedium{
		ReadFileFunc: func(path string) (string, error) {
			return "", assert.AnError
		},
	}
	dest := &MockMedium{}
	err := io.Copy(source, "source.txt", dest, "dest.txt")
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}
