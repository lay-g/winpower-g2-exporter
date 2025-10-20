package winpower

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/energy"
	"go.uber.org/zap"
)

// ClientInterface defines the interface for WinPower client operations
type ClientInterface interface {
	// Connect establishes connection and authenticates with the WinPower API
	Connect(ctx context.Context) error

	// Disconnect closes the connection and cleans up resources
	Disconnect() error

	// IsConnected checks if the client is connected and authenticated
	IsConnected() bool

	// GetConnectionStatus returns the current connection status
	GetConnectionStatus() ConnectionStatus

	// CollectDeviceData retrieves data for all devices
	CollectDeviceData(ctx context.Context) ([]*DeviceData, error)

	// CollectDeviceDataByID retrieves data for a specific device
	CollectDeviceDataByID(ctx context.Context, deviceID string) (*DeviceData, error)

	// CollectDeviceDataWithEnergy retrieves data and optionally calculates energy
	CollectDeviceDataWithEnergy(ctx context.Context, energyService energy.EnergyInterface) ([]*DeviceData, error)

	// GetLastCollectionTime returns the timestamp of the last successful data collection
	GetLastCollectionTime() time.Time

	// GetDeviceList returns a list of all available devices
	GetDeviceList(ctx context.Context) ([]*DeviceInfo, error)

	// Close closes the client and releases all resources
	Close() error
}

// ConnectionStatus represents the connection status of the client
type ConnectionStatus int

const (
	// StatusDisconnected indicates the client is not connected
	StatusDisconnected ConnectionStatus = iota

	// StatusConnecting indicates the client is in the process of connecting
	StatusConnecting

	// StatusConnected indicates the client is connected and authenticated
	StatusConnected

	// StatusReconnecting indicates the client is attempting to reconnect
	StatusReconnecting

	// StatusError indicates the client has encountered an error
	StatusError
)

// String returns the string representation of the connection status
func (cs ConnectionStatus) String() string {
	switch cs {
	case StatusDisconnected:
		return "disconnected"
	case StatusConnecting:
		return "connecting"
	case StatusConnected:
		return "connected"
	case StatusReconnecting:
		return "reconnecting"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// DeviceInfo contains basic information about a device
type DeviceInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Model        string            `json:"model"`
	SerialNumber string            `json:"serial_number"`
	Location     string            `json:"location"`
	Tags         map[string]string `json:"tags"`
}

// Client implements the ClientInterface for WinPower API communication
type Client struct {
	config       *Config
	httpClient   *HTTPClient
	tokenManager *TokenManager
	parser       *DataParser
	logger       *zap.Logger

	// Connection state
	status    ConnectionStatus
	mutex     sync.RWMutex
	lastError error

	// Collection state
	lastCollectionTime time.Time
	collectionMutex    sync.RWMutex

	// Statistics
	stats *ClientStats
}

// ClientStats tracks client statistics
type ClientStats struct {
	TotalConnections   int64     `json:"total_connections"`
	SuccessfulRequests int64     `json:"successful_requests"`
	FailedRequests     int64     `json:"failed_requests"`
	LastConnected      time.Time `json:"last_connected"`
	LastError          time.Time `json:"last_error"`
	mutex              sync.RWMutex
}

// NewClient creates a new WinPower client with the given configuration
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, NewConfigError("NewClient", "config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, NewConfigError("NewClient", fmt.Sprintf("invalid config: %v", err)).WithCause(err)
	}

	// Create HTTP client
	httpClient, err := NewHTTPClient(config)
	if err != nil {
		return nil, NewConfigError("NewClient", fmt.Sprintf("failed to create HTTP client: %v", err)).WithCause(err)
	}

	// Create token manager
	tokenManager, err := NewTokenManager(config, httpClient)
	if err != nil {
		httpClient.Close()
		return nil, NewConfigError("NewClient", fmt.Sprintf("failed to create token manager: %v", err)).WithCause(err)
	}

	// Create default logger if none provided
	logger := zap.NewNop()

	client := &Client{
		config:       config,
		httpClient:   httpClient,
		tokenManager: tokenManager,
		parser:       NewDataParser(),
		logger:       logger,
		status:       StatusDisconnected,
		stats:        &ClientStats{},
	}

	return client, nil
}

