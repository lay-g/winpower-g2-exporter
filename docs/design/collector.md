# Collector模块设计文档

## 概述

Collector模块是WinPower G2 Prometheus Exporter中的核心协调模块，负责从WinPower模块获取设备数据，并作为**唯一触发电能计算的模块**。该模块承担着数据采集协调、电能触发和指标更新的重要职责。

### 设计目标

- **统一协调**: 作为数据流的中枢，协调WinPower、Energy和Metrics模块
- **唯一触发**: 作为唯一触发电能计算的模块，避免多重触发导致的混乱
- **职责明确**: 清晰界定数据采集和计算触发的边界
- **可靠性**: 确保数据采集和电能计算的完整性和一致性

## 架构设计

### 模块职责

Collector模块在系统架构中处于协调位置，被调度器和Metrics模块调用，协调各个模块完成数据采集和计算：

```text
                调度器调用 (每5秒)               Metrics模块调用 (按需)
                           │                              │
                           ▼                              ▼
    ┌─────────────────────────────────────────────────────────────┐
    │                  Collector Module                          │
    │                (数据采集与触发协调)                          │
    │                                                             │
    │  ┌─────────────────────────────────────────────────────┐   │
    │  │              CollectorService                        │   │
    │  │                                                     │   │
    │  │  ┌─────────────────┐    ┌─────────────────────┐     │   │
    │  │  │                 │    │                     │     │   │
    │  │  │   数据获取协调   │    │    电能计算触发      │     │   │
    │  │  │                 │    │                     │     │   │
    │  │  │ • 调用WinPower  │    │ • 触发Energy模块    │     │   │
    │  │  │ • 解析响应数据   │    │ • 处理计算错误      │     │   │
    │  │  │ • 处理采集错误   │    │ • 记录执行日志      │     │   │
    │  │  │                 │    │                     │     │   │
    │  │  └─────────────────┘    └─────────────────────┘     │   │
    │  └─────────────────────────────────────────────────────┘   │
    └─────────────────────────────────────────────────────────────┘
                           │               │
                           ▼               ▼
    ┌─────────────────┐           ┌─────────────────┐
    │ WinPower Client │           │ Energy Service  │
    │                 │           │                 │
    │ • 设备数据采集   │           │ • 电能计算       │
    │ • 响应数据解析   │           │ • 结果持久化     │
    │ • 连接状态管理   │           │ • 错误处理       │
    └─────────────────┘           └─────────────────┘
```

### 数据流程图

```text
                调度器触发 (每5秒)               Metrics模块触发 (按需)
                           │                              │
                           ▼                              ▼
                ┌─────────────────────────────────────────────────────┐
                │         Collector.CollectDeviceData()              │
                │            (统一的数据采集入口)                      │
                └─────────────────────────────────────────────────────┘
                                          │
                                          ▼
                ┌─────────────────────┐
                │ 调用WinPower        │
                │ CollectDeviceData   │
                └─────────────────────┘
                           │
                           ▼
                ┌─────────────────────┐
                │ 获取设备数据         │
                │ • LoadTotalWatt     │
                │ • 设备状态信息       │
                └─────────────────────┘
                           │
                           ▼
                ┌─────────────────────┐
                │ 遍历每个设备         │
                │ 触发电能计算         │
                └─────────────────────┘
                           │
                           ▼
                ┌─────────────────────┐
                │ Energy.Calculate    │
                │ (deviceID, power)   │
                └─────────────────────┘
                           │
                           ▼
                ┌─────────────────────┐
                │ 记录执行日志         │
                │ 返回采集结果         │
                └─────────────────────┘
```

**关键流程说明：**

1. **双重触发机制**:
   - **调度器触发**: 每5秒定时触发一次Collector采集，用于常规的数据更新和电能累计
   - **Metrics触发**: 当Prometheus通过`/metrics`端点请求指标时，Collector会触发即时采集以确保返回最新的设备状态

2. **数据获取**: Collector调用WinPower获取最新的设备数据

3. **功率提取**: 直接使用WinPower返回的总负载有功功率(LoadTotalWatt)

4. **电能计算**: 为每个设备触发电能计算，Energy模块负责累计计算

5. **结果处理**: 记录详细的执行日志，失败直接返回错误由上层处理

