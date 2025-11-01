package winpower

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"go.uber.org/zap"
)

// HTTPClient handles HTTP communication with WinPower system.
type HTTPClient struct {
	client    *http.Client
	baseURL   string
	userAgent string
	logger    log.Logger
}

// NewHTTPClient creates a new HTTP client with the given configuration.
func NewHTTPClient(cfg *Config, logger log.Logger) *HTTPClient {
	// Configure TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.SkipSSLVerify, //nolint:gosec // User-configurable for self-signed certs
	}

	// Create HTTP client with connection pooling
	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
		},
	}

	return &HTTPClient{
		client:    client,
		baseURL:   cfg.BaseURL,
		userAgent: cfg.UserAgent,
		logger:    logger,
	}
}

// Login authenticates with WinPower and returns the login response.
func (c *HTTPClient) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	endpoint := fmt.Sprintf("%s/api/v1/auth/login", c.baseURL)

	c.logger.Debug("attempting login",
		zap.String("endpoint", endpoint),
		zap.String("username", username),
	)

	var loginResp LoginResponse
	err := c.postJSON(ctx, endpoint, loginReq, &loginResp)
	if err != nil {
		return nil, &AuthenticationError{
			Message: "login failed",
			Err:     err,
		}
	}

	// Check response code
	if loginResp.Code != "000000" {
		c.logger.Warn("login failed with error code",
			zap.String("code", loginResp.Code),
			zap.String("message", loginResp.Message),
		)
		return nil, &AuthenticationError{
			Message: fmt.Sprintf("login failed: %s", loginResp.Message),
		}
	}

	c.logger.Info("login successful",
		zap.String("device_id", loginResp.Data.DeviceID),
	)

	return &loginResp, nil
}

// GetDeviceData retrieves device data from WinPower system.
func (c *HTTPClient) GetDeviceData(ctx context.Context, token string) (*DeviceDataResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/deviceData/detail/list", c.baseURL)

	// Build query parameters
	params := map[string]string{
		"current":        "1",
		"pageSize":       "100",
		"areaId":         "00000000-0000-0000-0000-000000000000",
		"includeSubArea": "true",
		"pageNum":        "1",
		"deviceType":     "1",
	}

	// Construct URL with query parameters
	fullURL := endpoint + "?"
	first := true
	for key, value := range params {
		if !first {
			fullURL += "&"
		}
		fullURL += fmt.Sprintf("%s=%s", key, value)
		first = false
	}

	c.logger.Debug("fetching device data",
		zap.String("endpoint", fullURL),
	)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, &NetworkError{
			Message: "failed to create request",
			Err:     err,
		}
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-language", "zh-CN")

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	var resp DeviceDataResponse
	err = c.doRequest(req, &resp)
	if err != nil {
		return nil, &NetworkError{
			Message: "failed to fetch device data",
			Err:     err,
		}
	}

	// Check response code
	if resp.Code != "000000" {
		c.logger.Warn("device data request failed with error code",
			zap.String("code", resp.Code),
			zap.String("message", resp.Msg),
		)
		return nil, &NetworkError{
			Message: fmt.Sprintf("device data request failed: %s", resp.Msg),
		}
	}

	c.logger.Debug("device data fetched successfully",
		zap.Int("total", resp.Total),
		zap.Int("count", len(resp.Data)),
	)

	return &resp, nil
}

// postJSON sends a POST request with JSON body and decodes JSON response.
func (c *HTTPClient) postJSON(ctx context.Context, url string, body interface{}, result interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-language", "en")

	return c.doRequest(req, result)
}

// doRequest executes the HTTP request and decodes the response.
// It intelligently handles both successful responses and error responses where
// the 'data' field might be a string instead of the expected type.
func (c *HTTPClient) doRequest(req *http.Request, result interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("HTTP request failed",
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
			zap.Error(err),
		)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warn("failed to close response body", zap.Error(closeErr))
		}
	}()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to read response body",
			zap.Int("status_code", resp.StatusCode),
			zap.Error(err),
		)
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Warn("HTTP request returned non-2xx status",
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", string(bodyBytes)),
		)

		// Try to parse as error response for better error message
		if resp.StatusCode == http.StatusUnauthorized {
			var errResp ErrorResponse
			if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.Code == "401" {
				c.logger.Warn("authentication failed",
					zap.String("code", errResp.Code),
					zap.String("message", errResp.Message),
					zap.String("data", errResp.Data),
				)
				return ErrAuthenticationFailed
			}
			return ErrAuthenticationFailed
		}

		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Try to parse as error response first to detect application-level errors
	// This handles cases where HTTP status is 200 but the application returns an error
	var errResp ErrorResponse
	if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.Code != "" && errResp.Code != "000000" {
		c.logger.Warn("API returned error response",
			zap.String("code", errResp.Code),
			zap.String("message", errResp.Message),
			zap.String("data", errResp.Data),
		)

		// Check if it's an authentication error (code 401)
		if errResp.Code == "401" {
			return ErrAuthenticationFailed
		}

		// Return generic error for other error codes
		return fmt.Errorf("API error (code %s): %s", errResp.Code, errResp.Message)
	}

	// Parse as successful response
	if err := json.Unmarshal(bodyBytes, result); err != nil {
		c.logger.Error("failed to decode JSON response",
			zap.String("response_body", string(bodyBytes)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil
}

// Close closes the HTTP client and releases resources.
func (c *HTTPClient) Close() error {
	if c.client != nil {
		c.client.CloseIdleConnections()
	}
	return nil
}
