package winpower

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestClient creates a test client with a mock server
func setupTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server, func()) {
	t.Helper()

	// Create test server
	server := httptest.NewServer(handler)

	// Create test logger
	logger := log.NewTestLogger()

	// Create test config
	cfg := &Config{
		BaseURL:          server.URL,
		Username:         "testuser",
		Password:         "testpass",
		Timeout:          5 * time.Second,
		SkipSSLVerify:    true,
		RefreshThreshold: 5 * time.Minute,
		UserAgent:        "test-agent",
	}

	// Create client
	client, err := NewClient(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Return cleanup function
	cleanup := func() {
		_ = client.Close()
		server.Close()
	}

	return client, server, cleanup
}

// loadTestData loads test data from fixtures
func loadTestData(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile(fmt.Sprintf("fixtures/%s", filename))
	require.NoError(t, err, "failed to read test data file: %s", filename)
	return data
}

func TestNewClient(t *testing.T) {
	logger := log.NewTestLogger()

	tests := []struct {
		name    string
		config  *Config
		logger  log.Logger
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				BaseURL:          "https://example.com",
				Username:         "user",
				Password:         "pass",
				Timeout:          10 * time.Second,
				RefreshThreshold: 5 * time.Minute,
			},
			logger:  logger,
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			logger:  logger,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "nil logger",
			config: &Config{
				BaseURL:  "https://example.com",
				Username: "user",
				Password: "pass",
			},
			logger:  nil,
			wantErr: true,
			errMsg:  "logger cannot be nil",
		},
		{
			name: "invalid config - empty baseURL",
			config: &Config{
				BaseURL:  "",
				Username: "user",
				Password: "pass",
			},
			logger:  logger,
			wantErr: true,
			errMsg:  "invalid config",
		},
		{
			name: "invalid config - empty username",
			config: &Config{
				BaseURL:  "https://example.com",
				Username: "",
				Password: "pass",
			},
			logger:  logger,
			wantErr: true,
			errMsg:  "invalid config",
		},
		{
			name: "invalid config - empty password",
			config: &Config{
				BaseURL:  "https://example.com",
				Username: "user",
				Password: "",
			},
			logger:  logger,
			wantErr: true,
			errMsg:  "invalid config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config, tt.logger)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.httpClient)
				assert.NotNil(t, client.tokenManager)
				assert.NotNil(t, client.dataParser)
				assert.False(t, client.GetConnectionStatus())
				assert.Zero(t, client.GetLastCollectionTime())
			}
		})
	}
}

func TestClient_CollectDeviceData_Success(t *testing.T) {
	// Load test data
	deviceData := loadTestData(t, "device_data.json")

	// Create mock server
	loginCalled := false
	dataCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/auth/login":
			loginCalled = true
			w.Header().Set("Content-Type", "application/json")
			resp := LoginResponse{
				Code:    "000000",
				Message: "success",
			}
			resp.Data.Token = "test-token-123"
			resp.Data.DeviceID = "device-001"
			_ = json.NewEncoder(w).Encode(resp)

		case r.URL.Path == "/api/v1/deviceData/detail/list":
			dataCalled = true
			// Verify token
			auth := r.Header.Get("Authorization")
			assert.Equal(t, "Bearer test-token-123", auth)

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(deviceData)

		default:
			t.Errorf("unexpected request path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	client, _, cleanup := setupTestClient(t, handler)
	defer cleanup()

	// Collect device data
	ctx := context.Background()
	data, err := client.CollectDeviceData(ctx)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data, 1)

	// Verify API calls were made
	assert.True(t, loginCalled, "login endpoint should be called")
	assert.True(t, dataCalled, "data endpoint should be called")

	// Verify parsed data
	device := data[0]
	assert.Equal(t, "e156e6cb-41cb-4b35-b0dd-869929186a5c", device.DeviceID)
	assert.Equal(t, 1, device.DeviceType)
	assert.Equal(t, "ON-LINE", device.Model)
	assert.Equal(t, "C3K", device.Alias)
	assert.True(t, device.Connected)

	// Verify realtime data
	assert.Equal(t, 195.0, device.Realtime.LoadTotalWatt)
	assert.Equal(t, 236.8, device.Realtime.InputVolt1)
	assert.Equal(t, 220.1, device.Realtime.OutputVolt1)
	assert.Equal(t, 90.0, device.Realtime.BatCapacity)

	// Verify connection status
	assert.True(t, client.GetConnectionStatus())
	assert.NotZero(t, client.GetLastCollectionTime())

	// Verify statistics
	stats := client.GetStatistics()
	assert.Equal(t, int64(1), stats["collection_count"])
	assert.Equal(t, int64(1), stats["success_count"])
	assert.Equal(t, int64(0), stats["error_count"])
	assert.True(t, stats["connected"].(bool))
}

