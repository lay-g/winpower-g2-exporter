# energy-config Specification

## Purpose
TBD - created by archiving change implement-energy-module. Update Purpose after archive.
## Requirements
### Requirement: Energy模块配置结构
The system SHALL define its own configuration structure containing all configuration parameters related to energy calculation.

#### Scenario: 基本配置结构验证
- **GIVEN** 查看energy模块的配置定义
- **WHEN** 检查Config结构体字段
- **THEN** 应包含以下字段：计算精度、统计开关、超时时间、负功率允许等

```go
type Config struct {
    Precision           float64         `yaml:"precision" default:"0.01"`
    EnableStats         bool            `yaml:"enable_stats" default:"true"`
    MaxCalculationTime  time.Duration   `yaml:"max_calculation_time" default:"1s"`
    NegativePowerAllowed bool           `yaml:"negative_power_allowed" default:"true"`
}
```

#### Scenario: 默认配置生成
- **GIVEN** 调用energy.DefaultConfig()
- **WHEN** 返回配置实例
- **THEN** 所有字段应使用合理的默认值
- **AND** 配置应通过验证检查

```go
config := energy.DefaultConfig()
assert.Equal(t, 0.01, config.Precision)
assert.True(t, config.EnableStats)
assert.Equal(t, time.Second, config.MaxCalculationTime)
assert.True(t, config.NegativePowerAllowed)
```

### Requirement: 配置验证机制
The system SHALL provide configuration validation functionality to ensure the validity of configuration parameters.

#### Scenario: 有效配置验证
- **GIVEN** 创建包含有效参数的配置
- **WHEN** 调用config.Validate()
- **THEN** 应返回nil（无错误）

```go
config := &energy.Config{
    Precision:           0.001,
    EnableStats:         true,
    MaxCalculationTime:  2 * time.Second,
    NegativePowerAllowed: true,
}
err := config.Validate()
assert.NoError(t, err)
```

#### Scenario: 无效精度值验证
- **GIVEN** 创建配置，设置Precision为负数
- **WHEN** 调用config.Validate()
- **THEN** 应返回验证错误

```go
config := &energy.Config{Precision: -0.01}
err := config.Validate()
assert.Error(t, err)
assert.Contains(t, err.Error(), "precision must be positive")
```

### Requirement: 构造函数配置注入
The system SHALL receive configuration parameters through constructor functions without directly loading from files or environment variables.

#### Scenario: 构造函数接收配置
- **GIVEN** 已创建energy.Config实例
- **AND** 已创建storage.StorageManager实例
- **AND** 已创建logger实例
- **WHEN** 调用energy.NewEnergyService(storage, logger, config)
- **THEN** 应返回使用指定配置的EnergyService实例

```go
config := energy.DefaultConfig()
config.Precision = 0.001
service := energy.NewEnergyService(storageManager, logger, config)
assert.NotNil(t, service)
```

#### Scenario: 空配置使用默认值
- **GIVEN** 已创建storage和logger实例
- **AND** config参数为nil
- **WHEN** 调用energy.NewEnergyService(storage, logger, nil)
- **THEN** 应使用默认配置创建服务

```go
service := energy.NewEnergyService(storageManager, logger, nil)
assert.NotNil(t, service)
// 服务应使用默认配置正常工作
```

