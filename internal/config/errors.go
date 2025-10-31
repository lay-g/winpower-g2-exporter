package config

import (
	"errors"
	"fmt"
)

var (
	// ErrConfigNotFound 配置文件未找到
	ErrConfigNotFound = errors.New("config file not found")

	// ErrInvalidConfig 配置无效
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrConfigValidation 配置验证失败
	ErrConfigValidation = errors.New("configuration validation failed")

	// ErrConfigLoad 配置加载失败
	ErrConfigLoad = errors.New("failed to load configuration")

	// ErrConfigParse 配置解析失败
	ErrConfigParse = errors.New("failed to parse configuration")

	// ErrFlagBinding 命令行参数绑定失败
	ErrFlagBinding = errors.New("failed to bind command line flags")
)

// ConfigError 配置错误类型，提供详细的错误上下文
type ConfigError struct {
	// Field 错误相关的配置字段
	Field string

	// Message 错误消息
	Message string

	// Err 底层错误
	Err error
}

// Error 实现 error 接口
func (e *ConfigError) Error() string {
	if e.Field != "" {
		if e.Err != nil {
			return fmt.Sprintf("config error in field %q: %s: %v", e.Field, e.Message, e.Err)
		}
		return fmt.Sprintf("config error in field %q: %s", e.Field, e.Message)
	}

	if e.Err != nil {
		return fmt.Sprintf("config error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

// Unwrap 返回底层错误，支持 errors.Is 和 errors.As
func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError 创建一个新的配置错误
func NewConfigError(field, message string, err error) *ConfigError {
	return &ConfigError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}
