package metrics

import (
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestValidator_ValidateConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	tests := []struct {
		name     string
		config   MetricManagerConfig
		expected ValidationResult
	}{
		{
			name: "valid basic config",
			config: MetricManagerConfig{
				Namespace:                 "winpower",
				Subsystem:                 "exporter",
				RequestDurationBuckets:    []float64{0.1, 0.5, 1, 2, 5},
				CollectionDurationBuckets: []float64{0.5, 1, 2, 5, 10},
				APIResponseBuckets:        []float64{0.05, 0.1, 0.5, 1},
			},
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{},
			},
		},
		{
			name: "empty namespace",
			config: MetricManagerConfig{
				Namespace:                 "",
				Subsystem:                 "exporter",
				RequestDurationBuckets:    []float64{0.1, 0.5, 1, 2, 5},
				CollectionDurationBuckets: []float64{0.5, 1, 2, 5, 10},
				APIResponseBuckets:        []float64{0.05, 0.1, 0.5, 1},
			},
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"metrics namespace cannot be empty"},
				Warns:  []string{},
			},
		},
		{
			name: "invalid namespace characters",
			config: MetricManagerConfig{
				Namespace:                 "win-power",
				Subsystem:                 "exporter",
				RequestDurationBuckets:    []float64{0.1, 0.5, 1, 2, 5},
				CollectionDurationBuckets: []float64{0.5, 1, 2, 5, 10},
				APIResponseBuckets:        []float64{0.05, 0.1, 0.5, 1},
			},
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"invalid metrics namespace: win-power (must contain only letters, numbers, and underscores, and not start with a number)"},
				Warns:  []string{},
			},
		},
		{
			name: "empty buckets",
			config: MetricManagerConfig{
				Namespace:                 "winpower",
				Subsystem:                 "exporter",
				RequestDurationBuckets:    []float64{},
				CollectionDurationBuckets: []float64{0.5, 1, 2, 5, 10},
				APIResponseBuckets:        []float64{0.05, 0.1, 0.5, 1},
			},
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"request_duration buckets cannot be empty"},
				Warns:  []string{},
			},
		},
		{
			name: "non-increasing buckets",
			config: MetricManagerConfig{
				Namespace:                 "winpower",
				Subsystem:                 "exporter",
				RequestDurationBuckets:    []float64{0.5, 0.1, 1, 2, 5},
				CollectionDurationBuckets: []float64{0.5, 1, 2, 5, 10},
				APIResponseBuckets:        []float64{0.05, 0.1, 0.5, 1},
			},
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"request_duration buckets must be in increasing order: bucket[0] (0.500000) <= bucket[1] (0.100000)"},
				Warns:  []string{},
			},
		},
		{
			name: "negative bucket value",
			config: MetricManagerConfig{
				Namespace:                 "winpower",
				Subsystem:                 "exporter",
				RequestDurationBuckets:    []float64{-0.1, 0.5, 1, 2, 5},
				CollectionDurationBuckets: []float64{0.5, 1, 2, 5, 10},
				APIResponseBuckets:        []float64{0.05, 0.1, 0.5, 1},
			},
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"request_duration bucket[0] must be positive, got: -0.100000"},
				Warns:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateConfig(tt.config)

			if result.Valid != tt.expected.Valid {
				t.Errorf("ValidateConfig() validity = %v, expected %v", result.Valid, tt.expected.Valid)
			}

			if len(result.Errors) != len(tt.expected.Errors) {
				t.Errorf("ValidateConfig() errors count = %d, expected %d", len(result.Errors), len(tt.expected.Errors))
			}

			for i, err := range result.Errors {
				if i >= len(tt.expected.Errors) || err != tt.expected.Errors[i] {
					t.Errorf("ValidateConfig() error[%d] = %v, expected %v", i, err, tt.expected.Errors[i])
				}
			}
		})
	}
}

