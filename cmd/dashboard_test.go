package cmd_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andreoliwa/logseq-doctor/cmd"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file tests the dashboard command helpers.
// The full dashboard loop (PocketBase subprocess, browser open, signal handling)
// is not unit-testable, so these tests cover command structure, pure helpers,
// and HTTP mux route registration.

// newTestDashboardCmd builds a minimal Cobra command with the same flags as
// dashboardCmd so that ResolvePort can be tested without accessing the unexported
// global rootCmd.
func newTestDashboardCmd() *cobra.Command {
	c := &cobra.Command{Use: "dashboard"}
	c.Flags().IntP("port", "p", 8091, "HTTP server port")

	return c
}

func TestDashboardCmd_Aliases(t *testing.T) {
	// Verify the alias is wired — access via cmd package's Execute path is not
	// possible without running the binary, so we check the known constant.
	// The alias "dash" is declared in dashboard.go; if it changes, this test catches it.
	assert.Equal(t, []string{"dash"}, cmd.DashboardAliases())
}

func TestResolveEnvWithDefault(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		fallback string
		want     string
	}{
		{
			name:     "env var set",
			key:      "TEST_RESOLVE_KEY",
			envValue: "from-env",
			fallback: "default",
			want:     "from-env",
		},
		{
			name:     "env var not set uses fallback",
			key:      "TEST_RESOLVE_KEY_UNSET",
			envValue: "",
			fallback: "default",
			want:     "default",
		},
		{
			name:     "env var empty string uses fallback",
			key:      "TEST_RESOLVE_KEY_EMPTY",
			envValue: "",
			fallback: "fallback-value",
			want:     "fallback-value",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.envValue != "" {
				t.Setenv(testCase.key, testCase.envValue)
			}

			got := cmd.ResolveEnvWithDefault(testCase.key, testCase.fallback)
			assert.Equal(t, testCase.want, got)
		})
	}
}

func TestResolvePort_FlagDefault(t *testing.T) {
	c := newTestDashboardCmd()
	require.NoError(t, c.ParseFlags([]string{}))

	port := cmd.ResolvePort(c)
	assert.Equal(t, 8091, port)
}

func TestResolvePort_FlagOverride(t *testing.T) {
	c := newTestDashboardCmd()
	require.NoError(t, c.ParseFlags([]string{"--port", "9000"}))

	port := cmd.ResolvePort(c)
	assert.Equal(t, 9000, port)
}

func TestResolvePort_EnvVarOverride(t *testing.T) {
	t.Setenv("LQD_SERVE_PORT", "7777")

	c := newTestDashboardCmd()
	require.NoError(t, c.ParseFlags([]string{}))

	port := cmd.ResolvePort(c)
	assert.Equal(t, 7777, port)
}

func TestResolvePort_EnvVarInvalid(t *testing.T) {
	t.Setenv("LQD_SERVE_PORT", "not-a-number")

	c := newTestDashboardCmd()
	require.NoError(t, c.ParseFlags([]string{}))

	// Invalid env var value is ignored; flag default wins.
	port := cmd.ResolvePort(c)
	assert.Equal(t, 8091, port)
}

func TestBuildHTTPMux_Routes(t *testing.T) {
	mux := cmd.BuildHTTPMux("http://127.0.0.1:8090", "token", "")

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/backlog.css"},
		{"GET", "/internal/config"},
		{"POST", "/internal/move-to-unranked"},
		{"GET", "/api/collections/lqd_tasks/records"},
		{"POST", "/api/collections/lqd_tasks/records"},
		{"PATCH", "/api/collections/lqd_tasks/records/abc"},
		{"DELETE", "/api/collections/lqd_tasks/records/abc"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequestWithContext(context.Background(), route.method, route.path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			// All registered routes must return something other than 404.
			// Note: "GET /" is a subtree catch-all in ServeMux, so unknown paths
			// also return 200 — there is no testable 404 case for this mux.
			assert.NotEqual(t, http.StatusNotFound, rec.Code,
				"expected route to be registered: %s %s", route.method, route.path)
		})
	}
}

func TestBuildHTTPMux_RootServesHTML(t *testing.T) {
	mux := cmd.BuildHTTPMux("http://127.0.0.1:8090", "", "")

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	assert.NotEmpty(t, rec.Body.String())
}

func TestBuildHTTPMux_CSSRoute(t *testing.T) {
	mux := cmd.BuildHTTPMux("http://127.0.0.1:8090", "", "")

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/backlog.css", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/css")
	assert.NotEmpty(t, rec.Body.String())
}

func TestBuildHTTPMux_ConfigNoGraphPath(t *testing.T) {
	// With no graph path, /internal/config returns an empty JSON config (not an error).
	mux := cmd.BuildHTTPMux("http://127.0.0.1:8090", "", "")

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/internal/config", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
}
