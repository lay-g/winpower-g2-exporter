package winpower

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFlexibleTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "parse RFC3339 with timezone",
			input:   `"2025-10-13T08:37:57.048192Z"`,
			wantErr: false,
		},
		{
			name:    "parse without timezone (WinPower format)",
			input:   `"2025-10-13T08:37:57.048192"`,
			wantErr: false,
		},
		{
			name:    "parse without microseconds",
			input:   `"2025-10-13T08:37:57"`,
			wantErr: false,
		},
		{
			name:    "parse null value",
			input:   `null`,
			wantErr: false,
		},
		{
			name:    "parse empty string",
			input:   `""`,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   `"invalid-date"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ft FlexibleTime
			err := json.Unmarshal([]byte(tt.input), &ft)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlexibleTime.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFlexibleTime_MarshalJSON(t *testing.T) {
	testTime := time.Date(2025, 10, 13, 8, 37, 57, 48192000, time.UTC)
	ft := FlexibleTime{Time: testTime}

	data, err := json.Marshal(ft)
	if err != nil {
		t.Fatalf("FlexibleTime.MarshalJSON() error = %v", err)
	}

	expected := `"2025-10-13T08:37:57Z"`
	if string(data) != expected {
		t.Errorf("FlexibleTime.MarshalJSON() = %s, want %s", string(data), expected)
	}
}

func TestAssetDevice_UnmarshalJSON_WithFlexibleTime(t *testing.T) {
	// Test data from real WinPower response
	jsonData := `{
		"id": "e156e6cb-41cb-4b35-b0dd-869929186a5c",
		"deviceType": 1,
		"model": "ON-LINE",
		"alias": "C3K",
		"protocolId": 13,
		"connectType": 1,
		"comPort": "COM3",
		"baudRate": 2400,
		"areaId": "00000000-0000-0000-0000-000000000000",
		"isActive": true,
		"firmwareVersion": "03.09",
		"createTime": "2025-10-13T08:37:57.048192",
		"warrantyStatus": 0
	}`

	var device AssetDevice
	err := json.Unmarshal([]byte(jsonData), &device)
	if err != nil {
		t.Fatalf("Failed to unmarshal AssetDevice: %v", err)
	}

	// Verify the device was parsed correctly
	if device.ID != "e156e6cb-41cb-4b35-b0dd-869929186a5c" {
		t.Errorf("Expected ID to be 'e156e6cb-41cb-4b35-b0dd-869929186a5c', got '%s'", device.ID)
	}

	if device.CreateTime.IsZero() {
		t.Error("Expected CreateTime to be non-zero")
	}

	// Verify the time components
	expectedYear, expectedMonth, expectedDay := 2025, time.October, 13
	year, month, day := device.CreateTime.Date()
	if year != expectedYear || month != expectedMonth || day != expectedDay {
		t.Errorf("Expected date %d-%02d-%02d, got %d-%02d-%02d",
			expectedYear, expectedMonth, expectedDay, year, month, day)
	}
}
