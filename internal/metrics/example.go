package metrics

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// ExampleUsage demonstrates how to use the metrics module.
func ExampleUsage() {
	// Create a logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		_ = logger.Sync() // Ignore sync error in example
	}()

	// Create metrics configuration
	config := DefaultConfig()
	config.Registry = prometheus.NewRegistry()

	// Create metric manager
	manager, err := NewMetricManager(config, logger)
	if err != nil {
		log.Fatalf("Failed to create metric manager: %v", err)
	}

	// Set exporter status
	manager.SetUp(true)

	// Update some sample metrics (Exporter self-monitoring)
	manager.ObserveRequest("winpower.example.com", "GET", "200", 100*time.Millisecond)
	manager.IncScrapeError("winpower.example.com", "timeout")
	manager.ObserveCollection("winpower.example.com", "success", 500*time.Millisecond)
	manager.IncTokenRefresh("winpower.example.com", "success")
	manager.SetDeviceCount("winpower.example.com", "ups", 3.0)

	// TODO: Implement remaining metrics in future phases
	// manager.SetConnectionStatus("winpower.example.com", "https", 1.0)
	// manager.SetAuthStatus("winpower.example.com", "token", 1.0)
	// manager.SetDeviceConnected("device-001", "UPS-001", "ups", 1.0)
	// manager.SetPowerWatts("device-001", "UPS-001", "ups", 500.0)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle("/metrics", manager.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK")) // Ignore write error in example
	})

	server := &http.Server{
		Addr:    ":9090",
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down HTTP server")
	if err := server.Close(); err != nil {
		logger.Error("Failed to close HTTP server", zap.Error(err))
	}
}