**触发场景对比：**

| 触发源  | 频率  | 目的         | 数据用途                 |
| ------- | ----- | ------------ | ------------------------ |
| 调度器  | 每5秒 | 常规数据更新 | 电能累计、状态监控       |
| Metrics | 按需  | 实时指标提供 | Prometheus抓取、实时监控 |

## 核心组件设计

### 1. CollectorService

#### 职责

- 协调数据采集流程
- **作为唯一触发电能计算的模块**
- 使用WinPower提供的原始数据，直接处理
- 处理错误情况，失败直接返回

#### 数据结构

```go
// CollectorService 数据采集和协调服务
type CollectorService struct {
    winpowerClient  winpower.WinPowerClient  // WinPower客户端
    energyService   energy.EnergyInterface   // 电能计算服务
    logger          *zap.Logger              // 日志记录器
}

// CollectionResult 采集结果结构体
type CollectionResult struct {
    Success        bool                              `json:"success"`         // 采集是否成功
    DeviceCount    int                               `json:"device_count"`    // 采集到的设备数量
    Devices        map[string]*DeviceCollectionInfo  `json:"devices"`         // 设备采集信息，key为device_id
    CollectionTime time.Time                         `json:"collection_time"` // 采集时间
    Duration       time.Duration                     `json:"duration"`        // 采集耗时
    ErrorMessage   string                            `json:"error_message"`   // 错误信息（如果有）
}

// DeviceCollectionInfo 设备采集完整信息
type DeviceCollectionInfo struct {
    // 基本信息
    DeviceID       string    `json:"device_id"`        // 设备ID
    DeviceName     string    `json:"device_name"`      // 设备名称
    DeviceType     int       `json:"device_type"`      // 设备类型
    DeviceModel    string    `json:"device_model"`     // 设备型号
    Connected      bool      `json:"connected"`        // 连接状态
    LastUpdateTime time.Time `json:"last_update_time"` // 最后更新时间

    // 输入电气参数
    InputVolt1     float64 `json:"input_volt_1"`      // 输入电压 (相1) (V)
    InputFreq      float64 `json:"input_freq"`        // 输入频率 (Hz)

    // 输出电气参数
    OutputVolt1    float64 `json:"output_volt_1"`     // 输出电压 (相1) (V)
    OutputCurrent1 float64 `json:"output_current_1"`  // 输出电流 (相1) (A)
    OutputFreq     float64 `json:"output_freq"`       // 输出频率 (Hz)
    OutputVoltageType int   `json:"output_voltage_type"` // 输出电压类型

    // 负载和功率参数
    LoadPercent    float64 `json:"load_percent"`      // 负载百分比 (%)
    LoadTotalWatt  float64 `json:"load_total_watt"`   // 总负载有功功率 (W) - 能耗计算使用此字段
    LoadTotalVa    float64 `json:"load_total_va"`     // 总负载视在功率 (VA)
    LoadWatt1      float64 `json:"load_watt_1"`       // 负载有功功率 (相1) (W)
    LoadVa1        float64 `json:"load_va_1"`         // 负载视在功率 (相1) (VA)

    // 电池参数
    IsCharging     bool    `json:"is_charging"`       // 是否正在充电 (1=是, 0=否)
    BatVoltP       float64 `json:"bat_volt_p"`        // 电池电压百分比 (V)
    BatCapacity    float64 `json:"bat_capacity"`      // 电池容量 (%)
    BatRemainTime  int     `json:"bat_remain_time"`   // 电池剩余时间 (秒)
    BatteryStatus  int     `json:"battery_status"`    // 电池状态

    // UPS状态参数
    UpsTemperature float64 `json:"ups_temperature"`   // UPS 温度 (°C)
    Mode           int     `json:"mode"`              // UPS 工作模式
    Status         int     `json:"status"`            // 设备状态
    TestStatus     int     `json:"test_status"`       // 测试状态
    FaultCode      string  `json:"fault_code"`        // 故障代码

    // 其他参数
    InputTransformerType int `json:"input_transformer_type"` // 输入变压器类型

    // 能耗计算结果
    EnergyCalculated bool    `json:"energy_calculated"` // 电能计算是否成功
    EnergyValue      float64 `json:"energy_value"`      // 累计电能值 (Wh)

    // 错误信息
    ErrorMsg        string  `json:"error_msg"`         // 设备相关错误信息
}
```

