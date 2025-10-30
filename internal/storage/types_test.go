package storage

import (
	"testing"
)

func TestPowerData_Structure(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		energyWH  float64
	}{
		{
			name:      "zero values",
			timestamp: 0,
			energyWH:  0.0,
		},
		{
			name:      "positive values",
			timestamp: 1694678400000,
			energyWH:  15000.50,
		},
		{
			name:      "negative timestamp",
			timestamp: -1,
			energyWH:  100.0,
		},
		{
			name:      "negative energy",
			timestamp: 1694678400000,
			energyWH:  -100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &PowerData{
				Timestamp: tt.timestamp,
				EnergyWH:  tt.energyWH,
			}

			if data.Timestamp != tt.timestamp {
				t.Errorf("Timestamp = %v, want %v", data.Timestamp, tt.timestamp)
			}
			if data.EnergyWH != tt.energyWH {
				t.Errorf("EnergyWH = %v, want %v", data.EnergyWH, tt.energyWH)
			}
		})
	}
}
