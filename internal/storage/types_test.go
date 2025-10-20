package storage

import (
	"math"
	"testing"
	"time"
)

func TestNewPowerData(t *testing.T) {
	energy := 1500.75
	data := NewPowerData(energy)

	if data.EnergyWH != energy {
		t.Errorf("Expected EnergyWH to be %f, got %f", energy, data.EnergyWH)
	}

	if data.Timestamp <= 0 {
		t.Errorf("Expected Timestamp to be positive, got %d", data.Timestamp)
	}

	// Check that timestamp is recent (within last second)
	now := time.Now().UnixMilli()
	if data.Timestamp > now || data.Timestamp < now-1000 {
		t.Errorf("Expected Timestamp to be recent, got %d (now: %d)", data.Timestamp, now)
	}
}

func TestPowerData_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		data     *PowerData
		expected bool
	}{
		{
			name:     "nil data",
			data:     nil,
			expected: true,
		},
		{
			name: "zero values",
			data: &PowerData{
				Timestamp: 0,
				EnergyWH:  0,
			},
			expected: true,
		},
		{
			name: "non-zero timestamp",
			data: &PowerData{
				Timestamp: 1234567890,
				EnergyWH:  0,
			},
			expected: false,
		},
		{
			name: "non-zero energy",
			data: &PowerData{
				Timestamp: 0,
				EnergyWH:  100.5,
			},
			expected: false,
		},
		{
			name: "both non-zero",
			data: &PowerData{
				Timestamp: 1234567890,
				EnergyWH:  100.5,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.IsZero()
			if result != tt.expected {
				t.Errorf("IsZero() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestPowerData_Validate(t *testing.T) {
	tests := []struct {
		name        string
		data        *PowerData
		expectedErr error
	}{
		{
			name:        "nil data",
			data:        nil,
			expectedErr: ErrInvalidData,
		},
		{
			name: "valid data",
			data: &PowerData{
				Timestamp: 1694678400000,
				EnergyWH:  1500.75,
			},
			expectedErr: nil,
		},
		{
			name: "zero timestamp",
			data: &PowerData{
				Timestamp: 0,
				EnergyWH:  1500.75,
			},
			expectedErr: ErrInvalidTimestamp,
		},
		{
			name: "negative timestamp",
			data: &PowerData{
				Timestamp: -1,
				EnergyWH:  1500.75,
			},
			expectedErr: ErrInvalidTimestamp,
		},
		{
			name: "zero energy",
			data: &PowerData{
				Timestamp: 1694678400000,
				EnergyWH:  0,
			},
			expectedErr: nil,
		},
		{
			name: "positive energy",
			data: &PowerData{
				Timestamp: 1694678400000,
				EnergyWH:  1500.75,
			},
			expectedErr: nil,
		},
		{
			name: "negative energy (net energy)",
			data: &PowerData{
				Timestamp: 1694678400000,
				EnergyWH:  -500.25,
			},
			expectedErr: nil,
		},
		{
			name: "NaN energy",
			data: &PowerData{
				Timestamp: 1694678400000,
				EnergyWH:  math.NaN(),
			},
			expectedErr: ErrInvalidEnergyValue,
		},
		{
			name: "positive infinity energy",
			data: &PowerData{
				Timestamp: 1694678400000,
				EnergyWH:  math.Inf(1),
			},
			expectedErr: ErrInvalidEnergyValue,
		},
		{
			name: "negative infinity energy",
			data: &PowerData{
				Timestamp: 1694678400000,
				EnergyWH:  math.Inf(-1),
			},
			expectedErr: ErrInvalidEnergyValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.Validate()
			if err != tt.expectedErr {
				t.Errorf("Validate() error = %v, expectedErr = %v", err, tt.expectedErr)
			}
		})
	}
}

func TestIsFinite(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected bool
	}{
		{
			name:     "zero",
			value:    0,
			expected: true,
		},
		{
			name:     "positive number",
			value:    1500.75,
			expected: true,
		},
		{
			name:     "negative number",
			value:    -500.25,
			expected: true,
		},
		{
			name:     "NaN",
			value:    math.NaN(),
			expected: false,
		},
		{
			name:     "positive infinity",
			value:    math.Inf(1),
			expected: false,
		},
		{
			name:     "negative infinity",
			value:    math.Inf(-1),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFinite(tt.value)
			if result != tt.expected {
				t.Errorf("isFinite(%f) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}
