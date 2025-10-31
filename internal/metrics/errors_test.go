package metrics

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrCollectorNil", ErrCollectorNil},
		{"ErrLoggerNil", ErrLoggerNil},
		{"ErrCollectionFailed", ErrCollectionFailed},
		{"ErrMetricsUpdateFailed", ErrMetricsUpdateFailed},
		{"ErrDeviceNotFound", ErrDeviceNotFound},
		{"ErrInvalidCollectionResult", ErrInvalidCollectionResult},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.err)
			assert.NotEmpty(t, tt.err.Error())
		})
	}
}

func TestErrorIs(t *testing.T) {
	// Test that errors can be compared using errors.Is
	assert.True(t, errors.Is(ErrCollectorNil, ErrCollectorNil))
	assert.False(t, errors.Is(ErrCollectorNil, ErrLoggerNil))
}
