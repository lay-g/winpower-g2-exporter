package winpower

import (
	"context"
	"sync"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"go.uber.org/zap"
)

const (
	// tokenExpiry is the fixed token expiry duration (1 hour as per protocol)
	tokenExpiry = time.Hour
)

// TokenManager manages authentication tokens for WinPower system.
type TokenManager struct {
	httpClient       *HTTPClient
	username         string
	password         string
	refreshThreshold time.Duration
	logger           log.Logger

	mu    sync.RWMutex
	cache *TokenCache
}

// NewTokenManager creates a new token manager.
func NewTokenManager(httpClient *HTTPClient, username, password string, refreshThreshold time.Duration, logger log.Logger) *TokenManager {
	return &TokenManager{
		httpClient:       httpClient,
		username:         username,
		password:         password,
		refreshThreshold: refreshThreshold,
		logger:           logger,
	}
}

// GetToken returns a valid token, refreshing if necessary.
// This method is thread-safe and ensures only one login happens at a time.
func (tm *TokenManager) GetToken(ctx context.Context) (string, error) {
	// Fast path: check if we have a valid cached token (read lock)
	tm.mu.RLock()
	if tm.cache != nil && !tm.shouldRefresh() {
		token := tm.cache.Token
		tm.mu.RUnlock()

		tm.logger.Debug("using cached token",
			zap.Time("expires_at", tm.cache.ExpiresAt),
			zap.Duration("remaining", time.Until(tm.cache.ExpiresAt)),
		)

		return token, nil
	}
	tm.mu.RUnlock()

	// Slow path: need to refresh token (write lock)
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have refreshed)
	if tm.cache != nil && !tm.shouldRefresh() {
		token := tm.cache.Token

		tm.logger.Debug("using token refreshed by another goroutine",
			zap.Time("expires_at", tm.cache.ExpiresAt),
		)

		return token, nil
	}

	// Perform login
	tm.logger.Info("refreshing token",
		zap.String("username", tm.username),
		zap.Bool("has_cache", tm.cache != nil),
	)

	loginResp, err := tm.httpClient.Login(ctx, tm.username, tm.password)
	if err != nil {
		tm.logger.Error("failed to refresh token",
			zap.Error(err),
		)
		return "", err
	}

	// Cache the new token
	now := time.Now()
	tm.cache = &TokenCache{
		Token:     loginResp.Data.Token,
		ExpiresAt: now.Add(tokenExpiry),
		DeviceID:  loginResp.Data.DeviceID,
	}

	tm.logger.Info("token refreshed successfully",
		zap.String("device_id", loginResp.Data.DeviceID),
		zap.Time("expires_at", tm.cache.ExpiresAt),
		zap.Duration("valid_for", tokenExpiry),
	)

	return tm.cache.Token, nil
}

// shouldRefresh checks if the token should be refreshed.
// Must be called with at least a read lock held.
func (tm *TokenManager) shouldRefresh() bool {
	if tm.cache == nil {
		return true
	}

	// Refresh if we're within the threshold of expiry
	timeUntilExpiry := time.Until(tm.cache.ExpiresAt)
	shouldRefresh := timeUntilExpiry <= tm.refreshThreshold

	if shouldRefresh {
		tm.logger.Debug("token refresh needed",
			zap.Duration("time_until_expiry", timeUntilExpiry),
			zap.Duration("refresh_threshold", tm.refreshThreshold),
		)
	}

	return shouldRefresh
}

// GetCachedToken returns the cached token if available, without refreshing.
// Returns empty string if no token is cached.
// This method is thread-safe.
func (tm *TokenManager) GetCachedToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.cache == nil {
		return ""
	}

	return tm.cache.Token
}

// GetExpiresAt returns the expiration time of the cached token.
// Returns zero time if no token is cached.
// This method is thread-safe.
func (tm *TokenManager) GetExpiresAt() time.Time {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.cache == nil {
		return time.Time{}
	}

	return tm.cache.ExpiresAt
}

// GetDeviceID returns the device ID from the cached token.
// Returns empty string if no token is cached.
// This method is thread-safe.
func (tm *TokenManager) GetDeviceID() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.cache == nil {
		return ""
	}

	return tm.cache.DeviceID
}

// IsValid checks if the cached token is valid (not expired).
// This method is thread-safe.
func (tm *TokenManager) IsValid() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.cache == nil {
		return false
	}

	return time.Now().Before(tm.cache.ExpiresAt)
}

// ClearCache clears the cached token.
// This method is thread-safe.
func (tm *TokenManager) ClearCache() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.cache != nil {
		tm.logger.Info("clearing token cache")
		tm.cache = nil
	}
}
