package winpower

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

func TestNewHTTPClient(t *testing.T) {
	logger := log.NewTestLogger()
	cfg := DefaultConfig()
	cfg.BaseURL = "https://example.com"
	cfg.Username = "admin"
	cfg.Password = "secret"

	client := NewHTTPClient(cfg, logger)

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.baseURL != cfg.BaseURL {
		t.Errorf("expected baseURL %q, got %q", cfg.BaseURL, client.baseURL)
	}

	if client.userAgent != cfg.UserAgent {
		t.Errorf("expected userAgent %q, got %q", cfg.UserAgent, client.userAgent)
	}

	if client.client == nil {
		t.Error("expected non-nil http.Client")
	}
}

func TestHTTPClient_Login_Success(t *testing.T) {
	logger := log.NewTestLogger()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/auth/login" {
			t.Errorf("expected /api/v1/auth/login, got %s", r.URL.Path)
		}

		// Verify headers
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}

		// Verify request body
		var loginReq LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if loginReq.Username != "admin" {
			t.Errorf("expected username 'admin', got %q", loginReq.Username)
		}
		if loginReq.Password != "secret" {
			t.Errorf("expected password 'secret', got %q", loginReq.Password)
		}

		// Send success response
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

	client := NewHTTPClient(cfg, logger)
	ctx := context.Background()

	resp, err := client.Login(ctx, "admin", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Code != "000000" {
		t.Errorf("expected code '000000', got %q", resp.Code)
	}

	if resp.Data.Token != "test-token" {
		t.Errorf("expected token 'test-token', got %q", resp.Data.Token)
	}

	if resp.Data.DeviceID != "device-123" {
		t.Errorf("expected deviceID 'device-123', got %q", resp.Data.DeviceID)
	}
}

func TestHTTPClient_Login_Failure(t *testing.T) {
	logger := log.NewTestLogger()

	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := LoginResponse{
			Code:    "401",
			Message: "Invalid credentials",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "wrong"

	client := NewHTTPClient(cfg, logger)
	ctx := context.Background()

	_, err := client.Login(ctx, "admin", "wrong")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestHTTPClient_Login_Unauthorized(t *testing.T) {
	logger := log.NewTestLogger()

	// Create mock server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "secret"

	client := NewHTTPClient(cfg, logger)
	ctx := context.Background()

	_, err := client.Login(ctx, "admin", "secret")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestHTTPClient_GetDeviceData_Success(t *testing.T) {
	logger := log.NewTestLogger()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/deviceData/detail/list" {
			t.Errorf("expected /api/v1/deviceData/detail/list, got %s", r.URL.Path)
		}

		// Verify Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("expected 'Bearer test-token', got %q", authHeader)
		}

		// Verify query parameters
		if r.URL.Query().Get("current") != "1" {
			t.Errorf("expected current=1")
		}
		if r.URL.Query().Get("pageSize") != "100" {
			t.Errorf("expected pageSize=100")
		}

		// Send success response
		resp := DeviceDataResponse{
			Total:       1,
			PageSize:    100,
			CurrentPage: 1,
			Code:        "000000",
			Msg:         "OK",
			Data: []DeviceInfo{
				{
					AssetDevice: AssetDevice{
						ID:         "device-1",
						DeviceType: 1,
						Model:      "UPS-3000",
						Alias:      "Test UPS",
					},
					Connected: true,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "secret"

	client := NewHTTPClient(cfg, logger)
	ctx := context.Background()

	resp, err := client.GetDeviceData(ctx, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Code != "000000" {
		t.Errorf("expected code '000000', got %q", resp.Code)
	}

	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}

	if len(resp.Data) != 1 {
		t.Errorf("expected 1 device, got %d", len(resp.Data))
	}

	if resp.Data[0].AssetDevice.ID != "device-1" {
		t.Errorf("expected device ID 'device-1', got %q", resp.Data[0].AssetDevice.ID)
	}
}

func TestHTTPClient_GetDeviceData_Failure(t *testing.T) {
	logger := log.NewTestLogger()

	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := DeviceDataResponse{
			Code: "500",
			Msg:  "Internal Server Error",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "secret"

	client := NewHTTPClient(cfg, logger)
	ctx := context.Background()

	_, err := client.GetDeviceData(ctx, "test-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !IsNetworkError(err) {
		t.Errorf("expected NetworkError, got %T", err)
	}
}

func TestHTTPClient_Timeout(t *testing.T) {
	logger := log.NewTestLogger()

	// Create mock server that sleeps longer than timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "secret"
	cfg.Timeout = 100 * time.Millisecond // Very short timeout

	client := NewHTTPClient(cfg, logger)
	ctx := context.Background()

	_, err := client.Login(ctx, "admin", "secret")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestHTTPClient_ContextCancellation(t *testing.T) {
	logger := log.NewTestLogger()

	// Create mock server that sleeps
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "secret"

	client := NewHTTPClient(cfg, logger)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	_, err := client.Login(ctx, "admin", "secret")
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

func TestHTTPClient_SSLVerification(t *testing.T) {
	logger := log.NewTestLogger()

	// Create HTTPS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := LoginResponse{
			Code:    "000000",
			Message: "OK",
		}
		resp.Data.Token = "test-token"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Run("skip SSL verification", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.BaseURL = server.URL
		cfg.Username = "admin"
		cfg.Password = "secret"
		cfg.SkipSSLVerify = true // Skip SSL verification

		client := NewHTTPClient(cfg, logger)
		ctx := context.Background()

		_, err := client.Login(ctx, "admin", "secret")
		if err != nil {
			t.Fatalf("unexpected error with SkipSSLVerify=true: %v", err)
		}
	})
}

func TestHTTPClient_Close(t *testing.T) {
	logger := log.NewTestLogger()
	cfg := DefaultConfig()
	cfg.BaseURL = "https://example.com"
	cfg.Username = "admin"
	cfg.Password = "secret"

	client := NewHTTPClient(cfg, logger)

	err := client.Close()
	if err != nil {
		t.Errorf("unexpected error closing client: %v", err)
	}
}

func TestHTTPClient_InvalidJSON(t *testing.T) {
	logger := log.NewTestLogger()

	// Create mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Username = "admin"
	cfg.Password = "secret"

	client := NewHTTPClient(cfg, logger)
	ctx := context.Background()

	_, err := client.Login(ctx, "admin", "secret")
	if err == nil {
		t.Fatal("expected JSON decode error, got nil")
	}
}
