package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// stringToDurationHookFunc 将字符串转换为 time.Duration
func stringToDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Duration(0)) {
			return data, nil
		}

		// 转换字符串为 time.Duration
		return time.ParseDuration(data.(string))
	}
}

// ValidationFunc 验证函数类型
type ValidationFunc func(value interface{}) error

// Loader 配置加载器
type Loader struct {
	prefix          string                      // 环境变量前缀
	v               *viper.Viper                // Viper实例
	configMap       map[string]interface{}      // 配置映射
	mu              sync.RWMutex                // 读写锁
	cacheEnabled    bool                        // 缓存开关
	validationRules map[string][]ValidationFunc // 验证规则映射
	deepCopyEnabled bool                        // 深拷贝开关
}

// NewLoader 创建新的配置加载器
func NewLoader(prefix string) *Loader {
	v := viper.New()

	// 设置环境变量前缀
	v.SetEnvPrefix(prefix)

	// 设置环境变量替换器
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return &Loader{
		prefix:          prefix,
		v:               v,
		configMap:       make(map[string]interface{}),
		cacheEnabled:    true, // 默认启用缓存
		validationRules: make(map[string][]ValidationFunc),
		deepCopyEnabled: true, // 默认启用深拷贝
	}
}

// LoadModule 加载指定模块的配置
func (l *Loader) LoadModule(moduleName string, configStruct interface{}) error {
	l.mu.RLock()
	// 检查缓存
	if l.cacheEnabled {
		if cached, exists := l.configMap[moduleName]; exists {
			l.mu.RUnlock()
			return l.copyConfig(cached, configStruct)
		}
	}
	l.mu.RUnlock()

	// 读取配置文件
	configMap := make(map[string]interface{})
	if err := l.v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，不是致命错误，继续使用默认值
		if strings.Contains(err.Error(), "no such file or directory") ||
			strings.Contains(err.Error(), "Config File") && strings.Contains(err.Error(), "Not Found") {
			// 配置文件不存在，继续处理环境变量
		} else {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 获取模块配置
	moduleConfig := l.v.Sub(moduleName)
	if moduleConfig != nil {
		// 将模块配置的所有值复制到一个 map 中
		for key, value := range moduleConfig.AllSettings() {
			configMap[key] = value
		}
	}

	// 手动处理环境变量覆盖
	if err := l.applyEnvOverrides(moduleName, configStruct, configMap); err != nil {
		return fmt.Errorf("failed to apply environment variable overrides: %w", err)
	}

	// 使用 mapstructure 解组到结构体（如果配置映射不为空）
	if len(configMap) > 0 {
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:           configStruct,
			WeaklyTypedInput: true,
			TagName:          "yaml",
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				stringToDurationHookFunc(),
			),
		})
		if err != nil {
			return fmt.Errorf("failed to create decoder: %w", err)
		}

		if err := decoder.Decode(configMap); err != nil {
			return fmt.Errorf("failed to decode %s config: %w", moduleName, err)
		}
	}

	// 缓存配置
	l.mu.Lock()
	if l.cacheEnabled {
		if l.deepCopyEnabled {
			// 如果启用深拷贝，创建深拷贝后缓存
			deepCopy := l.deepCopyInterface(configStruct)
			l.configMap[moduleName] = deepCopy
		} else {
			l.configMap[moduleName] = configStruct
		}
	}
	l.mu.Unlock()

	return nil
}

// BindEnv 绑定环境变量
func (l *Loader) BindEnv(key string, envKeys ...string) error {
	// 构建完整的环境变量名称
	fullKey := l.prefix + "_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))

	// 如果提供了自定义环境变量键，使用第一个
	if len(envKeys) > 0 {
		fullKey = envKeys[0]
	}

	return l.v.BindEnv(key, fullKey)
}

// GetConfigPath 获取当前使用的配置文件路径
func (l *Loader) GetConfigPath() string {
	if l.v == nil {
		return ""
	}

	return l.v.ConfigFileUsed()
}

// GetString 获取字符串配置值
func (l *Loader) GetString(key string) string {
	return l.v.GetString(key)
}

// GetInt 获取整数配置值
func (l *Loader) GetInt(key string) int {
	return l.v.GetInt(key)
}

// GetBool 获取布尔配置值
func (l *Loader) GetBool(key string) bool {
	return l.v.GetBool(key)
}

// Set 设置配置值
func (l *Loader) Set(key string, value interface{}) {
	l.v.Set(key, value)
}

// SetConfigFile 设置配置文件路径
func (l *Loader) SetConfigFile(path string) {
	l.v.SetConfigFile(path)
}

// GetViper 获取 viper 实例（用于高级操作）
func (l *Loader) GetViper() *viper.Viper {
	return l.v
}

