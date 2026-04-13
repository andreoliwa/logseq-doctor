package pocketbase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const httpClientTimeout = 30 * time.Second

// Sentinel errors for PocketBase client operations.
var (
	ErrCannotConnect    = errors.New("cannot connect to PocketBase")
	ErrAuthFailed       = errors.New("PocketBase authentication failed")
	ErrUnexpectedStatus = errors.New("unexpected status from PocketBase")
)

// Client is a minimal PocketBase HTTP client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient authenticates with PocketBase and returns a ready-to-use client.
func NewClient(baseURL, username, password string) (*Client, error) {
	client := &Client{
		baseURL:    baseURL,
		token:      "",
		httpClient: &http.Client{Timeout: httpClientTimeout}, //nolint:exhaustruct // only Timeout needed
	}

	err := client.authenticate(username, password)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// NewClientWithToken returns a client pre-loaded with an existing auth token.
// Use when a token has already been obtained (e.g. at dashboard startup) to
// avoid a redundant authentication round-trip.
func NewClientWithToken(baseURL, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{Timeout: httpClientTimeout}, //nolint:exhaustruct // only Timeout needed
	}
}

// Token returns the current authentication token.
func (c *Client) Token() string {
	return c.token
}

// CollectionExists checks if a collection exists in PocketBase.
func (c *Client) CollectionExists(name string) (bool, error) {
	resp, err := c.doRequest(http.MethodGet, "/api/collections/"+name, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%w: status %d checking collection %s", ErrUnexpectedStatus, resp.StatusCode, name)
	}

	return true, nil
}

// CreateCollection creates a new collection with the given schema.
func (c *Client) CreateCollection(schema map[string]any) error {
	body, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	resp, err := c.doRequest(http.MethodPost, "/api/collections", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("%w: status %d, body: %s", ErrUnexpectedStatus, resp.StatusCode, respBody)
	}

	return nil
}

// DeleteCollection deletes a collection by name. It first fetches the collection to get its ID.
func (c *Client) DeleteCollection(name string) error {
	// First get the collection to find its ID.
	resp, err := c.doRequest(http.MethodGet, "/api/collections/"+name, nil)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}
	defer resp.Body.Close()

	var col struct {
		ID string `json:"id"`
	}

	err = json.NewDecoder(resp.Body).Decode(&col)
	if err != nil {
		return fmt.Errorf("failed to decode collection: %w", err)
	}

	// Delete by ID.
	deleteResp, err := c.doRequest(http.MethodDelete, "/api/collections/"+col.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusNoContent && deleteResp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: failed to delete collection %s, status %d", ErrUnexpectedStatus, name, deleteResp.StatusCode)
	}

	return nil
}

// FetchRecords fetches all records from a collection, handling pagination.
// Optional filter and sort parameters are passed as PB query params.
// If limit > 0, fetches at most that many records (single page).
func (c *Client) FetchRecords(collection, filter, sort string, limit ...int) ([]map[string]any, error) {
	var allRecords []map[string]any

	perPage := 500
	singlePage := false

	if len(limit) > 0 && limit[0] > 0 {
		perPage = limit[0]
		singlePage = true
	}

	for page := 1; ; page++ {
		path := fmt.Sprintf("/api/collections/%s/records?perPage=%d&page=%d", collection, perPage, page)
		if filter != "" {
			path += "&filter=" + url.QueryEscape(filter)
		}

		if sort != "" {
			path += "&sort=" + url.QueryEscape(sort)
		}

		resp, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch records page %d: %w", page, err)
		}

		var result struct {
			Page       int              `json:"page"`
			TotalPages int              `json:"totalPages"`
			Items      []map[string]any `json:"items"`
		}

		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			resp.Body.Close()

			return nil, fmt.Errorf("failed to decode records response: %w", err)
		}

		resp.Body.Close()

		allRecords = append(allRecords, result.Items...)

		if singlePage || result.Page >= result.TotalPages {
			break
		}
	}

	return allRecords, nil
}

// CreateRecord creates a record in the given collection.
func (c *Client) CreateRecord(collection string, data map[string]any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	resp, err := c.doRequest(http.MethodPost, "/api/collections/"+collection+"/records", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("%w: status %d creating record, body: %s", ErrUnexpectedStatus, resp.StatusCode, respBody)
	}

	return nil
}

// UpdateRecord updates a record by ID in the given collection.
func (c *Client) UpdateRecord(collection, recordID string, data map[string]any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal record update: %w", err)
	}

	path := fmt.Sprintf("/api/collections/%s/records/%s", collection, recordID)

	resp, err := c.doRequest(http.MethodPatch, path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf(
			"%w: status %d updating record %s, body: %s", ErrUnexpectedStatus, resp.StatusCode, recordID, respBody,
		)
	}

	return nil
}

// DeleteRecord deletes a record by ID from the given collection.
func (c *Client) DeleteRecord(collection, recordID string) error {
	path := fmt.Sprintf("/api/collections/%s/records/%s", collection, recordID)

	resp, err := c.doRequest(http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: failed to delete record %s, status %d", ErrUnexpectedStatus, recordID, resp.StatusCode)
	}

	return nil
}

func (c *Client) authenticate(username, password string) error {
	body, err := json.Marshal(map[string]string{
		"identity": username,
		"password": password,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	ctx := context.Background()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/collections/_superusers/auth-with-password",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w at %s. Is it running? Start with: pocketbase serve", ErrCannotConnect, c.baseURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w. Check POCKETBASE_USERNAME and POCKETBASE_PASSWORD", ErrAuthFailed)
	}

	var result struct {
		Token string `json:"token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.token = result.Token

	return nil
}

// doRequest sends an authenticated HTTP request to PocketBase.
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}
