package metrics

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ValidationResult 验证结果
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
	Warns  []string `json:"warns"`
}

// Validator 配置验证器
type Validator struct {
	logger *zap.Logger
}

// NewValidator 创建新的配置验证器
func NewValidator(logger *zap.Logger) *Validator {
	return &Validator{
		logger: logger,
	}
}

// ValidateConfig 验证完整的指标管理器配置
func (v *Validator) ValidateConfig(cfg MetricManagerConfig) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []string{},
		Warns:  []string{},
	}

	// 验证基础配置
	v.validateNamespace(cfg.Namespace, &result)
	v.validateSubsystem(cfg.Subsystem, &result)

	// 验证直方图桶配置
	v.validateBuckets(cfg.RequestDurationBuckets, "request_duration", &result)
	v.validateBuckets(cfg.CollectionDurationBuckets, "collection_duration", &result)
	v.validateBuckets(cfg.APIResponseBuckets, "api_response", &result)

	return result
}

// ValidateNamespace 验证命名空间
func (v *Validator) validateNamespace(namespace string, result *ValidationResult) {
	if namespace == "" {
		result.Errors = append(result.Errors, "metrics namespace cannot be empty")
		result.Valid = false
		return
	}

	if !isValidPrometheusName(namespace) {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid metrics namespace: %s (must contain only letters, numbers, and underscores, and not start with a number)", namespace))
		result.Valid = false
		return
	}

	if len(namespace) > 50 {
		result.Warns = append(result.Warns, fmt.Sprintf("metrics namespace is very long (%d characters), consider shortening it", len(namespace)))
	}
}

// ValidateSubsystem 验证子系统
func (v *Validator) validateSubsystem(subsystem string, result *ValidationResult) {
	if subsystem == "" {
		result.Errors = append(result.Errors, "metrics subsystem cannot be empty")
		result.Valid = false
		return
	}

	if !isValidPrometheusName(subsystem) {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid metrics subsystem: %s (must contain only letters, numbers, and underscores, and not start with a number)", subsystem))
		result.Valid = false
		return
	}

	if len(subsystem) > 50 {
		result.Warns = append(result.Warns, fmt.Sprintf("metrics subsystem is very long (%d characters), consider shortening it", len(subsystem)))
	}
}

// ValidateBuckets 验证直方图桶配置
func (v *Validator) validateBuckets(buckets []float64, bucketType string, result *ValidationResult) {
	if len(buckets) == 0 {
		result.Errors = append(result.Errors, fmt.Sprintf("%s buckets cannot be empty", bucketType))
		result.Valid = false
		return
	}

	// 验证桶数量
	if len(buckets) > 20 {
		result.Warns = append(result.Warns, fmt.Sprintf("%s has too many buckets (%d), consider reducing to 10 or fewer for better performance", bucketType, len(buckets)))
	}

	if len(buckets) < 3 {
		result.Warns = append(result.Warns, fmt.Sprintf("%s has too few buckets (%d), consider using at least 5 buckets for better granularity", bucketType, len(buckets)))
	}

	// 验证桶边界递增
	for i := 1; i < len(buckets); i++ {
		if buckets[i] <= buckets[i-1] {
			result.Errors = append(result.Errors, fmt.Sprintf("%s buckets must be in increasing order: bucket[%d] (%.6f) <= bucket[%d] (%.6f)",
				bucketType, i-1, buckets[i-1], i, buckets[i]))
			result.Valid = false
			return
		}
	}

	// 验证所有桶都为正数
	for i, bucket := range buckets {
		if bucket <= 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("%s bucket[%d] must be positive, got: %.6f", bucketType, i, bucket))
			result.Valid = false
			return
		}
	}

	// 验证桶边界的合理性
	v.validateBucketReasonableness(buckets, bucketType, result)
}

// ValidateBucketReasonableness 验证桶边界的合理性
func (v *Validator) validateBucketReasonableness(buckets []float64, bucketType string, result *ValidationResult) {
	// 根据桶类型提供特定的建议
	switch bucketType {
	case "request_duration":
		v.validateRequestDurationBuckets(buckets, result)
	case "collection_duration":
		v.validateCollectionDurationBuckets(buckets, result)
	case "api_response":
		v.validateAPIResponseBuckets(buckets, result)
	}
}

// ValidateRequestDurationBuckets 验证 HTTP 请求时延桶
func (v *Validator) validateRequestDurationBuckets(buckets []float64, result *ValidationResult) {
	// HTTP 请求时延通常在几毫秒到几秒之间
	minExpected := 0.001 // 1ms
	maxExpected := 30.0  // 30s

	firstBucket := buckets[0]
	lastBucket := buckets[len(buckets)-1]

	if firstBucket > minExpected*10 {
		result.Warns = append(result.Warns, fmt.Sprintf("request_duration first bucket (%.3fs) might be too large for fast requests, consider adding smaller buckets", firstBucket))
	}

	if lastBucket < maxExpected {
		result.Warns = append(result.Warns, fmt.Sprintf("request_duration last bucket (%.3fs) might be too small for slow requests, consider adding larger buckets", lastBucket))
	}

	// 检查桶的分布是否合理
	if len(buckets) >= 2 {
		ratio := buckets[1] / buckets[0]
		if ratio > 10 {
			result.Warns = append(result.Warns, fmt.Sprintf("request_duration buckets have large gaps (ratio %.1f), consider adding more intermediate buckets", ratio))
		}
	}
}

