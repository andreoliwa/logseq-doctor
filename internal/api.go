package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/andreoliwa/logseq-go"
	"io"
	"log"
	"net/http"
	"strings"
)

type LogseqAPI interface {
	PostQuery(query string) (string, error)
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

// OpenGraphFromPath opens a Logseq graph (from the path provided when the instance was created).
// It doesn't return an error and aborts the program if it fails because it's an internal function.
// This is done on purpose to avoid error handling boilerplate code throughout the package.
func OpenGraphFromPath(path string) *logseq.Graph {
	if path == "" {
		log.Fatalln("path is empty, maybe the LOGSEQ_GRAPH_PATH environment variable is not set")
	}

	ctx := context.Background()

	graph, err := logseq.Open(ctx, path)
	if err != nil {
		log.Fatalln("error opening graph: %w", err)
	}

	return graph
}

// PostQuery sends a query to the Logseq API and returns the result as JSON.
func (l *logseqAPIImpl) PostQuery(query string) (string, error) {
	if l.apiToken == "" || l.hostURL == "" {
		return "", ErrMissingConfig
	}

	client := &http.Client{} //nolint:exhaustruct

	jsonQuery, err := json.Marshal(query)
	if err != nil {
		return "", fmt.Errorf("failed to marshal query: %w", err)
	}

	ctx := context.Background()
	payload := fmt.Sprintf(`{"method":"logseq.db.q","args":[%s]}`, string(jsonQuery))

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
// It aborts the program in case of error because it's an internal function.
// Also, it's not common to have errors when opening a page.
func OpenPage(graph *logseq.Graph, pageTitle string) logseq.Page {
	page, err := graph.OpenPage(pageTitle)
	if err != nil {
		log.Fatalf("error opening page: %v", err)
	}

	return page
}
