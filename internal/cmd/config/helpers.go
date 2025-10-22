package config

// BoolValue 返回 *bool 的值，如果为 nil 则返回默认值
// 用于安全地获取可选 bool 字段的值
func BoolValue(b *bool, defaultValue bool) bool {
	if b == nil {
		return defaultValue
	}
	return *b
}

// BoolPtr 返回 bool 值的指针
// 用于创建 *bool 类型的值
func BoolPtr(b bool) *bool {
	return &b
}