// ValidateCollectionDurationBuckets 验证采集时延桶
func (v *Validator) validateCollectionDurationBuckets(buckets []float64, result *ValidationResult) {
	// 数据采集时延通常在几百毫秒到几分钟之间
	minExpected := 0.1   // 100ms
	maxExpected := 600.0 // 10分钟

	firstBucket := buckets[0]
	lastBucket := buckets[len(buckets)-1]

	if firstBucket > minExpected*5 {
		result.Warns = append(result.Warns, fmt.Sprintf("collection_duration first bucket (%.3fs) might be too large for fast collections", firstBucket))
	}

	if lastBucket < maxExpected/10 {
		result.Warns = append(result.Warns, fmt.Sprintf("collection_duration last bucket (%.3fs) might be too small for slow collections", lastBucket))
	}
}

// ValidateAPIResponseBuckets 验证 API 响应时延桶
func (v *Validator) validateAPIResponseBuckets(buckets []float64, result *ValidationResult) {
	// API 响应时延通常在几毫秒到几秒之间
	minExpected := 0.01 // 10ms
	maxExpected := 10.0 // 10s

	firstBucket := buckets[0]
	lastBucket := buckets[len(buckets)-1]

	if firstBucket > minExpected*10 {
		result.Warns = append(result.Warns, fmt.Sprintf("api_response first bucket (%.3fs) might be too large for fast APIs", firstBucket))
	}

	if lastBucket < maxExpected {
		result.Warns = append(result.Warns, fmt.Sprintf("api_response last bucket (%.3fs) might be too small for slow APIs", lastBucket))
	}
}

// ValidateEnvironmentConfig 验证环境变量配置
func (v *Validator) ValidateEnvironmentConfig(envVars map[string]string) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []string{},
		Warns:  []string{},
	}

	// 验证必需的环境变量
	requiredVars := []string{
		"WINPOWER_EXPORTER_METRICS_NAMESPACE",
		"WINPOWER_EXPORTER_METRICS_SUBSYSTEM",
	}

	for _, varName := range requiredVars {
		if value, exists := envVars[varName]; !exists || value == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("required environment variable %s is not set", varName))
			result.Valid = false
		}
	}

	// 验证桶配置的环境变量
	bucketVars := map[string]string{
		"WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS":    "request_duration",
		"WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS": "collection_duration",
		"WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS":        "api_response",
	}

	for varName, bucketType := range bucketVars {
		if value, exists := envVars[varName]; exists && value != "" {
			v.validateBucketsFromJSON(value, bucketType, &result)
		}
	}

	return result
}

// ValidateBucketsFromJSON 从 JSON 字符串验证桶配置
func (v *Validator) validateBucketsFromJSON(jsonStr, bucketType string, result *ValidationResult) {
	var buckets []float64
	if err := json.Unmarshal([]byte(jsonStr), &buckets); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid JSON format for %s buckets: %v", bucketType, err))
		result.Valid = false
		return
	}

	v.validateBuckets(buckets, bucketType, result)
}

// ValidateURL 验证 URL 格式
func (v *Validator) ValidateURL(urlStr, fieldName string) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []string{},
		Warns:  []string{},
	}

	if urlStr == "" {
		result.Errors = append(result.Errors, fmt.Sprintf("%s cannot be empty", fieldName))
		result.Valid = false
		return result
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s is not a valid URL: %v", fieldName, err))
		result.Valid = false
		return result
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		result.Errors = append(result.Errors, fmt.Sprintf("%s must use http or https scheme, got: %s", fieldName, parsedURL.Scheme))
		result.Valid = false
	}

	if parsedURL.Host == "" {
		result.Errors = append(result.Errors, fmt.Sprintf("%s must have a valid host", fieldName))
		result.Valid = false
	}

	// 检查端口是否在合理范围内
	if parsedURL.Port() != "" {
		if port, err := strconv.Atoi(parsedURL.Port()); err == nil {
			if port < 1 || port > 65535 {
				result.Errors = append(result.Errors, fmt.Sprintf("%s has invalid port number: %d", fieldName, port))
				result.Valid = false
			}
		}
	}

	return result
}

