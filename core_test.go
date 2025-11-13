package core

import (
	"embed"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestCore_New_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

// Mock service for testing
type MockService struct {
	Name string
}

func (m *MockService) GetName() string {
	return m.Name
}

func TestCore_WithService_Good(t *testing.T) {
	factory := func(c *Core) (any, error) {
		return &MockService{Name: "test"}, nil
	}
	c, err := New(WithService(factory))
	assert.NoError(t, err)
	svc := c.Service("core")
	assert.NotNil(t, svc)
	mockSvc, ok := svc.(*MockService)
	assert.True(t, ok)
	assert.Equal(t, "test", mockSvc.GetName())
}

func TestCore_WithService_Bad(t *testing.T) {
	factory := func(c *Core) (any, error) {
		return nil, assert.AnError
	}
	_, err := New(WithService(factory))
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

type MockConfigService struct{}

func (m *MockConfigService) Get(key string, out any) error { return nil }
func (m *MockConfigService) Set(key string, v any) error   { return nil }

type MockDisplayService struct{}

func (m *MockDisplayService) OpenWindow(opts ...WindowOption) error { return nil }

func TestCore_Services_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	err = c.RegisterService("config", &MockConfigService{})
	assert.NoError(t, err)

	err = c.RegisterService("display", &MockDisplayService{})
	assert.NoError(t, err)

	assert.NotNil(t, c.Config())
	assert.NotNil(t, c.Display())
}

func TestCore_Services_Ugly(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	assert.Panics(t, func() {
		c.Config()
	})
	assert.Panics(t, func() {
		c.Display()
	})
}

func TestCore_App_Good(t *testing.T) {
	app := &application.App{}
	c, err := New(WithWails(app))
	assert.NoError(t, err)

	// To test the global App() function, we need to set the global instance.
	originalInstance := instance
	instance = c
	defer func() { instance = originalInstance }()

	assert.Equal(t, app, App())
}

func TestCore_App_Ugly(t *testing.T) {
	// This test ensures that calling App() before the core is initialized panics.
	originalInstance := instance
	instance = nil
	defer func() { instance = originalInstance }()
	assert.Panics(t, func() {
		App()
	})
}

func TestCore_Core_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	assert.Equal(t, c, c.Core())
}

func TestFeatures_IsEnabled_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	c.Features.Flags = []string{"feature1", "feature2"}

	assert.True(t, c.Features.IsEnabled("feature1"))
	assert.True(t, c.Features.IsEnabled("feature2"))
	assert.False(t, c.Features.IsEnabled("feature3"))
}

type startupMessage struct{}
type shutdownMessage struct{}

func TestCore_ServiceLifecycle_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	var messageReceived Message
	handler := func(c *Core, msg Message) error {
		messageReceived = msg
		return nil
	}
	c.RegisterAction(handler)

	// Test Startup
	_ = c.ServiceStartup(nil, application.ServiceOptions{})
	_, ok := messageReceived.(ActionServiceStartup)
	assert.True(t, ok, "expected ActionServiceStartup message")

	// Test Shutdown
	_ = c.ServiceShutdown(nil)
	_, ok = messageReceived.(ActionServiceShutdown)
	assert.True(t, ok, "expected ActionServiceShutdown message")
}

func TestCore_WithWails_Good(t *testing.T) {
	app := &application.App{}
	c, err := New(WithWails(app))
	assert.NoError(t, err)
	assert.Equal(t, app, c.App)
}

//go:embed core/testdata
var testFS embed.FS

func TestCore_WithAssets_Good(t *testing.T) {
	c, err := New(WithAssets(testFS))
	assert.NoError(t, err)
	assets := c.Assets()
	file, err := assets.Open("core/testdata/test.txt")
	assert.NoError(t, err)
	defer file.Close()
	content, err := io.ReadAll(file)
	assert.NoError(t, err)
	assert.Equal(t, "hello from testdata\n", string(content))
}

func TestCore_WithServiceLock_Good(t *testing.T) {
	c, err := New(WithServiceLock())
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{})
	assert.Error(t, err)
}

func TestCore_RegisterService_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{Name: "test"})
	assert.NoError(t, err)
	svc := c.Service("test")
	assert.NotNil(t, svc)
	mockSvc, ok := svc.(*MockService)
	assert.True(t, ok)
	assert.Equal(t, "test", mockSvc.GetName())
}

func TestCore_RegisterService_Bad(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{})
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{})
	assert.Error(t, err)
	err = c.RegisterService("", &MockService{})
	assert.Error(t, err)
}

func TestCore_ServiceFor_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{Name: "test"})
	assert.NoError(t, err)
	svc, err := ServiceFor[*MockService](c, "test")
	assert.NoError(t, err)
	assert.Equal(t, "test", svc.GetName())
}

func TestCore_ServiceFor_Bad(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	_, err = ServiceFor[*MockService](c, "nonexistent")
	assert.Error(t, err)
	err = c.RegisterService("test", "not a service")
	assert.NoError(t, err)
	_, err = ServiceFor[*MockService](c, "test")
	assert.Error(t, err)
}

func TestCore_MustServiceFor_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	err = c.RegisterService("test", &MockService{Name: "test"})
	assert.NoError(t, err)
	svc := MustServiceFor[*MockService](c, "test")
	assert.Equal(t, "test", svc.GetName())
}

func TestCore_MustServiceFor_Ugly(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	assert.Panics(t, func() {
		MustServiceFor[*MockService](c, "nonexistent")
	})
	err = c.RegisterService("test", "not a service")
	assert.NoError(t, err)
	assert.Panics(t, func() {
		MustServiceFor[*MockService](c, "test")
	})
}

type MockAction struct {
	handled bool
}

func (a *MockAction) Handle(c *Core, msg Message) error {
	a.handled = true
	return nil
}

func TestCore_ACTION_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	action := &MockAction{}
	c.RegisterAction(action.Handle)
	err = c.ACTION(nil)
	assert.NoError(t, err)
	assert.True(t, action.handled)
}

func TestCore_RegisterActions_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	action1 := &MockAction{}
	action2 := &MockAction{}
	c.RegisterActions(action1.Handle, action2.Handle)
	err = c.ACTION(nil)
	assert.NoError(t, err)
	assert.True(t, action1.handled)
	assert.True(t, action2.handled)
}

func TestCore_WithName_Good(t *testing.T) {
	factory := func(c *Core) (any, error) {
		return &MockService{Name: "test"}, nil
	}
	c, err := New(WithName("my-service", factory))
	assert.NoError(t, err)
	svc := c.Service("my-service")
	assert.NotNil(t, svc)
	mockSvc, ok := svc.(*MockService)
	assert.True(t, ok)
	assert.Equal(t, "test", mockSvc.GetName())
}

func TestCore_WithName_Bad(t *testing.T) {
	factory := func(c *Core) (any, error) {
		return nil, assert.AnError
	}
	_, err := New(WithName("my-service", factory))
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}
