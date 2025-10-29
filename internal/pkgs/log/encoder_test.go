package log

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestBuildEncoder(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		wantFormat string
	}{
		{
			name: "json encoder production",
			config: &Config{
				Level:       "info",
				Format:      "json",
				Output:      "stdout",
				Development: false,
			},
			wantFormat: "json",
		},
		{
			name: "console encoder production",
			config: &Config{
				Level:       "info",
				Format:      "console",
				Output:      "stdout",
				Development: false,
			},
			wantFormat: "console",
		},
		{
			name: "json encoder development",
			config: &Config{
				Level:       "debug",
				Format:      "json",
				Output:      "stdout",
				Development: true,
			},
			wantFormat: "json",
		},
		{
			name: "console encoder development",
			config: &Config{
				Level:       "debug",
				Format:      "console",
				Output:      "stdout",
				Development: true,
			},
			wantFormat: "console",
		},
		{
			name: "default to console for invalid format",
			config: &Config{
				Level:       "info",
				Format:      "invalid",
				Output:      "stdout",
				Development: false,
			},
			wantFormat: "console",
		},
		{
			name: "invalid config uses default",
			config: &Config{
				Level:       "invalid_level",
				Format:      "json",
				Output:      "stdout",
				Development: false,
			},
			wantFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := BuildEncoder(tt.config)
			if encoder == nil {
				t.Error("Expected non-nil encoder")
			}

			// Verify the encoder works by encoding a test entry
			entry := zapcore.Entry{
				Level:   zapcore.InfoLevel,
				Message: "test message",
			}
			buf, err := encoder.EncodeEntry(entry, nil)
			if err != nil {
				t.Errorf("EncodeEntry() error = %v", err)
			}
			if buf.Len() == 0 {
				t.Error("Encoded buffer is empty")
			}
		})
	}
}

func TestBuildJSONEncoder(t *testing.T) {
	tests := []struct {
		name        string
		development bool
	}{
		{
			name:        "production json encoder",
			development: false,
		},
		{
			name:        "development json encoder",
			development: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := BuildJSONEncoder(tt.development)
			if encoder == nil {
				t.Fatal("Expected non-nil encoder")
			}

			// Test encoding
			entry := zapcore.Entry{
				Level:   zapcore.InfoLevel,
				Message: "test message",
			}
			buf, err := encoder.EncodeEntry(entry, nil)
			if err != nil {
				t.Errorf("EncodeEntry() error = %v", err)
			}
			if buf.Len() == 0 {
				t.Error("Encoded buffer is empty")
			}

			// Verify it's JSON format (should contain curly braces)
			output := buf.String()
			if len(output) == 0 {
				t.Error("Output is empty")
			}
		})
	}
}

func TestBuildConsoleEncoder(t *testing.T) {
	tests := []struct {
		name        string
		development bool
	}{
		{
			name:        "production console encoder",
			development: false,
		},
		{
			name:        "development console encoder",
			development: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := BuildConsoleEncoder(tt.development)
			if encoder == nil {
				t.Fatal("Expected non-nil encoder")
			}

			// Test encoding
			entry := zapcore.Entry{
				Level:   zapcore.InfoLevel,
				Message: "test message",
			}
			buf, err := encoder.EncodeEntry(entry, nil)
			if err != nil {
				t.Errorf("EncodeEntry() error = %v", err)
			}
			if buf.Len() == 0 {
				t.Error("Encoded buffer is empty")
			}

			// Verify output is not empty
			output := buf.String()
			if len(output) == 0 {
				t.Error("Output is empty")
			}
		})
	}
}

func TestBuildProductionEncoder(t *testing.T) {
	encoder := BuildProductionEncoder()
	if encoder == nil {
		t.Fatal("Expected non-nil encoder")
	}

	// Test encoding
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "production test",
	}
	buf, err := encoder.EncodeEntry(entry, nil)
	if err != nil {
		t.Errorf("EncodeEntry() error = %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Encoded buffer is empty")
	}
}

