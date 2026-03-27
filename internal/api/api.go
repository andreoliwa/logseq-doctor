// Package api provides the Logseq HTTP API client and graph-opening utilities.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/andreoliwa/logseq-go"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
)

// ErrFailedOpenGraph is returned when the graph cannot be opened.
var ErrFailedOpenGraph = errors.New("failed to open graph")

// ErrMissingConfig is returned when the Logseq API token or host URL is not set.
var ErrMissingConfig = errors.New("LOGSEQ_API_TOKEN and LOGSEQ_HOST_URL must be set")

// ErrInvalidResponseStatus is returned when the Logseq API returns a non-200 status code.
var ErrInvalidResponseStatus = errors.New("invalid response status")

// LogseqAPI is the interface for communicating with a running Logseq instance via its HTTP API.
type LogseqAPI interface {
	PostQuery(query string) (string, error)
	PostDatascriptQuery(query string) (string, error)
	// UpsertBlockProperty sets a block property via the Logseq Editor API.
	// This causes Logseq to write the property to the .md file immediately,
	// which is useful to force the id:: property onto disk for blocks that
	// Logseq has tracked internally but not yet written back.
	UpsertBlockProperty(uuid, key, value string) error
}

type logseqAPIImpl struct {
	path     string
	hostURL  string
	apiToken string
}

// NewLogseqAPI creates a new LogseqAPI instance.
func NewLogseqAPI(path, hostURL, apiToken string) LogseqAPI {
	return &logseqAPIImpl{
		path:     path,
		hostURL:  hostURL,
		apiToken: apiToken,
	}
}

// OpenGraphFromPath opens a Logseq graph from the given directory path.
// Delegates to logseqext.OpenGraphFromPath.
func OpenGraphFromPath(path string) *logseq.Graph {
	return logseqext.OpenGraphFromPath(path)
}

// PostQuery sends a query to the Logseq API and returns the result as JSON.
func (l *logseqAPIImpl) PostQuery(query string) (string, error) {
	return l.postAPI("logseq.db.q", query)
}

// PostDatascriptQuery sends a Datascript query ([:find ...]) to the Logseq API.
// Use this instead of PostQuery for queries that require pull syntax or complex patterns.
func (l *logseqAPIImpl) PostDatascriptQuery(query string) (string, error) {
	return l.postAPI("logseq.db.datascriptQuery", query)
}

// UpsertBlockProperty calls logseq.Editor.upsertBlockProperty to set a property on a block.
// Unlike PostQuery/PostDatascriptQuery which take a single query string, this method passes
// three separate args: uuid, key, value. Logseq then writes the property to the .md file,
// which forces the id:: property onto disk for blocks that haven't been persisted yet.
// Logseq lazy-writes block UUIDs: they exist in its DB but only hit .md files when something
// triggers a write (a backlink, an edit, or this Editor API call).
func (l *logseqAPIImpl) UpsertBlockProperty(uuid, key, value string) error {
	if l.apiToken == "" || l.hostURL == "" {
		return ErrMissingConfig
	}

	uuidJSON, err := json.Marshal(uuid)
	if err != nil {
		return fmt.Errorf("failed to marshal uuid: %w", err)
	}

	keyJSON, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("failed to marshal key: %w", err)
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	ctx := context.Background()
	payload := fmt.Sprintf(`{"method":"logseq.Editor.upsertBlockProperty","args":[%s,%s,%s]}`,
		string(uuidJSON), string(keyJSON), string(valueJSON))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.hostURL+"/api", strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+l.apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{} //nolint:exhaustruct

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error performing HTTP request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %s: %w", resp.Status, ErrInvalidResponseStatus)
	}

	return nil
}

// postAPI is the shared HTTP implementation for PostQuery and PostDatascriptQuery.
func (l *logseqAPIImpl) postAPI(method, query string) (string, error) {
	if l.apiToken == "" || l.hostURL == "" {
		return "", ErrMissingConfig
	}

	client := &http.Client{} //nolint:exhaustruct

	jsonQuery, err := json.Marshal(query)
	if err != nil {
		return "", fmt.Errorf("failed to marshal query: %w", err)
	}

	ctx := context.Background()
	payload := fmt.Sprintf(`{"method":%q,"args":[%s]}`, method, string(jsonQuery))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.hostURL+"/api", strings.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+l.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error performing HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %s with payload:\n%s: %w", resp.Status, payload, ErrInvalidResponseStatus)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// OpenPage opens a page in the Logseq graph.
// Delegates to logseqext.OpenPage.
func OpenPage(graph *logseq.Graph, pageTitle string) logseq.Page {
	return logseqext.OpenPage(graph, pageTitle)
}
