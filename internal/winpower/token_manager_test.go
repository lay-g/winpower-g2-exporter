package winpower

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

func TestNewTokenManager(t *testing.T) {
	logger := log.NewTestLogger()
	httpClient := &HTTPClient{}

	tm := NewTokenManager(httpClient, "admin", "secret", 5*time.Minute, logger)

	if tm == nil {
		t.Fatal("expected non-nil token manager")
	}

	if tm.username != "admin" {
		t.Errorf("expected username 'admin', got %q", tm.username)
	}

	if tm.password != "secret" {
		t.Errorf("expected password 'secret', got %q", tm.password)
	}

	if tm.refreshThreshold != 5*time.Minute {
		t.Errorf("expected refresh threshold 5m, got %v", tm.refreshThreshold)
	}
}

func TestTokenManager_GetToken_FirstLogin(t *testing.T) {
	logger := log.NewTestLogger()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := LoginResponse{
			Code:    "000000",
			Message: "OK",
		}
		resp.Data.DeviceID = "device-123"
		resp.Data.Token = "test-token"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "secret"

	httpClient := NewHTTPClient(cfg, logger)
	tm := NewTokenManager(httpClient, "admin", "secret", 5*time.Minute, logger)

	ctx := context.Background()
	token, err := tm.GetToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "test-token" {
		t.Errorf("expected token 'test-token', got %q", token)
	}

	// Verify cache is populated
	if tm.GetCachedToken() != "test-token" {
		t.Error("token should be cached")
	}

	if tm.GetDeviceID() != "device-123" {
		t.Errorf("expected device ID 'device-123', got %q", tm.GetDeviceID())
	}

	if !tm.IsValid() {
		t.Error("token should be valid")
	}
}

func TestTokenManager_GetToken_UseCached(t *testing.T) {
	logger := log.NewTestLogger()

	callCount := 0
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := LoginResponse{
			Code:    "000000",
			Message: "OK",
		}
		resp.Data.DeviceID = "device-123"
		resp.Data.Token = "test-token"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "secret"

	httpClient := NewHTTPClient(cfg, logger)
	tm := NewTokenManager(httpClient, "admin", "secret", 5*time.Minute, logger)

	ctx := context.Background()

	// First call - should login
	token1, err := tm.GetToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 login call, got %d", callCount)
	}

	// Second call - should use cache
	token2, err := tm.GetToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected still 1 login call (cached), got %d", callCount)
	}

	if token1 != token2 {
		t.Error("tokens should be the same (from cache)")
	}
}