// applyEnvOverrides 手动应用环境变量覆盖
func (l *Loader) applyEnvOverrides(moduleName string, configStruct interface{}, configMap map[string]interface{}) error {
	val := reflect.ValueOf(configStruct)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("configStruct must be a non-nil pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("configStruct must point to a struct")
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// 获取 YAML 标签和环境变量标签
		yamlTag := field.Tag.Get("yaml")
		envTag := field.Tag.Get("env")
		if yamlTag == "" {
			continue
		}

		// 确定要检查的环境变量名称
		envVarName := ""
		if envTag != "" {
			// 如果有 env 标签，使用它
			envVarName = envTag
		} else {
			// 否则构造标准的环境变量名称
			envVarName = fmt.Sprintf("%s_%s_%s", l.prefix, strings.ToUpper(moduleName), strings.ToUpper(field.Name))
		}

		// 检查环境变量是否存在
		if envValue := os.Getenv(envVarName); envValue != "" {
			// 尝试将环境变量值转换为正确的类型
			convertedValue, err := l.convertEnvValue(envValue, field.Type)
			if err != nil {
				return fmt.Errorf("failed to convert env value for %s: %w", yamlTag, err)
			}

			// 直接设置结构体字段值
			convertedReflectValue := reflect.ValueOf(convertedValue)
			if convertedReflectValue.Type().ConvertibleTo(fieldValue.Type()) {
				fieldValue.Set(convertedReflectValue.Convert(fieldValue.Type()))
			} else if convertedReflectValue.Type() == fieldValue.Type() {
				fieldValue.Set(convertedReflectValue)
			}

			// 同时更新配置映射（如果有配置文件的话）
			if configMap != nil {
				configMap[yamlTag] = convertedValue
			}
		}
	}

	return nil
}

// convertEnvValue 将环境变量值转换为正确的类型
func (l *Loader) convertEnvValue(envValue string, fieldType reflect.Type) (interface{}, error) {
	switch fieldType.Kind() {
	case reflect.String:
		return envValue, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fieldType == reflect.TypeOf(time.Duration(0)) {
			return time.ParseDuration(envValue)
		}
		return strconv.ParseInt(envValue, 10, 64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Special handling for os.FileMode
		if fieldType == reflect.TypeOf(os.FileMode(0)) {
			mode, err := strconv.ParseUint(envValue, 8, 32) // Parse as octal
			if err != nil {
				return nil, fmt.Errorf("invalid file mode '%s', expected octal format (e.g., 0644)", envValue)
			}
			return os.FileMode(mode), nil
		}
		return strconv.ParseUint(envValue, 10, 64)
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(envValue, 64)
	case reflect.Bool:
		return strconv.ParseBool(envValue)
	default:
		return envValue, nil
	}
}

// DisableCache 禁用配置缓存
func (l *Loader) DisableCache() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cacheEnabled = false
	// 清空现有缓存
	l.configMap = make(map[string]interface{})
}

// EnableCache 启用配置缓存
func (l *Loader) EnableCache() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cacheEnabled = true
}

// IsCacheEnabled 检查缓存是否启用
func (l *Loader) IsCacheEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.cacheEnabled
}

// ClearCache 清空配置缓存
func (l *Loader) ClearCache() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.configMap = make(map[string]interface{})
}

// AddValidationRule 添加自定义验证规则
func (l *Loader) AddValidationRule(fieldName string, rule ValidationFunc) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.validationRules == nil {
		l.validationRules = make(map[string][]ValidationFunc)
	}

	l.validationRules[fieldName] = append(l.validationRules[fieldName], rule)
}

// Validate 验证配置（增强版本）
func (l *Loader) Validate(config Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 首先执行配置对象的验证方法
	if err := config.Validate(); err != nil {
		return err
	}

	// 然后执行自定义验证规则
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.validateWithRules(config)
}

// validateWithRules 使用自定义规则验证配置
func (l *Loader) validateWithRules(config Config) error {
	if len(l.validationRules) == 0 {
		return nil
	}

	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "" {
			continue
		}

		// 检查是否有该字段的验证规则
		if rules, exists := l.validationRules[yamlTag]; exists {
			fieldValue := val.Field(i)
			for _, rule := range rules {
				if err := rule(fieldValue.Interface()); err != nil {
					return fmt.Errorf("validation failed for field %s: %w", yamlTag, err)
				}
			}
		}
	}

	return nil
}

// MergeConfig 合并两个配置对象
func (l *Loader) MergeConfig(base, overlay, result interface{}) error {
	baseVal := reflect.ValueOf(base)
	overlayVal := reflect.ValueOf(overlay)
	resultVal := reflect.ValueOf(result)

	// 确保所有参数都是指针
	if baseVal.Kind() != reflect.Ptr || overlayVal.Kind() != reflect.Ptr || resultVal.Kind() != reflect.Ptr {
		return fmt.Errorf("all parameters must be pointers")
	}

	if baseVal.IsNil() || overlayVal.IsNil() || resultVal.IsNil() {
		return fmt.Errorf("all parameters must be non-nil")
	}

	// 获取实际的结构体值
	baseElem := baseVal.Elem()
	overlayElem := overlayVal.Elem()
	resultElem := resultVal.Elem()

	if baseElem.Kind() != reflect.Struct || overlayElem.Kind() != reflect.Struct || resultElem.Kind() != reflect.Struct {
		return fmt.Errorf("all parameters must point to structs")
	}

	// 确保类型相同
	if baseElem.Type() != overlayElem.Type() || baseElem.Type() != resultElem.Type() {
		return fmt.Errorf("all parameters must be of the same type")
	}

	// 首先复制基础配置的所有字段
	l.copyStruct(baseElem, resultElem)

	// 然后用覆盖配置的非零值字段覆盖结果配置
	l.mergeStruct(overlayElem, resultElem)

	return nil
}

