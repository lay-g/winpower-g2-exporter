package winpower

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TokenInfo represents authentication token information
type TokenInfo struct {
	// Token is the authentication token string
	Token string

	// Expiry is the token expiration time
	Expiry time.Time

	// TokenType is the type of token (e.g., "Bearer")
	TokenType string
}

// IsExpired checks if the token is expired or will expire within the buffer time
func (t *TokenInfo) IsExpired() bool {
	if t == nil || t.Expiry.IsZero() {
		return true
	}

	// Consider token expired if it expires within 5 minutes
	bufferTime := 5 * time.Minute
	return time.Until(t.Expiry) <= bufferTime
}

// IsValid checks if the token is valid and not expired
func (t *TokenInfo) IsValid() bool {
	if t == nil {
		return false
	}

	return t.Token != "" && !t.IsExpired()
}

// TokenManager manages authentication tokens for WinPower API
type TokenManager struct {
	config *Config
	client *HTTPClient
	token  *TokenInfo
	mutex  sync.RWMutex
}

// NewTokenManager creates a new token manager
func NewTokenManager(config *Config, client *HTTPClient) (*TokenManager, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if client == nil {
		return nil, fmt.Errorf("HTTP client cannot be nil")
	}

	return &TokenManager{
		config: config,
		client: client,
		token:  nil,
	}, nil
}

// Login authenticates with the WinPower API and obtains a token
func (tm *TokenManager) Login(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	if ctx.Err() != nil {
		return fmt.Errorf("context error: %w", ctx.Err())
	}

	// For now, create a mock token
	// In a real implementation, this would make an HTTP request to the authentication endpoint
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Mock authentication - in real implementation, this would:
	// 1. Make HTTP POST request to login endpoint
	// 2. Send username and password
	// 3. Parse response to extract token and expiry
	// 4. Store token information

	mockToken := &TokenInfo{
		Token:     "mock-jwt-token-12345",
		TokenType: "Bearer",
		Expiry:    time.Now().Add(1 * time.Hour),
	}

	tm.token = mockToken
	return nil
}

// GetToken returns the current authentication token
func (tm *TokenManager) GetToken() (string, error) {
	if tm == nil {
		return "", fmt.Errorf("token manager is nil")
	}

	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if tm.token == nil || !tm.token.IsValid() {
		return "", fmt.Errorf("no valid token available")
	}

	return tm.token.Token, nil
}

// IsTokenValid checks if the current token is valid and not expired
func (tm *TokenManager) IsTokenValid() bool {
	if tm == nil {
		return false
	}

	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	return tm.token != nil && tm.token.IsValid()
}

// RefreshToken refreshes the current authentication token
func (tm *TokenManager) RefreshToken(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	if ctx.Err() != nil {
		return fmt.Errorf("context error: %w", ctx.Err())
	}

	// For now, just call Login to get a new token
	// In a real implementation, this might use a refresh token if available
	return tm.Login(ctx)
}

// GetTokenExpiry returns the expiration time of the current token
func (tm *TokenManager) GetTokenExpiry() time.Time {
	if tm == nil {
		return time.Time{}
	}

	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if tm.token == nil {
		return time.Time{}
	}

	return tm.token.Expiry
}

// GetTokenType returns the token type (e.g., "Bearer")
func (tm *TokenManager) GetTokenType() string {
	if tm == nil {
		return ""
	}

	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	if tm.token == nil {
		return ""
	}

	return tm.token.TokenType
}

// InvalidateToken manually invalidates the current token
func (tm *TokenManager) InvalidateToken() {
	if tm == nil {
		return
	}

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.token = nil
}

// GetAuthHeader returns the authorization header value for HTTP requests
func (tm *TokenManager) GetAuthHeader() (string, error) {
	token, err := tm.GetToken()
	if err != nil {
		return "", err
	}

	tokenType := tm.GetTokenType()
	if tokenType == "" {
		tokenType = "Bearer"
	}

	return fmt.Sprintf("%s %s", tokenType, token), nil
}

// SetToken manually sets the token (for testing purposes)
func (tm *TokenManager) SetToken(token *TokenInfo) {
	if tm == nil {
		return
	}

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.token = token
}

// TimeUntilExpiry returns the duration until the token expires
func (tm *TokenManager) TimeUntilExpiry() time.Duration {
	if tm == nil {
		return 0
	}

	expiry := tm.GetTokenExpiry()
	if expiry.IsZero() {
		return 0
	}

	return time.Until(expiry)
}

// WillExpireSoon checks if the token will expire within the specified duration
func (tm *TokenManager) WillExpireSoon(within time.Duration) bool {
	if tm == nil {
		return true
	}

	return tm.TimeUntilExpiry() <= within
}
