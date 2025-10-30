package storage

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// fileReader implements the FileReader interface.
type fileReader struct {
	config *Config
	logger log.Logger
}

// NewFileReader creates a new FileReader.
func NewFileReader(config *Config, logger log.Logger) FileReader {
	return &fileReader{
		config: config,
		logger: logger,
	}
}

// Read reads power data from a device file.
// If the file doesn't exist, it returns default initialized data.
func (r *fileReader) Read(deviceID string) (*PowerData, error) {
	filePath, err := buildFilePath(r.config.DataDir, deviceID)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Return default data for new devices
		r.logger.Debug("device file not found, returning default data",
			log.String("device_id", deviceID),
			log.String("path", filePath))

		return &PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  0.0,
		}, nil
	}

	// Open and read the file
	file, err := os.Open(filePath)
	if err != nil {
		r.logger.Error("failed to open device file",
			log.String("device_id", deviceID),
			log.String("path", filePath),
			log.Err(err))
		return nil, NewStorageError("read", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			r.logger.Warn("failed to close file",
				log.String("device_id", deviceID),
				log.String("path", filePath),
				log.Err(err))
		}
	}()

	// Parse the file format (two lines: timestamp, energy)
	scanner := bufio.NewScanner(file)

	// Read timestamp
	if !scanner.Scan() {
		err := fmt.Errorf("%w: missing timestamp line", ErrInvalidFormat)
		r.logger.Error("invalid file format",
			log.String("device_id", deviceID),
			log.String("path", filePath),
			log.Err(err))
		return nil, NewStorageError("read", filePath, err)
	}
	timestampStr := strings.TrimSpace(scanner.Text())

	// Read energy value
	if !scanner.Scan() {
		err := fmt.Errorf("%w: missing energy line", ErrInvalidFormat)
		r.logger.Error("invalid file format",
			log.String("device_id", deviceID),
			log.String("path", filePath),
			log.Err(err))
		return nil, NewStorageError("read", filePath, err)
	}
	energyStr := strings.TrimSpace(scanner.Text())

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		r.logger.Error("error reading file",
			log.String("device_id", deviceID),
			log.String("path", filePath),
			log.Err(err))
		return nil, NewStorageError("read", filePath, err)
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		err := fmt.Errorf("%w: invalid timestamp format: %v", ErrInvalidFormat, err)
		r.logger.Error("failed to parse timestamp",
			log.String("device_id", deviceID),
			log.String("timestamp", timestampStr),
			log.Err(err))
		return nil, NewStorageError("read", filePath, err)
	}

	// Parse energy value
	energy, err := strconv.ParseFloat(energyStr, 64)
	if err != nil {
		err := fmt.Errorf("%w: invalid energy format: %v", ErrInvalidFormat, err)
		r.logger.Error("failed to parse energy value",
			log.String("device_id", deviceID),
			log.String("energy", energyStr),
			log.Err(err))
		return nil, NewStorageError("read", filePath, err)
	}

	data := &PowerData{
		Timestamp: timestamp,
		EnergyWH:  energy,
	}

	// Validate the data
	if err := data.Validate(); err != nil {
		r.logger.Error("invalid data in file",
			log.String("device_id", deviceID),
			log.String("path", filePath),
			log.Err(err))
		return nil, NewStorageError("read", filePath, err)
	}

	r.logger.Debug("successfully read device data",
		log.String("device_id", deviceID),
		log.Int64("timestamp", data.Timestamp),
		log.Float64("energy_wh", data.EnergyWH))

	return data, nil
}