func TestClient_CollectDeviceData_AuthenticationFailure(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" {
			w.Header().Set("Content-Type", "application/json")
			resp := LoginResponse{
				Code:    "401001",
				Message: "authentication failed",
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	})

	client, _, cleanup := setupTestClient(t, handler)
	defer cleanup()

	// Attempt to collect device data
	ctx := context.Background()
	data, err := client.CollectDeviceData(ctx)

	// Verify error
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "authentication failed")

	// Verify connection status
	assert.False(t, client.GetConnectionStatus())

	// Verify statistics
	stats := client.GetStatistics()
	assert.Equal(t, int64(1), stats["collection_count"])
	assert.Equal(t, int64(0), stats["success_count"])
	assert.Equal(t, int64(1), stats["error_count"])
	assert.NotNil(t, stats["last_error"])
}

func TestClient_CollectDeviceData_NetworkError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/auth/login":
			w.Header().Set("Content-Type", "application/json")
			resp := LoginResponse{
				Code:    "000000",
				Message: "success",
			}
			resp.Data.Token = "test-token-123"
			resp.Data.DeviceID = "device-001"
			_ = json.NewEncoder(w).Encode(resp)

		case r.URL.Path == "/api/v1/deviceData/detail/list":
			// Simulate network error
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal server error"))

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	client, _, cleanup := setupTestClient(t, handler)
	defer cleanup()

	// Attempt to collect device data
	ctx := context.Background()
	data, err := client.CollectDeviceData(ctx)

	// Verify error
	assert.Error(t, err)
	assert.Nil(t, data)

	// Verify connection status
	assert.False(t, client.GetConnectionStatus())

	// Verify statistics
	stats := client.GetStatistics()
	assert.Equal(t, int64(1), stats["collection_count"])
	assert.Equal(t, int64(0), stats["success_count"])
	assert.Equal(t, int64(1), stats["error_count"])
}

func TestClient_CollectDeviceData_TokenCaching(t *testing.T) {
	deviceData := loadTestData(t, "device_data.json")

	loginCallCount := 0
	dataCallCount := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/auth/login":
			loginCallCount++
			w.Header().Set("Content-Type", "application/json")
			resp := LoginResponse{
				Code:    "000000",
				Message: "success",
			}
			resp.Data.Token = "test-token-123"
			resp.Data.DeviceID = "device-001"
			_ = json.NewEncoder(w).Encode(resp)

		case r.URL.Path == "/api/v1/deviceData/detail/list":
			dataCallCount++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(deviceData)

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	client, _, cleanup := setupTestClient(t, handler)
	defer cleanup()

	ctx := context.Background()

	// First collection
	data1, err := client.CollectDeviceData(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, data1)
	assert.Equal(t, 1, loginCallCount, "should login once")
	assert.Equal(t, 1, dataCallCount, "should fetch data once")

	// Second collection - token should be cached
	data2, err := client.CollectDeviceData(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, data2)
	assert.Equal(t, 1, loginCallCount, "should still be 1 (token cached)")
	assert.Equal(t, 2, dataCallCount, "should fetch data again")

	// Verify statistics
	stats := client.GetStatistics()
	assert.Equal(t, int64(2), stats["collection_count"])
	assert.Equal(t, int64(2), stats["success_count"])
	assert.Equal(t, int64(0), stats["error_count"])
}

func TestClient_CollectDeviceData_ContextCancellation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		resp := LoginResponse{
			Code:    "000000",
			Message: "success",
		}
		resp.Data.Token = "test-token-123"
		resp.Data.DeviceID = "device-001"
		_ = json.NewEncoder(w).Encode(resp)
	})

	client, _, cleanup := setupTestClient(t, handler)
	defer cleanup()

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Attempt to collect device data
	data, err := client.CollectDeviceData(ctx)

	// Verify error
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "context canceled")

	// Verify statistics
	stats := client.GetStatistics()
	assert.Equal(t, int64(1), stats["collection_count"])
	assert.Equal(t, int64(0), stats["success_count"])
	assert.Equal(t, int64(1), stats["error_count"])
}

func TestClient_CollectDeviceData_Concurrent(t *testing.T) {
	deviceData := loadTestData(t, "device_data.json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/auth/login":
			w.Header().Set("Content-Type", "application/json")
			resp := LoginResponse{
				Code:    "000000",
				Message: "success",
			}
			resp.Data.Token = "test-token-123"
			resp.Data.DeviceID = "device-001"
			_ = json.NewEncoder(w).Encode(resp)

		case r.URL.Path == "/api/v1/deviceData/detail/list":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(deviceData)

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	client, _, cleanup := setupTestClient(t, handler)
	defer cleanup()

	ctx := context.Background()

	// Run multiple concurrent collections
	const concurrency = 10
	errChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			data, err := client.CollectDeviceData(ctx)
			if err != nil {
				errChan <- err
			} else if len(data) != 1 {
				errChan <- errors.New("unexpected data count")
			} else {
				errChan <- nil
			}
		}()
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		err := <-errChan
		assert.NoError(t, err)
	}

	// Verify statistics
	stats := client.GetStatistics()
	assert.Equal(t, int64(concurrency), stats["collection_count"])
	assert.Equal(t, int64(concurrency), stats["success_count"])
	assert.Equal(t, int64(0), stats["error_count"])
}