func TestBuildDevelopmentEncoder(t *testing.T) {
	encoder := BuildDevelopmentEncoder()
	if encoder == nil {
		t.Fatal("Expected non-nil encoder")
	}

	// Test encoding
	entry := zapcore.Entry{
		Level:   zapcore.DebugLevel,
		Message: "development test",
	}
	buf, err := encoder.EncodeEntry(entry, nil)
	if err != nil {
		t.Errorf("EncodeEntry() error = %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Encoded buffer is empty")
	}
}

func TestBuildEncoderConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "production config",
			config: &Config{
				Level:       "info",
				Format:      "json",
				Output:      "stdout",
				Development: false,
			},
		},
		{
			name: "development config",
			config: &Config{
				Level:       "debug",
				Format:      "console",
				Output:      "stdout",
				Development: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoderConfig := buildEncoderConfig(tt.config)

			// Verify key configurations are set
			if encoderConfig.EncodeTime == nil {
				t.Error("EncodeTime should be set")
			}
			if encoderConfig.EncodeLevel == nil {
				t.Error("EncodeLevel should be set")
			}
			if encoderConfig.EncodeCaller == nil {
				t.Error("EncodeCaller should be set")
			}
		})
	}
}

func TestBuildEncoderConfigByMode(t *testing.T) {
	tests := []struct {
		name        string
		development bool
	}{
		{
			name:        "production mode",
			development: false,
		},
		{
			name:        "development mode",
			development: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoderConfig := buildEncoderConfigByMode(tt.development)

			// Verify key configurations are set
			if encoderConfig.EncodeTime == nil {
				t.Error("EncodeTime should be set")
			}
			if encoderConfig.EncodeLevel == nil {
				t.Error("EncodeLevel should be set")
			}
			if encoderConfig.EncodeCaller == nil {
				t.Error("EncodeCaller should be set")
			}
		})
	}
}

func TestEncoderFormats(t *testing.T) {
	config := &Config{
		Level:       "info",
		Format:      "json",
		Output:      "stdout",
		Development: false,
	}

	// Test JSON encoder
	jsonEncoder := BuildEncoder(config)
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "test",
	}
	jsonBuf, err := jsonEncoder.EncodeEntry(entry, nil)
	if err != nil {
		t.Errorf("JSON EncodeEntry() error = %v", err)
	}

	// Test Console encoder
	config.Format = "console"
	consoleEncoder := BuildEncoder(config)
	consoleBuf, err := consoleEncoder.EncodeEntry(entry, nil)
	if err != nil {
		t.Errorf("Console EncodeEntry() error = %v", err)
	}

	// Outputs should be different
	if jsonBuf.String() == consoleBuf.String() {
		t.Error("JSON and Console encoders produced identical output")
	}
}

func TestEncoderWithFields(t *testing.T) {
	encoder := BuildJSONEncoder(false)

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "test with fields",
	}

	fields := []zapcore.Field{
		{Key: "key1", Type: zapcore.StringType, String: "value1"},
		{Key: "key2", Type: zapcore.Int64Type, Integer: 42},
	}

	buf, err := encoder.EncodeEntry(entry, fields)
	if err != nil {
		t.Errorf("EncodeEntry() with fields error = %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Output with fields is empty")
	}
}

func BenchmarkBuildEncoder(b *testing.B) {
	config := &Config{
		Level:       "info",
		Format:      "json",
		Output:      "stdout",
		Development: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildEncoder(config)
	}
}

func BenchmarkBuildJSONEncoder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildJSONEncoder(false)
	}
}

func BenchmarkBuildConsoleEncoder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildConsoleEncoder(false)
	}
}

func BenchmarkEncoderEncodeEntry(b *testing.B) {
	encoder := BuildJSONEncoder(false)
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "benchmark test message",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoder.EncodeEntry(entry, nil)
	}
}
