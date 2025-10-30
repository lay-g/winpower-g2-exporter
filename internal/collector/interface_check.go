package collector

import (
	"github.com/lay-g/winpower-g2-exporter/internal/energy"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// Compile-time interface compliance checks.
// These declarations ensure that the concrete implementations satisfy the interfaces
// defined in this package. If they don't, the code will fail to compile.
//
// The underscore variable name means we're not actually using these variables;
// they exist solely for compile-time type checking.

var (
	// Verify that winpower.Client implements WinPowerClient interface
	_ WinPowerClient = (*winpower.Client)(nil)

	// Verify that energy.EnergyService implements EnergyCalculator interface
	_ EnergyCalculator = (*energy.EnergyService)(nil)

	// Verify that CollectorService implements CollectorInterface
	_ CollectorInterface = (*CollectorService)(nil)
)