// Connect establishes connection and authenticates with the WinPower API
func (c *Client) Connect(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if already connected
	if c.status == StatusConnected {
		return nil
	}

	// Set status to connecting
	c.setStatus(StatusConnecting)

	// Attempt to login
	if err := c.tokenManager.Login(ctx); err != nil {
		c.lastError = err
		c.setStatus(StatusError)
		c.stats.recordError()
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Mark as connected
	c.setStatus(StatusConnected)
	c.stats.recordConnection()
	c.lastError = nil

	return nil
}

// Disconnect closes the connection and cleans up resources
func (c *Client) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.status == StatusDisconnected {
		return nil
	}

	// Invalidate token
	c.tokenManager.InvalidateToken()

	// Set status to disconnected
	c.setStatus(StatusDisconnected)

	return nil
}

// IsConnected checks if the client is connected and authenticated
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.status == StatusConnected && c.tokenManager.IsTokenValid()
}

// GetConnectionStatus returns the current connection status
func (c *Client) GetConnectionStatus() ConnectionStatus {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// If we think we're connected but token is invalid, update status
	if c.status == StatusConnected && !c.tokenManager.IsTokenValid() {
		return StatusError
	}

	return c.status
}

// CollectDeviceData retrieves data for all devices
func (c *Client) CollectDeviceData(ctx context.Context) ([]*DeviceData, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client is not connected")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Mock implementation - in real implementation, this would:
	// 1. Make HTTP request to devices endpoint
	// 2. Parse response using parser
	// 3. Update last collection time
	// 4. Return device data

	c.collectionMutex.Lock()
	c.lastCollectionTime = time.Now()
	c.collectionMutex.Unlock()

	// Return mock device data for testing
	now := time.Now()
	return []*DeviceData{
		{
			ID:            "test-device-1",
			Name:          "Test Device 1",
			Type:          "UPS",
			Status:        "online",
			Timestamp:     now,
			Power:         1000.5,
			OutputVoltage: 230.0,
			OutputCurrent: 4.35,
		},
		{
			ID:            "test-device-2",
			Name:          "Test Device 2",
			Type:          "PDU",
			Status:        "online",
			Timestamp:     now,
			Power:         500.2,
			OutputVoltage: 230.0,
			OutputCurrent: 2.17,
		},
	}, nil
}

// CollectDeviceDataByID retrieves data for a specific device
func (c *Client) CollectDeviceDataByID(ctx context.Context, deviceID string) (*DeviceData, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client is not connected")
	}

	if deviceID == "" {
		return nil, fmt.Errorf("device ID cannot be empty")
	}

	// Mock implementation
	return nil, fmt.Errorf("device not found: %s", deviceID)
}

// CollectDeviceDataWithEnergy retrieves data and optionally calculates energy
func (c *Client) CollectDeviceDataWithEnergy(ctx context.Context, energyService energy.EnergyInterface) ([]*DeviceData, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client is not connected")
	}

	// First, collect device data
	devices, err := c.CollectDeviceData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect device data: %w", err)
	}

	// If energy service is provided, calculate energy for each device with panic recovery
	if energyService != nil {
		c.logger.Debug("Calculating energy for collected devices",
			zap.Int("device_count", len(devices)))

		for _, device := range devices {
			if device.Status == "online" && device.Power != 0 {
				// Safe energy calculation with panic recovery
				energy, calcErr := c.safeCalculateEnergy(energyService, device.ID, device.Power)
				if calcErr != nil {
					c.logger.Warn("Failed to calculate energy for device",
						zap.String("device_id", device.ID),
						zap.Float64("power", device.Power),
						zap.Error(calcErr))
					// Continue processing other devices even if one fails
					continue
				}

				// Validate calculated energy value
				if err := c.validateEnergyValue(energy); err != nil {
					c.logger.Warn("Invalid energy value calculated for device",
						zap.String("device_id", device.ID),
						zap.Float64("power", device.Power),
						zap.Float64("energy_wh", energy),
						zap.Error(err))
					continue
				}

				c.logger.Debug("Energy calculated for device",
					zap.String("device_id", device.ID),
					zap.Float64("power", device.Power),
					zap.Float64("energy_wh", energy))
			}
		}
	} else {
		c.logger.Debug("Energy service not provided, skipping energy calculation")
	}

	return devices, nil
}

