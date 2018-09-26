package stripe

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient(t *testing.T) (*Client, *http.ServeMux, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	c := &Client{
		baseURL: server.URL,
	}
	return c, mux, func() {
		server.Close()
	}
}
