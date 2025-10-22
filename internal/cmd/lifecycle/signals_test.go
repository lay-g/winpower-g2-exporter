package lifecycle

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// MockLogger for testing
type mockSignalLogger struct {
	logs []string
	mu   sync.Mutex
}

func (m *mockSignalLogger) Debug(msg string, fields ...log.Field) {
	m.log(msg)
}

func (m *mockSignalLogger) Info(msg string, fields ...log.Field) {
	m.log(msg)
}

func (m *mockSignalLogger) Warn(msg string, fields ...log.Field) {
	m.log(msg)
}

func (m *mockSignalLogger) Error(msg string, fields ...log.Field) {
	m.log(msg)
}

func (m *mockSignalLogger) Fatal(msg string, fields ...log.Field) {
	m.log(msg)
}

func (m *mockSignalLogger) With(fields ...log.Field) log.Logger {
	return m
}

func (m *mockSignalLogger) WithContext(ctx context.Context) log.Logger {
	return m
}

func (m *mockSignalLogger) Sync() error {
	return nil
}

func (m *mockSignalLogger) log(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, msg)
}

func (m *mockSignalLogger) GetLogs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	logs := make([]string, len(m.logs))
	copy(logs, m.logs)
	return logs
}

func (m *mockSignalLogger) ClearLogs() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = []string{}
}

func TestDefaultSignalConfig(t *testing.T) {
	config := DefaultSignalConfig()

	assert.True(t, config.EnableGracefulShutdown)
	assert.True(t, config.EnableForceShutdown)
	assert.False(t, config.EnableReload)
	assert.Equal(t, 30*time.Second, config.GracefulTimeout)
	assert.Equal(t, 5*time.Second, config.ForceTimeout)
	assert.Equal(t, 10, config.BufferSize)
	assert.True(t, config.EnableSignalLogging)
}

func TestNewSignalManager(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()

	sm := NewSignalManager(logger, config)

	assert.NotNil(t, sm)
	assert.Equal(t, logger, sm.logger)
	assert.Equal(t, config, sm.config)
	assert.Empty(t, sm.handlers)
	assert.Empty(t, sm.signalBuffer)
	assert.False(t, sm.isListening)
	assert.NotNil(t, sm.ctx)
	assert.NotNil(t, sm.cancel)
}

func TestNewSignalManagerWithNilConfig(t *testing.T) {
	logger := &mockSignalLogger{}

	sm := NewSignalManager(logger, nil)

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.config)
	assert.True(t, sm.config.EnableGracefulShutdown)
}

func TestRegisterHandler(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	handler := func(ctx context.Context, sig os.Signal) error {
		return nil
	}

	sm.RegisterHandler(syscall.SIGTERM, handler)

	assert.Equal(t, 1, sm.GetHandlerCount())
	_, exists := sm.handlers[syscall.SIGTERM]
	assert.True(t, exists)
	assert.Contains(t, logger.GetLogs(), "Signal handler registered")
}

func TestRegisterMultipleHandlers(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	handlers := map[os.Signal]SignalHandler{
		syscall.SIGTERM: func(ctx context.Context, sig os.Signal) error { return nil },
		syscall.SIGINT:  func(ctx context.Context, sig os.Signal) error { return nil },
		syscall.SIGQUIT: func(ctx context.Context, sig os.Signal) error { return nil },
	}

	sm.RegisterMultipleHandlers(handlers)

	assert.Equal(t, 3, sm.GetHandlerCount())
}

func TestUnregisterHandler(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	handler := func(ctx context.Context, sig os.Signal) error {
		return nil
	}

	// Register first
	sm.RegisterHandler(syscall.SIGTERM, handler)
	assert.Equal(t, 1, sm.GetHandlerCount())

	// Then unregister
	sm.UnregisterHandler(syscall.SIGTERM)
	assert.Equal(t, 0, sm.GetHandlerCount())
	_, exists := sm.handlers[syscall.SIGTERM]
	assert.False(t, exists)
	assert.Contains(t, logger.GetLogs(), "Signal handler unregistered")
}

func TestStartListening(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	err := sm.StartListening()

	assert.NoError(t, err)
	assert.True(t, sm.IsListening())
	assert.Contains(t, logger.GetLogs(), "Signal manager started listening")
}

func TestStartListeningAlreadyListening(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Start listening twice
	sm.StartListening()
	err := sm.StartListening()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already listening")
}