// ValidatePort 验证端口号
func (v *Validator) ValidatePort(portStr, fieldName string) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []string{},
		Warns:  []string{},
	}

	if portStr == "" {
		result.Warns = append(result.Warns, fmt.Sprintf("%s is not specified, will use default", fieldName))
		return result
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s is not a valid number: %v", fieldName, err))
		result.Valid = false
		return result
	}

	if port < 1 || port > 65535 {
		result.Errors = append(result.Errors, fmt.Sprintf("%s must be between 1 and 65535, got: %d", fieldName, port))
		result.Valid = false
	}

	// 检查常用端口
	wellKnownPorts := map[int]string{
		80:   "HTTP",
		443:  "HTTPS",
		8080: "HTTP Alternate",
		9090: "Prometheus",
		9091: "Prometheus Alternate",
	}

	if serviceName, exists := wellKnownPorts[port]; exists {
		result.Warns = append(result.Warns, fmt.Sprintf("%s is using well-known port %d (%s), ensure this is intentional", fieldName, port, serviceName))
	}

	return result
}

// ValidateTimeout 验证超时配置
func (v *Validator) ValidateTimeout(timeoutStr, fieldName string) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []string{},
		Warns:  []string{},
	}

	if timeoutStr == "" {
		result.Warns = append(result.Warns, fmt.Sprintf("%s is not specified, will use default", fieldName))
		return result
	}

	duration, err := time.ParseDuration(timeoutStr)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s is not a valid duration: %v", fieldName, err))
		result.Valid = false
		return result
	}

	// 检查超时时间的合理性
	switch fieldName {
	case "request_timeout", "connection_timeout":
		if duration < 5*time.Second {
			result.Warns = append(result.Warns, fmt.Sprintf("%s (%s) might be too short, consider using at least 5s", fieldName, timeoutStr))
		}
		if duration > 5*time.Minute {
			result.Warns = append(result.Warns, fmt.Sprintf("%s (%s) is very long, consider reducing to avoid hanging", fieldName, timeoutStr))
		}
	case "read_timeout", "write_timeout":
		if duration < 10*time.Second {
			result.Warns = append(result.Warns, fmt.Sprintf("%s (%s) might be too short for large responses", fieldName, timeoutStr))
		}
	}

	return result
}

// ValidateLogLevel 验证日志级别
func (v *Validator) ValidateLogLevel(level, fieldName string) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []string{},
		Warns:  []string{},
	}

	if level == "" {
		result.Warns = append(result.Warns, fmt.Sprintf("%s is not specified, will use default", fieldName))
		return result
	}

	validLevels := []string{"debug", "info", "warn", "error"}
	isValid := false
	for _, validLevel := range validLevels {
		if strings.ToLower(level) == validLevel {
			isValid = true
			break
		}
	}

	if !isValid {
		result.Errors = append(result.Errors, fmt.Sprintf("%s must be one of %v, got: %s", fieldName, validLevels, level))
		result.Valid = false
	}

	return result
}

// GenerateConfigSuggestions 生成配置建议
func (v *Validator) GenerateConfigSuggestions(cfg MetricManagerConfig) []string {
	suggestions := []string{}

	// 根据当前配置生成建议
	if len(cfg.RequestDurationBuckets) < 5 {
		suggestions = append(suggestions, "Consider adding more request_duration buckets for better granularity")
	}

	if len(cfg.CollectionDurationBuckets) < 5 {
		suggestions = append(suggestions, "Consider adding more collection_duration buckets for better granularity")
	}

	if cfg.RequestDurationBuckets[0] > 0.01 {
		suggestions = append(suggestions, "Consider adding smaller request_duration buckets for fast requests")
	}

	if cfg.CollectionDurationBuckets[len(cfg.CollectionDurationBuckets)-1] < 60 {
		suggestions = append(suggestions, "Consider adding larger collection_duration buckets for slow collections")
	}

	// 根据命名空间和子系统生成建议
	if strings.Contains(cfg.Namespace, "_") && len(strings.Split(cfg.Namespace, "_")) > 3 {
		suggestions = append(suggestions, "Consider simplifying namespace to reduce metric name length")
	}

	if strings.Contains(cfg.Subsystem, "_") && len(strings.Split(cfg.Subsystem, "_")) > 3 {
		suggestions = append(suggestions, "Consider simplifying subsystem to reduce metric name length")
	}

	return suggestions
}

// Helper functions

// isValidPrometheusName 检查是否为有效的 Prometheus 名称
func isValidPrometheusName(name string) bool {
	// Prometheus 指标名称规范：只能包含字母、数字、下划线，且不能以数字开头
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	return matched
}

// FormatValidationResult 格式化验证结果为字符串
func FormatValidationResult(result ValidationResult) string {
	if result.Valid {
		if len(result.Warns) == 0 {
			return "✅ Configuration is valid"
		}
		return fmt.Sprintf("✅ Configuration is valid (with %d warnings)", len(result.Warns))
	}

	return fmt.Sprintf("❌ Configuration is invalid (%d errors, %d warnings)", len(result.Errors), len(result.Warns))
}

// PrintValidationResult 打印验证结果
func (v *Validator) PrintValidationResult(result ValidationResult) {
	fmt.Println(FormatValidationResult(result))

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if len(result.Warns) > 0 {
		fmt.Println("\nWarnings:")
		for _, warn := range result.Warns {
			fmt.Printf("  - %s\n", warn)
		}
	}
}