func TestValidator_ValidateEnvironmentConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	tests := []struct {
		name     string
		envVars  map[string]string
		expected ValidationResult
	}{
		{
			name: "valid environment config",
			envVars: map[string]string{
				"WINPOWER_EXPORTER_METRICS_NAMESPACE":                   "winpower",
				"WINPOWER_EXPORTER_METRICS_SUBSYSTEM":                   "exporter",
				"WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS":    "[0.1, 0.5, 1, 2, 5]",
				"WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS": "[0.5, 1, 2, 5, 10]",
				"WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS":        "[0.05, 0.1, 0.5, 1]",
			},
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{},
			},
		},
		{
			name: "missing required variables",
			envVars: map[string]string{
				"WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS": "[0.1, 0.5, 1, 2, 5]",
			},
			expected: ValidationResult{
				Valid: false,
				Errors: []string{
					"required environment variable WINPOWER_EXPORTER_METRICS_NAMESPACE is not set",
					"required environment variable WINPOWER_EXPORTER_METRICS_SUBSYSTEM is not set",
				},
				Warns: []string{},
			},
		},
		{
			name: "invalid JSON in buckets",
			envVars: map[string]string{
				"WINPOWER_EXPORTER_METRICS_NAMESPACE":                "winpower",
				"WINPOWER_EXPORTER_METRICS_SUBSYSTEM":                "exporter",
				"WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS": "[0.1, 0.5, 1, 2, 5", // Missing closing bracket
			},
			expected: ValidationResult{
				Valid: false,
				Errors: []string{
					"invalid JSON format for request_duration buckets: invalid character ']' looking for beginning of value",
				},
				Warns: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateEnvironmentConfig(tt.envVars)

			if result.Valid != tt.expected.Valid {
				t.Errorf("ValidateEnvironmentConfig() validity = %v, expected %v", result.Valid, tt.expected.Valid)
			}

			if len(result.Errors) != len(tt.expected.Errors) {
				t.Errorf("ValidateEnvironmentConfig() errors count = %d, expected %d", len(result.Errors), len(tt.expected.Errors))
			}
		})
	}
}

func TestValidator_ValidateURL(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	tests := []struct {
		name     string
		url      string
		field    string
		expected ValidationResult
	}{
		{
			name:  "valid HTTP URL",
			url:   "http://example.com:8080",
			field: "test_url",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{},
			},
		},
		{
			name:  "valid HTTPS URL",
			url:   "https://secure.example.com",
			field: "test_url",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{},
			},
		},
		{
			name:  "empty URL",
			url:   "",
			field: "test_url",
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"test_url cannot be empty"},
				Warns:  []string{},
			},
		},
		{
			name:  "invalid scheme",
			url:   "ftp://example.com",
			field: "test_url",
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"test_url must use http or https scheme, got: ftp"},
				Warns:  []string{},
			},
		},
		{
			name:  "invalid port",
			url:   "http://example.com:99999",
			field: "test_url",
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"test_url has invalid port number: 99999"},
				Warns:  []string{},
			},
		},
		{
			name:  "missing host",
			url:   "http://",
			field: "test_url",
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"test_url must have a valid host"},
				Warns:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateURL(tt.url, tt.field)

			if result.Valid != tt.expected.Valid {
				t.Errorf("ValidateURL() validity = %v, expected %v", result.Valid, tt.expected.Valid)
			}

			if len(result.Errors) != len(tt.expected.Errors) {
				t.Errorf("ValidateURL() errors count = %d, expected %d", len(result.Errors), len(tt.expected.Errors))
			}
		})
	}
}

