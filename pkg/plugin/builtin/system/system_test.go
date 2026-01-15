package system

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Snider/Core/pkg/plugin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSystemPlugin_Info(t *testing.T) {
	router := plugin.NewRouter()
	ctx := context.Background()

	p := New()
	require.NoError(t, router.Register(ctx, p))

	req := httptest.NewRequest("GET", "/api/core/system/info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp InfoResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "Core", resp.Name)
	assert.NotEmpty(t, resp.GoVersion)
	assert.NotEmpty(t, resp.OS)
	assert.NotEmpty(t, resp.Arch)
}

func TestSystemPlugin_Health(t *testing.T) {
	router := plugin.NewRouter()
	ctx := context.Background()

	p := New()
	require.NoError(t, router.Register(ctx, p))

	req := httptest.NewRequest("GET", "/api/core/system/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "healthy", resp.Status)
	assert.NotEmpty(t, resp.Uptime)
}

func TestSystemPlugin_Runtime(t *testing.T) {
	router := plugin.NewRouter()
	ctx := context.Background()

	p := New()
	require.NoError(t, router.Register(ctx, p))

	req := httptest.NewRequest("GET", "/api/core/system/runtime", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp RuntimeResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Greater(t, resp.NumGoroutine, 0)
	assert.Greater(t, resp.NumCPU, 0)
	assert.Greater(t, resp.MemAlloc, uint64(0))
}

func TestSystemPlugin_Metadata(t *testing.T) {
	p := New()

	assert.Equal(t, "system", p.Name())
	assert.Equal(t, "core", p.Namespace())

	info := p.Info()
	assert.Equal(t, "Core system information and operations", info.Description)
	assert.Equal(t, "1.0.0", info.Version)
}
