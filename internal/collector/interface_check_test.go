package collector

import (
	"testing"

	"github.com/lay-g/winpower-g2-exporter/internal/energy"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// TestInterfaceCompliance verifies that concrete implementations satisfy our interfaces.
// This is a runtime verification that complements the compile-time checks in interface_check.go
func TestInterfaceCompliance(t *testing.T) {
	t.Run("WinPowerClient interface", func(t *testing.T) {
		// Create a real winpower.Client
		config := &winpower.Config{
			BaseURL:  "https://example.com",
			Username: "test",
			Password: "test",
		}
		logger := log.NewTestLogger()
		client, err := winpower.NewClient(config, logger)
		if err != nil {
			t.Fatalf("Failed to create winpower client: %v", err)
		}

		// Verify it can be used as WinPowerClient
		var _ WinPowerClient = client

		// This compiles and runs, proving winpower.Client implements WinPowerClient
		t.Log("✓ winpower.Client implements WinPowerClient interface")
	})

	t.Run("EnergyCalculator interface", func(t *testing.T) {
		// Create a real energy.EnergyService
		storageConfig := storage.DefaultConfig()
		storageConfig.DataDir = t.TempDir()
		logger := log.NewTestLogger()
		storageManager, err := storage.NewFileStorageManager(storageConfig, logger)
		if err != nil {
			t.Fatalf("Failed to create storage manager: %v", err)
		}

		energyService := energy.NewEnergyService(storageManager, logger)

		// Verify it can be used as EnergyCalculator
		var _ EnergyCalculator = energyService

		// This compiles and runs, proving energy.EnergyService implements EnergyCalculator
		t.Log("✓ energy.EnergyService implements EnergyCalculator interface")
	})

	t.Run("CollectorInterface implementation", func(t *testing.T) {
		// Create collector with mocks
		mockWinPower := &MockWinPowerClient{}
		mockEnergy := &MockEnergyCalculator{}
		logger := log.NewTestLogger()

		collector, err := NewCollectorService(mockWinPower, mockEnergy, logger)
		if err != nil {
			t.Fatalf("Failed to create collector: %v", err)
		}

		// Verify it can be used as CollectorInterface
		var _ CollectorInterface = collector

		// This compiles and runs, proving CollectorService implements CollectorInterface
		t.Log("✓ CollectorService implements CollectorInterface")
	})
}

// TestInterfaceCompatibility demonstrates that real implementations work with collector
func TestInterfaceCompatibility(t *testing.T) {
	t.Run("Real implementations can be used with collector", func(t *testing.T) {
		// Setup real WinPower client
		winpowerConfig := &winpower.Config{
			BaseURL:  "https://example.com",
			Username: "test",
			Password: "test",
		}
		logger := log.NewTestLogger()
		winpowerClient, err := winpower.NewClient(winpowerConfig, logger)
		if err != nil {
			t.Fatalf("Failed to create winpower client: %v", err)
		}

		// Setup real Energy service
		storageConfig := storage.DefaultConfig()
		storageConfig.DataDir = t.TempDir()
		storageManager, err := storage.NewFileStorageManager(storageConfig, logger)
		if err != nil {
			t.Fatalf("Failed to create storage manager: %v", err)
		}
		energyService := energy.NewEnergyService(storageManager, logger)

		// Create collector with real implementations
		collector, err := NewCollectorService(winpowerClient, energyService, logger)
		if err != nil {
			t.Fatalf("Failed to create collector with real implementations: %v", err)
		}

		if collector == nil {
			t.Error("Expected non-nil collector")
		}

		t.Log("✓ Collector successfully created with real WinPower and Energy implementations")
		t.Log("  This proves the interfaces are correctly defined and implemented")
	})
}