func TestValidator_ValidatePort(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	tests := []struct {
		name     string
		port     string
		field    string
		expected ValidationResult
	}{
		{
			name:  "valid port",
			port:  "8080",
			field: "test_port",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{},
			},
		},
		{
			name:  "empty port",
			port:  "",
			field: "test_port",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{"test_port is not specified, will use default"},
			},
		},
		{
			name:  "invalid number",
			port:  "invalid",
			field: "test_port",
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"test_port is not a valid number: strconv.Atoi: parsing \"invalid\": invalid syntax"},
				Warns:  []string{},
			},
		},
		{
			name:  "port too low",
			port:  "0",
			field: "test_port",
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"test_port must be between 1 and 65535, got: 0"},
				Warns:  []string{},
			},
		},
		{
			name:  "port too high",
			port:  "65536",
			field: "test_port",
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"test_port must be between 1 and 65535, got: 65536"},
				Warns:  []string{},
			},
		},
		{
			name:  "well-known port 80",
			port:  "80",
			field: "test_port",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{"test_port is using well-known port 80 (HTTP), ensure this is intentional"},
			},
		},
		{
			name:  "prometheus port 9090",
			port:  "9090",
			field: "test_port",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{"test_port is using well-known port 9090 (Prometheus), ensure this is intentional"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidatePort(tt.port, tt.field)

			if result.Valid != tt.expected.Valid {
				t.Errorf("ValidatePort() validity = %v, expected %v", result.Valid, tt.expected.Valid)
			}

			if len(result.Errors) != len(tt.expected.Errors) {
				t.Errorf("ValidatePort() errors count = %d, expected %d", len(result.Errors), len(tt.expected.Errors))
			}
		})
	}
}

func TestValidator_ValidateTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	tests := []struct {
		name     string
		timeout  string
		field    string
		expected ValidationResult
	}{
		{
			name:    "valid timeout",
			timeout: "30s",
			field:   "test_timeout",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{},
			},
		},
		{
			name:    "empty timeout",
			timeout: "",
			field:   "test_timeout",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{"test_timeout is not specified, will use default"},
			},
		},
		{
			name:    "invalid duration",
			timeout: "invalid",
			field:   "test_timeout",
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"test_timeout is not a valid duration: invalid duration"},
				Warns:  []string{},
			},
		},
		{
			name:    "short timeout for request",
			timeout: "2s",
			field:   "request_timeout",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{"request_timeout (2s) might be too short, consider using at least 5s"},
			},
		},
		{
			name:    "very long timeout",
			timeout: "10m",
			field:   "connection_timeout",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{"connection_timeout (10m0s) is very long, consider reducing to avoid hanging"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateTimeout(tt.timeout, tt.field)

			if result.Valid != tt.expected.Valid {
				t.Errorf("ValidateTimeout() validity = %v, expected %v", result.Valid, tt.expected.Valid)
			}

			if len(result.Errors) != len(tt.expected.Errors) {
				t.Errorf("ValidateTimeout() errors count = %d, expected %d", len(result.Errors), len(tt.expected.Errors))
			}
		})
	}
}

func TestValidator_ValidateLogLevel(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	tests := []struct {
		name     string
		level    string
		field    string
		expected ValidationResult
	}{
		{
			name:  "valid log level",
			level: "info",
			field: "log_level",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{},
			},
		},
		{
			name:  "empty log level",
			level: "",
			field: "log_level",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{"log_level is not specified, will use default"},
			},
		},
		{
			name:  "invalid log level",
			level: "verbose",
			field: "log_level",
			expected: ValidationResult{
				Valid:  false,
				Errors: []string{"log_level must be one of [debug info warn error], got: verbose"},
				Warns:  []string{},
			},
		},
		{
			name:  "uppercase log level",
			level: "DEBUG",
			field: "log_level",
			expected: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateLogLevel(tt.level, tt.field)

			if result.Valid != tt.expected.Valid {
				t.Errorf("ValidateLogLevel() validity = %v, expected %v", result.Valid, tt.expected.Valid)
			}

			if len(result.Errors) != len(tt.expected.Errors) {
				t.Errorf("ValidateLogLevel() errors count = %d, expected %d", len(result.Errors), len(tt.expected.Errors))
			}
		})
	}
}

