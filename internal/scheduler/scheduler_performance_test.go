package scheduler

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPerformance_MemoryUsage tests memory usage and ensures no memory leaks
func TestPerformance_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	config := &Config{
		CollectionInterval:      10 * time.Millisecond,
		GracefulShutdownTimeout: 1 * time.Second,
	}

	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			// Simulate light work
			time.Sleep(1 * time.Millisecond)
			return nil
		},
		isConnectedFunc: func() bool {
			return true
		},
	}

	logger := &MockLogger{}

	// Get initial memory stats
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Create and run scheduler for multiple cycles
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Let it run for a while
	time.Sleep(400 * time.Millisecond)

	err = scheduler.Stop()
	require.NoError(t, err)

	// Get final memory stats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Check memory usage
	allocDiff := m2.Alloc - m1.Alloc
	assert.Less(t, allocDiff, uint64(100*1024), "Memory allocation should be reasonable (<100KB)")

	// Verify no goroutine leaks
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running")
}

// TestPerformance_GoroutineManagement tests that goroutines are properly managed
func TestPerformance_GoroutineManagement(t *testing.T) {
	config := &Config{
		CollectionInterval:      50 * time.Millisecond,
		GracefulShutdownTimeout: 500 * time.Millisecond,
	}

	// Get initial goroutine count
	initialGoroutines := runtime.NumGoroutine()

	// Create multiple schedulers concurrently
	numSchedulers := 5
	var wg sync.WaitGroup
	schedulers := make([]*DefaultScheduler, numSchedulers)

	for i := 0; i < numSchedulers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			client := &MockWinPowerClient{
				collectDataFunc: func(ctx context.Context) error {
					time.Sleep(10 * time.Millisecond)
					return nil
				},
				isConnectedFunc: func() bool {
					return true
				},
			}

			logger := &MockLogger{}

			scheduler, err := NewDefaultScheduler(config, client, logger)
			require.NoError(t, err)

			schedulers[index] = scheduler

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			err = scheduler.Start(ctx)
			require.NoError(t, err)

			time.Sleep(150 * time.Millisecond)

			err = scheduler.Stop()
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Wait a bit for all goroutines to finish
	time.Sleep(100 * time.Millisecond)

	// Check goroutine count
	finalGoroutines := runtime.NumGoroutine()
	goroutineDiff := finalGoroutines - initialGoroutines

	// Allow some tolerance for test infrastructure
	assert.LessOrEqual(t, goroutineDiff, 3, "Should not have significant goroutine leaks")

	// Verify all schedulers are stopped
	for _, scheduler := range schedulers {
		if scheduler != nil {
			assert.False(t, scheduler.IsRunning(), "All schedulers should be stopped")
		}
	}
}

// TestPerformance_TickerPrecision tests the precision of the ticker mechanism
func TestPerformance_TickerPrecision(t *testing.T) {
	config := &Config{
		CollectionInterval:      50 * time.Millisecond,
		GracefulShutdownTimeout: 1 * time.Second,
	}

	collectionTimes := make([]time.Time, 0)
	var collectionMutex sync.Mutex

	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectionMutex.Lock()
			collectionTimes = append(collectionTimes, time.Now())
			collectionMutex.Unlock()
			return nil
		},
		isConnectedFunc: func() bool {
			return true
		},
	}

	logger := &MockLogger{}

	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for multiple collections
	time.Sleep(250 * time.Millisecond)

	err = scheduler.Stop()
	require.NoError(t, err)

	// Analyze collection timing
	collectionMutex.Lock()
	defer collectionMutex.Unlock()

	assert.GreaterOrEqual(t, len(collectionTimes), 3, "Should have collected at least 3 times")

	// Check intervals between collections
	for i := 1; i < len(collectionTimes); i++ {
		interval := collectionTimes[i].Sub(collectionTimes[i-1])
		// Allow some tolerance (±10ms)
		assert.GreaterOrEqual(t, interval, 40*time.Millisecond, "Collection interval should be close to expected")
		assert.LessOrEqual(t, interval, 60*time.Millisecond, "Collection interval should be close to expected")
	}
}

