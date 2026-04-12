package serve_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/serve"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProxyInjectsAuthHeader(t *testing.T) {
	var capturedAuth string

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")

		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	proxy := serve.NewProxy(backend.URL, "test-token-abc")
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/collections/lqd_tasks/records", nil)

	rr := httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Bearer test-token-abc", capturedAuth)
}

func TestNewProxyOverridesClientAuthHeader(t *testing.T) {
	var capturedAuth string

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")

		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	proxy := serve.NewProxy(backend.URL, "server-token")
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/collections/lqd_tasks/records", nil)
	req.Header.Set("Authorization", "Bearer client-supplied-token")

	rr := httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)

	// Server token must override any client-supplied token.
	assert.Equal(t, "Bearer server-token", capturedAuth)
}
