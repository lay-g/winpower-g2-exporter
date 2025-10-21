package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// DefaultScheduler implements the Scheduler interface
type DefaultScheduler struct {
	config    *Config
	client    WinPowerClient
	logger    log.Logger
	isRunning bool
	mu        sync.RWMutex // Protects isRunning field and concurrent access

	// Runtime fields
	ctx    context.Context
	cancel context.CancelFunc
	ticker *time.Ticker
	wg     sync.WaitGroup
}

// NewDefaultScheduler creates a new DefaultScheduler instance
func NewDefaultScheduler(config *Config, client WinPowerClient, logger log.Logger) (*DefaultScheduler, error) {
	// Validate parameters
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	if client == nil {
		return nil, errors.New("client cannot be nil")
	}
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create scheduler instance
	scheduler := &DefaultScheduler{
		config:    config,
		client:    client,
		logger:    logger,
		isRunning: false,
	}

	return scheduler, nil
}

// Start starts the scheduler
func (s *DefaultScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if scheduler is already running
	if s.isRunning {
		return nil // Already running, no-op
	}

	// Handle nil context by using context.Background()
	if ctx == nil {
		ctx = context.Background()
		s.logger.Warn("nil context provided, using background context")
	}

	// Create a new context with cancellation for this scheduler instance
	s.ctx, s.cancel = context.WithCancel(ctx)

	// Create ticker for periodic data collection
	s.ticker = time.NewTicker(s.config.CollectionInterval)

	// Mark scheduler as running
	s.isRunning = true

	// Start the collection goroutine
	s.wg.Add(1)
	go s.runCollectionLoop()

	// Log successful start
	s.logger.Info("scheduler started",
		log.String("collection_interval", s.config.CollectionInterval.String()))

	return nil
}

// Stop stops the scheduler
func (s *DefaultScheduler) Stop() error {
	// First, cancel the context to signal goroutines to stop
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return nil // Already stopped, no-op
	}

	// Cancel the context first (this will cause runCollectionLoop to exit)
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()

	// Give a brief moment for the goroutine to start exiting
	// This helps avoid the race condition where we set isRunning = false
	// while the goroutine is still trying to acquire a read lock
	time.Sleep(50 * time.Millisecond)

	// Now acquire the lock to update state and clean up
	s.mu.Lock()
	defer s.mu.Unlock()

	// Mark scheduler as not running
	s.isRunning = false

	// Stop the ticker
	if s.ticker != nil {
		s.ticker.Stop()
	}

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines finished gracefully
		s.logger.Info("scheduler stopped gracefully")
	case <-time.After(s.config.GracefulShutdownTimeout):
		// Timeout reached - force cleanup
		s.logger.Warn("scheduler stop timeout, forcing cleanup",
			log.String("timeout", s.config.GracefulShutdownTimeout.String()))
		return fmt.Errorf("scheduler stop timeout after %v", s.config.GracefulShutdownTimeout)
	}

	// Clear runtime fields
	s.ctx = nil
	s.cancel = nil
	s.ticker = nil

	return nil
}

