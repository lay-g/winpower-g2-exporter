package server

import "errors"

var (
	// ErrInvalidConfig indicates the configuration is invalid
	ErrInvalidConfig = errors.New("invalid server configuration")

	// ErrServerNotStarted indicates the server has not been started
	ErrServerNotStarted = errors.New("server not started")

	// ErrServerAlreadyRunning indicates the server is already running
	ErrServerAlreadyRunning = errors.New("server already running")

	// ErrMetricsServiceNil indicates the metrics service is nil
	ErrMetricsServiceNil = errors.New("metrics service cannot be nil")

	// ErrHealthServiceNil indicates the health service is nil
	ErrHealthServiceNil = errors.New("health service cannot be nil")

	// ErrLoggerNil indicates the logger is nil
	ErrLoggerNil = errors.New("logger cannot be nil")
)
