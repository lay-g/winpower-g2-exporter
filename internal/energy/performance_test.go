package energy

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

// newQuietLogger 创建一个静默的测试日志器，只输出错误级别
func newQuietLogger(t *testing.T) *zap.Logger {
	zapOpts := []zaptest.LoggerOption{
		zaptest.Level(zapcore.ErrorLevel), // 只记录 error 级别
		zaptest.WrapOptions(zap.AddCaller()), // 添加调用者信息用于调试
	}
	return zaptest.NewLogger(t, zapOpts...)
}

// newQuietLoggerForBenchmark 创建用于基准测试的静默日志器
func newQuietLoggerForBenchmark(b *testing.B) *zap.Logger {
	zapOpts := []zaptest.LoggerOption{
		zaptest.Level(zapcore.ErrorLevel), // 只记录 error 级别
		zaptest.WrapOptions(zap.AddCaller()), // 添加调用者信息用于调试
	}
	return zaptest.NewLogger(b, zapOpts...)
}

// TestEnergyService_Performance_SingleDevice tests single device calculation performance
func TestEnergyService_Performance_SingleDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Performance test constraint: total execution time should not exceed 5 seconds
	testStart := time.Now()
	defer func() {
		totalDuration := time.Since(testStart)
		assert.Less(t, totalDuration, 5*time.Second, "Performance test exceeded 5 second limit")
	}()

	logger := newQuietLogger(t)
	storage := NewMockStorage()
	service := NewEnergyService(storage, logger, DefaultConfig())

	deviceID := "perf-test-device"
	power := 500.0
	numCalculations := 100

	// Warm up
	_, _ = service.Calculate(deviceID, power)

	// Measure performance
	start := time.Now()

	for i := 0; i < numCalculations; i++ {
		_, err := service.Calculate(deviceID, power+float64(i))
		assert.NoError(t, err)
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(numCalculations)

	t.Logf("Single device performance: %d calculations in %v (avg: %v per calculation)",
		numCalculations, duration, avgDuration)

	// Performance requirement: average calculation should be less than 10ms
	assert.Less(t, avgDuration, 10*time.Millisecond,
		"Average calculation time should be less than 10ms")
}

// TestEnergyService_Performance_MultipleDevices tests multiple device calculation performance
func TestEnergyService_Performance_MultipleDevices(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Performance test constraint: total execution time should not exceed 5 seconds
	testStart := time.Now()
	defer func() {
		totalDuration := time.Since(testStart)
		assert.Less(t, totalDuration, 5*time.Second, "Performance test exceeded 5 second limit")
	}()

	logger := newQuietLogger(t)
	storage := NewMockStorage()
	service := NewEnergyService(storage, logger, DefaultConfig())

	const numDevices = 10
	const numCalculationsPerDevice = 20

	var wg sync.WaitGroup

	start := time.Now()

	// Start goroutines for each device
	for deviceIdx := 0; deviceIdx < numDevices; deviceIdx++ {
		wg.Add(1)
		go func(deviceIdx int) {
			defer wg.Done()

			deviceID := t.Name() + "-device-" + string(rune('A'+deviceIdx))
			power := float64(100 + deviceIdx*10)

			for calcIdx := 0; calcIdx < numCalculationsPerDevice; calcIdx++ {
				calcStart := time.Now()
				_, err := service.Calculate(deviceID, power+float64(calcIdx))
				calcDuration := time.Since(calcStart)

				assert.NoError(t, err)
				assert.Less(t, calcDuration, 10*time.Millisecond,
					"Individual calculation should be less than 10ms")
			}
		}(deviceIdx)
	}

	wg.Wait()
	duration := time.Since(start)

	totalCalculations := numDevices * numCalculationsPerDevice
	avgDuration := duration / time.Duration(totalCalculations)

	t.Logf("Multiple device performance: %d calculations across %d devices in %v (avg: %v per calculation)",
		totalCalculations, numDevices, duration, avgDuration)

	// Performance requirements
	assert.Less(t, avgDuration, 10*time.Millisecond,
		"Average calculation time should be less than 10ms")
	assert.Less(t, duration, 5*time.Second,
		"Total test time should be less than 5 seconds")
}