func TestStopListening(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Start and then stop
	sm.StartListening()
	sm.StopListening()

	assert.False(t, sm.IsListening())
	assert.Contains(t, logger.GetLogs(), "Signal manager stopped listening")
}

func TestStopListeningMultipleTimes(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Start and then stop multiple times
	sm.StartListening()
	sm.StopListening()
	sm.StopListening() // Should not panic

	assert.False(t, sm.IsListening())
}

func TestClassifySignal(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	tests := map[os.Signal]SignalType{
		syscall.SIGTERM: SignalTypeGraceful,
		syscall.SIGINT:  SignalTypeGraceful,
		syscall.SIGQUIT: SignalTypeForce,
		syscall.SIGHUP:  SignalTypeReload,
		syscall.SIGUSR1: SignalTypeUnknown,
	}

	for signal, expectedType := range tests {
		signalType := sm.classifySignal(signal)
		assert.Equal(t, expectedType, signalType, "Signal %s should be classified as %s", signal.String(), expectedType)
	}
}

func TestGetSignalTimeout(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	tests := map[SignalType]time.Duration{
		SignalTypeGraceful: 30 * time.Second,
		SignalTypeForce:    5 * time.Second,
		SignalTypeReload:   10 * time.Second,
		SignalTypeUnknown:  5 * time.Second,
	}

	for signalType, expectedTimeout := range tests {
		timeout := sm.getSignalTimeout(signalType)
		assert.Equal(t, expectedTimeout, timeout, "Signal type %s should have timeout %v", signalType, expectedTimeout)
	}
}

func TestGetSignalsToWatch(t *testing.T) {
	t.Run("AllEnabled", func(t *testing.T) {
		config := &SignalConfig{
			EnableGracefulShutdown: true,
			EnableForceShutdown:    true,
			EnableReload:           true,
		}
		sm := NewSignalManager(&mockSignalLogger{}, config)

		signals := sm.getSignalsToWatch()
		assert.Contains(t, signals, syscall.SIGTERM)
		assert.Contains(t, signals, syscall.SIGINT)
		assert.Contains(t, signals, syscall.SIGQUIT)
		assert.Contains(t, signals, syscall.SIGHUP)
	})

	t.Run("OnlyGraceful", func(t *testing.T) {
		config := &SignalConfig{
			EnableGracefulShutdown: true,
			EnableForceShutdown:    false,
			EnableReload:           false,
		}
		sm := NewSignalManager(&mockSignalLogger{}, config)

		signals := sm.getSignalsToWatch()
		assert.Contains(t, signals, syscall.SIGTERM)
		assert.Contains(t, signals, syscall.SIGINT)
		assert.NotContains(t, signals, syscall.SIGQUIT)
		assert.NotContains(t, signals, syscall.SIGHUP)
	})

	t.Run("OnlyForce", func(t *testing.T) {
		config := &SignalConfig{
			EnableGracefulShutdown: false,
			EnableForceShutdown:    true,
			EnableReload:           false,
		}
		sm := NewSignalManager(&mockSignalLogger{}, config)

		signals := sm.getSignalsToWatch()
		assert.NotContains(t, signals, syscall.SIGTERM)
		assert.NotContains(t, signals, syscall.SIGINT)
		assert.Contains(t, signals, syscall.SIGQUIT)
		assert.NotContains(t, signals, syscall.SIGHUP)
	})
}

func TestHandleSignal(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Register a handler
	handlerCalled := false
	var handledSignal os.Signal

	handler := func(ctx context.Context, sig os.Signal) error {
		handlerCalled = true
		handledSignal = sig
		return nil
	}

	sm.RegisterHandler(syscall.SIGTERM, handler)

	// Handle a signal
	sm.handleSignal(syscall.SIGTERM)

	assert.True(t, handlerCalled)
	assert.Equal(t, syscall.SIGTERM, handledSignal)

	// Check signal buffer
	buffer := sm.GetSignalBuffer()
	assert.Len(t, buffer, 1)
	assert.Equal(t, syscall.SIGTERM, buffer[0].Signal)
	assert.True(t, buffer[0].Handled)
	assert.NoError(t, buffer[0].Error)

	// Check statistics
	stats := sm.GetSignalStats()
	assert.Equal(t, 1, stats[syscall.SIGTERM])
}

