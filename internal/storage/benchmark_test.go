package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func BenchmarkFileStorageManager_Write(b *testing.B) {
	tempDir := b.TempDir()
	config := NewConfig()
	config.DataDir = tempDir
	config.SyncWrite = false // Disable sync for benchmark performance

	manager, err := NewFileStorageManager(config)
	if err != nil {
		b.Fatalf("NewFileStorageManager() error = %v", err)
	}
	deviceID := "benchmark-device"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data := NewPowerData(float64(i))
		err := manager.Write(deviceID, data)
		if err != nil {
			b.Fatalf("Write() error = %v", err)
		}
	}
}

func BenchmarkFileStorageManager_Read(b *testing.B) {
	tempDir := b.TempDir()
	config := NewConfig()
	config.DataDir = tempDir

	manager, err := NewFileStorageManager(config)
	if err != nil {
		b.Fatalf("NewFileStorageManager() error = %v", err)
	}
	deviceID := "benchmark-device"
	data := NewPowerData(100.0)

	// Pre-write data for reading
	err = manager.Write(deviceID, data)
	if err != nil {
		b.Fatalf("Write() error = %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := manager.Read(deviceID)
		if err != nil {
			b.Fatalf("Read() error = %v", err)
		}
	}
}

func BenchmarkFileStorageManager_WriteRead(b *testing.B) {
	tempDir := b.TempDir()
	config := NewConfig()
	config.DataDir = tempDir
	config.SyncWrite = false

	manager, err := NewFileStorageManager(config)
	if err != nil {
		b.Fatalf("NewFileStorageManager() error = %v", err)
	}
	deviceID := "benchmark-device"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data := NewPowerData(float64(i))

		// Write
		err := manager.Write(deviceID, data)
		if err != nil {
			b.Fatalf("Write() error = %v", err)
		}

		// Read
		_, err = manager.Read(deviceID)
		if err != nil {
			b.Fatalf("Read() error = %v", err)
		}
	}
}

func BenchmarkFileStorageManager_ConcurrentWrites(b *testing.B) {
	tempDir := b.TempDir()
	config := NewConfig()
	config.DataDir = tempDir
	config.SyncWrite = false

	manager, err := NewFileStorageManager(config)
	if err != nil {
		b.Fatalf("NewFileStorageManager() error = %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			deviceID := "device-" + string(rune('A'+(i%26)))
			data := NewPowerData(float64(i))

			err := manager.Write(deviceID, data)
			if err != nil {
				b.Fatalf("Write() error = %v", err)
			}
			i++
		}
	})
}

func BenchmarkFileStorageManager_ConcurrentReads(b *testing.B) {
	tempDir := b.TempDir()
	config := NewConfig()
	config.DataDir = tempDir

	manager, err := NewFileStorageManager(config)
	if err != nil {
		b.Fatalf("NewFileStorageManager() error = %v", err)
	}

	// Pre-write data for multiple devices
	for i := 0; i < 26; i++ {
		deviceID := "device-" + string(rune('A'+i))
		data := NewPowerData(float64(i))
		err := manager.Write(deviceID, data)
		if err != nil {
			b.Fatalf("Write() error = %v", err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			deviceID := "device-" + string(rune('A'+(i%26)))

			_, err := manager.Read(deviceID)
			if err != nil {
				b.Fatalf("Read() error = %v", err)
			}
			i++
		}
	})
}

func BenchmarkFileWriter_WriteDeviceFile(b *testing.B) {
	tempDir := b.TempDir()
	config := NewConfig()
	config.DataDir = tempDir
	config.SyncWrite = false
	writer := NewFileWriter(config)
	deviceID := "benchmark-device"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data := NewPowerData(float64(i))
		err := writer.WriteDeviceFile(deviceID, data)
		if err != nil {
			b.Fatalf("WriteDeviceFile() error = %v", err)
		}
	}
}

func BenchmarkFileReader_ReadAndParse(b *testing.B) {
	tempDir := b.TempDir()
	config := NewConfig()
	config.DataDir = tempDir
	config.SyncWrite = false
	writer := NewFileWriter(config)
	reader := NewFileReader(config)
	deviceID := "benchmark-device"
	data := NewPowerData(100.0)

	// Pre-write data
	err := writer.WriteDeviceFile(deviceID, data)
	if err != nil {
		b.Fatalf("WriteDeviceFile() error = %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := reader.ReadAndParse(deviceID)
		if err != nil {
			b.Fatalf("ReadAndParse() error = %v", err)
		}
	}
}

func BenchmarkFileReader_ParseData(b *testing.B) {
	tempDir := b.TempDir()
	reader := NewFileReaderWithPath(tempDir)
	content := []byte("1234567890\n100.5\n")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := reader.ParseData(content)
		if err != nil {
			b.Fatalf("ParseData() error = %v", err)
		}
	}
}

func BenchmarkPowerData_New(b *testing.B) {
	energy := 100.5

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = NewPowerData(energy)
	}
}

func BenchmarkPowerData_Validate(b *testing.B) {
	data := NewPowerData(100.5)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := data.Validate()
		if err != nil {
			b.Fatalf("Validate() error = %v", err)
		}
	}
}

func BenchmarkConfig_New(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = NewConfig()
	}
}

func BenchmarkConfig_Validate(b *testing.B) {
	config := NewConfig()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := config.Validate()
		if err != nil {
			b.Fatalf("Validate() error = %v", err)
		}
	}
}

func BenchmarkConfig_Clone(b *testing.B) {
	config := NewConfig()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = config.Clone()
	}
}

func BenchmarkStorageError_New(b *testing.B) {
	operation := "write"
	deviceID := "device123"
	path := "/tmp/device123.txt"
	cause := os.ErrExist

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = NewStorageError(operation, deviceID, path, cause)
	}
}

// Memory allocation benchmarks
func BenchmarkMemoryAllocation_PowerData(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data := &PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  float64(i),
		}
		_ = data
	}
}

func BenchmarkMemoryAllocation_Config(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		config := &Config{
			DataDir:         "./data",
			FilePermissions: 0644,
			DirPermissions:  0755,
			SyncWrite:       true,
			CreateDir:       true,
		}
		_ = config
	}
}

// Large data benchmarks
func BenchmarkLargeData_Write(b *testing.B) {
	tempDir := b.TempDir()
	config := NewConfig()
	config.DataDir = tempDir
	config.SyncWrite = false

	manager, err := NewFileStorageManager(config)
	if err != nil {
		b.Fatalf("NewFileStorageManager() error = %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Use large device IDs to test performance with longer names
		deviceID := "very-long-device-name-with-lots-of-characters-" + string(rune('A'+(i%26)))
		data := NewPowerData(123456.789012345)

		err := manager.Write(deviceID, data)
		if err != nil {
			b.Fatalf("Write() error = %v", err)
		}
	}
}

func BenchmarkInitialize(b *testing.B) {
	tempDir := b.TempDir()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Use a different subdirectory for each iteration
		config := NewConfig()
		config.DataDir = filepath.Join(tempDir, "init-"+string(rune('A'+(i%26))))
		config.CreateDir = true

		opts := &InitializeOptions{
			Config:        config,
			AutoCreateDir: true,
		}

		_, err := Initialize(opts)
		if err != nil {
			b.Fatalf("Initialize() error = %v", err)
		}
	}
}