// IsRunning returns true if the scheduler is currently running
func (s *DefaultScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// runCollectionLoop runs the main collection loop in a goroutine
func (s *DefaultScheduler) runCollectionLoop() {
	defer func() {
		s.logger.Debug("runCollectionLoop: calling wg.Done()")
		s.wg.Done()
		s.logger.Debug("runCollectionLoop: wg.Done() completed")
	}()

	s.logger.Debug("runCollectionLoop: starting collection loop")

	for {
		select {
		case <-s.ctx.Done():
			// Context cancelled, exit the loop
			s.logger.Info("scheduler collection loop stopped due to context cancellation")

			// Update running state when context is cancelled
			s.mu.Lock()
			s.isRunning = false
			s.mu.Unlock()

			s.logger.Debug("runCollectionLoop: exiting due to context cancellation")
			return

		case <-s.ticker.C:
			// Time to collect data - check if we should still be running
			s.mu.RLock()
			shouldRun := s.isRunning
			s.mu.RUnlock()

			if !shouldRun {
				s.logger.Debug("runCollectionLoop: scheduler marked as not running, exiting")
				return
			}

			s.logger.Debug("runCollectionLoop: ticker triggered, starting data collection")
			s.collectData()
			s.logger.Debug("runCollectionLoop: data collection completed")
		}
	}
}

// collectData performs the actual data collection
func (s *DefaultScheduler) collectData() {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("panic recovered in data collection",
				log.Any("panic", r),
				log.String("component", "scheduler"))
		}
	}()

	// Check if context is cancelled before starting collection
	select {
	case <-s.ctx.Done():
		s.logger.Debug("data collection skipped due to context cancellation",
			log.String("component", "scheduler"),
			log.String("reason", "context_cancelled"))
		return
	default:
	}

	// Record collection start time
	startTime := time.Now()

	s.logger.Debug("starting data collection",
		log.String("component", "scheduler"),
		log.String("collection_id", generateCollectionID()),
		log.Time("start_time", startTime))

	// Create a context with timeout for collection to avoid hanging
	// Use a shorter timeout than the collection interval but long enough for slow operations
	collectionTimeout := s.config.CollectionInterval / 2
	if collectionTimeout > 2*time.Second {
		collectionTimeout = 2 * time.Second
	}
	collectCtx, cancel := context.WithTimeout(s.ctx, collectionTimeout)
	defer cancel()

	// Collect device data with context awareness
	err := s.client.CollectDeviceData(collectCtx)
	duration := time.Since(startTime)

	if err != nil {
		// Determine error type for better categorization
		errorType := determineErrorType(err)

		s.logger.Error("data collection failed",
			log.String("component", "scheduler"),
			log.String("collection_id", generateCollectionID()),
			log.Error(err),
			log.String("error_type", errorType),
			log.Duration("duration", duration))

		// For certain error types, we might want to take additional actions
		handleCollectionError(err, s.logger)
		return
	}

	// Check if context is cancelled after collection
	select {
	case <-s.ctx.Done():
		s.logger.Debug("data collection completed but context cancelled",
			log.String("component", "scheduler"),
			log.String("collection_id", generateCollectionID()),
			log.Duration("duration", duration),
			log.String("reason", "context_cancelled_after_collection"))
		return
	default:
		s.logger.Info("data collection completed successfully",
			log.String("component", "scheduler"),
			log.String("collection_id", generateCollectionID()),
			log.Duration("duration", duration))
	}
}

// Global counter for collection IDs
var collectionCounter uint64

// generateCollectionID generates a unique collection ID for tracking
func generateCollectionID() string {
	id := atomic.AddUint64(&collectionCounter, 1)
	return fmt.Sprintf("collection-%d", id)
}

// determineErrorType categorizes errors for better monitoring and handling
func determineErrorType(err error) string {
	if err == nil {
		return "unknown"
	}

	errMsg := strings.ToLower(err.Error())

	// Network-related errors
	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "network is unreachable") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "timeout") {
		return "network_error"
	}

	// Authentication errors
	if strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "authentication failed") ||
		strings.Contains(errMsg, "invalid credentials") ||
		strings.Contains(errMsg, "forbidden") {
		return "auth_error"
	}

	// Context-related errors
	if errors.Is(err, context.Canceled) {
		return "context_cancelled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout_error"
	}

	// Data parsing errors
	if strings.Contains(errMsg, "json") ||
		strings.Contains(errMsg, "parse") ||
		strings.Contains(errMsg, "invalid format") ||
		strings.Contains(errMsg, "malformed") {
		return "data_error"
	}

	// Rate limiting
	if strings.Contains(errMsg, "rate limit") ||
		strings.Contains(errMsg, "too many requests") ||
		strings.Contains(errMsg, "quota exceeded") {
		return "rate_limit_error"
	}

	// Default
	return "unknown_error"
}

// handleCollectionError performs error-specific handling and recovery actions
func handleCollectionError(err error, logger log.Logger) {
	errorType := determineErrorType(err)

	switch errorType {
	case "network_error":
		// For network errors, we might want to implement backoff logic
		// For now, just log a warning that the service might be unavailable
		logger.Warn("network connectivity issue detected",
			log.String("error_type", errorType),
			log.String("suggestion", "check network connectivity and service availability"))

	case "auth_error":
		// For authentication errors, this is more serious
		logger.Error("authentication failure detected",
			log.String("error_type", errorType),
			log.String("suggestion", "check credentials and authentication configuration"))

	case "timeout_error":
		// For timeout errors, the service might be slow
		logger.Warn("data collection timeout",
			log.String("error_type", errorType),
			log.String("suggestion", "service might be overloaded, consider increasing timeout"))

	case "data_error":
		// For data parsing errors
		logger.Error("data parsing or format error",
			log.String("error_type", errorType),
			log.String("suggestion", "check API response format compatibility"))

	case "rate_limit_error":
		// For rate limiting
		logger.Warn("rate limit exceeded",
			log.String("error_type", errorType),
			log.String("suggestion", "reduce request frequency or check rate limit settings"))

	default:
		// For unknown errors
		logger.Error("unexpected error during data collection",
			log.String("error_type", errorType),
			log.String("suggestion", "investigate the error cause and implement appropriate handling"))
	}
}
