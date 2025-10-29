package log

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestBuildWriter(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantErr   bool
		checkFunc func(*testing.T, *WriterCloser)
	}{
		{
			name: "stdout output",
			config: &Config{
				Level:  "info",
				Format: "console",
				Output: "stdout",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, wc *WriterCloser) {
				if wc == nil {
					t.Error("Expected non-nil WriterCloser")
				}
				if len(wc.closers) != 0 {
					t.Errorf("Expected no closers for stdout, got %d", len(wc.closers))
				}
			},
		},
		{
			name: "stderr output",
			config: &Config{
				Level:  "info",
				Format: "console",
				Output: "stderr",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, wc *WriterCloser) {
				if wc == nil {
					t.Error("Expected non-nil WriterCloser")
				}
				if len(wc.closers) != 0 {
					t.Errorf("Expected no closers for stderr, got %d", len(wc.closers))
				}
			},
		},
		{
			name: "file output",
			config: &Config{
				Level:      "info",
				Format:     "json",
				Output:     "file",
				FilePath:   filepath.Join(t.TempDir(), "test.log"),
				MaxSize:    10,
				MaxAge:     7,
				MaxBackups: 3,
				Compress:   false,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, wc *WriterCloser) {
				if wc == nil {
					t.Error("Expected non-nil WriterCloser")
				}
				if len(wc.closers) != 1 {
					t.Errorf("Expected 1 closer for file output, got %d", len(wc.closers))
				}
			},
		},
		{
			name: "both output",
			config: &Config{
				Level:      "info",
				Format:     "json",
				Output:     "both",
				FilePath:   filepath.Join(t.TempDir(), "test.log"),
				MaxSize:    10,
				MaxAge:     7,
				MaxBackups: 3,
				Compress:   true,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, wc *WriterCloser) {
				if wc == nil {
					t.Error("Expected non-nil WriterCloser")
				}
				if len(wc.closers) != 1 {
					t.Errorf("Expected 1 closer for both output, got %d", len(wc.closers))
				}
			},
		},
		{
			name: "invalid output returns error",
			config: &Config{
				Level:  "info",
				Format: "console",
				Output: "invalid",
			},
			wantErr: true,
		},
		{
			name: "file output without filepath returns error",
			config: &Config{
				Level:  "info",
				Format: "json",
				Output: "file",
			},
			wantErr: true,
		},
		{
			name: "invalid config",
			config: &Config{
				Level:  "invalid",
				Format: "console",
				Output: "stdout",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wc, err := BuildWriter(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, wc)
			}
			if wc != nil && len(wc.closers) > 0 {
				_ = wc.Close()
			}
		})
	}
}

func TestWriterCloserClose(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	config := &Config{
		Level:      "info",
		Format:     "json",
		Output:     "file",
		FilePath:   filepath.Join(tmpDir, "test.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   false,
	}

	wc, err := BuildWriter(config)
	if err != nil {
		t.Fatalf("BuildWriter() error = %v", err)
	}

	// Write something to actually create the file
	testData := []byte("test log entry\n")
	_, err = wc.Write(testData)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}

	// Test Close
	err = wc.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(config.FilePath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestBuildStdoutWriter(t *testing.T) {
	writer := BuildStdoutWriter()
	if writer == nil {
		t.Error("Expected non-nil writer")
	}

	// Test writing
	testData := []byte("test message\n")
	n, err := writer.Write(testData)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != len(testData) {
		t.Errorf("Write() wrote %d bytes, want %d", n, len(testData))
	}
}

func TestBuildStderrWriter(t *testing.T) {
	writer := BuildStderrWriter()
	if writer == nil {
		t.Error("Expected non-nil writer")
	}

	// Test writing
	testData := []byte("test error message\n")
	n, err := writer.Write(testData)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != len(testData) {
		t.Errorf("Write() wrote %d bytes, want %d", n, len(testData))
	}
}

func TestBuildFileWriterWithRotation(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	writer := BuildFileWriterWithRotation(filePath, 10, 7, 3, false)
	if writer == nil {
		t.Fatal("Expected non-nil writer")
	}
	defer writer.Close()

	// Test writing
	testData := []byte("test log entry\n")
	n, err := writer.Write(testData)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != len(testData) {
		t.Errorf("Write() wrote %d bytes, want %d", n, len(testData))
	}

	// Verify the file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestBuildFileWriter(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		Level:      "info",
		Format:     "json",
		Output:     "file",
		FilePath:   filepath.Join(tmpDir, "test.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   true,
	}

	fileWriter := buildFileWriter(config)
	if fileWriter == nil {
		t.Fatal("Expected non-nil file writer")
	}
	defer fileWriter.Close()

	// Verify configuration
	if fileWriter.Filename != config.FilePath {
		t.Errorf("Filename = %v, want %v", fileWriter.Filename, config.FilePath)
	}
	if fileWriter.MaxSize != config.MaxSize {
		t.Errorf("MaxSize = %v, want %v", fileWriter.MaxSize, config.MaxSize)
	}
	if fileWriter.MaxAge != config.MaxAge {
		t.Errorf("MaxAge = %v, want %v", fileWriter.MaxAge, config.MaxAge)
	}
	if fileWriter.MaxBackups != config.MaxBackups {
		t.Errorf("MaxBackups = %v, want %v", fileWriter.MaxBackups, config.MaxBackups)
	}
	if fileWriter.Compress != config.Compress {
		t.Errorf("Compress = %v, want %v", fileWriter.Compress, config.Compress)
	}
	if !fileWriter.LocalTime {
		t.Error("LocalTime should be true")
	}
}

func TestWriterIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	config := &Config{
		Level:      "info",
		Format:     "json",
		Output:     "file",
		FilePath:   filepath.Join(tmpDir, "integration.log"),
		MaxSize:    1,
		MaxAge:     1,
		MaxBackups: 2,
		Compress:   false,
	}

	wc, err := BuildWriter(config)
	if err != nil {
		t.Fatalf("BuildWriter() error = %v", err)
	}
	defer wc.Close()

	// Write test data
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	})

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "test message",
	}

	buf, err := encoder.EncodeEntry(entry, nil)
	if err != nil {
		t.Fatalf("EncodeEntry() error = %v", err)
	}

	_, err = wc.Write(buf.Bytes())
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}

	// Sync and close
	if err := wc.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}
	if err := wc.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify the file was created and contains data
	data, err := os.ReadFile(config.FilePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("Log file is empty")
	}
	if !bytes.Contains(data, []byte("test message")) {
		t.Error("Log file does not contain expected message")
	}
}

func BenchmarkBuildWriter(b *testing.B) {
	config := &Config{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wc, _ := BuildWriter(config)
		if wc != nil {
			_ = wc.Close()
		}
	}
}

func BenchmarkBuildFileWriter(b *testing.B) {
	tmpDir := b.TempDir()
	config := &Config{
		Level:      "info",
		Format:     "json",
		Output:     "file",
		FilePath:   filepath.Join(tmpDir, "bench.log"),
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wc, _ := BuildWriter(config)
		if wc != nil {
			_ = wc.Close()
		}
	}
}