#### 核心方法

```go
// NewCollectorService 创建采集服务
func NewCollectorService(
    winpowerClient winpower.WinPowerClient,
    energyService energy.EnergyInterface,
    logger *zap.Logger,
) *CollectorService

// CollectDeviceData 采集设备数据并触发电能计算（统一入口）
func (c *CollectorService) CollectDeviceData(ctx context.Context) (*CollectionResult, error)
```

#### 实现要点

```go
func (c *CollectorService) CollectDeviceData(ctx context.Context) (*CollectionResult, error)
// 实现逻辑（统一入口）：
// 1. 记录开始时间和创建组件专用日志器
// 2. 调用WinPower客户端采集设备数据
// 3. 遍历所有设备，直接使用WinPower返回的总负载有功功率(LoadTotalWatt)触发电能计算
// 4. 记录采集完成日志并返回CollectionResult结构体
// 注意：失败直接返回错误，不进行重试

func (c *CollectorService) collectFromWinPower(ctx context.Context) (*winpower.ParsedDeviceData, error)
// 实现逻辑：
// 1. 调用WinPower客户端的CollectDeviceData方法
// 2. 返回采集到的设备数据或错误

func (c *CollectorService) processDeviceData(ctx context.Context, data *winpower.ParsedDeviceData) (*CollectionResult, error)
// 实现逻辑：
// 1. 遍历所有设备，触发电能计算
// 2. 统一处理设备数据和错误
// 3. 返回CollectionResult结构体包含采集结果

func NewCollectorService(
    winpowerClient winpower.WinPowerClient,
    energyService energy.EnergyInterface,
    logger *zap.Logger,
) *CollectorService
// 实现逻辑：
// 1. 保存WinPower客户端引用
// 2. 保存Energy服务引用
// 3. 保存日志器
// 4. 返回服务实例
```

### 数据转换映射

CollectionResult结构体中的字段与WinPower API返回值的对应关系如下：

**CollectionResult → WinPower API 映射表**

| CollectionResult字段 | WinPower API字段 | 数据来源路径               | 说明                       |
| -------------------- | ---------------- | -------------------------- | -------------------------- |
| Success              | -                | 计算得出                   | 基于整体采集和计算的成功性 |
| DeviceCount          | total            | response.data.total        | 采集到的设备总数           |
| Devices              | data             | response.data              | 设备详细信息映射           |
| CollectionTime       | -                | 系统生成                   | Collector完成采集的时间戳  |
| Duration             | -                | 计算得出                   | 采集开始到结束的总耗时     |
| ErrorMessage         | code/msg         | response.code/response.msg | 当采集失败时的错误描述     |

**DeviceCollectionInfo → WinPower API 完整映射表**

