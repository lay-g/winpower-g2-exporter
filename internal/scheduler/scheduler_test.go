package scheduler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// MockCollector is a mock implementation of CollectorInterface for testing.
type MockCollector struct {
	mu sync.Mutex

	// CollectDeviceDataFunc is the function to call when CollectDeviceData is invoked
	CollectDeviceDataFunc func(ctx context.Context) (*CollectionResult, error)

	// Tracking
	CollectDeviceDataCalls []context.Context
	CallCount              int
}

// CollectDeviceData implements CollectorInterface.
func (m *MockCollector) CollectDeviceData(ctx context.Context) (*CollectionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallCount++
	m.CollectDeviceDataCalls = append(m.CollectDeviceDataCalls, ctx)

	if m.CollectDeviceDataFunc != nil {
		return m.CollectDeviceDataFunc(ctx)
	}

	return &CollectionResult{
		Success:     true,
		DeviceCount: 0,
	}, nil
}

// GetCallCount returns the number of times CollectDeviceData was called.
func (m *MockCollector) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.CallCount
}

// MockLogger is a mock implementation of Logger for testing.
type MockLogger struct {
	mu sync.Mutex

	// Collected logs
	InfoLogs  []LogEntry
	ErrorLogs []LogEntry
	WarnLogs  []LogEntry
	DebugLogs []LogEntry
}

// LogEntry represents a single log entry.
type LogEntry struct {
	Message string
	Fields  []interface{}
}

// Info implements Logger.
func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.InfoLogs = append(m.InfoLogs, LogEntry{Message: msg, Fields: fields})
}

// Error implements Logger.
func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorLogs = append(m.ErrorLogs, LogEntry{Message: msg, Fields: fields})
}

// Warn implements Logger.
func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WarnLogs = append(m.WarnLogs, LogEntry{Message: msg, Fields: fields})
}

// Debug implements Logger.
func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DebugLogs = append(m.DebugLogs, LogEntry{Message: msg, Fields: fields})
}

// HasInfoLog checks if an info log with the given message exists.
func (m *MockLogger) HasInfoLog(msg string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, log := range m.InfoLogs {
		if log.Message == msg {
			return true
		}
	}
	return false
}

// HasErrorLog checks if an error log with the given message exists.
func (m *MockLogger) HasErrorLog(msg string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, log := range m.ErrorLogs {
		if log.Message == msg {
			return true
		}
	}
	return false
}

// HasWarnLog checks if a warn log with the given message exists.
func (m *MockLogger) HasWarnLog(msg string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, log := range m.WarnLogs {
		if log.Message == msg {
			return true
		}
	}
	return false
}

func TestNewDefaultScheduler(t *testing.T) {
	config := DefaultConfig()
	collector := &MockCollector{}
	logger := &MockLogger{}

	tests := []struct {
		name      string
		config    *Config
		collector CollectorInterface
		logger    Logger
		wantErr   error
	}{
		{
			name:      "valid scheduler",
			config:    config,
			collector: collector,
			logger:    logger,
			wantErr:   nil,
		},
		{
			name:      "nil config",
			config:    nil,
			collector: collector,
			logger:    logger,
			wantErr:   ErrNilConfig,
		},
		{
			name:      "nil collector",
			config:    config,
			collector: nil,
			logger:    logger,
			wantErr:   ErrNilCollector,
		},
		{
			name:      "nil logger",
			config:    config,
			collector: collector,
			logger:    nil,
			wantErr:   ErrNilLogger,
		},
		{
			name: "invalid config",
			config: &Config{
				CollectionInterval:      0,
				GracefulShutdownTimeout: 5 * time.Second,
			},
			collector: collector,
			logger:    logger,
			wantErr:   errors.New("collection_interval must be positive"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler, err := NewDefaultScheduler(tt.config, tt.collector, tt.logger)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("NewDefaultScheduler() expected error, got nil")
					return
				}
				// Check error message prefix for validation errors
				if tt.wantErr != ErrNilConfig && tt.wantErr != ErrNilCollector && tt.wantErr != ErrNilLogger {
					if err.Error()[:len(tt.wantErr.Error())] != tt.wantErr.Error() {
						t.Errorf("NewDefaultScheduler() error = %v, want error containing %v", err, tt.wantErr)
					}
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("NewDefaultScheduler() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewDefaultScheduler() unexpected error = %v", err)
				return
			}

			if scheduler == nil {
				t.Error("NewDefaultScheduler() returned nil scheduler")
				return
			}

			if scheduler.config != tt.config {
				t.Error("NewDefaultScheduler() config not set correctly")
			}

			if scheduler.collector != tt.collector {
				t.Error("NewDefaultScheduler() collector not set correctly")
			}

			if scheduler.logger != tt.logger {
				t.Error("NewDefaultScheduler() logger not set correctly")
			}

			if scheduler.IsRunning() {
				t.Error("NewDefaultScheduler() scheduler should not be running initially")
			}
		})
	}
}