func TestClient_GetConnectionStatus(t *testing.T) {
	logger := log.NewTestLogger()
	cfg := &Config{
		BaseURL:          "https://example.com",
		Username:         "user",
		Password:         "pass",
		Timeout:          10 * time.Second,
		RefreshThreshold: 5 * time.Minute,
	}

	client, err := NewClient(cfg, logger)
	require.NoError(t, err)

	// Initial status should be false
	assert.False(t, client.GetConnectionStatus())

	// Simulate successful collection
	client.recordSuccess(1)
	assert.True(t, client.GetConnectionStatus())

	// Simulate error
	client.recordError(errors.New("test error"))
	assert.False(t, client.GetConnectionStatus())
}

func TestClient_GetLastCollectionTime(t *testing.T) {
	logger := log.NewTestLogger()
	cfg := &Config{
		BaseURL:          "https://example.com",
		Username:         "user",
		Password:         "pass",
		Timeout:          10 * time.Second,
		RefreshThreshold: 5 * time.Minute,
	}

	client, err := NewClient(cfg, logger)
	require.NoError(t, err)

	// Initial time should be zero
	assert.Zero(t, client.GetLastCollectionTime())

	// Simulate successful collection
	before := time.Now()
	client.recordSuccess(1)
	after := time.Now()

	lastTime := client.GetLastCollectionTime()
	assert.True(t, lastTime.After(before) || lastTime.Equal(before))
	assert.True(t, lastTime.Before(after) || lastTime.Equal(after))
}

func TestClient_GetStatistics(t *testing.T) {
	logger := log.NewTestLogger()
	cfg := &Config{
		BaseURL:          "https://example.com",
		Username:         "user",
		Password:         "pass",
		Timeout:          10 * time.Second,
		RefreshThreshold: 5 * time.Minute,
	}

	client, err := NewClient(cfg, logger)
	require.NoError(t, err)

	// Initial statistics
	stats := client.GetStatistics()
	assert.Equal(t, int64(0), stats["collection_count"])
	assert.Equal(t, int64(0), stats["success_count"])
	assert.Equal(t, int64(0), stats["error_count"])
	assert.False(t, stats["connected"].(bool))

	// Simulate operations
	client.incrementCollectionCount()
	client.recordSuccess(1)

	client.incrementCollectionCount()
	client.recordError(errors.New("test error"))

	// Updated statistics
	stats = client.GetStatistics()
	assert.Equal(t, int64(2), stats["collection_count"])
	assert.Equal(t, int64(1), stats["success_count"])
	assert.Equal(t, int64(1), stats["error_count"])
	assert.False(t, stats["connected"].(bool)) // Last operation was error
	assert.NotNil(t, stats["last_error"])
}

func TestClient_Close(t *testing.T) {
	deviceData := loadTestData(t, "device_data.json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/auth/login":
			w.Header().Set("Content-Type", "application/json")
			resp := LoginResponse{
				Code:    "000000",
				Message: "success",
			}
			resp.Data.Token = "test-token-123"
			resp.Data.DeviceID = "device-001"
			_ = json.NewEncoder(w).Encode(resp)

		case r.URL.Path == "/api/v1/deviceData/detail/list":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(deviceData)
		}
	})

	client, server, _ := setupTestClient(t, handler)
	defer server.Close()

	// Perform a collection to cache token
	ctx := context.Background()
	_, err := client.CollectDeviceData(ctx)
	require.NoError(t, err)

	// Verify token is cached
	assert.True(t, client.tokenManager.IsValid())

	// Close client
	err = client.Close()
	assert.NoError(t, err)

	// Verify token cache is cleared
	assert.False(t, client.tokenManager.IsValid())
}

func TestClient_PerformanceBenchmark(t *testing.T) {
	deviceData := loadTestData(t, "device_data.json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/auth/login":
			w.Header().Set("Content-Type", "application/json")
			resp := LoginResponse{
				Code:    "000000",
				Message: "success",
			}
			resp.Data.Token = "test-token-123"
			resp.Data.DeviceID = "device-001"
			_ = json.NewEncoder(w).Encode(resp)

		case r.URL.Path == "/api/v1/deviceData/detail/list":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(deviceData)
		}
	})

	client, _, cleanup := setupTestClient(t, handler)
	defer cleanup()

	ctx := context.Background()

	// Measure collection time
	start := time.Now()
	data, err := client.CollectDeviceData(ctx)
	elapsed := time.Since(start)

	// Verify success
	assert.NoError(t, err)
	assert.NotNil(t, data)

	// Verify performance (should be < 2 seconds as per requirements)
	t.Logf("Collection completed in %v", elapsed)
	if elapsed > 2*time.Second {
		t.Errorf("Collection took too long: %v (expected < 2s)", elapsed)
	}
}