| DeviceCollectionInfo字段 | WinPower API字段     | 数据来源路径                                  | 类型转换                 |
| ------------------------ | -------------------- | --------------------------------------------- | ------------------------ |
| **基本信息**             |                      |                                               |                          |
| DeviceID                 | id                   | response.data[].assetDevice.id                | string → string          |
| DeviceName               | alias                | response.data[].assetDevice.alias             | string → string          |
| DeviceType               | deviceType           | response.data[].assetDevice.deviceType        | int → int                |
| DeviceModel              | model                | response.data[].assetDevice.model             | string → string          |
| Connected                | connected            | response.data[].connected                     | bool → bool              |
| LastUpdateTime           | -                    | 系统生成                                      | time.Time                |
| **输入电气参数**         |                      |                                               |                          |
| InputVolt1               | inputVolt1           | response.data[].realtime.inputVolt1           | string → float64         |
| InputFreq                | inputFreq            | response.data[].realtime.inputFreq            | string → float64         |
| **输出电气参数**         |                      |                                               |                          |
| OutputVolt1              | outputVolt1          | response.data[].realtime.outputVolt1          | string → float64         |
| OutputCurrent1           | outputCurrent1       | response.data[].realtime.outputCurrent1       | string → float64         |
| OutputFreq               | outputFreq           | response.data[].realtime.outputFreq           | string → float64         |
| OutputVoltageType        | outputVoltageType    | response.data[].realtime.outputVoltageType    | string → int             |
| **负载和功率参数**       |                      |                                               |                          |
| LoadPercent              | loadPercent          | response.data[].realtime.loadPercent          | string → float64         |
| LoadTotalWatt            | loadTotalWatt        | response.data[].realtime.loadTotalWatt        | string → float64         |
| LoadTotalVa              | loadTotalVa          | response.data[].realtime.loadTotalVa          | string → float64         |
| LoadWatt1                | loadWatt1            | response.data[].realtime.loadWatt1            | string → float64         |
| LoadVa1                  | loadVa1              | response.data[].realtime.loadVa1              | string → float64         |
| **电池参数**             |                      |                                               |                          |
| IsCharging               | isCharging           | response.data[].realtime.isCharging           | string → bool            |
| BatVoltP                 | batVoltP             | response.data[].realtime.batVoltP             | string → float64         |
| BatCapacity              | batCapacity          | response.data[].realtime.batCapacity          | string → float64         |
| BatRemainTime            | batRemainTime        | response.data[].realtime.batRemainTime        | string → int             |
| BatteryStatus            | batteryStatus        | response.data[].realtime.batteryStatus        | string → int             |
| **UPS状态参数**          |                      |                                               |                          |
| UpsTemperature           | upsTemperature       | response.data[].realtime.upsTemperature       | string → float64         |
| Mode                     | mode                 | response.data[].realtime.mode                 | string → int             |
| Status                   | status               | response.data[].realtime.status               | string → int             |
| TestStatus               | testStatus           | response.data[].realtime.testStatus           | string → int             |
| FaultCode                | faultCode            | response.data[].realtime.faultCode            | string → string          |
| **其他参数**             |                      |                                               |                          |
| InputTransformerType     | inputTransformerType | response.data[].realtime.inputTransformerType | string → int             |
| **能耗计算结果**         |                      |                                               |                          |
| EnergyCalculated         | -                    | 计算得出                                      | Energy模块计算成功标志   |
| EnergyValue              | -                    | Energy模块返回                                | float64 → float64        |
| **错误信息**             |                      |                                               |                          |
| ErrorMsg                 | -                    | 错误处理                                      | 设备处理过程中的错误信息 |

**关键字段转换说明：**

1. **功率相关字段 (能耗计算核心)**
   ```go
   // WinPower API 返回 (字符串格式)
   "realtime": {
       "loadTotalWatt": "500.5",    // 总负载有功功率 (W)
       "loadTotalVa": "600.0",      // 总负载视在功率 (VA)
       "loadWatt1": "250.25",       // 相1有功功率 (W)
       "loadVa1": "300.0"           // 相1视在功率 (VA)
   }

   // 转换为 DeviceCollectionInfo (数值格式)
   LoadTotalWatt: 500.5,    // 用于能耗计算的核心字段
   LoadTotalVa: 600.0,
   LoadWatt1: 250.25,
   LoadVa1: 300.0
   ```

2. **电气参数字段**
   ```go
   // WinPower API 返回
   "realtime": {
       "inputVolt1": "220.5",      // 输入电压
       "outputVolt1": "219.8",     // 输出电压
       "outputCurrent1": "2.15",   // 输出电流
       "inputFreq": "50.0",        // 输入频率
       "outputFreq": "49.9"        // 输出频率
   }

   // 转换为 DeviceCollectionInfo
   InputVolt1: 220.5,
   OutputVolt1: 219.8,
   OutputCurrent1: 2.15,
   InputFreq: 50.0,
   OutputFreq: 49.9
   ```

3. **电池状态字段**
   ```go
   // WinPower API 返回
   "realtime": {
       "isCharging": "1",          // 充电状态 (1=是, 0=否)
       "batCapacity": "85.5",      // 电池容量
       "batRemainTime": "1800",    // 电池剩余时间(秒)
       "batteryStatus": "2"        // 电池状态码
   }

   // 转换为 DeviceCollectionInfo
   IsCharging: true,              // 字符串转布尔
   BatCapacity: 85.5,
   BatRemainTime: 1800,
   BatteryStatus: 2
   ```

