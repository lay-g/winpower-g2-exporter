package winpower

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Timeout != 15*time.Second {
		t.Errorf("expected timeout 15s, got %v", cfg.Timeout)
	}

	if cfg.SkipSSLVerify != false {
		t.Errorf("expected skip_ssl_verify false, got %v", cfg.SkipSSLVerify)
	}

	if cfg.RefreshThreshold != 5*time.Minute {
		t.Errorf("expected refresh_threshold 5m, got %v", cfg.RefreshThreshold)
	}

	if cfg.UserAgent == "" {
		t.Error("expected non-empty user_agent")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: &Config{
				BaseURL:          "https://winpower.example.com",
				Username:         "admin",
				Password:         "secret",
				Timeout:          15 * time.Second,
				RefreshThreshold: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "empty base_url",
			cfg: &Config{
				BaseURL:          "",
				Username:         "admin",
				Password:         "secret",
				Timeout:          15 * time.Second,
				RefreshThreshold: 5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "base_url",
		},
		{
			name: "invalid base_url format",
			cfg: &Config{
				BaseURL:          "not a url",
				Username:         "admin",
				Password:         "secret",
				Timeout:          15 * time.Second,
				RefreshThreshold: 5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "base_url",
		},
		{
			name: "invalid URL scheme",
			cfg: &Config{
				BaseURL:          "ftp://winpower.example.com",
				Username:         "admin",
				Password:         "secret",
				Timeout:          15 * time.Second,
				RefreshThreshold: 5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "scheme",
		},
		{
			name: "empty username",
			cfg: &Config{
				BaseURL:          "https://winpower.example.com",
				Username:         "",
				Password:         "secret",
				Timeout:          15 * time.Second,
				RefreshThreshold: 5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "username",
		},
		{
			name: "empty password",
			cfg: &Config{
				BaseURL:          "https://winpower.example.com",
				Username:         "admin",
				Password:         "",
				Timeout:          15 * time.Second,
				RefreshThreshold: 5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "password",
		},
		{
			name: "zero timeout",
			cfg: &Config{
				BaseURL:          "https://winpower.example.com",
				Username:         "admin",
				Password:         "secret",
				Timeout:          0,
				RefreshThreshold: 5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "timeout",
		},
		{
			name: "negative timeout",
			cfg: &Config{
				BaseURL:          "https://winpower.example.com",
				Username:         "admin",
				Password:         "secret",
				Timeout:          -1 * time.Second,
				RefreshThreshold: 5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "timeout",
		},
		{
			name: "refresh threshold too short",
			cfg: &Config{
				BaseURL:          "https://winpower.example.com",
				Username:         "admin",
				Password:         "secret",
				Timeout:          15 * time.Second,
				RefreshThreshold: 30 * time.Second,
			},
			wantErr: true,
			errMsg:  "refresh_threshold",
		},
		{
			name: "refresh threshold too long",
			cfg: &Config{
				BaseURL:          "https://winpower.example.com",
				Username:         "admin",
				Password:         "secret",
				Timeout:          15 * time.Second,
				RefreshThreshold: 2 * time.Hour,
			},
			wantErr: true,
			errMsg:  "refresh_threshold",
		},
		{
			name: "http URL allowed",
			cfg: &Config{
				BaseURL:          "http://winpower.example.com",
				Username:         "admin",
				Password:         "secret",
				Timeout:          15 * time.Second,
				RefreshThreshold: 5 * time.Minute,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("error message should contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfig_WithDefaults(t *testing.T) {
	cfg := &Config{
		BaseURL:  "https://winpower.example.com",
		Username: "admin",
		Password: "secret",
		// Omit optional fields
	}

	cfg = cfg.WithDefaults()

	if cfg.Timeout == 0 {
		t.Error("expected timeout to be filled with default value")
	}

	if cfg.RefreshThreshold == 0 {
		t.Error("expected refresh_threshold to be filled with default value")
	}

	if cfg.UserAgent == "" {
		t.Error("expected user_agent to be filled with default value")
	}

	// Test that existing values are not overwritten
	cfg2 := &Config{
		BaseURL:          "https://winpower.example.com",
		Username:         "admin",
		Password:         "secret",
		Timeout:          30 * time.Second,
		RefreshThreshold: 10 * time.Minute,
		UserAgent:        "Custom Agent",
	}

	cfg2 = cfg2.WithDefaults()

	if cfg2.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", cfg2.Timeout)
	}

	if cfg2.RefreshThreshold != 10*time.Minute {
		t.Errorf("expected refresh_threshold 10m, got %v", cfg2.RefreshThreshold)
	}

	if cfg2.UserAgent != "Custom Agent" {
		t.Errorf("expected user_agent 'Custom Agent', got %q", cfg2.UserAgent)
	}
}

func TestConfig_Clone(t *testing.T) {
	original := &Config{
		BaseURL:          "https://winpower.example.com",
		Username:         "admin",
		Password:         "secret",
		Timeout:          15 * time.Second,
		SkipSSLVerify:    true,
		RefreshThreshold: 5 * time.Minute,
		UserAgent:        "Test Agent",
	}

	cloned := original.Clone()

	// Verify all fields are copied
	if cloned.BaseURL != original.BaseURL {
		t.Error("BaseURL not cloned correctly")
	}
	if cloned.Username != original.Username {
		t.Error("Username not cloned correctly")
	}
	if cloned.Password != original.Password {
		t.Error("Password not cloned correctly")
	}
	if cloned.Timeout != original.Timeout {
		t.Error("Timeout not cloned correctly")
	}
	if cloned.SkipSSLVerify != original.SkipSSLVerify {
		t.Error("SkipSSLVerify not cloned correctly")
	}
	if cloned.RefreshThreshold != original.RefreshThreshold {
		t.Error("RefreshThreshold not cloned correctly")
	}
	if cloned.UserAgent != original.UserAgent {
		t.Error("UserAgent not cloned correctly")
	}

	// Verify it's a different instance
	cloned.Password = "modified"
	if original.Password == "modified" {
		t.Error("modifying clone affected original")
	}
}

func TestConfig_Sanitize(t *testing.T) {
	cfg := &Config{
		BaseURL:          "https://winpower.example.com",
		Username:         "admin",
		Password:         "secret123",
		Timeout:          15 * time.Second,
		SkipSSLVerify:    true,
		RefreshThreshold: 5 * time.Minute,
		UserAgent:        "Test Agent",
	}

	sanitized := cfg.Sanitize()

	// Check that password is redacted
	if password, ok := sanitized["password"].(string); !ok || password == "secret123" {
		t.Errorf("password should be redacted, got %v", sanitized["password"])
	}

	// Check that other fields are present
	if sanitized["username"] != "admin" {
		t.Error("username should not be sanitized")
	}

	if sanitized["base_url"] != "https://winpower.example.com" {
		t.Error("base_url should not be sanitized")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