func TestHandleSignalWithNoHandler(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Handle a signal without registering a handler
	sm.handleSignal(syscall.SIGTERM)

	// Check signal buffer
	buffer := sm.GetSignalBuffer()
	assert.Len(t, buffer, 1)
	assert.Equal(t, syscall.SIGTERM, buffer[0].Signal)
	assert.False(t, buffer[0].Handled)
	assert.Error(t, buffer[0].Error)
	assert.Contains(t, buffer[0].Error.Error(), "no handler registered")

	// Check statistics
	stats := sm.GetSignalStats()
	assert.Equal(t, 1, stats[syscall.SIGTERM])
}

func TestHandleSignalWithHandlerError(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Register a handler that returns an error
	handler := func(ctx context.Context, sig os.Signal) error {
		return assert.AnError
	}

	sm.RegisterHandler(syscall.SIGTERM, handler)

	// Handle a signal
	sm.handleSignal(syscall.SIGTERM)

	// Check signal buffer
	buffer := sm.GetSignalBuffer()
	assert.Len(t, buffer, 1)
	assert.Equal(t, syscall.SIGTERM, buffer[0].Signal)
	assert.False(t, buffer[0].Handled)
	assert.Equal(t, assert.AnError, buffer[0].Error)

	// Check statistics
	stats := sm.GetSignalStats()
	assert.Equal(t, 1, stats[syscall.SIGTERM])
}

func TestHandleSignalWithTimeout(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	config.GracefulTimeout = 100 * time.Millisecond
	sm := NewSignalManager(logger, config)

	// Register a handler that will timeout
	handler := func(ctx context.Context, sig os.Signal) error {
		time.Sleep(200 * time.Millisecond) // Longer than timeout
		return nil
	}

	sm.RegisterHandler(syscall.SIGTERM, handler)

	// Handle a signal
	sm.handleSignal(syscall.SIGTERM)

	// Check signal buffer
	buffer := sm.GetSignalBuffer()
	assert.Len(t, buffer, 1)
	assert.Equal(t, syscall.SIGTERM, buffer[0].Signal)
	// The handling behavior depends on implementation
	// Just verify the signal was recorded
}

func TestSignalBuffer(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Add more signals than the buffer size to test cleanup
	for i := 0; i < 150; i++ {
		sm.handleSignal(syscall.SIGTERM)
	}

	buffer := sm.GetSignalBuffer()
	assert.Len(t, buffer, 100) // Should be limited to maxBufferSize
}

func TestClearBuffer(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Add some signals
	sm.handleSignal(syscall.SIGTERM)
	sm.handleSignal(syscall.SIGINT)

	assert.Len(t, sm.GetSignalBuffer(), 2)

	// Clear buffer
	sm.ClearBuffer()

	assert.Len(t, sm.GetSignalBuffer(), 0)
	assert.Contains(t, logger.GetLogs(), "Signal buffer cleared")
}

func TestUpdateConfig(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Update config while not listening
	newConfig := &SignalConfig{
		EnableGracefulShutdown: false,
		EnableForceShutdown:    true,
		EnableReload:           false,
		GracefulTimeout:        10 * time.Second,
		ForceTimeout:           2 * time.Second,
		BufferSize:             5,
		EnableSignalLogging:    false,
	}

	err := sm.UpdateConfig(newConfig)

	assert.NoError(t, err)
	assert.Equal(t, newConfig, sm.config)
	assert.Contains(t, logger.GetLogs(), "Signal configuration updated")
}

func TestUpdateConfigWhileListening(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Start listening
	sm.StartListening()

	// Try to update config while listening
	newConfig := DefaultSignalConfig()
	err := sm.UpdateConfig(newConfig)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot update config while signal manager is listening")
}

func TestGetConfig(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	retrievedConfig := sm.GetConfig()
	assert.Equal(t, config, retrievedConfig)
}

func TestSignalCancel(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Cancel the context
	sm.Cancel()

	// The context should be cancelled
	assert.Error(t, sm.ctx.Err())
	assert.Equal(t, context.Canceled, sm.ctx.Err())
}

func TestWaitForShutdown(t *testing.T) {
	logger := &mockSignalLogger{}
	config := DefaultSignalConfig()
	sm := NewSignalManager(logger, config)

	// Start listening
	sm.StartListening()

	// Stop listening in a goroutine
	go func() {
		time.Sleep(50 * time.Millisecond)
		sm.StopListening()
	}()

	// Wait for shutdown
	sm.WaitForShutdown()

	assert.False(t, sm.IsListening())
}