// GetLastCollectionTime returns the timestamp of the last successful data collection
func (c *Client) GetLastCollectionTime() time.Time {
	c.collectionMutex.RLock()
	defer c.collectionMutex.RUnlock()
	return c.lastCollectionTime
}

// GetDeviceList returns a list of all available devices
func (c *Client) GetDeviceList(ctx context.Context) ([]*DeviceInfo, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client is not connected")
	}

	// Mock implementation
	return []*DeviceInfo{}, nil
}

// Close closes the client and releases all resources
func (c *Client) Close() error {
	// Disconnect first
	if err := c.Disconnect(); err != nil {
		return err
	}

	// Close HTTP client
	if c.httpClient != nil {
		c.httpClient.Close()
	}

	return nil
}

// setStatus sets the connection status (must be called with mutex lock held)
func (c *Client) setStatus(status ConnectionStatus) {
	c.status = status
}

// GetStats returns a copy of the client statistics
func (c *Client) GetStats() *ClientStats {
	if c == nil || c.stats == nil {
		return &ClientStats{}
	}

	c.stats.mutex.RLock()
	defer c.stats.mutex.RUnlock()

	return &ClientStats{
		TotalConnections:   c.stats.TotalConnections,
		SuccessfulRequests: c.stats.SuccessfulRequests,
		FailedRequests:     c.stats.FailedRequests,
		LastConnected:      c.stats.LastConnected,
		LastError:          c.stats.LastError,
	}
}

// recordConnection records a successful connection
func (cs *ClientStats) recordConnection() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.TotalConnections++
	cs.LastConnected = time.Now()
}


// recordError records a failed request
func (cs *ClientStats) recordError() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.FailedRequests++
	cs.LastError = time.Now()
}

// safeCalculateEnergy safely calculates energy with panic recovery
func (c *Client) safeCalculateEnergy(energyService energy.EnergyInterface, deviceID string, power float64) (float64, error) {
	// Use defer and recover to handle panics from energy service
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("Energy service panicked during calculation",
				zap.String("device_id", deviceID),
				zap.Float64("power", power),
				zap.Any("panic", r))
		}
	}()

	energy, err := energyService.Calculate(deviceID, power)
	if err != nil {
		return 0, fmt.Errorf("energy calculation failed: %w", err)
	}

	return energy, nil
}

// validateEnergyValue validates the calculated energy value
func (c *Client) validateEnergyValue(energy float64) error {
	// Check for NaN
	if math.IsNaN(energy) {
		return fmt.Errorf("calculated energy is NaN")
	}

	// Check for infinity
	if math.IsInf(energy, 0) {
		return fmt.Errorf("calculated energy is infinite")
	}

	// Check for reasonable energy values (in Watt-hours)
	// A typical device shouldn't generate more than 1 MWh in a single interval
	const maxReasonableEnergy = 1000000.0 // 1 MWh
	if math.Abs(energy) > maxReasonableEnergy {
		return fmt.Errorf("calculated energy %f Wh is unreasonably large", energy)
	}

	// Check for negative energy (could be valid for some scenarios, but log warning)
	if energy < 0 {
		c.logger.Debug("Negative energy value calculated",
			zap.Float64("energy_wh", energy))
	}

	return nil
}
