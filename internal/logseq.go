package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/andreoliwa/logseq-go"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func OpenGraphFromDirOrEnv(dir string) *logseq.Graph {
	if dir == "" {
		dir = os.Getenv("LOGSEQ_GRAPH_PATH")
		if dir == "" {
			log.Fatalln("LOGSEQ_GRAPH_PATH environment variable is not set.")
		}
	}

	ctx := context.Background()

	graph, err := logseq.Open(ctx, dir)
	if err != nil {
		log.Fatalln("error opening graph: %w", err)
	}

	return graph
}

// QueryLogseqAPI sends a query to the Logseq API and returns the result as JSON.
func QueryLogseqAPI(query string) (string, error) {
	apiToken := os.Getenv("LOGSEQ_API_TOKEN")

	hostURL := os.Getenv("LOGSEQ_HOST_URL")
	if apiToken == "" || hostURL == "" {
		return "", ErrMissingConfig
	}

	client := &http.Client{} //nolint:exhaustruct

	jsonQuery, err := json.Marshal(query)
	if err != nil {
		return "", fmt.Errorf("failed to marshal query: %w", err)
	}

	ctx := context.Background()
	payload := fmt.Sprintf(`{"method":"logseq.db.q","args":[%s]}`, string(jsonQuery))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hostURL+"/api", strings.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
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
