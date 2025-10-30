package storage

import (
	"fmt"
	"os"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// fileWriter implements the FileWriter interface.
type fileWriter struct {
	config *Config
	logger log.Logger
}

// NewFileWriter creates a new FileWriter.
func NewFileWriter(config *Config, logger log.Logger) FileWriter {
	return &fileWriter{
		config: config,
		logger: logger,
	}
}

// Write writes power data to a device file atomically.
func (w *fileWriter) Write(deviceID string, data *PowerData) error {
	if err := data.Validate(); err != nil {
		w.logger.Error("invalid data for write",
			log.String("device_id", deviceID),
			log.Err(err))
		return err
	}

	filePath, err := buildFilePath(w.config.DataDir, deviceID)
	if err != nil {
		return err
	}

	// Ensure the data directory exists
	if err := os.MkdirAll(w.config.DataDir, 0755); err != nil {
		w.logger.Error("failed to create data directory",
			log.String("dir", w.config.DataDir),
			log.Err(err))
		return NewStorageError("write", filePath, err)
	}

	// Format the data (two lines: timestamp, energy)
	content := fmt.Sprintf("%d\n%.2f\n", data.Timestamp, data.EnergyWH)

	// Write atomically using a temporary file
	tempPath := filePath + ".tmp"

	// Write to temporary file
	if err := os.WriteFile(tempPath, []byte(content), w.config.FilePermissions); err != nil {
		w.logger.Error("failed to write temporary file",
			log.String("device_id", deviceID),
			log.String("temp_path", tempPath),
			log.Err(err))
		return NewStorageError("write", filePath, err)
	}

	// Sync to ensure data is written to disk
	file, err := os.OpenFile(tempPath, os.O_RDWR, w.config.FilePermissions)
	if err == nil {
		_ = file.Sync()
		if err := file.Close(); err != nil {
			w.logger.Warn("failed to close temporary file",
				log.String("device_id", deviceID),
				log.String("temp_path", tempPath),
				log.Err(err))
		}
	}

	// Atomically rename the temporary file to the final file
	if err := os.Rename(tempPath, filePath); err != nil {
		// Clean up temp file on error
		_ = os.Remove(tempPath)
		w.logger.Error("failed to rename temporary file",
			log.String("device_id", deviceID),
			log.String("temp_path", tempPath),
			log.String("final_path", filePath),
			log.Err(err))
		return NewStorageError("write", filePath, err)
	}

	w.logger.Debug("successfully wrote device data",
		log.String("device_id", deviceID),
		log.String("path", filePath),
		log.Int64("timestamp", data.Timestamp),
		log.Float64("energy_wh", data.EnergyWH))

	return nil
}
