package logseq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the Logseq HTTP API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL sets the API base URL (default: http://127.0.0.1:12315).
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithToken sets the Bearer token for authentication.
func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// NewClient creates a new Logseq API client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: "http://127.0.0.1:12315",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// apiRequest is the JSON-RPC request body.
type apiRequest struct {
	Method string `json:"method"`
	Args   []any  `json:"args"`
}

// apiError represents an error response from Logseq.
type apiError struct {
	Error string `json:"error"`
}

// CallAPI makes a raw API call and returns the response body as json.RawMessage.
func (c *Client) CallAPI(ctx context.Context, method string, args ...any) (json.RawMessage, error) {
	if args == nil {
		args = []any{}
	}
	reqBody := apiRequest{
		Method: method,
		Args:   args,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, friendlyConnectionError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("logseq api: authentication failed (HTTP %d) — check your API token", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error != "" {
			return nil, fmt.Errorf("logseq api: %s", apiErr.Error)
		}
		return nil, fmt.Errorf("logseq api: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Check for error in response body
	var apiErr apiError
	if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error != "" {
		return nil, fmt.Errorf("logseq api: %s", apiErr.Error)
	}

	return json.RawMessage(respBody), nil
}

// decode is a helper to unmarshal a json.RawMessage into a typed result.
func decode[T any](raw json.RawMessage, err error) (T, error) {
	var zero T
	if err != nil {
		return zero, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return zero, nil
	}
	var result T
	if err := json.Unmarshal(raw, &result); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}

// friendlyConnectionError wraps low-level connection errors with user-friendly messages.
func friendlyConnectionError(err error) error {
	msg := err.Error()
	if strings.Contains(msg, "connection refused") {
		return fmt.Errorf("cannot connect to Logseq — is it running with the HTTP API server enabled? (%w)", err)
	}
	if strings.Contains(msg, "no such host") || strings.Contains(msg, "dial tcp") {
		return fmt.Errorf("cannot reach Logseq API server — check the host and port configuration (%w)", err)
	}
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded") {
		return fmt.Errorf("connection to Logseq timed out — the server may be busy or unreachable (%w)", err)
	}
	return fmt.Errorf("failed to connect to Logseq: %w", err)
}
