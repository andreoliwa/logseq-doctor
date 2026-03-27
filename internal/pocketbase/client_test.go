package pocketbase_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andreoliwa/logseq-doctor/internal/pocketbase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/api/collections/_superusers/auth-with-password", request.URL.Path)
		assert.Equal(t, http.MethodPost, request.Method)

		var body map[string]string
		assert.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, "admin@test.com", body["identity"])
		assert.Equal(t, "secret", body["password"])

		writer.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(writer).Encode(map[string]string{"token": "test-token-123"})
		assert.NoError(t, err)
	}))
	defer server.Close()

	client, err := pocketbase.NewClient(server.URL, "admin@test.com", "secret")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestAuthenticate_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusUnauthorized)

		err := json.NewEncoder(writer).Encode(map[string]string{"message": "invalid credentials"})
		assert.NoError(t, err)
	}))
	defer server.Close()

	client, err := pocketbase.NewClient(server.URL, "bad@test.com", "wrong")
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestAuthenticate_Unreachable(t *testing.T) {
	client, err := pocketbase.NewClient("http://127.0.0.1:19999", "a@b.com", "x")
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "cannot connect")
}

func newTestClient(t *testing.T, handler http.HandlerFunc) (*pocketbase.Client, *httptest.Server) {
	t.Helper()

	authCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/collections/_superusers/auth-with-password" && !authCalled {
			authCalled = true

			writer.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(writer).Encode(map[string]string{"token": "test-token"})
			assert.NoError(t, err)

			return
		}

		handler(writer, request)
	}))

	client, err := pocketbase.NewClient(server.URL, "a@b.com", "pass")
	require.NoError(t, err)

	return client, server
}

func TestCollectionExists_True(t *testing.T) {
	client, server := newTestClient(t, func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/api/collections/lqd_tasks", request.URL.Path)
		assert.Equal(t, http.MethodGet, request.Method)

		writer.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(writer).Encode(map[string]string{"id": "abc123", "name": "lqd_tasks"})
		assert.NoError(t, err)
	})
	defer server.Close()

	exists, err := client.CollectionExists("lqd_tasks")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestCollectionExists_False(t *testing.T) {
	client, server := newTestClient(t, func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	exists, err := client.CollectionExists("lqd_tasks")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestDeleteCollection(t *testing.T) {
	client, server := newTestClient(t, func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet {
			writer.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(writer).Encode(map[string]string{"id": "abc123", "name": "lqd_tasks"})
			assert.NoError(t, err)

			return
		}

		assert.Equal(t, http.MethodDelete, request.Method)
		assert.Equal(t, "/api/collections/abc123", request.URL.Path)

		writer.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteCollection("lqd_tasks")
	require.NoError(t, err)
}

func TestCreateCollection(t *testing.T) {
	client, server := newTestClient(t, func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/api/collections", request.URL.Path)
		assert.Equal(t, http.MethodPost, request.Method)

		writer.WriteHeader(http.StatusOK)

		_, err := writer.Write([]byte(`{"id":"new123","name":"lqd_tasks"}`))
		assert.NoError(t, err)
	})
	defer server.Close()

	schema := map[string]any{"name": "lqd_tasks", "type": "base"}
	err := client.CreateCollection(schema)
	require.NoError(t, err)
}

func TestFetchRecords_Paginated(t *testing.T) {
	page := 0

	client, server := newTestClient(t, func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Contains(t, request.URL.Path, "/api/collections/lqd_tasks/records")

		page++

		writer.Header().Set("Content-Type", "application/json")

		if page == 1 {
			err := json.NewEncoder(writer).Encode(map[string]any{
				"page":       1,
				"totalPages": 2,
				"items": []map[string]any{
					{"id": "uuid-1", "name": "Task 1"},
					{"id": "uuid-2", "name": "Task 2"},
				},
			})
			assert.NoError(t, err)
		} else {
			err := json.NewEncoder(writer).Encode(map[string]any{
				"page":       2,
				"totalPages": 2,
				"items": []map[string]any{
					{"id": "uuid-3", "name": "Task 3"},
				},
			})
			assert.NoError(t, err)
		}
	})
	defer server.Close()

	records, err := client.FetchRecords("lqd_tasks", "", "")
	require.NoError(t, err)
	assert.Len(t, records, 3)
	assert.Equal(t, "uuid-1", records[0]["id"])
	assert.Equal(t, "uuid-3", records[2]["id"])
}

func TestCreateRecord(t *testing.T) {
	client, server := newTestClient(t, func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/api/collections/lqd_tasks/records", request.URL.Path)

		var body map[string]any

		assert.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		assert.Equal(t, "abc-123", body["id"])

		writer.WriteHeader(http.StatusOK)

		err := json.NewEncoder(writer).Encode(body)
		assert.NoError(t, err)
	})
	defer server.Close()

	err := client.CreateRecord("lqd_tasks", map[string]any{"id": "abc-123", "name": "Test"})
	require.NoError(t, err)
}

func TestUpdateRecord(t *testing.T) {
	client, server := newTestClient(t, func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPatch, request.Method)
		assert.Equal(t, "/api/collections/lqd_tasks/records/abc-123", request.URL.Path)

		writer.WriteHeader(http.StatusOK)

		_, err := writer.Write([]byte(`{}`))
		assert.NoError(t, err)
	})
	defer server.Close()

	err := client.UpdateRecord("lqd_tasks", "abc-123", map[string]any{"name": "Updated"})
	require.NoError(t, err)
}

func TestDeleteRecord(t *testing.T) {
	client, server := newTestClient(t, func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodDelete, request.Method)
		assert.Equal(t, "/api/collections/lqd_tasks/records/abc-123", request.URL.Path)

		writer.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteRecord("lqd_tasks", "abc-123")
	require.NoError(t, err)
}
