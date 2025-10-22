package config

import (
	"fmt"
)

// ConfigError 配置错误类型
// 这个结构体提供了配置相关的详细错误信息，包括模块、字段、值等上下文。
// 它包装了底层错误，提供了更好的错误追踪和调试能力。
type ConfigError struct {
	Module  string      // 模块名称，如 "storage", "winpower" 等
	Field   string      // 字段名称，如 "data_dir", "url" 等
	Value   interface{} // 导致错误的字段值
	Message string      // 错误消息
	Code    string      // 错误代码
	Cause   error       // 底层错误原因（可选）
}

// Error 返回错误的字符串表示
// 实现了 error 接口
func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("config error in %s.%s: %s (value: %v, code: %s): %v",
			e.Module, e.Field, e.Message, e.Value, e.Code, e.Cause)
	}
	return fmt.Sprintf("config error in %s.%s: %s (value: %v, code: %s)",
		e.Module, e.Field, e.Message, e.Value, e.Code)
}

// Unwrap 返回底层错误
// 支持错误链 unwrap 操作，可以使用 errors.Is 和 errors.As
func (e *ConfigError) Unwrap() error {
	return e.Cause
}

// Is 检查错误是否匹配目标错误
// 实现了 errors.Is 接口，支持错误比较
func (e *ConfigError) Is(target error) bool {
	if t, ok := target.(*ConfigError); ok {
		return e.Code == t.Code && e.Module == t.Module
	}
	return false
}

// ValidationError 验证错误类型
// 这个结构体专门用于配置验证相关的错误，提供了字段级别的错误信息。
type ValidationError struct {
	Field   string      // 字段名称
	Value   interface{} // 导致错误的字段值
	Message string      // 错误消息
	Code    string      // 错误代码
}

// Error 返回错误的字符串表示
// 实现了 error 接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s (value: %v, code: %s)",
		e.Field, e.Message, e.Value, e.Code)
}

// Is 检查错误是否匹配目标错误
// 实现了 errors.Is 接口，支持错误比较
func (e *ValidationError) Is(target error) bool {
	if t, ok := target.(*ValidationError); ok {
		return e.Code == t.Code && e.Field == t.Field
	}
	return false
}

// 错误代码常量
// 定义了所有可能的配置错误代码，用于错误分类和处理
const (
	// ErrCodeRequiredField 必需字段缺失错误
	ErrCodeRequiredField = "REQUIRED_FIELD"

	// ErrCodeInvalidType 类型错误
	ErrCodeInvalidType = "INVALID_TYPE"

	// ErrCodeInvalidRange 数值范围错误
	ErrCodeInvalidRange = "INVALID_RANGE"

	// ErrCodeInvalidFormat 格式错误（如URL格式、文件路径格式等）
	ErrCodeInvalidFormat = "INVALID_FORMAT"

	// ErrCodeFileNotFound 配置文件不存在错误
	ErrCodeFileNotFound = "FILE_NOT_FOUND"

	// ErrCodeParseError 配置文件解析错误
	ErrCodeParseError = "PARSE_ERROR"

	// ErrCodePermissionError 文件权限错误
	ErrCodePermissionError = "PERMISSION_ERROR"

	// ErrCodeEnvVariableMissing 环境变量缺失错误
	ErrCodeEnvVariableMissing = "ENV_VARIABLE_MISSING"

	// ErrCodeEnvVariableInvalid 环境变量值无效错误
	ErrCodeEnvVariableInvalid = "ENV_VARIABLE_INVALID"

	// ErrCodeModuleLoadError 模块加载错误
	ErrCodeModuleLoadError = "MODULE_LOAD_ERROR"

	// ErrCodeDependencyError 依赖配置错误
	ErrCodeDependencyError = "DEPENDENCY_ERROR"
)

// NewConfigError 创建新的配置错误
// 这是一个便利函数，用于创建标准的配置错误
func NewConfigError(module, field string, value interface{}, message, code string, cause error) *ConfigError {
	return &ConfigError{
		Module:  module,
		Field:   field,
		Value:   value,
		Message: message,
		Code:    code,
		Cause:   cause,
	}
}

// NewValidationError 创建新的验证错误
// 这是一个便利函数，用于创建标准的验证错误
func NewValidationError(field string, value interface{}, message, code string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Code:    code,
	}
}

// IsConfigError 检查错误是否为配置错误
// 这是一个便利函数，用于错误类型检查
func IsConfigError(err error) (*ConfigError, bool) {
	var configErr *ConfigError
	if ok := As(err, &configErr); ok {
		return configErr, true
	}
	return nil, false
}

// IsValidationError 检查错误是否为验证错误
// 这是一个便利函数，用于错误类型检查
func IsValidationError(err error) (*ValidationError, bool) {
	var validationErr *ValidationError
	if ok := As(err, &validationErr); ok {
		return validationErr, true
	}
	return nil, false
}

// As 是 errors.As 的别名，用于错误类型断言
// 提供了一个便捷的错误类型检查方法
func As(err error, target interface{}) bool {
	// 这里使用标准库的 errors.As 功能
	// 为了避免循环依赖，我们使用简单的类型检查
	switch t := target.(type) {
	case **ConfigError:
		if configErr, ok := err.(*ConfigError); ok {
			*t = configErr
			return true
		}
	case **ValidationError:
		if validationErr, ok := err.(*ValidationError); ok {
			*t = validationErr
			return true
		}
	}
	return false
}