4. **设备基本信息**
   ```go
   // WinPower API 返回
   "assetDevice": {
       "id": "device-001",
       "alias": "主UPS设备",
       "deviceType": 1,
       "model": "WinPower G2-1000VA"
   },
   "connected": true

   // 转换为 DeviceCollectionInfo
   DeviceID: "device-001",
   DeviceName: "主UPS设备",
   DeviceType: 1,
   DeviceModel: "WinPower G2-1000VA",
   Connected: true
   ```

5. **状态和故障信息**
   ```go
   // WinPower API 返回
   "realtime": {
       "mode": "1",                // UPS工作模式
       "status": "0",              // 设备状态
       "faultCode": "000",         // 故障代码
       "upsTemperature": "35.5"    // UPS温度
   }

   // 转换为 DeviceCollectionInfo
   Mode: 1,
   Status: 0,
   FaultCode: "000",
   UpsTemperature: 35.5
   ```

**数据转换实现示例：**

```go
// convertToCollectionResult 将WinPower数据转换为CollectionResult
func (c *CollectorService) convertToCollectionResult(
    winpowerData []winpower.ParsedDeviceData,
    energyResults map[string]float64,
    startTime time.Time,
    collectionErr error,
) *CollectionResult {
    result := &CollectionResult{
        Devices:        make(map[string]*DeviceCollectionInfo),
        CollectionTime: time.Now(),
        Duration:       time.Since(startTime),
    }

    if collectionErr != nil {
        result.Success = false
        result.ErrorMessage = collectionErr.Error()
        return result
    }

    result.Success = true
    result.DeviceCount = len(winpowerData)

    for _, device := range winpowerData {
        deviceInfo := &DeviceCollectionInfo{
            // 基本信息
            DeviceID:        device.DeviceID,
            DeviceName:      device.DeviceName,
            DeviceType:      device.DeviceType,
            DeviceModel:     device.DeviceModel,
            Connected:       device.Connected,
            LastUpdateTime:  device.Timestamp,

            // 输入电气参数
            InputVolt1:      device.RealtimeData.InputVolt1,
            InputFreq:       device.RealtimeData.InputFreq,

            // 输出电气参数
            OutputVolt1:     device.RealtimeData.OutputVolt1,
            OutputCurrent1:  device.RealtimeData.OutputCurrent1,
            OutputFreq:      device.RealtimeData.OutputFreq,
            OutputVoltageType: device.RealtimeData.OutputVoltageType,

            // 负载和功率参数
            LoadPercent:     device.RealtimeData.LoadPercent,
            LoadTotalWatt:   device.RealtimeData.LoadTotalWatt,
            LoadTotalVa:     device.RealtimeData.LoadTotalVa,
            LoadWatt1:       device.RealtimeData.LoadWatt1,
            LoadVa1:         device.RealtimeData.LoadVa1,

            // 电池参数
            IsCharging:      device.RealtimeData.IsCharging,
            BatVoltP:        device.RealtimeData.BatVoltP,
            BatCapacity:     device.RealtimeData.BatCapacity,
            BatRemainTime:   device.RealtimeData.BatRemainTime,
            BatteryStatus:   device.RealtimeData.BatteryStatus,

            // UPS状态参数
            UpsTemperature:  device.RealtimeData.UpsTemperature,
            Mode:            device.RealtimeData.Mode,
            Status:          device.RealtimeData.Status,
            TestStatus:      device.RealtimeData.TestStatus,
            FaultCode:       device.RealtimeData.FaultCode,

            // 其他参数
            InputTransformerType: device.RealtimeData.InputTransformerType,
        }

        // 添加电能计算结果
        if energyValue, calculated := energyResults[device.DeviceID]; calculated {
            deviceInfo.EnergyCalculated = true
            deviceInfo.EnergyValue = energyValue
        } else {
            deviceInfo.EnergyCalculated = false
            deviceInfo.ErrorMsg = "能量计算失败"
        }

        result.Devices[device.DeviceID] = deviceInfo
    }

    return result
}
```

## 接口设计

### Collector Interface

```go
// CollectorInterface 采集器接口
type CollectorInterface interface {
    // CollectDeviceData 采集设备数据并触发电能计算（统一入口）
    CollectDeviceData(ctx context.Context) (*CollectionResult, error)
}
```

### 依赖接口