// TestPerformance_HighFrequencyCollection tests performance under high frequency collection
func TestPerformance_HighFrequencyCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high frequency test in short mode")
	}

	config := &Config{
		CollectionInterval:      1 * time.Millisecond, // Very high frequency
		GracefulShutdownTimeout: 2 * time.Second,
	}

	collectionCount := 0
	var collectionMutex sync.Mutex

	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectionMutex.Lock()
			collectionCount++
			collectionMutex.Unlock()
			// Very light work
			return nil
		},
		isConnectedFunc: func() bool {
			return true
		},
	}

	logger := &MockLogger{}

	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	startTime := time.Now()
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Let it run
	time.Sleep(80 * time.Millisecond)

	err = scheduler.Stop()
	require.NoError(t, err)

	duration := time.Since(startTime)

	collectionMutex.Lock()
	defer collectionMutex.Unlock()

	// Should have made many collections
	expectedCollections := int(duration / config.CollectionInterval)
	actualCollections := collectionCount

	// Allow some tolerance for timing variations
	assert.GreaterOrEqual(t, actualCollections, expectedCollections/2,
		"Should have made approximately expected number of collections")

	t.Logf("Made %d collections in %v (expected ~%d)", actualCollections, duration, expectedCollections)
}

// TestPerformance_ConcurrentAccess tests performance under concurrent access
func TestPerformance_ConcurrentAccess(t *testing.T) {
	config := &Config{
		CollectionInterval:      20 * time.Millisecond,
		GracefulShutdownTimeout: 1 * time.Second,
	}

	collectionCount := 0
	var collectionMutex sync.Mutex

	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectionMutex.Lock()
			collectionCount++
			collectionMutex.Unlock()
			time.Sleep(5 * time.Millisecond) // Simulate some work
			return nil
		},
		isConnectedFunc: func() bool {
			return true
		},
	}

	logger := &MockLogger{}

	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Perform concurrent operations
	var wg sync.WaitGroup
	numOperations := 10

	// Concurrent status checks
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				_ = scheduler.IsRunning()
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	// Wait for operations and scheduler runtime
	time.Sleep(150 * time.Millisecond)
	wg.Wait()

	err = scheduler.Stop()
	require.NoError(t, err)

	// Verify collections occurred
	collectionMutex.Lock()
	assert.Greater(t, collectionCount, 0, "Collections should have occurred")
	collectionMutex.Unlock()

	// Verify scheduler is properly stopped
	assert.False(t, scheduler.IsRunning(), "Scheduler should be stopped")
}

// TestPerformance_LongRunningStability tests stability over longer periods
func TestPerformance_LongRunningStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running stability test in short mode")
	}

	config := &Config{
		CollectionInterval:      100 * time.Millisecond,
		GracefulShutdownTimeout: 2 * time.Second,
	}

	collectionCount := 0
	errorCount := 0
	var collectionMutex sync.Mutex

	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectionMutex.Lock()
			collectionCount++

			// Occasionally simulate errors for robustness testing
			if collectionCount%5 == 0 {
				errorCount++
				collectionMutex.Unlock()
				return assert.AnError
			}
			collectionMutex.Unlock()

			// Simulate variable work time
			time.Sleep(time.Duration(5+collectionCount%10) * time.Millisecond)
			return nil
		},
		isConnectedFunc: func() bool {
			return true
		},
	}

	logger := &MockLogger{}

	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Run for a longer period
	runTime := 1 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), runTime+200*time.Millisecond)
	defer cancel()

	startTime := time.Now()
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Let it run
	time.Sleep(runTime)

	err = scheduler.Stop()
	require.NoError(t, err)

	actualRunTime := time.Since(startTime)

	// Verify performance
	collectionMutex.Lock()
	defer collectionMutex.Unlock()

	expectedCollections := int(actualRunTime / config.CollectionInterval)
	actualCollections := collectionCount

	assert.GreaterOrEqual(t, actualCollections, expectedCollections/2,
		"Should have made approximately expected number of collections")
	assert.Greater(t, errorCount, 0, "Should have experienced some errors for robustness testing")

	t.Logf("Long-running test: %d collections in %v with %d errors",
		actualCollections, actualRunTime, errorCount)
}
