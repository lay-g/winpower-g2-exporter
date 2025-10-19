package log

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   Level
		want    string
		wantErr bool
	}{
		{
			name:    "debug level",
			level:   DebugLevel,
			want:    "debug",
			wantErr: false,
		},
		{
			name:    "info level",
			level:   InfoLevel,
			want:    "info",
			wantErr: false,
		},
		{
			name:    "warn level",
			level:   WarnLevel,
			want:    "warn",
			wantErr: false,
		},
		{
			name:    "error level",
			level:   ErrorLevel,
			want:    "error",
			wantErr: false,
		},
		{
			name:    "fatal level",
			level:   FatalLevel,
			want:    "fatal",
			wantErr: false,
		},
		{
			name:    "invalid level",
			level:   Level("invalid"),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.String() != tt.want {
				t.Errorf("parseLevel() = %v, want %v", result.String(), tt.want)
			}
		})
	}
}

func TestErrInvalidLevel(t *testing.T) {
	err := &ErrInvalidLevel{level: "invalid"}
	expected := "invalid log level: invalid"

	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestBuildJSONEncoderConfig(t *testing.T) {
	config := &Config{
		Level: InfoLevel,
	}

	encoderConfig := buildJSONEncoderConfig(config)

	// 验证基本配置
	if encoderConfig.TimeKey != "timestamp" {
		t.Errorf("Expected TimeKey to be 'timestamp', got '%s'", encoderConfig.TimeKey)
	}

	if encoderConfig.LevelKey != "level" {
		t.Errorf("Expected LevelKey to be 'level', got '%s'", encoderConfig.LevelKey)
	}

	if encoderConfig.MessageKey != "message" {
		t.Errorf("Expected MessageKey to be 'message', got '%s'", encoderConfig.MessageKey)
	}

	// Info级别不应该启用函数键
	if encoderConfig.FunctionKey != zapcore.OmitKey {
		t.Errorf("Expected FunctionKey to be omitted for info level, got '%s'", encoderConfig.FunctionKey)
	}

	// 验证编码器函数存在
	if encoderConfig.EncodeLevel == nil {
		t.Error("Expected EncodeLevel function to be set")
	}

	if encoderConfig.EncodeTime == nil {
		t.Error("Expected EncodeTime function to be set")
	}
}

func TestBuildJSONEncoderConfigDebug(t *testing.T) {
	config := &Config{
		Level: DebugLevel,
	}

	encoderConfig := buildJSONEncoderConfig(config)

	// Debug级别应该启用函数键
	if encoderConfig.FunctionKey == zapcore.OmitKey {
		t.Error("Expected FunctionKey to be set for debug level")
	}

	if encoderConfig.FunctionKey != "function" {
		t.Errorf("Expected FunctionKey to be 'function' for debug level, got '%s'", encoderConfig.FunctionKey)
	}
}

func TestBuildConsoleEncoderConfig(t *testing.T) {
	config := &Config{
		Level: InfoLevel,
	}

	encoderConfig := buildConsoleEncoderConfig(config)

	// 验证基本配置
	if encoderConfig.TimeKey != "T" {
		t.Errorf("Expected TimeKey to be 'T', got '%s'", encoderConfig.TimeKey)
	}

	if encoderConfig.LevelKey != "L" {
		t.Errorf("Expected LevelKey to be 'L', got '%s'", encoderConfig.LevelKey)
	}

	if encoderConfig.MessageKey != "M" {
		t.Errorf("Expected MessageKey to be 'M', got '%s'", encoderConfig.MessageKey)
	}

	// Info级别不应该启用函数键
	if encoderConfig.FunctionKey != zapcore.OmitKey {
		t.Errorf("Expected FunctionKey to be omitted for info level, got '%s'", encoderConfig.FunctionKey)
	}
}

func TestBuildConsoleEncoderConfigDebug(t *testing.T) {
	config := &Config{
		Level: DebugLevel,
	}

	encoderConfig := buildConsoleEncoderConfig(config)

	// Debug级别应该启用函数键
	if encoderConfig.FunctionKey == zapcore.OmitKey {
		t.Error("Expected FunctionKey to be set for debug level")
	}

	if encoderConfig.FunctionKey != "F" {
		t.Errorf("Expected FunctionKey to be 'F' for debug level, got '%s'", encoderConfig.FunctionKey)
	}
}

func TestParseLevelToZapLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		expected zapcore.Level
	}{
		{"debug", DebugLevel, zapcore.DebugLevel},
		{"info", InfoLevel, zapcore.InfoLevel},
		{"warn", WarnLevel, zapcore.WarnLevel},
		{"error", ErrorLevel, zapcore.ErrorLevel},
		{"fatal", FatalLevel, zapcore.FatalLevel},
		{"invalid", Level("invalid"), zapcore.InfoLevel}, // 默认为info级别
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLevelToZapLevel(tt.level)
			if result != tt.expected {
				t.Errorf("parseLevelToZapLevel() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildZapConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "json format",
			config: &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: StdoutOutput,
			},
			wantErr: false,
		},
		{
			name: "console format",
			config: &Config{
				Level:  DebugLevel,
				Format: ConsoleFormat,
				Output: StdoutOutput,
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
			name: "caller enabled",
			config: &Config{
				Level:        InfoLevel,
				Format:       JSONFormat,
				Output:       StdoutOutput,
				EnableCaller: true,
			},
			wantErr: false,
		},
		{
			name: "error level with stacktrace",
			config: &Config{
				Level:  ErrorLevel,
				Format: JSONFormat,
				Output: StdoutOutput,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zapConfig, err := buildZapConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildZapConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if zapConfig == nil {
					t.Error("Expected zap config to be created")
				} else {
					// 验证基本配置
					if tt.config.Format == JSONFormat && zapConfig.Encoding != "json" {
						t.Errorf("Expected encoding to be 'json', got '%s'", zapConfig.Encoding)
					}
					if tt.config.Format == ConsoleFormat && zapConfig.Encoding != "console" {
						t.Errorf("Expected encoding to be 'console', got '%s'", zapConfig.Encoding)
					}
					if zapConfig.DisableCaller != !tt.config.EnableCaller {
						t.Errorf("Expected DisableCaller to be %v, got %v", !tt.config.EnableCaller, zapConfig.DisableCaller)
					}
				}
			}
		})
	}
}
