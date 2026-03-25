// Package client provides a thin HTTP client for the agent registry v0 API.
// It is intentionally decoupled from the agentregistry module so this CLI
// can iterate independently.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client talks to the agent registry HTTP API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// New creates a client for the given registry server URL.
func New(server, token string) *Client {
	baseURL := strings.TrimRight(server, "/")
	if !strings.HasSuffix(baseURL, "/v0") {
		baseURL += "/v0"
	}
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{},
	}
}

// List fetches a list of resources at the given API path.
// Returns the raw JSON response as a map.
func (c *Client) List(apiPath string) (map[string]any, error) {
	return c.do(http.MethodGet, apiPath, nil)
}

// Get fetches a single resource by name.
func (c *Client) Get(apiPath, name string) (map[string]any, error) {
	path := fmt.Sprintf("%s/%s", apiPath, name)
	return c.do(http.MethodGet, path, nil)
}

// GetVersion fetches a specific version of a resource.
func (c *Client) GetVersion(apiPath, name, version string) (map[string]any, error) {
	path := fmt.Sprintf("%s/%s/versions/%s", apiPath, name, version)
	return c.do(http.MethodGet, path, nil)
}

// Create posts a new resource. The body is JSON-encoded.
func (c *Client) Create(apiPath string, body any) (map[string]any, error) {
	return c.do(http.MethodPost, apiPath, body)
}

// Delete removes a resource version.
func (c *Client) Delete(apiPath, name, version string) error {
	path := fmt.Sprintf("%s/%s/versions/%s", apiPath, name, version)
	_, err := c.do(http.MethodDelete, path, nil)
	return err
}

// Ping checks connectivity to the registry.
func (c *Client) Ping() error {
	url := strings.TrimSuffix(c.baseURL, "/v0") + "/ping"
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping returned %s", resp.Status)
	}
	return nil
}

func (c *Client) do(method, path string, body any) (map[string]any, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encoding request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseError(resp.StatusCode, respBody)
	}

	// For DELETE with no body, return nil map.
	if len(respBody) == 0 {
		return nil, nil
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response JSON: %w", err)
	}
	return result, nil
}

func parseError(status int, body []byte) error {
	// Try to parse Huma-style error response.
	var errResp struct {
		Title  string `json:"title"`
		Detail string `json:"detail"`
		Status int    `json:"status"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Title != "" {
		if errResp.Detail != "" {
			return fmt.Errorf("%s: %s (HTTP %d)", errResp.Title, errResp.Detail, status)
		}
		return fmt.Errorf("%s (HTTP %d)", errResp.Title, status)
	}
	return fmt.Errorf("HTTP %d: %s", status, string(body))
}
