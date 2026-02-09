package plane

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Client handles communication with the Plane.so API
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
	workspace  string
}

// ClientOption allows customizing the client
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithWorkspace sets the default workspace
func WithWorkspace(workspace string) ClientOption {
	return func(c *Client) {
		c.workspace = workspace
	}
}

// NewClient creates a new Plane API client
func NewClient(baseURL, apiToken string, options ...ClientOption) (*Client, error) {
	// Validate inputs
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if apiToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	// Parse and normalize base URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	client := &Client{
		baseURL:  parsedURL.String(),
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Apply options
	for _, opt := range options {
		opt(client)
	}

	return client, nil
}

// SetWorkspace sets the workspace for subsequent API calls
func (c *Client) SetWorkspace(workspace string) {
	c.workspace = workspace
}

// doRequest makes an HTTP request to the API
func (c *Client) doRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	// Build full URL
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	// Check if endpoint has trailing slash before path.Join strips it
	hasTrailingSlash := strings.HasSuffix(endpoint, "/")
	u.Path = path.Join(u.Path, endpoint)
	// Restore trailing slash if it was present
	if hasTrailingSlash && !strings.HasSuffix(u.Path, "/") {
		u.Path = u.Path + "/"
	}

	fmt.Printf("DEBUG: URL: %s\n", u.String())

	// Marshal body if provided
	var bodyReader io.Reader
	if body != nil {
		// Standard JSON marshaling
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
		fmt.Printf("DEBUG: Sending JSON (%d bytes): %s\n", len(jsonBody), string(jsonBody))
	}

	// Create request
	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("X-API-Key", c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Explicitly set Content-Length for bytes.Reader
	if bodyReader != nil {
		if br, ok := bodyReader.(*bytes.Reader); ok {
			req.ContentLength = int64(br.Len())
		}
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// get makes a GET request
func (c *Client) get(endpoint string, result interface{}) error {
	resp, err := c.doRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// getWithQuery makes a GET request with query parameters
func (c *Client) getWithQuery(endpoint string, query url.Values, result interface{}) error {
	// Build full URL
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, endpoint)
	u.RawQuery = query.Encode()

	// Create request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("X-API-Key", c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// post makes a POST request
func (c *Client) post(endpoint string, body, result interface{}) error {
	resp, err := c.doRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// patch makes a PATCH request
func (c *Client) patch(endpoint string, body, result interface{}) error {
	resp, err := c.doRequest(http.MethodPatch, endpoint, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// delete makes a DELETE request
func (c *Client) delete(endpoint string) error {
	resp, err := c.doRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
