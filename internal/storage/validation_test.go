package storage

import (
	"math"
	"testing"
	"time"
)

func TestPowerData_Validate(t *testing.T) {
	now := time.Now().UnixMilli()

	tests := []struct {
		name    string
		data    *PowerData
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil data",
			data:    nil,
			wantErr: true,
			errMsg:  "PowerData cannot be nil",
		},
		{
			name: "negative timestamp",
			data: &PowerData{
				Timestamp: -1,
				EnergyWH:  100.0,
			},
			wantErr: true,
			errMsg:  "timestamp cannot be negative",
		},
		{
			name: "timestamp too far in future",
			data: &PowerData{
				Timestamp: now + (48 * 60 * 60 * 1000), // 2 days in future
				EnergyWH:  100.0,
			},
			wantErr: true,
			errMsg:  "timestamp is too far in the future",
		},
		{
			name: "NaN energy",
			data: &PowerData{
				Timestamp: now,
				EnergyWH:  math.NaN(),
			},
			wantErr: true,
			errMsg:  "energy value must be finite",
		},
		{
			name: "Inf energy",
			data: &PowerData{
				Timestamp: now,
				EnergyWH:  math.Inf(1),
			},
			wantErr: true,
			errMsg:  "energy value must be finite",
		},
		{
			name: "negative energy",
			data: &PowerData{
				Timestamp: now,
				EnergyWH:  -100.0,
			},
			wantErr: true,
			errMsg:  "energy value cannot be negative",
		},
		{
			name: "valid data",
			data: &PowerData{
				Timestamp: now,
				EnergyWH:  15000.50,
			},
			wantErr: false,
		},
		{
			name: "valid data with zero energy",
			data: &PowerData{
				Timestamp: now,
				EnergyWH:  0.0,
			},
			wantErr: false,
		},
		{
			name: "valid data with old timestamp",
			data: &PowerData{
				Timestamp: now - (365 * 24 * 60 * 60 * 1000), // 1 year ago
				EnergyWH:  1000.0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			}
		})
	}
}
