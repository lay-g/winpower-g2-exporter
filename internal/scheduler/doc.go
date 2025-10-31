// Package scheduler provides a simple and reliable scheduling mechanism for periodic data collection.
//
// The scheduler is responsible for triggering the Collector module at fixed intervals (5 seconds by default)
// to collect device data. It follows a simplified design principle, focusing on reliability and ease of testing.
//
// Key Features:
//   - Fixed interval triggering (5 seconds)
//   - Graceful start and stop operations
//   - Error resilience (errors in one cycle don't affect subsequent cycles)
//   - Structured logging integration
//   - Thread-safe state management
//
// Design Principles:
//   - Simplicity over flexibility: Fixed interval, no dynamic adjustment
//   - Reliability over performance: Single-threaded sequential execution
//   - Clear responsibility: Only triggers collection, doesn't handle collection logic
//   - Test-friendly: Interface-based design for easy mocking
//
// Example Usage:
//
//	config := scheduler.DefaultConfig()
//	sched := scheduler.NewDefaultScheduler(config, collector, logger)
//
//	// Start scheduler
//	if err := sched.Start(ctx); err != nil {
//	    log.Fatal("failed to start scheduler", err)
//	}
//
//	// Stop scheduler gracefully
//	if err := sched.Stop(ctx); err != nil {
//	    log.Error("failed to stop scheduler", err)
//	}
package scheduler
