package config

import (
	"errors"
	"testing"
)

func TestConfigError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *ConfigError
		want string
	}{
		{
			name: "error without cause",
			err: &ConfigError{
				Module:  "storage",
				Field:   "data_dir",
				Value:   "",
				Message: "data directory is required",
				Code:    ErrCodeRequiredField,
			},
			want: "config error in storage.data_dir: data directory is required (value: , code: REQUIRED_FIELD)",
		},
		{
			name: "error with cause",
			err: &ConfigError{
				Module:  "storage",
				Field:   "data_dir",
				Value:   "/invalid/path",
				Message: "failed to create data directory",
				Code:    ErrCodePermissionError,
				Cause:   errors.New("permission denied"),
			},
			want: "config error in storage.data_dir: failed to create data directory (value: /invalid/path, code: PERMISSION_ERROR): permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("ConfigError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &ConfigError{
		Module:  "test",
		Field:   "field",
		Value:   "value",
		Message: "message",
		Code:    "CODE",
		Cause:   cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("ConfigError.Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Test without cause
	errWithoutCause := &ConfigError{
		Module:  "test",
		Field:   "field",
		Value:   "value",
		Message: "message",
		Code:    "CODE",
	}

	if unwrapped := errWithoutCause.Unwrap(); unwrapped != nil {
		t.Errorf("ConfigError.Unwrap() without cause = %v, want nil", unwrapped)
	}
}

func TestConfigError_Is(t *testing.T) {
	err1 := &ConfigError{
		Module: "storage",
		Field:  "data_dir",
		Code:   ErrCodeRequiredField,
	}

	err2 := &ConfigError{
		Module: "storage",
		Field:  "data_dir",
		Code:   ErrCodeRequiredField,
	}

	err3 := &ConfigError{
		Module: "winpower",
		Field:  "url",
		Code:   ErrCodeInvalidFormat,
	}

	// Test same errors
	if !err1.Is(err2) {
		t.Errorf("ConfigError.Is() should return true for same module and code")
	}

	// Test different errors
	if err1.Is(err3) {
		t.Errorf("ConfigError.Is() should return false for different module or code")
	}

	// Test non-ConfigError error
	otherErr := errors.New("other error")
	if err1.Is(otherErr) {
		t.Errorf("ConfigError.Is() should return false for non-ConfigError")
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "port",
		Value:   -1,
		Message: "port must be positive",
		Code:    ErrCodeInvalidRange,
	}

	want := "validation error for field 'port': port must be positive (value: -1, code: INVALID_RANGE)"
	if got := err.Error(); got != want {
		t.Errorf("ValidationError.Error() = %v, want %v", got, want)
	}
}

func TestValidationError_Is(t *testing.T) {
	err1 := &ValidationError{
		Field: "port",
		Code:  ErrCodeInvalidRange,
	}

	err2 := &ValidationError{
		Field: "port",
		Code:  ErrCodeInvalidRange,
	}

	err3 := &ValidationError{
		Field: "timeout",
		Code:  ErrCodeInvalidType,
	}

	// Test same errors
	if !err1.Is(err2) {
		t.Errorf("ValidationError.Is() should return true for same field and code")
	}

	// Test different errors
	if err1.Is(err3) {
		t.Errorf("ValidationError.Is() should return false for different field or code")
	}

	// Test non-ValidationError error
	otherErr := errors.New("other error")
	if err1.Is(otherErr) {
		t.Errorf("ValidationError.Is() should return false for non-ValidationError")
	}
}

func TestNewConfigError(t *testing.T) {
	cause := errors.New("test cause")
	err := NewConfigError("test_module", "test_field", "test_value", "test message", "TEST_CODE", cause)

	if err.Module != "test_module" {
		t.Errorf("NewConfigError() Module = %v, want %v", err.Module, "test_module")
	}
	if err.Field != "test_field" {
		t.Errorf("NewConfigError() Field = %v, want %v", err.Field, "test_field")
	}
	if err.Value != "test_value" {
		t.Errorf("NewConfigError() Value = %v, want %v", err.Value, "test_value")
	}
	if err.Message != "test message" {
		t.Errorf("NewConfigError() Message = %v, want %v", err.Message, "test message")
	}
	if err.Code != "TEST_CODE" {
		t.Errorf("NewConfigError() Code = %v, want %v", err.Code, "TEST_CODE")
	}
	if err.Cause != cause {
		t.Errorf("NewConfigError() Cause = %v, want %v", err.Cause, cause)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("test_field", "test_value", "test message", "TEST_CODE")

	if err.Field != "test_field" {
		t.Errorf("NewValidationError() Field = %v, want %v", err.Field, "test_field")
	}
	if err.Value != "test_value" {
		t.Errorf("NewValidationError() Value = %v, want %v", err.Value, "test_value")
	}
	if err.Message != "test message" {
		t.Errorf("NewValidationError() Message = %v, want %v", err.Message, "test message")
	}
	if err.Code != "TEST_CODE" {
		t.Errorf("NewValidationError() Code = %v, want %v", err.Code, "TEST_CODE")
	}
}

func TestIsConfigError(t *testing.T) {
	configErr := &ConfigError{
		Module: "test",
		Field:  "field",
		Code:   "CODE",
	}

	validationErr := &ValidationError{
		Field: "field",
		Code:  "CODE",
	}

	// Test ConfigError
	if err, ok := IsConfigError(configErr); !ok {
		t.Errorf("IsConfigError() should return true for ConfigError")
	} else if err != configErr {
		t.Errorf("IsConfigError() should return the original error")
	}

	// Test ValidationError
	if _, ok := IsConfigError(validationErr); ok {
		t.Errorf("IsConfigError() should return false for ValidationError")
	}

	// Test nil error
	if _, ok := IsConfigError(nil); ok {
		t.Errorf("IsConfigError() should return false for nil error")
	}
}

func TestIsValidationError(t *testing.T) {
	configErr := &ConfigError{
		Module: "test",
		Field:  "field",
		Code:   "CODE",
	}

	validationErr := &ValidationError{
		Field: "field",
		Code:  "CODE",
	}

	// Test ValidationError
	if err, ok := IsValidationError(validationErr); !ok {
		t.Errorf("IsValidationError() should return true for ValidationError")
	} else if err != validationErr {
		t.Errorf("IsValidationError() should return the original error")
	}

	// Test ConfigError
	if _, ok := IsValidationError(configErr); ok {
		t.Errorf("IsValidationError() should return false for ConfigError")
	}

	// Test nil error
	if _, ok := IsValidationError(nil); ok {
		t.Errorf("IsValidationError() should return false for nil error")
	}
}

func TestAs(t *testing.T) {
	configErr := &ConfigError{
		Module: "test",
		Field:  "field",
		Code:   "CODE",
	}

	validationErr := &ValidationError{
		Field: "field",
		Code:  "CODE",
	}

	// Test ConfigError
	var targetConfigErr *ConfigError
	if !As(configErr, &targetConfigErr) {
		t.Errorf("As() should return true for ConfigError")
	}
	if targetConfigErr != configErr {
		t.Errorf("As() should set target to the original ConfigError")
	}

	// Test ValidationError
	var targetValidationErr *ValidationError
	if !As(validationErr, &targetValidationErr) {
		t.Errorf("As() should return true for ValidationError")
	}
	if targetValidationErr != validationErr {
		t.Errorf("As() should set target to the original ValidationError")
	}

	// Test mismatched type
	var targetConfigErr2 *ConfigError
	if As(validationErr, &targetConfigErr2) {
		t.Errorf("As() should return false for mismatched type")
	}
}

func TestErrorCodes(t *testing.T) {
	// Test that all error codes are defined and non-empty
	errorCodes := []string{
		ErrCodeRequiredField,
		ErrCodeInvalidType,
		ErrCodeInvalidRange,
		ErrCodeInvalidFormat,
		ErrCodeFileNotFound,
		ErrCodeParseError,
		ErrCodePermissionError,
		ErrCodeEnvVariableMissing,
		ErrCodeEnvVariableInvalid,
		ErrCodeModuleLoadError,
		ErrCodeDependencyError,
	}

	for _, code := range errorCodes {
		if code == "" {
			t.Errorf("Error code should not be empty")
		}
	}
}

// Test error types implement error interface
func TestErrorInterfaceImplementation(t *testing.T) {
	var _ error = &ConfigError{}
	var _ error = &ValidationError{}

	// Test that they can be used as error interface
	err := &ConfigError{
		Module:  "test",
		Field:   "field",
		Value:   "value",
		Message: "message",
		Code:    "CODE",
	}

	var errorInterface error = err
	if errorInterface.Error() != err.Error() {
		t.Errorf("ConfigError should properly implement error interface")
	}

	validationErr := &ValidationError{
		Field:   "field",
		Value:   "value",
		Message: "message",
		Code:    "CODE",
	}

	errorInterface = validationErr
	if errorInterface.Error() != validationErr.Error() {
		t.Errorf("ValidationError should properly implement error interface")
	}
}
