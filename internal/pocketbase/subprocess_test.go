package pocketbase_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andreoliwa/logseq-doctor/internal/pocketbase"
	"github.com/stretchr/testify/require"
)

func TestWaitForReadySucceeds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	err := pocketbase.WaitForReady(srv.URL, 2*time.Second)
	require.NoError(t, err)
}

func TestWaitForReadyTimesOut(t *testing.T) {
	// Port 19999 should not be listening; the call must time out quickly.
	err := pocketbase.WaitForReady("http://127.0.0.1:19999", 300*time.Millisecond)
	require.Error(t, err)
	require.ErrorIs(t, err, pocketbase.ErrWaitTimeout)
	require.ErrorContains(t, err, "last health check")
}

func TestWaitForReadyRetriesNonOKResponses(t *testing.T) {
	attempts := 0

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			writer.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		writer.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	require.NoError(t, pocketbase.WaitForReady(srv.URL, time.Second))
	require.Equal(t, 3, attempts)
}