// copyStruct 复制结构体
func (l *Loader) copyStruct(src, dst reflect.Value) {
	srcType := src.Type()
	for i := 0; i < srcType.NumField(); i++ {
		field := srcType.Field(i)
		if !field.IsExported() {
			continue
		}

		srcField := src.Field(i)
		dstField := dst.Field(i)

		if dstField.CanSet() {
			dstField.Set(srcField)
		}
	}
}

// mergeStruct 合并结构体（仅覆盖非零值）
func (l *Loader) mergeStruct(src, dst reflect.Value) {
	srcType := src.Type()
	for i := 0; i < srcType.NumField(); i++ {
		field := srcType.Field(i)
		if !field.IsExported() {
			continue
		}

		srcField := src.Field(i)
		dstField := dst.Field(i)

		if !dstField.CanSet() {
			continue
		}

		// 检查源字段是否为零值
		if !l.isZeroValue(srcField) {
			dstField.Set(srcField)
		}
	}
}

// isZeroValue 检查值是否为零值
func (l *Loader) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Chan, reflect.Func:
		return v.IsNil()
	case reflect.Struct:
		// 特殊处理 time.Duration
		if v.Type() == reflect.TypeOf(time.Duration(0)) {
			return v.Interface().(time.Duration) == 0
		}
		return false
	default:
		return false
	}
}

// copyConfig 复制配置对象
func (l *Loader) copyConfig(src, dst interface{}) error {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	if srcVal.Kind() != reflect.Ptr || dstVal.Kind() != reflect.Ptr {
		return fmt.Errorf("both src and dst must be pointers")
	}

	if srcVal.IsNil() || dstVal.IsNil() {
		return fmt.Errorf("both src and dst must be non-nil")
	}

	srcElem := srcVal.Elem()
	dstElem := dstVal.Elem()

	if srcElem.Kind() != reflect.Struct || dstElem.Kind() != reflect.Struct {
		return fmt.Errorf("both src and dst must point to structs")
	}

	if srcElem.Type() != dstElem.Type() {
		return fmt.Errorf("src and dst must be of the same type")
	}

	l.copyStruct(srcElem, dstElem)
	return nil
}

// deepCopyInterface 创建接口的深拷贝
func (l *Loader) deepCopyInterface(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	// 使用 JSON 序列化/反序列化实现深拷贝
	// 这是一种简单但有效的方式
	// 对于性能敏感的场景，可以考虑使用其他方法
	return l.deepCopyValue(reflect.ValueOf(src)).Interface()
}

// deepCopyValue 创建值的深拷贝
func (l *Loader) deepCopyValue(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return reflect.Value{}
	}

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		// 创建新指针
		newPtr := reflect.New(v.Type().Elem())
		// 递归复制指针指向的值
		newPtr.Elem().Set(l.deepCopyValue(v.Elem()))
		return newPtr

	case reflect.Struct:
		// 特殊处理 time.Duration
		if v.Type() == reflect.TypeOf(time.Duration(0)) {
			return v
		}

		// 创建新结构体
		newStruct := reflect.New(v.Type()).Elem()
		// 递归复制每个字段
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).PkgPath != "" { // 非导出字段
				continue
			}
			newStruct.Field(i).Set(l.deepCopyValue(v.Field(i)))
		}
		return newStruct

	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		// 创建新切片
		newSlice := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
		// 递归复制每个元素
		for i := 0; i < v.Len(); i++ {
			newSlice.Index(i).Set(l.deepCopyValue(v.Index(i)))
		}
		return newSlice

	case reflect.Map:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		// 创建新映射
		newMap := reflect.MakeMap(v.Type())
		// 递归复制每个键值对
		for _, key := range v.MapKeys() {
			newKey := l.deepCopyValue(key)
			newVal := l.deepCopyValue(v.MapIndex(key))
			newMap.SetMapIndex(newKey, newVal)
		}
		return newMap

	case reflect.Array:
		// 创建新数组
		newArray := reflect.New(v.Type()).Elem()
		// 递归复制每个元素
		for i := 0; i < v.Len(); i++ {
			newArray.Index(i).Set(l.deepCopyValue(v.Index(i)))
		}
		return newArray

	case reflect.Interface:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		// 递归复制接口包含的值
		return l.deepCopyValue(v.Elem())

	default:
		// 基本类型（int, string, bool等）直接复制
		return v
	}
}