func TestDefaultScheduler_Start(t *testing.T) {
	t.Run("successful start", func(t *testing.T) {
		config := DefaultConfig()
		collector := &MockCollector{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Start(ctx)
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		if !scheduler.IsRunning() {
			t.Error("Start() scheduler should be running")
		}

		// Clean up
		_ = scheduler.Stop(context.Background())
	})

	t.Run("double start error", func(t *testing.T) {
		config := DefaultConfig()
		collector := &MockCollector{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Start(ctx)
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		// Try to start again
		err = scheduler.Start(ctx)
		if !errors.Is(err, ErrAlreadyRunning) {
			t.Errorf("Start() error = %v, want %v", err, ErrAlreadyRunning)
		}

		// Clean up
		_ = scheduler.Stop(context.Background())
	})

	t.Run("logs startup", func(t *testing.T) {
		config := DefaultConfig()
		collector := &MockCollector{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Start(ctx)
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		if !logger.HasInfoLog("scheduler started") {
			t.Error("Start() should log 'scheduler started'")
		}

		// Clean up
		_ = scheduler.Stop(context.Background())
	})
}

func TestDefaultScheduler_Stop(t *testing.T) {
	t.Run("successful stop", func(t *testing.T) {
		config := DefaultConfig()
		collector := &MockCollector{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Start(ctx)
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		err = scheduler.Stop(ctx)
		if err != nil {
			t.Errorf("Stop() error = %v", err)
		}

		if scheduler.IsRunning() {
			t.Error("Stop() scheduler should not be running")
		}
	})

	t.Run("stop without start error", func(t *testing.T) {
		config := DefaultConfig()
		collector := &MockCollector{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Stop(ctx)
		if !errors.Is(err, ErrNotRunning) {
			t.Errorf("Stop() error = %v, want %v", err, ErrNotRunning)
		}
	})

	t.Run("logs shutdown", func(t *testing.T) {
		config := DefaultConfig()
		collector := &MockCollector{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Start(ctx)
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		err = scheduler.Stop(ctx)
		if err != nil {
			t.Errorf("Stop() error = %v", err)
		}

		if !logger.HasInfoLog("scheduler stopped gracefully") {
			t.Error("Stop() should log 'scheduler stopped gracefully'")
		}
	})

	t.Run("stop with context deadline", func(t *testing.T) {
		config := DefaultConfig()
		collector := &MockCollector{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Start(ctx)
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		// Stop with a short deadline (should still succeed)
		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err = scheduler.Stop(stopCtx)
		if err != nil {
			t.Errorf("Stop() error = %v", err)
		}
	})
}

func TestDefaultScheduler_Collection(t *testing.T) {
	t.Run("collects data at interval", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      1 * time.Second, // Short interval for testing
			GracefulShutdownTimeout: 5 * time.Second,
		}
		collector := &MockCollector{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Start(ctx)
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		// Wait for at least 2 collections
		time.Sleep(2500 * time.Millisecond)

		callCount := collector.GetCallCount()
		if callCount < 2 {
			t.Errorf("Expected at least 2 collections, got %d", callCount)
		}

		// Clean up
		_ = scheduler.Stop(context.Background())
	})

	t.Run("continues after collection error", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      1 * time.Second,
			GracefulShutdownTimeout: 5 * time.Second,
		}

		collector := &MockCollector{}
		// Make the first call fail
		firstCall := true
		collector.CollectDeviceDataFunc = func(ctx context.Context) (*CollectionResult, error) {
			if firstCall {
				firstCall = false
				return nil, errors.New("test error")
			}
			return &CollectionResult{Success: true, DeviceCount: 1}, nil
		}

		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Start(ctx)
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		// Wait for multiple collections
		time.Sleep(2500 * time.Millisecond)

		callCount := collector.GetCallCount()
		if callCount < 2 {
			t.Errorf("Expected at least 2 collections (error recovery), got %d", callCount)
		}

		// Should have logged the error
		if !logger.HasErrorLog("collection failed") {
			t.Error("Should have logged collection error")
		}

		// Clean up
		_ = scheduler.Stop(context.Background())
	})

	t.Run("logs collection results", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      1 * time.Second,
			GracefulShutdownTimeout: 5 * time.Second,
		}

		collector := &MockCollector{}
		collector.CollectDeviceDataFunc = func(ctx context.Context) (*CollectionResult, error) {
			return &CollectionResult{
				Success:     true,
				DeviceCount: 3,
			}, nil
		}

		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Start(ctx)
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		// Wait for at least one collection
		time.Sleep(1500 * time.Millisecond)

		if !logger.HasInfoLog("collection completed") {
			t.Error("Should have logged 'collection completed'")
		}

		// Clean up
		_ = scheduler.Stop(context.Background())
	})

	t.Run("handles nil collection result", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      1 * time.Second,
			GracefulShutdownTimeout: 5 * time.Second,
		}

		collector := &MockCollector{}
		collector.CollectDeviceDataFunc = func(ctx context.Context) (*CollectionResult, error) {
			return nil, nil
		}

		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Start(ctx)
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		// Wait for at least one collection
		time.Sleep(1500 * time.Millisecond)

		if !logger.HasWarnLog("collection returned nil result") {
			t.Error("Should have logged warning for nil result")
		}

		// Clean up
		_ = scheduler.Stop(context.Background())
	})

	t.Run("handles unsuccessful collection", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      1 * time.Second,
			GracefulShutdownTimeout: 5 * time.Second,
		}

		collector := &MockCollector{}
		collector.CollectDeviceDataFunc = func(ctx context.Context) (*CollectionResult, error) {
			return &CollectionResult{
				Success:      false,
				DeviceCount:  0,
				ErrorMessage: "device connection failed",
			}, nil
		}

		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("NewDefaultScheduler() error = %v", err)
		}

		ctx := context.Background()
		err = scheduler.Start(ctx)
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		// Wait for at least one collection
		time.Sleep(1500 * time.Millisecond)

		if !logger.HasErrorLog("collection unsuccessful") {
			t.Error("Should have logged error for unsuccessful collection")
		}

		// Clean up
		_ = scheduler.Stop(context.Background())
	})
}

func TestDefaultScheduler_IsRunning(t *testing.T) {
	config := DefaultConfig()
	collector := &MockCollector{}
	logger := &MockLogger{}

	scheduler, err := NewDefaultScheduler(config, collector, logger)
	if err != nil {
		t.Fatalf("NewDefaultScheduler() error = %v", err)
	}

	// Initially not running
	if scheduler.IsRunning() {
		t.Error("IsRunning() should be false initially")
	}

	// Start and check
	ctx := context.Background()
	err = scheduler.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	if !scheduler.IsRunning() {
		t.Error("IsRunning() should be true after Start()")
	}

	// Stop and check
	err = scheduler.Stop(ctx)
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	if scheduler.IsRunning() {
		t.Error("IsRunning() should be false after Stop()")
	}
}
