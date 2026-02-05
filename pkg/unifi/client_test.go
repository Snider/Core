package unifi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Mock UniFi controller response for login/initialization
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"meta":{"rc":"ok"}, "data": []}`)
	}))
	defer ts.Close()

	// Test basic client creation
	client, err := New(ts.URL, "user", "pass", "", true)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, ts.URL, client.URL())
	assert.NotNil(t, client.API())

	if client.API().Client != nil && client.API().Client.Transport != nil {
		if tr, ok := client.API().Client.Transport.(*http.Transport); ok {
			assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
		} else {
			t.Errorf("expected *http.Transport, got %T", client.API().Client.Transport)
		}
	} else {
		t.Errorf("client or transport is nil")
	}

	// Test with insecure false
	client, err = New(ts.URL, "user", "pass", "", false)
	assert.NoError(t, err)
	if tr, ok := client.API().Client.Transport.(*http.Transport); ok {
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
	}
}

func TestNew_Error(t *testing.T) {
	// uf.NewUnifi fails if URL is invalid (e.g. missing scheme)
	client, err := New("localhost:8443", "user", "pass", "", false)
	assert.Error(t, err)
	assert.Nil(t, client)
}
