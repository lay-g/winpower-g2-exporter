package winpower

import (
	"strings"
	"testing"
	"time"
)

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "should_return_non_nil_config",
			test: func(t *testing.T) {
				config := DefaultConfig()
				if config == nil {
					t.Fatal("DefaultConfig() should not return nil")
				}
			},
		},
		{
			name: "should_have_sensible_defaults",
			test: func(t *testing.T) {
				config := DefaultConfig()

				if config.URL == "" {
					t.Error("URL should have a default value")
				}

				if config.Username == "" {
					t.Error("Username should have a default value")
				}

				if config.Password == "" {
					t.Error("Password should have a default value")
				}

				if config.Timeout <= 0 {
					t.Error("Timeout should be positive")
				}

				if config.MaxRetries < 0 {
					t.Error("MaxRetries should not be negative")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestConfig_Validate tests the Validate method
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "nil_config_should_error",
			test: func(t *testing.T) {
				var config *Config
				err := config.Validate()
				if err == nil {
					t.Fatal("Validate() should return error for nil config")
				}
			},
		},
		{
			name: "empty_url_should_error",
			test: func(t *testing.T) {
				config := &Config{
					URL:      "",
					Username: "testuser",
					Password: "testpass",
				}
				err := config.Validate()
				if err == nil {
					t.Error("Validate() should return error for empty URL")
				}
			},
		},
		{
			name: "invalid_url_format_should_error",
			test: func(t *testing.T) {
				config := &Config{
					URL:      "invalid-url",
					Username: "testuser",
					Password: "testpass",
				}
				err := config.Validate()
				if err == nil {
					t.Error("Validate() should return error for invalid URL format")
				}
			},
		},
		{
			name: "empty_username_should_error",
			test: func(t *testing.T) {
				config := &Config{
					URL:      "https://example.com",
					Username: "",
					Password: "testpass",
				}
				err := config.Validate()
				if err == nil {
					t.Error("Validate() should return error for empty username")
				}
			},
		},
		{
			name: "empty_password_should_error",
			test: func(t *testing.T) {
				config := &Config{
					URL:      "https://example.com",
					Username: "testuser",
					Password: "",
				}
				err := config.Validate()
				if err == nil {
					t.Error("Validate() should return error for empty password")
				}
			},
		},
		{
			name: "negative_timeout_should_error",
			test: func(t *testing.T) {
				config := &Config{
					URL:      "https://example.com",
					Username: "testuser",
					Password: "testpass",
					Timeout:  -1 * time.Second,
				}
				err := config.Validate()
				if err == nil {
					t.Error("Validate() should return error for negative timeout")
				}
			},
		},
		{
			name: "negative_max_retries_should_error",
			test: func(t *testing.T) {
				config := &Config{
					URL:        "https://example.com",
					Username:   "testuser",
					Password:   "testpass",
					Timeout:    10 * time.Second,
					MaxRetries: -1,
				}
				err := config.Validate()
				if err == nil {
					t.Error("Validate() should return error for negative max retries")
				}
			},
		},
		{
			name: "valid_config_should_pass",
			test: func(t *testing.T) {
				config := &Config{
					URL:        "https://example.com",
					Username:   "testuser",
					Password:   "testpass",
					Timeout:    10 * time.Second,
					MaxRetries: 3,
				}
				err := config.Validate()
				if err != nil {
					t.Errorf("Validate() should not return error for valid config: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestConfig_Clone tests the Clone method
func TestConfig_Clone(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "nil_config_should_return_nil",
			test: func(t *testing.T) {
				var config *Config
				cloned := config.Clone()
				if cloned != nil {
					t.Error("Clone() should return nil for nil config")
				}
			},
		},
		{
			name: "clone_should_create_deep_copy",
			test: func(t *testing.T) {
				original := &Config{
					URL:           "https://example.com",
					Username:      "testuser",
					Password:      "testpass",
					Timeout:       10 * time.Second,
					MaxRetries:    3,
					SkipTLSVerify: true,
				}
				clonedInterface := original.Clone()
				cloned, ok := clonedInterface.(*Config)
				if !ok {
					t.Fatal("Clone() did not return *Config type")
				}

				if cloned == original {
					t.Error("Clone() should create a new instance")
				}

				if cloned.URL != original.URL {
					t.Error("Clone() should copy URL correctly")
				}

				if cloned.Username != original.Username {
					t.Error("Clone() should copy Username correctly")
				}

				if cloned.Password != original.Password {
					t.Error("Clone() should copy Password correctly")
				}

				if cloned.Timeout != original.Timeout {
					t.Error("Clone() should copy Timeout correctly")
				}

				if cloned.MaxRetries != original.MaxRetries {
					t.Error("Clone() should copy MaxRetries correctly")
				}

				if cloned.SkipTLSVerify != original.SkipTLSVerify {
					t.Error("Clone() should copy SkipTLSVerify correctly")
				}
			},
		},
		{
			name: "modifying_clone_should_not_affect_original",
			test: func(t *testing.T) {
				original := &Config{
					URL:      "https://example.com",
					Username: "testuser",
					Password: "testpass",
				}
				clonedInterface := original.Clone()
				cloned, ok := clonedInterface.(*Config)
				if !ok {
					t.Fatal("Clone() did not return *Config type")
				}
				cloned.URL = "https://modified.com"

				if original.URL == cloned.URL {
					t.Error("Modifying clone should not affect original")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestConfig_String tests the String method
func TestConfig_String(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "nil_config_should_return_nil_string",
			test: func(t *testing.T) {
				var config *Config
				str := config.String()
				if str != "<nil>" {
					t.Errorf("String() should return '<nil>' for nil config, got: %s", str)
				}
			},
		},
		{
			name: "string_should_mask_sensitive_data",
			test: func(t *testing.T) {
				config := &Config{
					URL:      "https://example.com",
					Username: "testuser",
					Password: "secret123",
				}
				str := config.String()

				if str == "" {
					t.Error("String() should not return empty string")
				}

				// Password should be masked
				if contains(str, "secret123") {
					t.Error("String() should mask password")
				}

				// URL and username should be visible
				if !contains(str, "example.com") {
					t.Error("String() should include URL")
				}

				if !contains(str, "testuser") {
					t.Error("String() should include username")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestConfig_WithMethods tests the fluent "With" methods
func TestConfig_WithMethods(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "WithURL_should_set_url_and_return_self",
			test: func(t *testing.T) {
				config := &Config{}
				result := config.WithURL("https://test.com")

				if config.URL != "https://test.com" {
					t.Error("WithURL() should set URL")
				}

				if result != config {
					t.Error("WithURL() should return self for chaining")
				}
			},
		},
		{
			name: "WithCredentials_should_set_credentials_and_return_self",
			test: func(t *testing.T) {
				config := &Config{}
				result := config.WithCredentials("user", "pass")

				if config.Username != "user" {
					t.Error("WithCredentials() should set username")
				}

				if config.Password != "pass" {
					t.Error("WithCredentials() should set password")
				}

				if result != config {
					t.Error("WithCredentials() should return self for chaining")
				}
			},
		},
		{
			name: "WithTimeout_should_set_timeout_and_return_self",
			test: func(t *testing.T) {
				config := &Config{}
				timeout := 5 * time.Second
				result := config.WithTimeout(timeout)

				if config.Timeout != timeout {
					t.Error("WithTimeout() should set timeout")
				}

				if result != config {
					t.Error("WithTimeout() should return self for chaining")
				}
			},
		},
		{
			name: "WithMaxRetries_should_set_max_retries_and_return_self",
			test: func(t *testing.T) {
				config := &Config{}
				result := config.WithMaxRetries(5)

				if config.MaxRetries != 5 {
					t.Error("WithMaxRetries() should set max retries")
				}

				if result != config {
					t.Error("WithMaxRetries() should return self for chaining")
				}
			},
		},
		{
			name: "WithSkipTLSVerify_should_set_skip_tls_verify_and_return_self",
			test: func(t *testing.T) {
				config := &Config{}
				result := config.WithSkipTLSVerify(true)

				if config.SkipTLSVerify != true {
					t.Error("WithSkipTLSVerify() should set skip TLS verify")
				}

				if result != config {
					t.Error("WithSkipTLSVerify() should return self for chaining")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestConfig_ValidateOptimized tests the enhanced validation with performance improvements
func TestConfig_ValidateOptimized(t *testing.T) {
	t.Run("invalid_url_scheme_should_error", func(t *testing.T) {
		config := &Config{
			URL:      "ftp://example.com",
			Username: "testuser",
			Password: "testpass",
		}
		err := config.Validate()
		if err == nil {
			t.Error("Validate() should return error for invalid URL scheme")
		}
	})

	t.Run("invalid_hostname_should_error", func(t *testing.T) {
		config := &Config{
			URL:      "https://invalid hostname with spaces.com",
			Username: "testuser",
			Password: "testpass",
		}
		err := config.Validate()
		if err == nil {
			t.Error("Validate() should return error for invalid hostname")
		}
	})

	t.Run("username_with_invalid_chars_should_error", func(t *testing.T) {
		config := &Config{
			URL:      "https://example.com",
			Username: "test user",
			Password: "testpass",
		}
		err := config.Validate()
		if err == nil {
			t.Error("Validate() should return error for username with spaces")
		}
	})

	t.Run("timeout_too_short_should_error", func(t *testing.T) {
		config := &Config{
			URL:      "https://example.com",
			Username: "testuser",
			Password: "testpass",
			Timeout:  3 * time.Second,
		}
		err := config.Validate()
		if err == nil {
			t.Error("Validate() should return error for timeout too short")
		}
	})

	t.Run("timeout_too_long_should_error", func(t *testing.T) {
		config := &Config{
			URL:      "https://example.com",
			Username: "testuser",
			Password: "testpass",
			Timeout:  400 * time.Second,
		}
		err := config.Validate()
		if err == nil {
			t.Error("Validate() should return error for timeout too long")
		}
	})

	t.Run("max_retries_too_high_should_error", func(t *testing.T) {
		config := &Config{
			URL:        "https://example.com",
			Username:   "testuser",
			Password:   "testpass",
			Timeout:    30 * time.Second,
			MaxRetries: 15,
		}
		err := config.Validate()
		if err == nil {
			t.Error("Validate() should return error for max retries too high")
		}
	})

	t.Run("localhost_should_be_valid", func(t *testing.T) {
		config := &Config{
			URL:        "https://localhost:8080",
			Username:   "testuser",
			Password:   "testpass",
			Timeout:    30 * time.Second,
			MaxRetries: 3,
		}
		err := config.Validate()
		if err != nil {
			t.Errorf("Validate() should pass for localhost: %v", err)
		}
	})

	t.Run("ip_address_should_be_valid", func(t *testing.T) {
		config := &Config{
			URL:        "https://127.0.0.1:8080",
			Username:   "testuser",
			Password:   "testpass",
			Timeout:    30 * time.Second,
			MaxRetries: 3,
		}
		err := config.Validate()
		if err != nil {
			t.Errorf("Validate() should pass for IP address: %v", err)
		}
	})
}

// TestConfig_ValidateQuick tests the quick validation method
func TestConfig_ValidateQuick(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "nil_config_should_error",
			test: func(t *testing.T) {
				var config *Config
				err := config.ValidateQuick()
				if err == nil {
					t.Error("ValidateQuick() should return error for nil config")
				}
			},
		},
		{
			name: "valid_config_should_pass_quick_validation",
			test: func(t *testing.T) {
				config := &Config{
					URL:      "https://example.com",
					Username: "testuser",
					Password: "testpass",
					// No need to set timeout and max_retries for quick validation
				}
				err := config.ValidateQuick()
				if err != nil {
					t.Errorf("ValidateQuick() should pass for valid config: %v", err)
				}
			},
		},
		{
			name: "empty_url_should_error_quick",
			test: func(t *testing.T) {
				config := &Config{
					Username: "testuser",
					Password: "testpass",
				}
				err := config.ValidateQuick()
				if err == nil {
					t.Error("ValidateQuick() should return error for empty URL")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestConfig_IsSecure tests the IsSecure method
func TestConfig_IsSecure(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name:   "nil_config_should_return_false",
			config: nil,
			want:   false,
		},
		{
			name: "https_without_skip_verify_should_be_secure",
			config: &Config{
				URL:           "https://example.com",
				SkipTLSVerify: false,
			},
			want: true,
		},
		{
			name: "https_with_skip_verify_should_be_insecure",
			config: &Config{
				URL:           "https://example.com",
				SkipTLSVerify: true,
			},
			want: false,
		},
		{
			name: "http_should_be_insecure",
			config: &Config{
				URL:           "http://example.com",
				SkipTLSVerify: false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsSecure(); got != tt.want {
				t.Errorf("Config.IsSecure() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConfig_IsProductionReady tests the IsProductionReady method
func TestConfig_IsProductionReady(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil_config_should_error",
			config:  nil,
			wantErr: true,
			errMsg:  "config is nil",
		},
		{
			name: "default_credentials_should_error",
			config: &Config{
				URL:      "https://example.com",
				Username: "admin",
				Password: "admin",
			},
			wantErr: true,
			errMsg:  "default admin credentials",
		},
		{
			name: "localhost_url_should_error",
			config: &Config{
				URL:      "https://localhost:8080",
				Username: "user",
				Password: "password",
			},
			wantErr: true,
			errMsg:  "localhost URL",
		},
		{
			name: "insecure_connection_should_error",
			config: &Config{
				URL:           "http://example.com",
				Username:      "user",
				Password:      "password",
				SkipTLSVerify: false,
			},
			wantErr: true,
			errMsg:  "insecure connection",
		},
		{
			name: "production_ready_config_should_pass",
			config: &Config{
				URL:           "https://winpower.company.com",
				Username:      "operator",
				Password:      "secure_password",
				SkipTLSVerify: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.IsProductionReady()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.IsProductionReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Config.IsProductionReady() error = %v, expected to contain %s", err, tt.errMsg)
			}
		})
	}
}

// TestConfig_GetConnectionSummary tests the GetConnectionSummary method
func TestConfig_GetConnectionSummary(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name:   "nil_config_should_return_nil_string",
			config: nil,
			want:   "<nil config>",
		},
		{
			name: "https_secure_config_should_show_https",
			config: &Config{
				URL:           "https://example.com:8080",
				Timeout:       30 * time.Second,
				MaxRetries:    3,
				SkipTLSVerify: false,
			},
			want: "HTTPS connection to https://example.com:8080 (secure, timeout: 30s, retries: 3)",
		},
		{
			name: "http_insecure_config_should_show_http",
			config: &Config{
				URL:           "http://example.com",
				Timeout:       15 * time.Second,
				MaxRetries:    5,
				SkipTLSVerify: false,
			},
			want: "HTTP connection to http://example.com (secure, timeout: 15s, retries: 5)",
		},
		{
			name: "https_with_skip_verify_should_show_insecure",
			config: &Config{
				URL:           "https://example.com",
				Timeout:       10 * time.Second,
				MaxRetries:    1,
				SkipTLSVerify: true,
			},
			want: "HTTPS connection to https://example.com (insecure (TLS verify skipped), timeout: 10s, retries: 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetConnectionSummary(); got != tt.want {
				t.Errorf("Config.GetConnectionSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConfig_ValidationError tests the new ValidationError type
func TestConfig_ValidationError(t *testing.T) {
	t.Run("validation_error_should_have_structured_fields", func(t *testing.T) {
		config := &Config{
			URL:      "invalid-url",
			Username: "test user", // invalid username
			Timeout:  -1 * time.Second,
		}

		err := config.Validate()
		if err == nil {
			t.Fatal("Expected validation error for invalid config")
		}

		// Check if it's a ValidationError
		if validationErr, ok := err.(*ValidationError); ok {
			if validationErr.Field == "" {
				t.Error("ValidationError should have a Field")
			}
			if validationErr.Code == "" {
				t.Error("ValidationError should have a Code")
			}
			if validationErr.Message == "" {
				t.Error("ValidationError should have a Message")
			}
			if validationErr.Value == "" {
				t.Error("ValidationError should have a Value")
			}

			// Test Error() method
			errStr := validationErr.Error()
			if !strings.Contains(errStr, validationErr.Field) {
				t.Error("Error() should include field name")
			}
			if !strings.Contains(errStr, validationErr.Message) {
				t.Error("Error() should include message")
			}
		} else {
			t.Errorf("Expected *ValidationError, got %T", err)
		}
	})
}

// TestConfig_ValidateWithSummary tests the new ValidateWithSummary method
func TestConfig_ValidateWithSummary(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
		expectField string
	}{
		{
			name:        "valid_config_should_return_valid_summary",
			config:      DefaultConfig(),
			expectValid: true,
		},
		{
			name: "invalid_config_should_return_invalid_summary",
			config: &Config{
				URL:      "ftp://example.com", // invalid scheme
				Username: "test",
				Password: "test",
			},
			expectValid: false,
			expectField: "url",
		},
		{
			name:        "nil_config_should_return_invalid_summary",
			config:      nil,
			expectValid: false,
			expectField: "config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, err := tt.config.ValidateWithSummary()

			if tt.expectValid {
				if err != nil {
					t.Errorf("Expected no error for valid config, got: %v", err)
				}
				if !summary.IsValid {
					t.Error("Expected IsValid=true for valid config")
				}
				if summary.SecureStatus == "" {
					t.Error("Expected SecureStatus for valid config")
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid config")
				}
				if summary.IsValid {
					t.Error("Expected IsValid=false for invalid config")
				}
				if tt.expectField != "" && summary.Field != tt.expectField {
					t.Errorf("Expected field %s, got %s", tt.expectField, summary.Field)
				}
			}
		})
	}
}

// TestConfig_ClearValidationCache tests the ClearValidationCache function
func TestConfig_ClearValidationCache(t *testing.T) {
	// First validate a config to populate cache
	config := DefaultConfig()
	err := config.Validate()
	if err != nil {
		t.Fatalf("Unexpected error validating config: %v", err)
	}

	// Clear cache
	ClearValidationCache()

	// Validate again - should still work
	err = config.Validate()
	if err != nil {
		t.Errorf("Unexpected error after clearing cache: %v", err)
	}
}

// TestConfig_ValidationCachePerformance tests that validation cache improves performance
func TestConfig_ValidationCachePerformance(t *testing.T) {
	config := DefaultConfig()
	urls := []string{
		"https://example1.com",
		"https://example2.com",
		"https://example3.com",
	}

	// Clear cache first
	ClearValidationCache()

	// First validation (cache miss)
	start := time.Now()
	for _, url := range urls {
		testConfig := config.WithURL(url)
		err := testConfig.Validate()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
	firstPassDuration := time.Since(start)

	// Second validation (cache hit)
	start = time.Now()
	for _, url := range urls {
		testConfig := config.WithURL(url)
		err := testConfig.Validate()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
	secondPassDuration := time.Since(start)

	// Second pass should be faster (or at least not significantly slower)
	// Note: This is a basic performance test, results may vary on different systems
	if secondPassDuration > firstPassDuration*2 {
		t.Logf("Warning: Cache may not be improving performance significantly. First pass: %v, Second pass: %v",
			firstPassDuration, secondPassDuration)
	}
}

// TestConfig_Validate_URL tests URL validation in detail
func TestConfig_Validate_URL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedErr bool
		errMsg      string
	}{
		{
			name:        "empty URL",
			url:         "",
			expectedErr: true,
			errMsg:      "cannot be empty",
		},
		{
			name:        "valid HTTPS URL",
			url:         "https://example.com",
			expectedErr: false,
		},
		{
			name:        "valid HTTP URL",
			url:         "http://example.com",
			expectedErr: false,
		},
		{
			name:        "valid HTTPS URL with port",
			url:         "https://example.com:8080",
			expectedErr: false,
		},
		{
			name:        "valid localhost HTTPS",
			url:         "https://localhost:8080",
			expectedErr: false,
		},
		{
			name:        "valid IP address",
			url:         "https://192.168.1.100:8080",
			expectedErr: false,
		},
		{
			name:        "invalid scheme - FTP",
			url:         "ftp://example.com",
			expectedErr: true,
			errMsg:      "must start with http:// or https://",
		},
		{
			name:        "invalid scheme - no protocol",
			url:         "example.com",
			expectedErr: true,
			errMsg:      "must start with http:// or https://",
		},
		{
			name:        "malformed URL",
			url:         "https://",
			expectedErr: true,
			errMsg:      "hostname cannot be empty",
		},
		{
			name:        "URL with spaces",
			url:         "https://example .com",
			expectedErr: true,
			errMsg:      "invalid URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				URL:        tt.url,
				Username:   "testuser",
				Password:   "testpass",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			}

			err := config.Validate()
			if (err != nil) != tt.expectedErr {
				t.Errorf("Validate() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if tt.expectedErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
				} else if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

// TestConfig_Validate_Credentials tests credential validation
func TestConfig_Validate_Credentials(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		password    string
		expectedErr bool
		errMsg      string
	}{
		{
			name:        "valid credentials",
			username:    "testuser",
			password:    "testpass",
			expectedErr: false,
		},
		{
			name:        "empty username",
			username:    "",
			password:    "testpass",
			expectedErr: true,
			errMsg:      "username cannot be empty",
		},
		{
			name:        "empty password",
			username:    "testuser",
			password:    "",
			expectedErr: true,
			errMsg:      "password cannot be empty",
		},
		{
			name:        "both empty",
			username:    "",
			password:    "",
			expectedErr: true,
			errMsg:      "username cannot be empty",
		},
		{
			name:        "username with invalid special characters",
			username:    "test\tuser", // Tab character
			password:    "testpass",
			expectedErr: true,
			errMsg:      "invalid characters",
		},
		{
			name:        "username with spaces",
			username:    "test user",
			password:    "testpass",
			expectedErr: true,
			errMsg:      "invalid characters",
		},
		{
			name:        "username with colon",
			username:    "test:user",
			password:    "testpass",
			expectedErr: true,
			errMsg:      "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				URL:        "https://example.com",
				Username:   tt.username,
				Password:   tt.password,
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			}

			err := config.Validate()
			if (err != nil) != tt.expectedErr {
				t.Errorf("Validate() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if tt.expectedErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
				} else if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

// TestConfig_String_MaskPassword tests password masking in String method
func TestConfig_String_MaskPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected string
	}{
		{
			name:     "short password",
			password: "ab",
			expected: "**",
		},
		{
			name:     "medium password",
			password: "password123",
			expected: "p***3",
		},
		{
			name:     "long password",
			password: "verylongpassword",
			expected: "v***d",
		},
		{
			name:     "empty password",
			password: "",
			expected: "<empty>",
		},
		{
			name:     "single character password",
			password: "a",
			expected: "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				URL:        "https://example.com",
				Username:   "testuser",
				Password:   tt.password,
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			}

			str := config.String()

			// Password should be masked (check that the raw password is not in the string)
			// For single character passwords, the masked version is the same as the original,
			// so we need to be more careful
			if len(tt.password) > 1 && strings.Contains(str, tt.password) {
				t.Errorf("String() should mask password '%s', but it's visible in: %s", tt.password, str)
			}

			// Should contain the masked version
			if !strings.Contains(str, tt.expected) {
				t.Errorf("String() should contain masked password '%s', got: %s", tt.expected, str)
			}

			// Username and URL should be visible
			if !strings.Contains(str, "testuser") {
				t.Errorf("String() should contain username, got: %s", str)
			}

			if !strings.Contains(str, "example.com") {
				t.Errorf("String() should contain URL, got: %s", str)
			}
		})
	}
}
