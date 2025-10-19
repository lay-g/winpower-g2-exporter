package log

import (
	"os"
	"runtime"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != InfoLevel {
		t.Errorf("Expected default level to be %s, got %s", InfoLevel, config.Level)
	}

	if config.Format != JSONFormat {
		t.Errorf("Expected default format to be %s, got %s", JSONFormat, config.Format)
	}

	if config.Output != StdoutOutput {
		t.Errorf("Expected default output to be %s, got %s", StdoutOutput, config.Output)
	}

	if config.MaxSize != 100 {
		t.Errorf("Expected default MaxSize to be 100, got %d", config.MaxSize)
	}

	if config.MaxAge != 30 {
		t.Errorf("Expected default MaxAge to be 30, got %d", config.MaxAge)
	}

	if config.MaxBackups != 10 {
		t.Errorf("Expected default MaxBackups to be 10, got %d", config.MaxBackups)
	}

	if !config.Compress {
		t.Error("Expected default Compress to be true")
	}
}

func TestDevelopmentDefaults(t *testing.T) {
	config := DevelopmentDefaults()

	if config.Level != DebugLevel {
		t.Errorf("Expected development level to be %s, got %s", DebugLevel, config.Level)
	}

	if config.Format != ConsoleFormat {
		t.Errorf("Expected development format to be %s, got %s", ConsoleFormat, config.Format)
	}

	if config.Output != StdoutOutput {
		t.Errorf("Expected development output to be %s, got %s", StdoutOutput, config.Output)
	}

	if !config.EnableColor {
		t.Error("Expected development EnableColor to be true")
	}

	if !config.EnableCaller {
		t.Error("Expected development EnableCaller to be true")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "valid development config",
			config: &Config{
				Level:        DebugLevel,
				Format:       ConsoleFormat,
				Output:       StdoutOutput,
				EnableColor:  true,
				EnableCaller: true,
			},
			wantErr: false,
		},
		{
			name: "invalid level",
			config: &Config{
				Level:  Level("invalid"),
				Format: JSONFormat,
				Output: StdoutOutput,
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			config: &Config{
				Level:  InfoLevel,
				Format: Format("invalid"),
				Output: StdoutOutput,
			},
			wantErr: true,
		},
		{
			name: "invalid output",
			config: &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: Output("invalid"),
			},
			wantErr: true,
		},
		{
			name: "file output without filename",
			config: &Config{
				Level:    InfoLevel,
				Format:   JSONFormat,
				Output:   FileOutput,
				Filename: "",
			},
			wantErr: true,
		},
		{
			name: "both output without filename",
			config: &Config{
				Level:    InfoLevel,
				Format:   JSONFormat,
				Output:   BothOutput,
				Filename: "",
			},
			wantErr: true,
		},
		{
			name: "valid file output",
			config: &Config{
				Level:    InfoLevel,
				Format:   JSONFormat,
				Output:   FileOutput,
				Filename: "/tmp/test.log",
			},
			wantErr: false,
		},
		{
			name: "valid both output",
			config: &Config{
				Level:    InfoLevel,
				Format:   JSONFormat,
				Output:   BothOutput,
				Filename: "/tmp/test.log",
			},
			wantErr: false,
		},
		{
			name: "zero values should be fixed",
			config: &Config{
				Level:      InfoLevel,
				Format:     JSONFormat,
				Output:     StdoutOutput,
				MaxSize:    0,
				MaxAge:     0,
				MaxBackups: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// 验证零值是否被修正
				if tt.config.MaxSize <= 0 {
					t.Errorf("Expected MaxSize to be corrected to positive value, got %d", tt.config.MaxSize)
				}
				if tt.config.MaxAge <= 0 {
					t.Errorf("Expected MaxAge to be corrected to positive value, got %d", tt.config.MaxAge)
				}
				if tt.config.MaxBackups <= 0 {
					t.Errorf("Expected MaxBackups to be corrected to positive value, got %d", tt.config.MaxBackups)
				}
			}
		})
	}
}

func TestConfigClone(t *testing.T) {
	original := &Config{
		Level:      DebugLevel,
		Format:     ConsoleFormat,
		Output:     FileOutput,
		Filename:   "/tmp/test.log",
		MaxSize:    200,
		MaxAge:     60,
		MaxBackups: 20,
		Compress:   false,
	}

	clone := original.Clone()

	// 验证克隆的值
	if clone.Level != original.Level {
		t.Errorf("Clone level mismatch")
	}
	if clone.Format != original.Format {
		t.Errorf("Clone format mismatch")
	}
	if clone.Filename != original.Filename {
		t.Errorf("Clone filename mismatch")
	}

	// 修改克隆不应该影响原始
	clone.Level = ErrorLevel
	clone.Filename = "/tmp/modified.log"

	if original.Level == clone.Level {
		t.Error("Modifying clone affected original")
	}
	if original.Filename == clone.Filename {
		t.Error("Modifying clone affected original")
	}
}

func TestConfigSetters(t *testing.T) {
	config := DefaultConfig()

	// 测试 SetLevel
	config.SetLevel(ErrorLevel)
	if config.Level != ErrorLevel {
		t.Errorf("SetLevel failed, expected %s, got %s", ErrorLevel, config.Level)
	}

	// 测试 SetFormat
	config.SetFormat(ConsoleFormat)
	if config.Format != ConsoleFormat {
		t.Errorf("SetFormat failed, expected %s, got %s", ConsoleFormat, config.Format)
	}

	// 测试 SetOutput
	config.SetOutput(FileOutput)
	if config.Output != FileOutput {
		t.Errorf("SetOutput failed, expected %s, got %s", FileOutput, config.Output)
	}

	// 测试 SetFile
	config.SetFile("/tmp/test.log", 500, 90, 30, false)
	if config.Filename != "/tmp/test.log" {
		t.Errorf("SetFile failed for filename")
	}
	if config.MaxSize != 500 {
		t.Errorf("SetFile failed for MaxSize")
	}
	if config.MaxAge != 90 {
		t.Errorf("SetFile failed for MaxAge")
	}
	if config.MaxBackups != 30 {
		t.Errorf("SetFile failed for MaxBackups")
	}
	if config.Compress != false {
		t.Errorf("SetFile failed for Compress")
	}
}

func TestConfigValidateFileCreation(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	filename := tempDir + "/subdir/test.log"

	config := &Config{
		Level:    InfoLevel,
		Format:   JSONFormat,
		Output:   FileOutput,
		Filename: filename,
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Expected no error creating directory, got %v", err)
	}

	// 验证目录是否被创建
	if _, err := os.Stat(tempDir + "/subdir"); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}

func TestConfigConsoleAutoColor(t *testing.T) {
	config := &Config{
		Level:  InfoLevel,
		Format: ConsoleFormat,
		Output: StdoutOutput,
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}

	// 在非Windows平台上应该自动启用颜色
	if runtime.GOOS != "windows" && !config.EnableColor {
		t.Error("Expected EnableColor to be true on non-Windows platform for console format")
	}
}

func TestConfigNilClone(t *testing.T) {
	var config *Config
	clone := config.Clone()

	if clone != nil {
		t.Error("Expected nil config clone to be nil")
	}
}

func TestConfigValidationFilePermissions(t *testing.T) {
	// 测试在只读目录中创建文件的情况（如果可能）
	tempDir := t.TempDir()
	filename := tempDir + "/test.log"

	// 先创建目录
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	config := &Config{
		Level:    InfoLevel,
		Format:   JSONFormat,
		Output:   FileOutput,
		Filename: filename,
	}

	err = config.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid file path, got %v", err)
	}
}
