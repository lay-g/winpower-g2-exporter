package lifecycle

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// SignalType represents different types of signals
type SignalType string

const (
	SignalTypeGraceful SignalType = "graceful"
	SignalTypeForce    SignalType = "force"
	SignalTypeReload   SignalType = "reload"
	SignalTypeUnknown  SignalType = "unknown"
)

// SignalHandler defines the function signature for signal handlers
type SignalHandler func(ctx context.Context, sig os.Signal) error

// SignalConfig contains configuration for signal handling
type SignalConfig struct {
	EnableGracefulShutdown bool
	EnableForceShutdown    bool
	EnableReload           bool
	GracefulTimeout        time.Duration
	ForceTimeout           time.Duration
	BufferSize             int
	EnableSignalLogging    bool
}

// DefaultSignalConfig returns a default signal configuration
func DefaultSignalConfig() *SignalConfig {
	return &SignalConfig{
		EnableGracefulShutdown: true,
		EnableForceShutdown:    true,
		EnableReload:           false, // Reload functionality can be added later
		GracefulTimeout:        30 * time.Second,
		ForceTimeout:           5 * time.Second,
		BufferSize:             10,
		EnableSignalLogging:    true,
	}
}

// SignalInfo contains information about a received signal
type SignalInfo struct {
	Signal     os.Signal
	Type       SignalType
	Timestamp  time.Time
	ProcessPID int
	Handled    bool
	Error      error
	Duration   time.Duration
	TimeoutHit bool
}

// SignalManager manages signal handling for the application
type SignalManager struct {
	// Dependencies
	logger    log.Logger
	config    *SignalConfig

	// Signal handling
	signalChan    chan os.Signal
	handlers      map[os.Signal]SignalHandler
	handlerMutex  sync.RWMutex
	signalBuffer  []SignalInfo
	bufferMutex   sync.RWMutex
	maxBufferSize int

	// State management
	isListening   bool
	listenMutex   sync.RWMutex
	shutdownChan  chan struct{}
	shutdownOnce  sync.Once
	listenWG      sync.WaitGroup

	// Context management
	ctx    context.Context
	cancel context.CancelFunc

	// Statistics
	signalsReceived map[os.Signal]int
	statsMutex      sync.RWMutex
}

