# energy-calculator Specification

## Purpose
TBD - created by archiving change implement-energy-module. Update Purpose after archive.
## Requirements
### Requirement: 电能累计计算
The system SHALL calculate accumulated energy consumption based on device power readings using power×time interval integration.

#### Scenario: 首次设备电能计算
- **GIVEN** 系统接收到设备"ups-001"的功率读数500W
- **WHEN** 调用energy.Calculate("ups-001", 500.0)
- **THEN** 系统应返回累计电能值0.0Wh
- **AND** 系统应将计算结果持久化到storage

```go
// 创建电能服务
energyService := energy.NewEnergyService(storageManager, logger, config)

// 首次计算
totalEnergy, err := energyService.Calculate("ups-001", 500.0)
assert.NoError(t, err)
assert.Equal(t, 0.0, totalEnergy) // 首次计算从0开始
```

#### Scenario: 存在历史数据的电能累计
- **GIVEN** 设备"ups-001"历史累计电能为1000Wh
- **AND** 上次更新时间为5分钟前
- **WHEN** 接收到当前功率读数600W并调用energy.Calculate("ups-001", 600.0)
- **THEN** 系统应计算间隔电能：600W × (5分钟/60分钟) = 50Wh
- **AND** 系统应返回新的累计电能：1000Wh + 50Wh = 1050Wh
- **AND** 系统应更新storage中的电能数据和时间戳

```go
// 模拟历史数据
storageManager.Write("ups-001", &storage.PowerData{
    Timestamp: time.Now().Add(-5 * time.Minute).UnixMilli(),
    EnergyWH:  1000.0,
})

// 计算新的累计电能
totalEnergy, err := energyService.Calculate("ups-001", 600.0)
assert.NoError(t, err)
assert.Equal(t, 1050.0, totalEnergy)
```

### Requirement: 负功率处理
The system SHALL handle negative power readings to represent energy feedback or net energy reduction.

#### Scenario: 负功率电能计算
- **GIVEN** 设备"ups-001"历史累计电能为2000Wh
- **AND** 当前功率读数为-100W（能量反馈）
- **WHEN** 调用energy.Calculate("ups-001", -100.0)
- **THEN** 系统应计算负的间隔电能
- **AND** 系统应返回减少后的累计电能值
- **AND** 累计电能值允许为负数，表示净能量

```go
// 设置负功率计算
storageManager.Write("ups-001", &storage.PowerData{
    Timestamp: time.Now().Add(-10 * time.Minute).UnixMilli(),
    EnergyWH:  2000.0,
})

totalEnergy, err := energyService.Calculate("ups-001", -100.0)
assert.NoError(t, err)
assert.Less(t, totalEnergy, 2000.0) // 电能应该减少
```

### Requirement: 数据查询接口
The system SHALL provide an interface to query the latest accumulated energy data for devices.

#### Scenario: 查询设备电能数据
- **GIVEN** 设备"ups-001"已有累计电能数据
- **WHEN** 调用energy.Get("ups-001")
- **THEN** 系统应返回最新的累计电能值
- **AND** 返回值应与storage中的数据一致

```go
// 查询电能数据
currentEnergy, err := energyService.Get("ups-001")
assert.NoError(t, err)
assert.Greater(t, currentEnergy, 0.0)
```

#### Scenario: 查询不存在设备的电能数据
- **GIVEN** 设备"new-device"从未进行过电能计算
- **WHEN** 调用energy.Get("new-device")
- **THEN** 系统应返回0.0Wh

```go
// 查询不存在设备
energy, err := energyService.Get("new-device")
assert.NoError(t, err)
assert.Equal(t, 0.0, energy)
```

### Requirement: 串行执行保证
The system SHALL ensure all energy calculation operations execute serially through a global lock mechanism to avoid concurrency issues.

#### Scenario: 并发计算安全性
- **GIVEN** 多个goroutine同时调用energy.Calculate不同设备
- **WHEN** 系统处理这些并发请求
- **THEN** 所有计算必须串行执行，不产生数据竞争
- **AND** 每个设备的电能数据独立计算，互不影响

```go
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(deviceID string, power float64) {
        defer wg.Done()
        _, err := energyService.Calculate(deviceID, power)
        assert.NoError(t, err)
    }(fmt.Sprintf("device-%d", i%3), float64(100+i*10))
}
wg.Wait()
```

### Requirement: 输入参数验证
The system SHALL validate input parameters and handle various edge cases.

#### Scenario: 空设备ID处理
- **GIVEN** 调用energy.Calculate("", 500.0)
- **WHEN** 系统执行计算
- **THEN** 应返回参数错误

```go
_, err := energyService.Calculate("", 500.0)
assert.Error(t, err)
assert.Contains(t, err.Error(), "device ID cannot be empty")
```

#### Scenario: 无效功率值处理
- **GIVEN** 调用energy.Calculate("ups-001", math.NaN())
- **WHEN** 系统执行计算
- **THEN** 应返回参数错误

```go
_, err := energyService.Calculate("ups-001", math.NaN())
assert.Error(t, err)
assert.Contains(t, err.Error(), "invalid power value")
```

### Requirement: 计算统计信息
The system SHALL collect and maintain basic calculation statistics.

#### Scenario: 统计功能验证
- **GIVEN** 系统已完成多次电能计算
- **WHEN** 调用energy.GetStats()
- **THEN** 应返回正确的统计信息

```go
// 执行多次计算
for i := 0; i < 5; i++ {
    energyService.Calculate("test-device", float64(100+i))
}

stats := energyService.GetStats()
assert.Equal(t, int64(5), stats.TotalCalculations)
assert.Equal(t, int64(0), stats.TotalErrors)
```

### Requirement: 性能约束
The system SHALL meet performance requirements for energy calculation operations.

#### Scenario: 单次计算性能要求
- **GIVEN** 执行单次电能计算操作
- **WHEN** 调用energy.Calculate()方法
- **THEN** 计算延迟应小于10ms
- **AND** 性能测试执行时长不超过5秒

```go
start := time.Now()
_, err := energyService.Calculate("test-device", 500.0)
duration := time.Since(start)

assert.NoError(t, err)
assert.Less(t, duration, 10*time.Millisecond)
```

#### Scenario: 并发计算性能要求
- **GIVEN** 多设备并发计算场景
- **WHEN** 多个goroutine同时调用energy.Calculate()
- **THEN** 整体性能测试应在5秒内完成
- **AND** 单个计算延迟仍满足10ms要求

```go
// 性能测试约束：总执行时间不超过5秒
testStart := time.Now()
defer func() {
    totalDuration := time.Since(testStart)
    assert.Less(t, totalDuration, 5*time.Second, "Performance test exceeded 5 second limit")
}()

var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(deviceID string, power float64) {
        defer wg.Done()
        start := time.Now()
        _, err := energyService.Calculate(deviceID, power)
        duration := time.Since(start)

        assert.NoError(t, err)
        assert.Less(t, duration, 10*time.Millisecond)
    }(fmt.Sprintf("device-%d", i), float64(100+i*10))
}
wg.Wait()
```

