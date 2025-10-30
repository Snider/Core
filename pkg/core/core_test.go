package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockServiceInterface is an interface that MockService implements.
type MockServiceInterface interface {
	GetName() string
}

// MockService is a simple struct to act as a mock service.
type MockService struct {
	Name string
}

func (m *MockService) GetName() string {
	return m.Name
}

// MockConfig is a mock implementation of the Config interface.
type MockConfig struct{}

func (m *MockConfig) Get(key string, out any) error { return nil }
func (m *MockConfig) Set(key string, v any) error   { return nil }

func TestNew(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.NotNil(t, c.services)
	assert.False(t, c.servicesLocked)
}

func TestWithService(t *testing.T) {
	// Test successful service registration
	t.Run("successful service registration", func(t *testing.T) {
		c, err := New()
		assert.NoError(t, err)

		factory := func(c *Core) (any, error) {
			return &MockService{Name: "testService"}, nil
		}

		err = WithService(factory)(c)
		assert.NoError(t, err)

		// The service name is derived from the package path of MockService, which is "core"
		svc := c.Service("core")
		assert.NotNil(t, svc)
		mockSvc, ok := svc.(*MockService)
		assert.True(t, ok)
		assert.Equal(t, "testService", mockSvc.Name)
	})

	// Test service registration with factory error
	t.Run("service registration with factory error", func(t *testing.T) {
		c, err := New()
		assert.NoError(t, err)

		factory := func(c *Core) (any, error) {
			return nil, errors.New("factory error")
		}

		err = WithService(factory)(c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "factory error")
	})

	// Test service registration when services are locked
	t.Run("service registration when locked", func(t *testing.T) {
		c, err := New(WithServiceLock())
		assert.NoError(t, err)

		factory := func(c *Core) (any, error) {
			return &MockService{Name: "lockedService"}, nil
		}

		err = WithService(factory)(c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not permitted by the serviceLock setting")
	})
}

func TestServiceFor(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// Register a mock service
	err = c.RegisterService("mockservice", &MockService{Name: "testService"})
	assert.NoError(t, err)

	// Test successful retrieval as an interface
	t.Run("successful retrieval as interface", func(t *testing.T) {
		svc, err := ServiceFor[MockServiceInterface](c, "mockservice")
		assert.NoError(t, err)
		assert.Equal(t, "testService", svc.GetName())
	})

	// Test service not found
	t.Run("service not found", func(t *testing.T) {
		_, err := ServiceFor[MockServiceInterface](c, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service 'nonexistent' not found")
	})

	// Test type mismatch
	t.Run("type mismatch", func(t *testing.T) {
		err := c.RegisterService("stringservice", "hello")
		assert.NoError(t, err)
		_, err = ServiceFor[MockServiceInterface](c, "stringservice")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is of type string, but expected <nil>")
	})
}

func TestMustServiceFor(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// Register a mock service
	assert.NoError(t, c.RegisterService("mockservice", &MockService{Name: "testService"}))

	// Test successful retrieval as an interface
	assert.NotPanics(t, func() {
		svc := MustServiceFor[MockServiceInterface](c, "mockservice")
		assert.Equal(t, "testService", svc.GetName())
	})

	// Test service not found (should panic)
	assert.Panics(t, func() {
		MustServiceFor[MockServiceInterface](c, "nonexistent")
	})

	// Test type mismatch (should panic)
	assert.NoError(t, c.RegisterService("stringservice", "hello"))
	assert.Panics(t, func() {
		MustServiceFor[MockServiceInterface](c, "stringservice")
	})
}

func TestRegisterService(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// Test successful registration
	err = c.RegisterService("myservice", &MockService{})
	assert.NoError(t, err)
	assert.NotNil(t, c.Service("myservice"))

	// Test duplicate registration
	err = c.RegisterService("myservice", &MockService{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test empty name
	err = c.RegisterService("", &MockService{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service name cannot be empty")

	// Test registration when locked
	lockedCore, err := New(WithServiceLock())
	assert.NoError(t, err)
	err = lockedCore.RegisterService("lockedservice", &MockService{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not permitted by the serviceLock setting")
}

func TestCoreConfig(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// Register a mock config service
	mockCfg := &MockConfig{}
	err = c.RegisterService("config", mockCfg)
	assert.NoError(t, err)

	// Test successful retrieval of Config service
	cfg := c.Config()
	assert.NotNil(t, cfg)
	assert.Implements(t, (*Config)(nil), cfg)

	// Test panic if Config service not registered
	coreWithoutConfig, err := New()
	assert.NoError(t, err)
	assert.Panics(t, func() {
		coreWithoutConfig.Config()
	})
}