// TestEnergyService_Performance_ConcurrentAccess tests concurrent access performance
func TestEnergyService_Performance_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Performance test constraint: total execution time should not exceed 5 seconds
	testStart := time.Now()
	defer func() {
		totalDuration := time.Since(testStart)
		assert.Less(t, totalDuration, 5*time.Second, "Performance test exceeded 5 second limit")
	}()

	logger := newQuietLogger(t)
	storage := NewMockStorage()
	service := NewEnergyService(storage, logger, DefaultConfig())

	const numGoroutines = 20
	const numCalculationsPerGoroutine = 10

	errors := make(chan error, numGoroutines*numCalculationsPerGoroutine)
	var calculationsDone int64

	start := time.Now()

	var wg sync.WaitGroup

	// Start concurrent calculations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numCalculationsPerGoroutine; j++ {
				deviceID := t.Name() + "-device-" + string(rune('A'+goroutineID%5))
				power := float64(100 + j)

				calcStart := time.Now()
				_, err := service.Calculate(deviceID, power)
				calcDuration := time.Since(calcStart)

				if err != nil {
					errors <- err
					return
				}

				// Verify individual calculation performance
				assert.Less(t, calcDuration, 10*time.Millisecond,
					"Individual calculation should be less than 10ms")

				calculationsDone++
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	duration := time.Since(start)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent calculation error: %v", err)
	}

	expectedCalculations := int64(numGoroutines * numCalculationsPerGoroutine)
	assert.Equal(t, expectedCalculations, calculationsDone, "All calculations should complete")

	avgDuration := duration / time.Duration(expectedCalculations)

	t.Logf("Concurrent access performance: %d calculations in %v (avg: %v per calculation)",
		expectedCalculations, duration, avgDuration)

	// Performance requirements
	assert.Less(t, avgDuration, 10*time.Millisecond,
		"Average calculation time should be less than 10ms")
	assert.Less(t, duration, 5*time.Second,
		"Total test time should be less than 5 seconds")

	// Verify statistics
	stats := service.GetStats()
	assert.Equal(t, expectedCalculations, stats.TotalCalculations)
	assert.Equal(t, int64(0), stats.TotalErrors)
}

// TestEnergyService_Performance_GetMethod tests Get method performance
func TestEnergyService_Performance_GetMethod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	logger := newQuietLogger(t)
	storage := NewMockStorage()
	service := NewEnergyService(storage, logger, DefaultConfig())

	// Set up some devices with data
	deviceIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		deviceIDs[i] = "device-" + string(rune('A'+i%26))
		_ = storage.Write(deviceIDs[i], &PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  float64(1000 + i),
		})
	}

	const numReads = 1000

	// Measure Get performance
	start := time.Now()

	for i := 0; i < numReads; i++ {
		deviceID := deviceIDs[i%len(deviceIDs)]
		energy, err := service.Get(deviceID)

		assert.NoError(t, err)
		assert.Greater(t, energy, 0.0)
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(numReads)

	t.Logf("Get method performance: %d reads in %v (avg: %v per read)",
		numReads, duration, avgDuration)

	// Get operations should be very fast since they're just reads
	assert.Less(t, avgDuration, 1*time.Millisecond,
		"Average Get operation should be less than 1ms")
}

// BenchmarkEnergyService_Calculate benchmarks the Calculate method
func BenchmarkEnergyService_Calculate(b *testing.B) {
	logger := newQuietLoggerForBenchmark(b)
	storage := NewMockStorage()
	service := NewEnergyService(storage, logger, DefaultConfig())

	deviceID := "benchmark-device"
	power := 500.0

	// Warm up
	_, _ = service.Calculate(deviceID, power)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.Calculate(deviceID, power+float64(i))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEnergyService_Get benchmarks the Get method
func BenchmarkEnergyService_Get(b *testing.B) {
	logger := newQuietLoggerForBenchmark(b)
	storage := NewMockStorage()
	service := NewEnergyService(storage, logger, DefaultConfig())

	deviceID := "benchmark-device"
	_ = storage.Write(deviceID, &PowerData{
		Timestamp: time.Now().UnixMilli(),
		EnergyWH:  1000.0,
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.Get(deviceID)
		if err != nil {
			b.Fatal(err)
		}
	}
}