```go
// WinPowerClient WinPower客户端接口（由winpower模块提供）
type WinPowerClient interface {
    CollectDeviceData(ctx context.Context) (*winpower.ParsedDeviceData, error)
    GetConnectionStatus() bool
    GetLastCollectionTime() time.Time
}

// EnergyInterface 电能计算接口（由energy模块提供）
type EnergyInterface interface {
    Calculate(deviceID string, power float64) (float64, error)
    Get(deviceID string) (float64, error)
}
```


## 错误处理策略

### 分层错误处理

1. **WinPower采集错误**: 记录错误并直接返回，由上层调度器决定是否重试
2. **电能计算错误**: 单个设备失败不影响其他设备，记录详细错误日志并继续处理
3. **其他错误**: 记录错误详情，直接返回，不进行自动重试

### 错误分类

Collector模块的错误类型会传递给Metrics模块进行统计：

```go
// ErrorType 错误类型（用于Metrics模块统计）
type ErrorType int

const (
    ErrorTypeWinPowerCollection ErrorType = iota  // WinPower采集错误
    ErrorTypeEnergyCalculation                   // 电能计算错误
)

// CollectionError 采集错误
type CollectionError struct {
    Type      ErrorType `json:"type"`
    DeviceID  string    `json:"device_id,omitempty"`
    Message   string    `json:"message"`
    Cause     error     `json:"cause,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}

func (e *CollectionError) Error() string {
    if e.DeviceID != "" {
        return fmt.Sprintf("[%s] %s (device: %s): %s", e.Type, e.Message, e.DeviceID, e.Cause)
    }
    return fmt.Sprintf("[%s] %s: %s", e.Type, e.Message, e.Cause)
}

// 错误类型会映射到Metrics模块的error_type标签，用于错误统计：
// - "winpower_collection" - WinPower采集错误
// - "energy_calculation" - 电能计算错误
```

## 与其他模块的集成

### 与Scheduler模块集成

```go
// 调度器调用Collector采集数据
func (s *Scheduler) runCollection(ctx context.Context) {
    logger := s.logger.With(zap.String("task", "collection"))

    if err := s.collector.CollectDeviceData(ctx); err != nil {
        logger.Error("Device data collection failed", zap.Error(err))
        // 调度器根据错误类型决定是否重试
        s.handleCollectionError(err)
        return
    }

    logger.Debug("Device data collection completed successfully")
}
```

### 与Energy模块集成

```go
// Collector作为唯一触发点调用Energy模块
func (c *CollectorService) processDevice(deviceID string, deviceData *winpower.DeviceData) error {
    // 直接使用WinPower返回的总负载有功功率
    power := deviceData.PowerInfo.LoadTotalWatt
    totalEnergy, err := c.energyService.Calculate(deviceID, power)
    if err != nil {
        return fmt.Errorf("energy calculation failed for device %s: %w", deviceID, err)
    }

    c.logger.Debug("Energy calculation completed",
        zap.String("device_id", deviceID),
        zap.Float64("power", power),
        zap.Float64("total_energy", totalEnergy))

    return nil
}
```



## 测试设计

### 单元测试

```go
// collector_test.go
package collector

