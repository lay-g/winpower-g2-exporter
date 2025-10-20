package energy

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnergyError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		cause := errors.New("underlying error")
		ee := NewEnergyError("Calculate", "device-001", "context", cause)

		expected := "energy Calculate operation failed for device 'device-001': underlying error (context)"
		assert.Equal(t, expected, ee.Error())
		assert.Equal(t, cause, ee.Unwrap())
		assert.False(t, ee.Is(errors.New("different")))
		assert.True(t, ee.Is(cause))
	})

	t.Run("error without context", func(t *testing.T) {
		cause := errors.New("underlying error")
		ee := NewEnergyError("Get", "device-002", "", cause)

		expected := "energy Get operation failed for device 'device-002': underlying error"
		assert.Equal(t, expected, ee.Error())
		assert.Equal(t, cause, ee.Unwrap())
	})

	t.Run("error matching", func(t *testing.T) {
		cause := ErrInvalidDeviceID
		ee := NewEnergyError("Calculate", "device-003", "validation", cause)

		assert.True(t, ee.Is(ErrInvalidDeviceID))
		assert.False(t, ee.Is(ErrInvalidPowerValue))
	})
}

func TestIsValidPower(t *testing.T) {
	t.Run("valid positive power", func(t *testing.T) {
		assert.True(t, IsValidPower(100.0))
		assert.True(t, IsValidPower(0.0))
		assert.True(t, IsValidPower(0.0001))
	})

	t.Run("valid negative power", func(t *testing.T) {
		assert.True(t, IsValidPower(-100.0))
		assert.True(t, IsValidPower(-0.0001))
	})

	t.Run("invalid NaN power", func(t *testing.T) {
		assert.False(t, IsValidPower(math.NaN()))
	})

	t.Run("invalid +Inf power", func(t *testing.T) {
		assert.False(t, IsValidPower(math.Inf(1)))
	})

	t.Run("invalid -Inf power", func(t *testing.T) {
		assert.False(t, IsValidPower(math.Inf(-1)))
	})
}

func TestValidatePower(t *testing.T) {
	t.Run("valid power with nil config", func(t *testing.T) {
		err := ValidatePower(100.0, nil)
		assert.NoError(t, err)
	})

	t.Run("valid positive power", func(t *testing.T) {
		config := DefaultConfig()
		err := ValidatePower(100.0, config)
		assert.NoError(t, err)
	})

	t.Run("valid negative power when allowed", func(t *testing.T) {
		config := DefaultConfig()
		config.NegativePowerAllowed = true
		err := ValidatePower(-100.0, config)
		assert.NoError(t, err)
	})

	t.Run("invalid negative power when not allowed", func(t *testing.T) {
		config := DefaultConfig()
		config.NegativePowerAllowed = false
		err := ValidatePower(-100.0, config)
		assert.Error(t, err)
		assert.Equal(t, ErrNegativePowerNotAllowed, err)
	})

	t.Run("invalid NaN power", func(t *testing.T) {
		config := DefaultConfig()
		err := ValidatePower(math.NaN(), config)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPowerValue, err)
	})

	t.Run("invalid Inf power", func(t *testing.T) {
		config := DefaultConfig()
		err := ValidatePower(math.Inf(1), config)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPowerValue, err)
	})
}

func TestValidateDeviceID(t *testing.T) {
	t.Run("valid device ID", func(t *testing.T) {
		validIDs := []string{
			"device-001",
			"ups-001",
			"A1",
			"device_with_underscores",
			"device-with-dashes",
			"device123",
		}

		for _, id := range validIDs {
			err := ValidateDeviceID(id)
			assert.NoError(t, err, "Device ID %s should be valid", id)
		}
	})

	t.Run("invalid device ID", func(t *testing.T) {
		invalidIDs := []string{
			"",
			"   ", // spaces only
			"\t",  // tab only
			"\n",  // newline only
		}

		for _, id := range invalidIDs {
			err := ValidateDeviceID(id)
			assert.Error(t, err, "Device ID %q should be invalid", id)
			assert.Equal(t, ErrInvalidDeviceID, err)
		}
	})
}

func TestErrorConstants(t *testing.T) {
	// Test that error constants are properly defined
	assert.NotNil(t, ErrInvalidDeviceID)
	assert.NotNil(t, ErrInvalidPowerValue)
	assert.NotNil(t, ErrNegativePowerNotAllowed)
	assert.NotNil(t, ErrCalculationTimeout)
	assert.NotNil(t, ErrStorageOperation)
	assert.NotNil(t, ErrInvalidConfiguration)

	// Test error messages
	assert.Equal(t, "device ID cannot be empty", ErrInvalidDeviceID.Error())
	assert.Equal(t, "power value must be a finite number", ErrInvalidPowerValue.Error())
	assert.Equal(t, "negative power is not allowed by configuration", ErrNegativePowerNotAllowed.Error())
	assert.Equal(t, "energy calculation timed out", ErrCalculationTimeout.Error())
	assert.Equal(t, "storage operation failed", ErrStorageOperation.Error())
	assert.Equal(t, "invalid energy module configuration", ErrInvalidConfiguration.Error())
}
