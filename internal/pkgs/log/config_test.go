package log

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// 验证生产环境默认值
	if config.Level != "info" {
		t.Errorf("Expected Level to be 'info', got '%s'", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("Expected Format to be 'json', got '%s'", config.Format)
	}
	if config.Output != "stdout" {
		t.Errorf("Expected Output to be 'stdout', got '%s'", config.Output)
	}
	if config.Development {
		t.Error("Expected Development to be false")
	}
	if config.EnableCaller {
		t.Error("Expected EnableCaller to be false")
	}
	if config.EnableStacktrace {
		t.Error("Expected EnableStacktrace to be false")
	}
	if config.MaxSize != 100 {
		t.Errorf("Expected MaxSize to be 100, got %d", config.MaxSize)
	}
	if config.MaxAge != 30 {
		t.Errorf("Expected MaxAge to be 30, got %d", config.MaxAge)
	}
	if config.MaxBackups != 10 {
		t.Errorf("Expected MaxBackups to be 10, got %d", config.MaxBackups)
	}
	if !config.Compress {
		t.Error("Expected Compress to be true")
	}
}

func TestDevelopmentDefaults(t *testing.T) {
	config := DevelopmentDefaults()

	// 验证开发环境默认值
	if config.Level != "debug" {
		t.Errorf("Expected Level to be 'debug', got '%s'", config.Level)
	}
	if config.Format != "console" {
		t.Errorf("Expected Format to be 'console', got '%s'", config.Format)
	}
	if config.Output != "stdout" {
		t.Errorf("Expected Output to be 'stdout', got '%s'", config.Output)
	}
	if !config.Development {
		t.Error("Expected Development to be true")
	}
	if !config.EnableCaller {
		t.Error("Expected EnableCaller to be true")
	}
	if !config.EnableStacktrace {
		t.Error("Expected EnableStacktrace to be true")
	}
	if config.MaxAge != 7 {
		t.Errorf("Expected MaxAge to be 7, got %d", config.MaxAge)
	}
	if config.MaxBackups != 3 {
		t.Errorf("Expected MaxBackups to be 3, got %d", config.MaxBackups)
	}
	if config.Compress {
		t.Error("Expected Compress to be false")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid default config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name:      "valid development config",
			config:    DevelopmentDefaults(),
			wantError: false,
		},
		{
			name: "invalid log level",
			config: &Config{
				Level:  "invalid",
				Format: "json",
				Output: "stdout",
			},
			wantError: true,
			errorMsg:  "invalid log level",
		},
		{
			name: "invalid format",
			config: &Config{
				Level:  "info",
				Format: "xml",
				Output: "stdout",
			},
			wantError: true,
			errorMsg:  "invalid log format",
		},
		{
			name: "invalid output",
			config: &Config{
				Level:  "info",
				Format: "json",
				Output: "invalid",
			},
			wantError: true,
			errorMsg:  "invalid log output",
		},
		{
			name: "file output without path",
			config: &Config{
				Level:    "info",
				Format:   "json",
				Output:   "file",
				FilePath: "",
			},
			wantError: true,
			errorMsg:  "file_path is required",
		},
		{
			name: "both output without path",
			config: &Config{
				Level:    "info",
				Format:   "json",
				Output:   "both",
				FilePath: "",
			},
			wantError: true,
			errorMsg:  "file_path is required",
		},
		{
			name: "negative max size",
			config: &Config{
				Level:   "info",
				Format:  "json",
				Output:  "stdout",
				MaxSize: -1,
			},
			wantError: true,
			errorMsg:  "max_size must be non-negative",
		},
		{
			name: "negative max age",
			config: &Config{
				Level:  "info",
				Format: "json",
				Output: "stdout",
				MaxAge: -1,
			},
			wantError: true,
			errorMsg:  "max_age must be non-negative",
		},
		{
			name: "negative max backups",
			config: &Config{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxBackups: -1,
			},
			wantError: true,
			errorMsg:  "max_backups must be non-negative",
		},
		{
			name: "valid file output with path",
			config: &Config{
				Level:    "info",
				Format:   "json",
				Output:   "file",
				FilePath: "/var/log/app.log",
				MaxSize:  100,
				MaxAge:   30,
			},
			wantError: false,
		},
		{
			name: "case insensitive level",
			config: &Config{
				Level:  "INFO",
				Format: "json",
				Output: "stdout",
			},
			wantError: false,
		},
		{
			name: "case insensitive format",
			config: &Config{
				Level:  "info",
				Format: "JSON",
				Output: "stdout",
			},
			wantError: false,
		},
		{
			name: "case insensitive output",
			config: &Config{
				Level:  "info",
				Format: "json",
				Output: "STDOUT",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestConfigValidateNormalization(t *testing.T) {
	// 测试配置标准化
	config := &Config{
		Level:  "INFO",
		Format: "JSON",
		Output: "STDOUT",
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证已标准化为小写
	if config.Level != "info" {
		t.Errorf("Expected Level to be normalized to 'info', got '%s'", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("Expected Format to be normalized to 'json', got '%s'", config.Format)
	}
	if config.Output != "stdout" {
		t.Errorf("Expected Output to be normalized to 'stdout', got '%s'", config.Output)
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