import (
    "context"
    "testing"
    "time"
    "go.uber.org/zap/zaptest"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockWinPowerClient Mock WinPower客户端
type MockWinPowerClient struct {
    mock.Mock
}

func (m *MockWinPowerClient) CollectDeviceData(ctx context.Context) (*winpower.ParsedDeviceData, error) {
    args := m.Called(ctx)
    return args.Get(0).(*winpower.ParsedDeviceData), args.Error(1)
}

// MockEnergyService Mock电能服务
type MockEnergyService struct {
    mock.Mock
}

func (m *MockEnergyService) Calculate(deviceID string, power float64) (float64, error) {
    args := m.Called(deviceID, power)
    return args.Get(0).(float64), args.Error(1)
}

func TestCollectorService_CollectDeviceData(t *testing.T) {
    logger := zaptest.NewLogger(t)

    // 创建Mock依赖
    mockWinPower := new(MockWinPowerClient)
    mockEnergy := new(MockEnergyService)

    // 设置Mock数据
    mockDeviceData := &winpower.ParsedDeviceData{
        Devices: map[string]*winpower.DeviceData{
            "device-001": {
                DeviceID: "device-001",
                PowerInfo: winpower.PowerInfo{
                    LoadTotalWatt: 500.0,
                },
            },
        },
    }

    // 设置Mock预期
    mockWinPower.On("CollectDeviceData", mock.Anything).Return(mockDeviceData, nil)
    mockEnergy.On("Calculate", "device-001", 500.0).Return(1500.0, nil)

    // 创建Collector服务
    collector := NewCollectorService(mockWinPower, mockEnergy, logger)

    // 执行采集（统一入口，可用于调度器或Metrics触发）
    result, err := collector.CollectDeviceData(context.Background())
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.True(t, result.Success)
    assert.Equal(t, 1, result.DeviceCount)

    // 验证Mock调用
    mockWinPower.AssertExpectations(t)
    mockEnergy.AssertExpectations(t)
}

func TestCollectorService_EnergyCalculationFailure(t *testing.T) {
    logger := zaptest.NewLogger(t)

    mockWinPower := new(MockWinPowerClient)
    mockEnergy := new(MockEnergyService)

    mockDeviceData := &winpower.ParsedDeviceData{
        Devices: map[string]*winpower.DeviceData{
            "device-001": {
                DeviceID: "device-001",
                PowerInfo: winpower.PowerInfo{
                    LoadTotalWatt: 500.0,
                },
            },
            "device-002": {
                DeviceID: "device-002",
                PowerInfo: winpower.PowerInfo{
                    LoadTotalWatt: 300.0,
                },
            },
        },
    }

    // 设置Mock：第一个设备电能计算失败，第二个成功
    mockWinPower.On("CollectDeviceData", mock.Anything).Return(mockDeviceData, nil)
    mockEnergy.On("Calculate", "device-001", 500.0).Return(0.0, assert.AnError)
    mockEnergy.On("Calculate", "device-002", 300.0).Return(800.0, nil)

    collector := NewCollectorService(mockWinPower, mockEnergy, logger)

    result, err := collector.CollectDeviceData(context.Background())

    // 整体采集应该成功，尽管有部分设备失败
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.True(t, result.Success)
    assert.Equal(t, 2, result.DeviceCount)

    // 验证Mock调用
    mockWinPower.AssertExpectations(t)
    mockEnergy.AssertExpectations(t)
}
```

### 集成测试

```go
// collector_integration_test.go
// +build integration

package collector

import (
    "context"
    "testing"
    "time"
    "go.uber.org/zap/zaptest"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCollectorService_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    logger := zaptest.NewLogger(t)

    // 创建真实的依赖服务
    // ...（集成测试需要真实的WinPower、Energy、Metrics服务）

    collector := NewCollectorService(realWinPower, realEnergy, logger)

    // 执行采集
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    result, err := collector.CollectDeviceData(ctx)
    require.NoError(t, err)
    require.NotNil(t, result)
    require.True(t, result.Success)

    // 验证采集成功完成
}
```

## 监控指标

Collector模块的监控指标统一由Metrics模块管理，包括：

- **错误统计指标**：`winpower_exporter_scrape_errors_total` - 采集错误总数，标签包含`error_type`
- **采集性能指标**：`winpower_exporter_collection_duration_seconds` - 采集+计算整体耗时
- **设备状态指标**：反映设备的连接状态和基本状态信息
- **能量相关指标**：由Energy模块在完成计算后直接更新到Metrics模块

**注意**：Collector模块本身不维护任何统计信息，所有监控指标由Metrics模块统一创建和管理。错误计数等统计功能已集成到Metrics模块的设计中。

## 最佳实践

### 1. 错误处理

- 优雅降级：单设备失败不影响整体采集
- 详细日志：记录足够的信息用于问题排查
- 指标跟踪：通过指标反映系统健康状况


### 3. 指标管理

- **统一管理**: 所有监控指标由Metrics模块统一创建和管理
- **错误统计**: 采集错误通过`winpower_exporter_scrape_errors_total`计数，按`error_type`分类
- **性能监控**: 采集耗时通过`winpower_exporter_collection_duration_seconds`记录
- **设备状态**: 设备连接和状态信息通过Metrics模块的设备指标暴露

### 4. 可维护性

- 接口设计：通过接口解耦依赖
- 单元测试：保证核心逻辑的正确性