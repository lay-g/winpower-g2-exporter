package winpower

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"
	"time"
)

// TestNewHTTPClient tests the NewHTTPClient constructor
func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "nil_config_should_error",
			test: func(t *testing.T) {
				client, err := NewHTTPClient(nil)
				if err == nil {
					t.Fatal("NewHTTPClient() should return error for nil config")
				}
				if client != nil {
					t.Error("NewHTTPClient() should return nil client when config is nil")
				}
			},
		},
		{
			name: "invalid_config_should_error",
			test: func(t *testing.T) {
				config := &Config{
					URL:     "invalid-url",
					Timeout: 10 * time.Second,
				}
				client, err := NewHTTPClient(config)
				if err == nil {
					t.Fatal("NewHTTPClient() should return error for invalid config")
				}
				if client != nil {
					t.Error("NewHTTPClient() should return nil client when config is invalid")
				}
			},
		},
		{
			name: "valid_config_should_succeed",
			test: func(t *testing.T) {
				config := DefaultConfig().WithURL("https://example.com")
				client, err := NewHTTPClient(config)
				if err != nil {
					t.Fatalf("NewHTTPClient() should not return error for valid config: %v", err)
				}
				if client == nil {
					t.Fatal("NewHTTPClient() should return non-nil client for valid config")
				}
			},
		},
		{
			name: "should_create_http_client_with_correct_timeout",
			test: func(t *testing.T) {
				config := DefaultConfig().WithURL("https://example.com").WithTimeout(5 * time.Second)
				client, err := NewHTTPClient(config)
				if err != nil {
					t.Fatalf("NewHTTPClient() failed: %v", err)
				}

				if client.httpClient == nil {
					t.Fatal("HTTPClient should have an underlying httpClient")
				}

				if client.httpClient.Timeout != 5*time.Second {
					t.Errorf("Expected timeout %v, got %v", 5*time.Second, client.httpClient.Timeout)
				}
			},
		},
		{
			name: "should_create_http_client_with_tls_settings",
			test: func(t *testing.T) {
				config := DefaultConfig().WithURL("https://example.com").WithSkipTLSVerify(true)
				client, err := NewHTTPClient(config)
				if err != nil {
					t.Fatalf("NewHTTPClient() failed: %v", err)
				}

				if client.httpClient.Transport == nil {
					t.Fatal("HTTPClient should have a transport configured")
				}

				transport, ok := client.httpClient.Transport.(*http.Transport)
				if !ok {
					t.Fatal("Transport should be of type *http.Transport")
				}

				if transport.TLSClientConfig.InsecureSkipVerify != true {
					t.Error("TLSClientConfig should have InsecureSkipVerify set to true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestHTTPClient_Get tests the Get method
func TestHTTPClient_Get(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "get_with_invalid_url_should_error",
			test: func(t *testing.T) {
				client := &HTTPClient{}
				resp, err := client.Get(context.Background(), "invalid-url")
				if err == nil {
					t.Error("Get() should return error for invalid URL")
				}
				if resp != nil {
					t.Error("Get() should return nil response for invalid URL")
				}
			},
		},
		{
			name: "get_with_cancelled_context_should_error",
			test: func(t *testing.T) {
				client := &HTTPClient{}
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Immediately cancel the context

				resp, err := client.Get(ctx, "https://example.com")
				if err == nil {
					t.Error("Get() should return error for cancelled context")
				}
				if resp != nil {
					t.Error("Get() should return nil response for cancelled context")
				}
			},
		},
		{
			name: "get_with_empty_path_should_error",
			test: func(t *testing.T) {
				client := &HTTPClient{}
				resp, err := client.Get(context.Background(), "")
				if err == nil {
					t.Error("Get() should return error for empty path")
				}
				if resp != nil {
					t.Error("Get() should return nil response for empty path")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestHTTPClient_Post tests the Post method
func TestHTTPClient_Post(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "post_with_invalid_url_should_error",
			test: func(t *testing.T) {
				client := &HTTPClient{}
				resp, err := client.Post(context.Background(), "invalid-url", "application/json", []byte("{}"))
				if err == nil {
					t.Error("Post() should return error for invalid URL")
				}
				if resp != nil {
					t.Error("Post() should return nil response for invalid URL")
				}
			},
		},
		{
			name: "post_with_cancelled_context_should_error",
			test: func(t *testing.T) {
				client := &HTTPClient{}
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Immediately cancel the context

				resp, err := client.Post(ctx, "https://example.com", "application/json", []byte("{}"))
				if err == nil {
					t.Error("Post() should return error for cancelled context")
				}
				if resp != nil {
					t.Error("Post() should return nil response for cancelled context")
				}
			},
		},
		{
			name: "post_with_empty_path_should_error",
			test: func(t *testing.T) {
				client := &HTTPClient{}
				resp, err := client.Post(context.Background(), "", "application/json", []byte("{}"))
				if err == nil {
					t.Error("Post() should return error for empty path")
				}
				if resp != nil {
					t.Error("Post() should return nil response for empty path")
				}
			},
		},
		{
			name: "post_with_nil_body_should_error",
			test: func(t *testing.T) {
				client := &HTTPClient{}
				resp, err := client.Post(context.Background(), "https://example.com", "application/json", nil)
				if err == nil {
					t.Error("Post() should return error for nil body")
				}
				if resp != nil {
					t.Error("Post() should return nil response for nil body")
				}
			},
		},
		{
			name: "post_with_empty_content_type_should_error",
			test: func(t *testing.T) {
				client := &HTTPClient{}
				resp, err := client.Post(context.Background(), "https://example.com", "", []byte("{}"))
				if err == nil {
					t.Error("Post() should return error for empty content type")
				}
				if resp != nil {
					t.Error("Post() should return nil response for empty content type")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestHTTPClient_Close tests the Close method
func TestHTTPClient_Close(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "close_nil_client_should_not_panic",
			test: func(t *testing.T) {
				var client *HTTPClient
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Close() should not panic for nil client, got: %v", r)
					}
				}()
				client.Close()
			},
		},
		{
			name: "close_client_with_nil_http_client_should_not_panic",
			test: func(t *testing.T) {
				client := &HTTPClient{httpClient: nil}
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Close() should not panic for client with nil httpClient, got: %v", r)
					}
				}()
				client.Close()
			},
		},
		{
			name: "close_client_with_transport_should_close_idle_connections",
			test: func(t *testing.T) {
				transport := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				httpClient := &http.Client{
					Transport: transport,
					Timeout:   10 * time.Second,
				}
				client := &HTTPClient{httpClient: httpClient}

				// This should not panic
				client.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestHTTPClient_buildURL tests the buildURL method
func TestHTTPClient_buildURL(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "build_url_with_empty_path_should_return_base_url",
			test: func(t *testing.T) {
				client := &HTTPClient{
					config: &Config{URL: "https://example.com"},
				}
				url := client.buildURL("")
				if url != "https://example.com" {
					t.Errorf("Expected 'https://example.com', got: %s", url)
				}
			},
		},
		{
			name: "build_url_with_path_should_combine_correctly",
			test: func(t *testing.T) {
				client := &HTTPClient{
					config: &Config{URL: "https://example.com"},
				}
				url := client.buildURL("/api/v1/devices")
				if url != "https://example.com/api/v1/devices" {
					t.Errorf("Expected 'https://example.com/api/v1/devices', got: %s", url)
				}
			},
		},
		{
			name: "build_url_with_base_url_trailing_slash_should_not_duplicate_slash",
			test: func(t *testing.T) {
				client := &HTTPClient{
					config: &Config{URL: "https://example.com/"},
				}
				url := client.buildURL("/api/v1/devices")
				if url != "https://example.com/api/v1/devices" {
					t.Errorf("Expected 'https://example.com/api/v1/devices', got: %s", url)
				}
			},
		},
		{
			name: "build_url_with_path_without_leading_slash_should_add_slash",
			test: func(t *testing.T) {
				client := &HTTPClient{
					config: &Config{URL: "https://example.com"},
				}
				url := client.buildURL("api/v1/devices")
				if url != "https://example.com/api/v1/devices" {
					t.Errorf("Expected 'https://example.com/api/v1/devices', got: %s", url)
				}
			},
		},
		{
			name: "build_url_with_query_params_should_preserve_them",
			test: func(t *testing.T) {
				client := &HTTPClient{
					config: &Config{URL: "https://example.com"},
				}
				url := client.buildURL("/api/v1/devices?type=all&active=true")
				if url != "https://example.com/api/v1/devices?type=all&active=true" {
					t.Errorf("Expected 'https://example.com/api/v1/devices?type=all&active=true', got: %s", url)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
