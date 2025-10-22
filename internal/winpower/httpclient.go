package winpower

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// HTTPClient wraps an HTTP client with WinPower-specific configurations and behavior.
type HTTPClient struct {
	httpClient *http.Client
	config     *Config
}

// NewHTTPClient creates a new HTTP client with the given configuration.
func NewHTTPClient(config *Config) (*HTTPClient, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create transport with connection pooling and TLS settings
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   config.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SkipTLSVerify,
		},
	}

	// Create HTTP client
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &HTTPClient{
		httpClient: httpClient,
		config:     config,
	}, nil
}

// Get performs an HTTP GET request to the specified path.
func (c *HTTPClient) Get(ctx context.Context, path string) (*http.Response, error) {
	if c == nil {
		return nil, fmt.Errorf("HTTPClient is nil")
	}

	if c.httpClient == nil {
		return nil, fmt.Errorf("HTTPClient's httpClient is nil")
	}

	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	url := c.buildURL(path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "WinPower-G2-Exporter/1.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET request failed: %w", err)
	}

	return resp, nil
}

// Post performs an HTTP POST request to the specified path with the given body.
func (c *HTTPClient) Post(ctx context.Context, path, contentType string, body []byte) (*http.Response, error) {
	if c == nil {
		return nil, fmt.Errorf("HTTPClient is nil")
	}

	if c.httpClient == nil {
		return nil, fmt.Errorf("HTTPClient's httpClient is nil")
	}

	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	if contentType == "" {
		return nil, fmt.Errorf("content type cannot be empty")
	}

	if body == nil {
		return nil, fmt.Errorf("body cannot be nil")
	}

	url := c.buildURL(path)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "WinPower-G2-Exporter/1.0")
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST request failed: %w", err)
	}

	return resp, nil
}

// Close closes the HTTP client and cleans up resources.
func (c *HTTPClient) Close() {
	if c != nil && c.httpClient != nil {
		// Close idle connections if transport supports it
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
}

// buildURL constructs a full URL from the base config URL and the given path.
func (c *HTTPClient) buildURL(path string) string {
	if c == nil || c.config == nil {
		return path
	}

	// If path is empty, return the base URL as-is
	if path == "" {
		return c.config.URL
	}

	baseURL := strings.TrimSuffix(c.config.URL, "/")

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return baseURL + path
}

// GetConfig returns a copy of the client configuration.
func (c *HTTPClient) GetConfig() *Config {
	if c == nil || c.config == nil {
		return nil
	}
	configClone := c.config.Clone()
	clonedConfig, ok := configClone.(*Config)
	if !ok {
		// This should never happen, but return a new config if it does
		return DefaultConfig()
	}
	return clonedConfig
}

// GetTimeout returns the configured timeout.
func (c *HTTPClient) GetTimeout() time.Duration {
	if c == nil || c.config == nil {
		return 0
	}
	return c.config.Timeout
}

// GetMaxRetries returns the configured maximum retries.
func (c *HTTPClient) GetMaxRetries() int {
	if c == nil || c.config == nil {
		return 0
	}
	return c.config.MaxRetries
}

// IsTLSVerifySkipped returns true if TLS verification is disabled.
func (c *HTTPClient) IsTLSVerifySkipped() bool {
	if c == nil || c.config == nil {
		return false
	}
	return c.config.SkipTLSVerify
}
