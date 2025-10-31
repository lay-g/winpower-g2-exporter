package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConfigError
		expected string
	}{
		{
			name: "error with field and underlying error",
			err: &ConfigError{
				Field:   "server.port",
				Message: "invalid value",
				Err:     errors.New("must be positive"),
			},
			expected: `config error in field "server.port": invalid value: must be positive`,
		},
		{
			name: "error with field only",
			err: &ConfigError{
				Field:   "server.port",
				Message: "invalid value",
			},
			expected: `config error in field "server.port": invalid value`,
		},
		{
			name: "error with underlying error only",
			err: &ConfigError{
				Message: "configuration failed",
				Err:     errors.New("file not found"),
			},
			expected: "config error: configuration failed: file not found",
		},
		{
			name: "error with message only",
			err: &ConfigError{
				Message: "configuration failed",
			},
			expected: "config error: configuration failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestConfigError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &ConfigError{
		Message: "config error",
		Err:     underlying,
	}

	assert.Equal(t, underlying, errors.Unwrap(err))
	assert.True(t, errors.Is(err, underlying))
}

func TestNewConfigError(t *testing.T) {
	underlying := errors.New("test error")
	err := NewConfigError("test.field", "test message", underlying)

	assert.Equal(t, "test.field", err.Field)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, underlying, err.Err)
}

func TestConfigErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrConfigNotFound", ErrConfigNotFound},
		{"ErrInvalidConfig", ErrInvalidConfig},
		{"ErrConfigValidation", ErrConfigValidation},
		{"ErrConfigLoad", ErrConfigLoad},
		{"ErrConfigParse", ErrConfigParse},
		{"ErrFlagBinding", ErrFlagBinding},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.err)
			assert.NotEmpty(t, tt.err.Error())
		})
	}
}
