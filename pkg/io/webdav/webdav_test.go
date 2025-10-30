package webdav

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockWebDAVServer creates a test HTTP server that mimics a WebDAV server.
func mockWebDAVServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			if r.URL.Path == "/" {
				w.WriteHeader(http.StatusMultiStatus)
				return
			}
			// For IsFile test
			if r.URL.Path == "/test.txt" {
				w.WriteHeader(http.StatusMultiStatus)
				fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<D:multistatus xmlns:D="DAV:">
  <D:response>
    <D:href>/test.txt</D:href>
    <D:propstat>
      <D:prop>
        <D:resourcetype/>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>`)
				return
			}
			if r.URL.Path == "/testdir/" {
				w.WriteHeader(http.StatusMultiStatus)
				fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<D:multistatus xmlns:D="DAV:">
  <D:response>
    <D:href>/testdir/</D:href>
    <D:propstat>
      <D:prop>
        <D:resourcetype><D:collection/></D:resourcetype>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>`)
				return
			}
			http.NotFound(w, r)
		case "GET":
			if r.URL.Path == "/test.txt" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "Hello, WebDAV!")
				return
			}
			http.NotFound(w, r)
		case "PUT":
			if r.URL.Path == "/test.txt" {
				w.WriteHeader(http.StatusCreated)
				return
			}
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		case "MKCOL":
			if r.URL.Path == "/testdir/" {
				w.WriteHeader(http.StatusCreated)
				return
			}
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
	return httptest.NewServer(handler)
}

func TestNew_Success(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	cfg := ConnectionConfig{
		URL:      server.URL,
		User:     "user",
		Password: "password",
	}

	medium, err := New(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, medium)
}

func TestRead(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	medium := &Medium{client: server.Client(), baseURL: server.URL}
	content, err := medium.Read("test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "Hello, WebDAV!", content)
}

func TestWrite(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	medium := &Medium{client: server.Client(), baseURL: server.URL}
	err := medium.Write("test.txt", "Hello, WebDAV!")
	assert.NoError(t, err)
}

func TestEnsureDir(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	medium := &Medium{client: server.Client(), baseURL: server.URL}
	err := medium.EnsureDir("testdir")
	assert.NoError(t, err)
}

func TestIsFile(t *testing.T) {
	server := mockWebDAVServer()
	defer server.Close()

	medium := &Medium{client: server.Client(), baseURL: server.URL}
	assert.True(t, medium.IsFile("test.txt"))
	assert.False(t, medium.IsFile("testdir"))
}
