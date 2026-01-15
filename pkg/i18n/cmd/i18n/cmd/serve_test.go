package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServe(t *testing.T) {
	// Create a new test server with the application's router
	ts := httptest.NewServer(setupRouter())
	defer ts.Close()

	// Make a request to the demo endpoint
	res, err := http.Get(ts.URL + "/api/v1/demo")
	assert.NoError(t, err)
	defer res.Body.Close()

	// Check the status code
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Check the response body
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, world!", string(body))
}