// NewSignalManager creates a new signal manager
func NewSignalManager(logger log.Logger, config *SignalConfig) *SignalManager {
	if config == nil {
		config = DefaultSignalConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &SignalManager{
		logger:         logger,
		config:         config,
		signalChan:     make(chan os.Signal, config.BufferSize),
		handlers:       make(map[os.Signal]SignalHandler),
		signalBuffer:   make([]SignalInfo, 0),
		maxBufferSize:  100, // Keep last 100 signals in buffer
		shutdownChan:   make(chan struct{}),
		ctx:            ctx,
		cancel:         cancel,
		signalsReceived: make(map[os.Signal]int),
	}
}

// RegisterHandler registers a handler for a specific signal
func (sm *SignalManager) RegisterHandler(sig os.Signal, handler SignalHandler) {
	sm.handlerMutex.Lock()
	defer sm.handlerMutex.Unlock()

	sm.handlers[sig] = handler
	sm.logger.Debug("Signal handler registered",
		zap.String("signal", sig.String()),
	)
}

// RegisterMultipleHandlers registers multiple signal handlers
func (sm *SignalManager) RegisterMultipleHandlers(handlers map[os.Signal]SignalHandler) {
	sm.handlerMutex.Lock()
	defer sm.handlerMutex.Unlock()

	for sig, handler := range handlers {
		sm.handlers[sig] = handler
		sm.logger.Debug("Signal handler registered",
			zap.String("signal", sig.String()),
		)
	}
}

// UnregisterHandler removes a handler for a specific signal
func (sm *SignalManager) UnregisterHandler(sig os.Signal) {
	sm.handlerMutex.Lock()
	defer sm.handlerMutex.Unlock()

	delete(sm.handlers, sig)
	sm.logger.Debug("Signal handler unregistered",
		zap.String("signal", sig.String()),
	)
}

// StartListening starts listening for signals
func (sm *SignalManager) StartListening() error {
	sm.listenMutex.Lock()
	defer sm.listenMutex.Unlock()

	if sm.isListening {
		return fmt.Errorf("signal manager is already listening")
	}

	sm.isListening = true

	// Register for signals
	signalsToWatch := sm.getSignalsToWatch()
	signal.Notify(sm.signalChan, signalsToWatch...)

	sm.logger.Info("Signal manager started listening",
		zap.Strings("signals", sm.getSignalNames(signalsToWatch)),
		zap.Int("buffer_size", sm.config.BufferSize),
	)

	// Start the signal handling goroutine
	sm.listenWG.Add(1)
	go sm.signalHandlingLoop()

	return nil
}

// StopListening stops listening for signals
func (sm *SignalManager) StopListening() {
	sm.shutdownOnce.Do(func() {
		sm.listenMutex.Lock()
		sm.isListening = false
		sm.listenMutex.Unlock()

		// Stop receiving signals
		signal.Stop(sm.signalChan)

		// Signal the handling loop to stop
		close(sm.shutdownChan)

		// Wait for the handling goroutine to finish
		sm.listenWG.Wait()

		sm.logger.Info("Signal manager stopped listening")
	})
}

// signalHandlingLoop is the main signal handling loop
func (sm *SignalManager) signalHandlingLoop() {
	defer sm.listenWG.Done()

	for {
		select {
		case sig := <-sm.signalChan:
			sm.handleSignal(sig)

		case <-sm.shutdownChan:
			sm.logger.Info("Signal manager shutting down")
			return

		case <-sm.ctx.Done():
			sm.logger.Info("Signal manager context cancelled")
			return
		}
	}
}

// handleSignal handles a received signal
func (sm *SignalManager) handleSignal(sig os.Signal) {
	startTime := time.Now()

	signalInfo := SignalInfo{
		Signal:     sig,
		Timestamp:  startTime,
		ProcessPID: os.Getpid(),
		Handled:    false,
	}

	// Update statistics
	sm.statsMutex.Lock()
	sm.signalsReceived[sig]++
	sm.statsMutex.Unlock()

	// Log the signal
	if sm.config.EnableSignalLogging {
		sm.logger.Info("Signal received",
			zap.String("signal", sig.String()),
			zap.Int("pid", signalInfo.ProcessPID),
			zap.Time("timestamp", startTime),
		)
	}

	// Determine signal type
	signalType := sm.classifySignal(sig)
	signalInfo.Type = signalType

	// Get handler
	sm.handlerMutex.RLock()
	handler, exists := sm.handlers[sig]
	sm.handlerMutex.RUnlock()

	if !exists {
		sm.logger.Warn("No handler registered for signal",
			zap.String("signal", sig.String()),
		)
		signalInfo.Error = fmt.Errorf("no handler registered")
		sm.addToBuffer(signalInfo)
		return
	}

	// Handle the signal
	sm.logger.Info("Handling signal",
		zap.String("signal", sig.String()),
		zap.String("type", string(signalType)),
	)

	// Create context with timeout based on signal type
	timeout := sm.getSignalTimeout(signalType)
	ctx, cancel := context.WithTimeout(sm.ctx, timeout)
	defer cancel()

	// Execute handler
	err := handler(ctx, sig)
	signalInfo.Duration = time.Since(startTime)

	if err != nil {
		sm.logger.Error("Signal handler failed",
			zap.String("signal", sig.String()),
			zap.Duration("duration", signalInfo.Duration),
			zap.Error(err),
		)
		signalInfo.Error = err
		if ctx.Err() == context.DeadlineExceeded {
			signalInfo.TimeoutHit = true
		}
	} else {
		signalInfo.Handled = true
		sm.logger.Info("Signal handled successfully",
			zap.String("signal", sig.String()),
			zap.Duration("duration", signalInfo.Duration),
		)
	}

	// Add to buffer
	sm.addToBuffer(signalInfo)
}

// classifySignal classifies a signal into different types
func (sm *SignalManager) classifySignal(sig os.Signal) SignalType {
	switch sig {
	case syscall.SIGTERM, syscall.SIGINT:
		return SignalTypeGraceful
	case syscall.SIGQUIT:
		return SignalTypeForce
	case syscall.SIGHUP:
		return SignalTypeReload
	default:
		return SignalTypeUnknown
	}
}

// getSignalTimeout returns the timeout for a signal type
func (sm *SignalManager) getSignalTimeout(signalType SignalType) time.Duration {
	switch signalType {
	case SignalTypeGraceful:
		return sm.config.GracefulTimeout
	case SignalTypeForce:
		return sm.config.ForceTimeout
	case SignalTypeReload:
		return 10 * time.Second // Default reload timeout
	default:
		return 5 * time.Second
	}
}

// getSignalsToWatch returns the list of signals to watch based on configuration
func (sm *SignalManager) getSignalsToWatch() []os.Signal {
	var signals []os.Signal

	if sm.config.EnableGracefulShutdown {
		signals = append(signals, syscall.SIGTERM, syscall.SIGINT)
	}

	if sm.config.EnableForceShutdown {
		signals = append(signals, syscall.SIGQUIT)
	}

	if sm.config.EnableReload {
		signals = append(signals, syscall.SIGHUP)
	}

	return signals
}

// getSignalNames returns string names for signals
func (sm *SignalManager) getSignalNames(signals []os.Signal) []string {
	names := make([]string, len(signals))
	for i, sig := range signals {
		names[i] = sig.String()
	}
	return names
}

// addToBuffer adds a signal info to the buffer
func (sm *SignalManager) addToBuffer(info SignalInfo) {
	sm.bufferMutex.Lock()
	defer sm.bufferMutex.Unlock()

	sm.signalBuffer = append(sm.signalBuffer, info)

	// Keep only the last maxBufferSize signals
	if len(sm.signalBuffer) > sm.maxBufferSize {
		sm.signalBuffer = sm.signalBuffer[1:]
	}
}

// GetSignalBuffer returns the signal buffer
func (sm *SignalManager) GetSignalBuffer() []SignalInfo {
	sm.bufferMutex.RLock()
	defer sm.bufferMutex.RUnlock()

	// Return a copy to prevent external modification
	buffer := make([]SignalInfo, len(sm.signalBuffer))
	copy(buffer, sm.signalBuffer)
	return buffer
}

// GetSignalStats returns statistics about received signals
func (sm *SignalManager) GetSignalStats() map[os.Signal]int {
	sm.statsMutex.RLock()
	defer sm.statsMutex.RUnlock()

	stats := make(map[os.Signal]int)
	for sig, count := range sm.signalsReceived {
		stats[sig] = count
	}
	return stats
}

// IsListening returns whether the signal manager is currently listening
func (sm *SignalManager) IsListening() bool {
	sm.listenMutex.RLock()
	defer sm.listenMutex.RUnlock()
	return sm.isListening
}

// GetConfig returns the current signal configuration
func (sm *SignalManager) GetConfig() *SignalConfig {
	return sm.config
}

// UpdateConfig updates the signal configuration
func (sm *SignalManager) UpdateConfig(config *SignalConfig) error {
	if sm.IsListening() {
		return fmt.Errorf("cannot update config while signal manager is listening")
	}

	sm.config = config
	sm.logger.Info("Signal configuration updated")
	return nil
}

// ClearBuffer clears the signal buffer
func (sm *SignalManager) ClearBuffer() {
	sm.bufferMutex.Lock()
	defer sm.bufferMutex.Unlock()

	sm.signalBuffer = make([]SignalInfo, 0)
	sm.logger.Debug("Signal buffer cleared")
}

// GetHandlerCount returns the number of registered handlers
func (sm *SignalManager) GetHandlerCount() int {
	sm.handlerMutex.RLock()
	defer sm.handlerMutex.RUnlock()
	return len(sm.handlers)
}

// Cancel cancels the signal manager context
func (sm *SignalManager) Cancel() {
	sm.cancel()
}

// WaitForShutdown waits for the signal manager to shutdown
func (sm *SignalManager) WaitForShutdown() {
	sm.listenWG.Wait()
}