func TestTokenManager_GetToken_AutoRefresh(t *testing.T) {
	logger := log.NewTestLogger()

	callCount := 0
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := LoginResponse{
			Code:    "000000",
			Message: "OK",
		}
		resp.Data.DeviceID = "device-123"
		resp.Data.Token = "test-token-" + string(rune('0'+callCount))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "secret"

	httpClient := NewHTTPClient(cfg, logger)
	// Set a very short refresh threshold for testing
	tm := NewTokenManager(httpClient, "admin", "secret", 2*time.Hour, logger)

	ctx := context.Background()

	// First call - should login
	token1, err := tm.GetToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token1 != "test-token-1" {
		t.Errorf("expected token 'test-token-1', got %q", token1)
	}

	// Manually set cache to be near expiry
	tm.mu.Lock()
	tm.cache.ExpiresAt = time.Now().Add(1 * time.Minute) // Within refresh threshold
	tm.mu.Unlock()

	// Second call - should refresh
	token2, err := tm.GetToken(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token2 != "test-token-2" {
		t.Errorf("expected token 'test-token-2', got %q", token2)
	}

	if callCount != 2 {
		t.Errorf("expected 2 login calls, got %d", callCount)
	}
}

func TestTokenManager_GetToken_LoginFailure(t *testing.T) {
	logger := log.NewTestLogger()

	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "wrong"

	httpClient := NewHTTPClient(cfg, logger)
	tm := NewTokenManager(httpClient, "admin", "wrong", 5*time.Minute, logger)

	ctx := context.Background()
	_, err := tm.GetToken(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}

	// Verify cache is not populated
	if tm.GetCachedToken() != "" {
		t.Error("token should not be cached on failure")
	}

	if tm.IsValid() {
		t.Error("token should not be valid on failure")
	}
}

func TestTokenManager_ConcurrentAccess(t *testing.T) {
	logger := log.NewTestLogger()

	var callCount int32
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		// Add a small delay to simulate network latency
		time.Sleep(50 * time.Millisecond)

		resp := LoginResponse{
			Code:    "000000",
			Message: "OK",
		}
		resp.Data.DeviceID = "device-123"
		resp.Data.Token = "test-token"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "secret"

	httpClient := NewHTTPClient(cfg, logger)
	tm := NewTokenManager(httpClient, "admin", "secret", 5*time.Minute, logger)

	ctx := context.Background()

	// Launch multiple goroutines requesting token simultaneously
	const numGoroutines = 10
	var wg sync.WaitGroup
	tokens := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			token, err := tm.GetToken(ctx)
			tokens[idx] = token
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	// All goroutines should succeed
	for i, err := range errors {
		if err != nil {
			t.Errorf("goroutine %d failed: %v", i, err)
		}
	}

	// All tokens should be the same
	firstToken := tokens[0]
	for i, token := range tokens {
		if token != firstToken {
			t.Errorf("goroutine %d got different token: %q vs %q", i, token, firstToken)
		}
	}

	// Should only call login once (other goroutines wait or use cache)
	finalCallCount := atomic.LoadInt32(&callCount)
	if finalCallCount != 1 {
		t.Errorf("expected 1 login call, got %d", finalCallCount)
	}
}

func TestTokenManager_GetCachedToken(t *testing.T) {
	logger := log.NewTestLogger()
	httpClient := &HTTPClient{}
	tm := NewTokenManager(httpClient, "admin", "secret", 5*time.Minute, logger)

	// Initially no cache
	if token := tm.GetCachedToken(); token != "" {
		t.Errorf("expected empty token, got %q", token)
	}

	// Set cache manually
	tm.mu.Lock()
	tm.cache = &TokenCache{
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		DeviceID:  "device-123",
	}
	tm.mu.Unlock()

	// Should return cached token
	if token := tm.GetCachedToken(); token != "test-token" {
		t.Errorf("expected 'test-token', got %q", token)
	}
}

func TestTokenManager_GetExpiresAt(t *testing.T) {
	logger := log.NewTestLogger()
	httpClient := &HTTPClient{}
	tm := NewTokenManager(httpClient, "admin", "secret", 5*time.Minute, logger)

	// Initially no cache
	if !tm.GetExpiresAt().IsZero() {
		t.Error("expected zero time for no cache")
	}

	// Set cache manually
	expiresAt := time.Now().Add(1 * time.Hour)
	tm.mu.Lock()
	tm.cache = &TokenCache{
		Token:     "test-token",
		ExpiresAt: expiresAt,
		DeviceID:  "device-123",
	}
	tm.mu.Unlock()

	// Should return expiry time
	if got := tm.GetExpiresAt(); !got.Equal(expiresAt) {
		t.Errorf("expected %v, got %v", expiresAt, got)
	}
}

func TestTokenManager_GetDeviceID(t *testing.T) {
	logger := log.NewTestLogger()
	httpClient := &HTTPClient{}
	tm := NewTokenManager(httpClient, "admin", "secret", 5*time.Minute, logger)

	// Initially no cache
	if deviceID := tm.GetDeviceID(); deviceID != "" {
		t.Errorf("expected empty device ID, got %q", deviceID)
	}

	// Set cache manually
	tm.mu.Lock()
	tm.cache = &TokenCache{
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		DeviceID:  "device-123",
	}
	tm.mu.Unlock()

	// Should return device ID
	if deviceID := tm.GetDeviceID(); deviceID != "device-123" {
		t.Errorf("expected 'device-123', got %q", deviceID)
	}
}

func TestTokenManager_IsValid(t *testing.T) {
	logger := log.NewTestLogger()
	httpClient := &HTTPClient{}
	tm := NewTokenManager(httpClient, "admin", "secret", 5*time.Minute, logger)

	// Initially no cache
	if tm.IsValid() {
		t.Error("expected invalid for no cache")
	}

	// Set valid cache
	tm.mu.Lock()
	tm.cache = &TokenCache{
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		DeviceID:  "device-123",
	}
	tm.mu.Unlock()

	if !tm.IsValid() {
		t.Error("expected valid token")
	}

	// Set expired cache
	tm.mu.Lock()
	tm.cache.ExpiresAt = time.Now().Add(-1 * time.Hour)
	tm.mu.Unlock()

	if tm.IsValid() {
		t.Error("expected invalid for expired token")
	}
}

func TestTokenManager_ClearCache(t *testing.T) {
	logger := log.NewTestLogger()
	httpClient := &HTTPClient{}
	tm := NewTokenManager(httpClient, "admin", "secret", 5*time.Minute, logger)

	// Set cache manually
	tm.mu.Lock()
	tm.cache = &TokenCache{
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		DeviceID:  "device-123",
	}
	tm.mu.Unlock()

	// Verify cache exists
	if tm.GetCachedToken() == "" {
		t.Error("cache should exist before clear")
	}

	// Clear cache
	tm.ClearCache()

	// Verify cache is cleared
	if tm.GetCachedToken() != "" {
		t.Error("cache should be cleared")
	}

	if tm.IsValid() {
		t.Error("token should not be valid after clear")
	}
}
