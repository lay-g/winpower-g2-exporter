package energy

import (
	"testing"
	"time"
)

func TestStats_GetTotalCalculations(t *testing.T) {
	stats := &Stats{
		TotalCalculations: 10,
	}

	got := stats.GetTotalCalculations()
	if got != 10 {
		t.Errorf("GetTotalCalculations() = %v, want %v", got, 10)
	}
}

func TestStats_GetTotalErrors(t *testing.T) {
	stats := &Stats{
		TotalErrors: 5,
	}

	got := stats.GetTotalErrors()
	if got != 5 {
		t.Errorf("GetTotalErrors() = %v, want %v", got, 5)
	}
}

func TestStats_GetLastUpdateTime(t *testing.T) {
	now := time.Now()
	stats := &Stats{
		LastUpdateTime: now,
	}

	got := stats.GetLastUpdateTime()
	if !got.Equal(now) {
		t.Errorf("GetLastUpdateTime() = %v, want %v", got, now)
	}
}

func TestStats_GetAvgCalculationTime(t *testing.T) {
	duration := 100 * time.Millisecond
	stats := &Stats{
		AvgCalculationTime: duration,
	}

	got := stats.GetAvgCalculationTime()
	if got != duration {
		t.Errorf("GetAvgCalculationTime() = %v, want %v", got, duration)
	}
}

func TestStats_ConcurrentAccess(t *testing.T) {
	stats := &Stats{}

	// Test concurrent reads and writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			stats.GetTotalCalculations()
			stats.GetTotalErrors()
			stats.GetLastUpdateTime()
			stats.GetAvgCalculationTime()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
