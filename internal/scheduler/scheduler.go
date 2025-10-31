package scheduler

import (
	"context"
	"sync"
	"time"
)

// DefaultScheduler implements the Scheduler interface with a simple fixed-interval design.
type DefaultScheduler struct {
	config    *Config
	collector CollectorInterface
	logger    Logger

	// Runtime state
	ticker  *time.Ticker
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running bool
	mu      sync.RWMutex
}

// NewDefaultScheduler creates a new DefaultScheduler with the given configuration and dependencies.
func NewDefaultScheduler(config *Config, collector CollectorInterface, logger Logger) (*DefaultScheduler, error) {
	if config == nil {
		return nil, ErrNilConfig
	}
	if collector == nil {
		return nil, ErrNilCollector
	}
	if logger == nil {
		return nil, ErrNilLogger
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &DefaultScheduler{
		config:    config,
		collector: collector,
		logger:    logger,
	}, nil
}

// Start starts the scheduler and begins triggering data collection at configured intervals.
func (s *DefaultScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return ErrAlreadyRunning
	}

	// Create a cancellable context
	s.ctx, s.cancel = context.WithCancel(ctx)

	// Create ticker with configured interval
	s.ticker = time.NewTicker(s.config.CollectionInterval)

	// Mark as running
	s.running = true

	// Start the collection loop in a goroutine
	s.wg.Add(1)
	go s.collectionLoop()

	s.logger.Info("scheduler started",
		"collection_interval", s.config.CollectionInterval,
	)

	return nil
}

// Stop stops the scheduler gracefully.
func (s *DefaultScheduler) Stop(ctx context.Context) error {
	s.mu.Lock()

	if !s.running {
		s.mu.Unlock()
		return ErrNotRunning
	}

	// Cancel the context to signal goroutine to stop
	s.cancel()

	// Stop the ticker
	if s.ticker != nil {
		s.ticker.Stop()
	}

	// Mark as not running
	s.running = false
	s.mu.Unlock()

	// Wait for goroutine to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// Determine timeout from context or config
	timeout := s.config.GracefulShutdownTimeout
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining < timeout {
			timeout = remaining
		}
	}

	select {
	case <-done:
		s.logger.Info("scheduler stopped gracefully")
		return nil
	case <-time.After(timeout):
		s.logger.Error("scheduler shutdown timeout",
			"timeout", timeout,
		)
		return ErrShutdownTimeout
	}
}

// collectionLoop runs the periodic collection in a separate goroutine.
func (s *DefaultScheduler) collectionLoop() {
	defer s.wg.Done()

	s.logger.Debug("collection loop started")

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Debug("collection loop stopped")
			return

		case <-s.ticker.C:
			s.runCollection()
		}
	}
}

// runCollection executes a single collection cycle.
func (s *DefaultScheduler) runCollection() {
	start := time.Now()

	// Create a context with timeout for this collection cycle
	ctx, cancel := context.WithTimeout(context.Background(), s.config.CollectionInterval)
	defer cancel()

	// Execute collection
	result, err := s.collector.CollectDeviceData(ctx)

	duration := time.Since(start)

	if err != nil {
		s.logger.Error("collection failed",
			"error", err,
			"duration", duration,
		)
		return
	}

	if result == nil {
		s.logger.Warn("collection returned nil result",
			"duration", duration,
		)
		return
	}

	// Log collection result
	if result.Success {
		s.logger.Info("collection completed",
			"device_count", result.DeviceCount,
			"duration", duration,
		)
	} else {
		s.logger.Error("collection unsuccessful",
			"device_count", result.DeviceCount,
			"error_message", result.ErrorMessage,
			"duration", duration,
		)
	}
}

// IsRunning returns whether the scheduler is currently running.
func (s *DefaultScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
