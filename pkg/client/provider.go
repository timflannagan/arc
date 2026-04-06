package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// ListProviders fetches providers, optionally filtered by platform (aws or gcp).
func (c *Client) ListProviders(platform string) (map[string]any, error) {
	path := "/providers"
	if platform != "" {
		path += "?platform=" + url.QueryEscape(platform)
	}
	return c.do(http.MethodGet, path, nil)
}

// GetProvider fetches a single provider by ID.
func (c *Client) GetProvider(id string) (map[string]any, error) {
	path := fmt.Sprintf("/providers/%s", escapePathSegment(id))
	return c.do(http.MethodGet, path, nil)
}

// CreateProvider creates a new provider connection.
func (c *Client) CreateProvider(body any) (map[string]any, error) {
	return c.do(http.MethodPost, "/providers", body)
}

// UpdateProvider updates an existing provider connection by ID.
func (c *Client) UpdateProvider(id string, body any) (map[string]any, error) {
	path := fmt.Sprintf("/providers/%s", escapePathSegment(id))
	return c.do(http.MethodPut, path, body)
}

// DeleteProvider removes a provider connection by ID.
func (c *Client) DeleteProvider(id string) error {
	path := fmt.Sprintf("/providers/%s", escapePathSegment(id))
	_, err := c.do(http.MethodDelete, path, nil)
	return err
}

// PostSetup calls a platform setup endpoint (e.g., /v0/platforms/aws/setup).
func (c *Client) PostSetup(path string, body any) (map[string]any, error) {
	return c.do(http.MethodPost, path, body)
}
