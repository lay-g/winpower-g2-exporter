package winpower

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"go.uber.org/zap"
)

// Client implements the WinPowerClient interface.
type Client struct {
	config       *Config
	httpClient   *HTTPClient
	tokenManager *TokenManager
	dataParser   *DataParser
	logger       log.Logger

	// Connection state management
	mu                 sync.RWMutex
	connected          bool
	lastCollectionTime time.Time
	lastError          error
	collectionCount    int64
	successCount       int64
	errorCount         int64
}

// Ensure Client implements WinPowerClient interface
var _ WinPowerClient = (*Client)(nil)

// NewClient creates a new WinPower client with the given configuration.
func NewClient(cfg *Config, logger log.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Validate and apply defaults
	cfg = cfg.WithDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create HTTP client
	httpClient := NewHTTPClient(cfg, logger)

	// Create token manager
	tokenManager := NewTokenManager(
		httpClient,
		cfg.Username,
		cfg.Password,
		cfg.RefreshThreshold,
		logger,
	)

	// Create data parser
	// DataParser requires a *zap.Logger, so we get the underlying logger
	zapLogger := zap.NewNop()
	if logger.Core() != nil {
		zapLogger = zap.New(logger.Core())
	}
	dataParser := NewDataParser(zapLogger)

	client := &Client{
		config:       cfg,
		httpClient:   httpClient,
		tokenManager: tokenManager,
		dataParser:   dataParser,
		logger:       logger,
		connected:    false,
	}

	logger.Info("WinPower client created",
		zap.String("base_url", cfg.BaseURL),
		zap.String("username", cfg.Username),
		zap.Duration("timeout", cfg.Timeout),
		zap.Bool("skip_ssl_verify", cfg.SkipSSLVerify),
	)

	return client, nil
}

// CollectDeviceData collects device data from WinPower system.
// This is the main entry point for data collection.
func (c *Client) CollectDeviceData(ctx context.Context) ([]ParsedDeviceData, error) {
	startTime := time.Now()

	c.logger.Debug("starting device data collection")

	// Update collection count
	c.incrementCollectionCount()

	// Step 1: Check if we're healthy enough to proceed
	if !c.isHealthy() {
		c.logger.Warn("client not healthy, attempting to proceed anyway")
	}

	// Step 2: Get valid token
	token, err := c.tokenManager.GetToken(ctx)
	if err != nil {
		c.recordError(err)
		c.logger.Error("failed to get authentication token",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)),
		)
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	c.logger.Debug("authentication token obtained",
		zap.Duration("elapsed", time.Since(startTime)),
	)

	// Step 3: Fetch device data
	response, err := c.httpClient.GetDeviceData(ctx, token)
	if err != nil {
		c.recordError(err)
		c.logger.Error("failed to fetch device data",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)),
		)

		// If authentication failed, clear token cache to force re-login next time
		if IsAuthenticationError(err) {
			c.logger.Warn("authentication error detected, clearing token cache")
			c.tokenManager.ClearCache()
		}

		return nil, fmt.Errorf("data fetch failed: %w", err)
	}

	c.logger.Debug("device data fetched successfully",
		zap.Int("total", response.Total),
		zap.Duration("elapsed", time.Since(startTime)),
	)

	// Step 4: Parse response data
	data, err := c.dataParser.ParseResponse(response)
	if err != nil {
		c.recordError(err)
		c.logger.Error("failed to parse device data",
			zap.Error(err),
			zap.Int("raw_count", len(response.Data)),
			zap.Duration("elapsed", time.Since(startTime)),
		)
		return nil, fmt.Errorf("data parsing failed: %w", err)
	}

	// Step 5: Update collection status
	c.recordSuccess(len(data))

	elapsedTime := time.Since(startTime)
	c.logger.Info("device data collection completed",
		zap.Int("device_count", len(data)),
		zap.Duration("elapsed", elapsedTime),
		zap.Int64("total_collections", c.getCollectionCount()),
		zap.Int64("success_count", c.getSuccessCount()),
		zap.Int64("error_count", c.getErrorCount()),
	)

	// Log warning if collection took too long (> 2 seconds)
	if elapsedTime > 2*time.Second {
		c.logger.Warn("device data collection took longer than expected",
			zap.Duration("elapsed", elapsedTime),
			zap.Duration("threshold", 2*time.Second),
		)
	}

	return data, nil
}

// GetConnectionStatus returns the current connection status.
func (c *Client) GetConnectionStatus() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetLastCollectionTime returns the timestamp of the last successful collection.
func (c *Client) GetLastCollectionTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastCollectionTime
}

// GetLastError returns the last error encountered, if any.
func (c *Client) GetLastError() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastError
}

// GetStatistics returns collection statistics.
func (c *Client) GetStatistics() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"connected":            c.connected,
		"last_collection_time": c.lastCollectionTime,
		"collection_count":     c.collectionCount,
		"success_count":        c.successCount,
		"error_count":          c.errorCount,
		"last_error":           c.lastError,
		"token_valid":          c.tokenManager.IsValid(),
		"token_expires_at":     c.tokenManager.GetExpiresAt(),
	}
}

// Close closes the client and releases resources.
func (c *Client) Close() error {
	c.logger.Info("closing WinPower client")

	if err := c.httpClient.Close(); err != nil {
		c.logger.Warn("error closing HTTP client", zap.Error(err))
		return err
	}

	c.tokenManager.ClearCache()

	c.logger.Info("WinPower client closed")
	return nil
}

// Internal helper methods for state management

// isHealthy checks if the client is in a healthy state.
// Currently, we consider the client healthy if we have a valid token.
func (c *Client) isHealthy() bool {
	return c.tokenManager.IsValid()
}

// recordSuccess updates state after a successful collection.
func (c *Client) recordSuccess(deviceCount int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = true
	c.lastCollectionTime = time.Now()
	c.lastError = nil
	c.successCount++

	c.logger.Debug("collection success recorded",
		zap.Int("device_count", deviceCount),
		zap.Int64("total_success", c.successCount),
	)
}

// recordError updates state after a failed collection.
func (c *Client) recordError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false
	c.lastError = err
	c.errorCount++

	c.logger.Debug("collection error recorded",
		zap.Error(err),
		zap.Int64("total_errors", c.errorCount),
	)
}

// incrementCollectionCount increments the total collection count.
func (c *Client) incrementCollectionCount() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collectionCount++
}

// getCollectionCount returns the total collection count.
func (c *Client) getCollectionCount() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.collectionCount
}

// getSuccessCount returns the success count.
func (c *Client) getSuccessCount() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.successCount
}

// getErrorCount returns the error count.
func (c *Client) getErrorCount() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.errorCount
}

// GetTokenExpiresAt returns the expiration time of the current token.
func (c *Client) GetTokenExpiresAt() time.Time {
	return c.tokenManager.GetExpiresAt()
}

// IsTokenValid checks if the current token is valid.
func (c *Client) IsTokenValid() bool {
	return c.tokenManager.IsValid()
}