func TestValidator_GenerateConfigSuggestions(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	tests := []struct {
		name        string
		config      MetricManagerConfig
		minExpected int
	}{
		{
			name: "good config",
			config: MetricManagerConfig{
				Namespace:                 "winpower",
				Subsystem:                 "exporter",
				RequestDurationBuckets:    []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
				CollectionDurationBuckets: []float64{0.1, 0.5, 1, 2, 5, 15, 60},
				APIResponseBuckets:        []float64{0.01, 0.05, 0.1, 0.5, 1, 2},
			},
			minExpected: 0,
		},
		{
			name: "config with few buckets",
			config: MetricManagerConfig{
				Namespace:                 "winpower",
				Subsystem:                 "exporter",
				RequestDurationBuckets:    []float64{0.1, 1, 5},
				CollectionDurationBuckets: []float64{1, 5, 10},
				APIResponseBuckets:        []float64{0.1, 1, 5},
			},
			minExpected: 2, // Should suggest more buckets for request and collection
		},
		{
			name: "config with large first bucket",
			config: MetricManagerConfig{
				Namespace:                 "winpower",
				Subsystem:                 "exporter",
				RequestDurationBuckets:    []float64{1, 5, 10, 30},
				CollectionDurationBuckets: []float64{0.5, 1, 2, 5, 10, 30},
				APIResponseBuckets:        []float64{0.05, 0.1, 0.5, 1},
			},
			minExpected: 1, // Should suggest smaller buckets for request duration
		},
		{
			name: "config with small last bucket",
			config: MetricManagerConfig{
				Namespace:                 "winpower",
				Subsystem:                 "exporter",
				RequestDurationBuckets:    []float64{0.01, 0.05, 0.1, 0.5},
				CollectionDurationBuckets: []float64{0.1, 0.5, 1, 2, 5},
				APIResponseBuckets:        []float64{0.05, 0.1, 0.5, 1},
			},
			minExpected: 1, // Should suggest larger buckets for collection duration
		},
		{
			name: "complex namespace and subsystem",
			config: MetricManagerConfig{
				Namespace:                 "very_long_company_service_name",
				Subsystem:                 "exporter_production_metrics",
				RequestDurationBuckets:    []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
				CollectionDurationBuckets: []float64{0.1, 0.5, 1, 2, 5, 15, 60},
				APIResponseBuckets:        []float64{0.01, 0.05, 0.1, 0.5, 1, 2},
			},
			minExpected: 1, // Should suggest simplifying namespace (subsystem has only 3 parts)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := validator.GenerateConfigSuggestions(tt.config)

			if len(suggestions) < tt.minExpected {
				t.Errorf("GenerateConfigSuggestions() returned %d suggestions, expected at least %d", len(suggestions), tt.minExpected)
			}

			// Verify suggestions are not empty strings
			for _, suggestion := range suggestions {
				if suggestion == "" {
					t.Error("GenerateConfigSuggestions() returned empty suggestion")
				}
			}
		})
	}
}

func TestFormatValidationResult(t *testing.T) {
	tests := []struct {
		name     string
		result   ValidationResult
		expected string
	}{
		{
			name: "valid with no warnings",
			result: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{},
			},
			expected: "✅ Configuration is valid",
		},
		{
			name: "valid with warnings",
			result: ValidationResult{
				Valid:  true,
				Errors: []string{},
				Warns:  []string{"warning1", "warning2"},
			},
			expected: "✅ Configuration is valid (with 2 warnings)",
		},
		{
			name: "invalid with errors and warnings",
			result: ValidationResult{
				Valid:  false,
				Errors: []string{"error1", "error2"},
				Warns:  []string{"warning1"},
			},
			expected: "❌ Configuration is invalid (2 errors, 1 warnings)",
		},
		{
			name: "invalid with only errors",
			result: ValidationResult{
				Valid:  false,
				Errors: []string{"error1"},
				Warns:  []string{},
			},
			expected: "❌ Configuration is invalid (1 errors, 0 warnings)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValidationResult(tt.result)

			if result != tt.expected {
				t.Errorf("FormatValidationResult() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
