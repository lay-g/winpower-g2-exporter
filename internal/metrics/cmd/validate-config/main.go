package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.yaml.in/yaml/v2"

	"github.com/lay-g/winpower-g2-exporter/internal/metrics"
)

func main() {
	var (
		configFile = flag.String("config", "", "Configuration file to validate (YAML or JSON)")
		envOnly    = flag.Bool("env-only", false, "Only validate environment variables")
		outputJSON = flag.Bool("json", false, "Output results in JSON format")
		verbose    = flag.Bool("verbose", false, "Show detailed validation results")
		help       = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	logger := zap.NewExample()
	defer func() {
		_ = logger.Sync()
	}()

	validator := metrics.NewValidator(logger)

	var overallValid = true
	var results []ValidationResultJSON

	if *envOnly {
		// 只验证环境变量
		envVars := getEnvironmentVariables()
		result := validator.ValidateEnvironmentConfig(envVars)

		resultJSON := ValidationResultJSON{
			Type:   "environment",
			Valid:  result.Valid,
			Errors: result.Errors,
			Warns:  result.Warns,
		}
		results = append(results, resultJSON)

		if !result.Valid {
			overallValid = false
		}

		if *verbose || !result.Valid {
			printResult("Environment Variables", result, *outputJSON)
		}
	} else {
		if *configFile == "" {
			fmt.Fprintf(os.Stderr, "Error: Configuration file is required when not using --env-only\n")
			fmt.Fprintf(os.Stderr, "Use --help for more information\n")
			os.Exit(1)
		}

		// 验证配置文件
		config, err := loadConfigFile(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration file: %v\n", err)
			os.Exit(1)
		}

		result := validator.ValidateConfig(config)

		resultJSON := ValidationResultJSON{
			Type:   "config_file",
			File:   *configFile,
			Valid:  result.Valid,
			Errors: result.Errors,
			Warns:  result.Warns,
		}
		results = append(results, resultJSON)

		if !result.Valid {
			overallValid = false
		}

		printResult("Configuration File", result, *outputJSON)

		// 如果配置文件有效，生成建议
		if result.Valid {
			suggestions := validator.GenerateConfigSuggestions(config)
			if len(suggestions) > 0 {
				fmt.Printf("\n📋 Suggestions:\n")
				for _, suggestion := range suggestions {
					fmt.Printf("  • %s\n", suggestion)
				}
			}
		}

		// 验证环境变量
		envVars := getEnvironmentVariables()
		envResult := validator.ValidateEnvironmentConfig(envVars)

		envResultJSON := ValidationResultJSON{
			Type:   "environment",
			Valid:  envResult.Valid,
			Errors: envResult.Errors,
			Warns:  envResult.Warns,
		}
		results = append(results, envResultJSON)

		if !envResult.Valid {
			overallValid = false
		}

		if *verbose || !envResult.Valid {
			printResult("Environment Variables", envResult, *outputJSON)
		}
	}

	// 输出 JSON 结果（如果请求）
	if *outputJSON {
		jsonOutput := map[string]interface{}{
			"overall_valid": overallValid,
			"results":       results,
		}

		jsonData, err := json.MarshalIndent(jsonOutput, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON output: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonData))
	}

	// 设置退出码
	if !overallValid {
		os.Exit(1)
	}
}

// ValidationResultJSON 用于 JSON 输出的结构
type ValidationResultJSON struct {
	Type   string   `json:"type"`
	File   string   `json:"file,omitempty"`
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
	Warns  []string `json:"warnings"`
}

// ConfigFile 配置文件结构
type ConfigFile struct {
	Metrics metrics.MetricManagerConfig `yaml:"metrics" json:"metrics"`
}

// loadConfigFile 加载配置文件
func loadConfigFile(filename string) (metrics.MetricManagerConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return metrics.MetricManagerConfig{}, fmt.Errorf("reading file: %w", err)
	}

	ext := filepath.Ext(filename)
	var config ConfigFile

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return metrics.MetricManagerConfig{}, fmt.Errorf("parsing YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return metrics.MetricManagerConfig{}, fmt.Errorf("parsing JSON: %w", err)
		}
	default:
		return metrics.MetricManagerConfig{}, fmt.Errorf("unsupported file format: %s", ext)
	}

	return config.Metrics, nil
}

// getEnvironmentVariables 获取相关的环境变量
func getEnvironmentVariables() map[string]string {
	envVars := make(map[string]string)

	// 获取所有以 WINPOWER_EXPORTER_ 开头的环境变量
	for _, env := range os.Environ() {
		if key, value, found := strings.Cut(env, "="); found && strings.HasPrefix(key, "WINPOWER_EXPORTER_") {
			envVars[key] = value
		}
	}

	return envVars
}

// printResult 打印验证结果
func printResult(title string, result metrics.ValidationResult, outputJSON bool) {
	if outputJSON {
		return // JSON 输出由主函数处理
	}

	fmt.Printf("\n🔍 %s Validation:\n", title)
	fmt.Printf("%s\n", metrics.FormatValidationResult(result))

	if len(result.Errors) > 0 {
		fmt.Printf("\n❌ Errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  • %s\n", err)
		}
	}

	if len(result.Warns) > 0 {
		fmt.Printf("\n⚠️  Warnings:\n")
		for _, warn := range result.Warns {
			fmt.Printf("  • %s\n", warn)
		}
	}
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Printf(`WinPower G2 Exporter Metrics Configuration Validator

USAGE:
  validate-config [OPTIONS]

OPTIONS:
  -config string     Configuration file to validate (YAML or JSON)
  -env-only          Only validate environment variables
  -json              Output results in JSON format
  -verbose           Show detailed validation results
  -help              Show this help message

EXAMPLES:
  # Validate configuration file
  validate-config -config config.yaml

  # Validate only environment variables
  validate-config -env-only -verbose

  # Validate both config file and environment, output JSON
  validate-config -config config.yaml -json

  # Validate with verbose output
  validate-config -config config.yaml -verbose

ENVIRONMENT VARIABLES:
  The validator checks for the following environment variables:

  Required:
    WINPOWER_EXPORTER_METRICS_NAMESPACE
    WINPOWER_EXPORTER_METRICS_SUBSYSTEM

  Optional:
    WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS
    WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS
    WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS

EXIT CODES:
  0  All validations passed
  1  One or more validations failed

CONFIGURATION FORMAT:
  The configuration file should contain a 'metrics' section with the following structure:

  metrics:
    namespace: "winpower"
    subsystem: "exporter"
    request_duration_buckets: [0.05, 0.1, 0.2, 0.5, 1, 2, 5]
    collection_duration_buckets: [0.1, 0.2, 0.5, 1, 2, 5, 10]
    api_response_buckets: [0.05, 0.1, 0.2, 0.5, 1]

For more information, see the metrics module documentation.
`)
}
