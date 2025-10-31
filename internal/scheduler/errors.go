package scheduler

import "errors"

var (
	// ErrAlreadyRunning is returned when attempting to start an already running scheduler.
	ErrAlreadyRunning = errors.New("scheduler is already running")

	// ErrNotRunning is returned when attempting to stop a scheduler that is not running.
	ErrNotRunning = errors.New("scheduler is not running")

	// ErrShutdownTimeout is returned when graceful shutdown exceeds the configured timeout.
	ErrShutdownTimeout = errors.New("scheduler shutdown timeout exceeded")

	// ErrNilCollector is returned when a nil collector is provided.
	ErrNilCollector = errors.New("collector cannot be nil")

	// ErrNilLogger is returned when a nil logger is provided.
	ErrNilLogger = errors.New("logger cannot be nil")

	// ErrNilConfig is returned when a nil config is provided.
	ErrNilConfig = errors.New("config cannot be nil")
)
