package gocache

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPPool(t *testing.T) {
	p := NewHTTPPool("localhost:8080")
	NewGroup("identity", 0, GetterFunc(
		func(key string) ([]byte, error) {
			return []byte(key), nil
		}))

	resp := poolFetch(p, "identity", "key")
	resp.Body.Close()
	if b, _ := io.ReadAll(resp.Body); resp.StatusCode != http.StatusOK || string(b) != "key" {
		t.Fatalf("HTTP Get failed (expected: key, got: %s)", b)
	}

	resp = poolFetch(p, "nonexistent", "key")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("HTTP Get failed (expected: %v, got: %v)", http.StatusNotFound, resp.StatusCode)
	}
}

// poolFetch sends a GET request to the HTTP pool and returns the response.
func poolFetch(p *HTTPPool, group string, key string) *http.Response {
	endpoint := fmt.Sprintf("%s/%s/%s", p.basePath, group, key)
	req := httptest.NewRequest("GET", endpoint, nil)
	w := httptest.NewRecorder()
	p.GetHTTPHandler().ServeHTTP(w, req)
	resp := w.Result()
	return resp
}
