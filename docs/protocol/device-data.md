# WinPower G2 设备数据查询协议

## 概述

WinPower G2 系统提供设备数据的实时查询接口，支持获取设备的详细信息、实时数据、配置参数、设置信息和告警状态。该接口采用分页查询方式，支持按区域和设备类型过滤。

## API 接口

### 设备数据详情查询

#### 请求信息

- **URL**: `/api/v1/deviceData/detail/list`
- **方法**: `GET`
- **Content-Type**: `application/json`
- **认证**: 需要有效的 JWT Token

#### 请求头

```
GET /api/v1/deviceData/detail/list?{parameters} HTTP/1.1
Host: {host}:{port}
Connection: keep-alive
Authorization: Bearer {jwt_token}
User-Agent: Mozilla/5.0 (compatible; WinPower-Exporter/1.0)
Accept: application/json, text/plain, */*
Content-language: zh-CN
```

#### 查询参数

| 参数名         | 类型    | 必填 | 默认值                                 | 说明                   |
| -------------- | ------- | ---- | -------------------------------------- | ---------------------- |
| current        | integer | 否   | 1                                      | 当前页码               |
| pageSize       | integer | 否   | 100                                    | 每页记录数             |
| areaId         | string  | 否   | "00000000-0000-0000-0000-000000000000" | 区域ID                 |
| includeSubArea | boolean | 否   | true                                   | 是否包含子区域         |
| pageNum        | integer | 否   | 1                                      | 页码 (与 current 相同) |
| deviceType     | integer | 否   | -                                      | 设备类型过滤           |

**设备类型说明:**
- 1: UPS (不间断电源)
- 其他数值根据具体系统定义

#### 请求示例

```
GET /api/v1/deviceData/detail/list?current=1&pageSize=100&areaId=00000000-0000-0000-0000-000000000000&includeSubArea=true&pageNum=1&deviceType=1 HTTP/1.1
```

## 响应数据结构

### 成功响应 (HTTP 200)

```json
{
    "total": 1,
    "pageSize": 20,
    "currentPage": 1,
    "data": [
        {
            "assetDevice": { /* 设备基本信息 */ },
            "realtime": { /* 实时数据 */ },
            "config": { /* 配置参数 */ },
            "setting": { /* 设置信息 */ },
            "activeAlarms": [ /* 活动告警 */ ],
            "controlSupported": { /* 支持的控制功能 */ },
            "connected": true
        }
    ],
    "code": "000000",
    "msg": "OK"
}
```

### 响应字段说明

| 字段名      | 类型    | 说明                          |
| ----------- | ------- | ----------------------------- |
| total       | integer | 总记录数                      |
| pageSize    | integer | 每页记录数                    |
| currentPage | integer | 当前页码                      |
| data        | array   | 设备数据列表                  |
| code        | string  | 响应状态码，"000000" 表示成功 |
| msg         | string  | 响应消息                      |

## 设备数据详细结构

### 1. assetDevice (设备基本信息)

```json
{
    "id": "e156e6cb-41cb-4b35-b0dd-869929186a5c",
    "deviceType": 1,
    "model": "ON-LINE",
    "alias": "C3K",
    "protocolId": 13,
    "connectType": 1,
    "comPort": "COM3",
    "baudRate": 2400,
    "areaId": "00000000-0000-0000-0000-000000000000",
    "isActive": true,
    "firmwareVersion": "03.09",
    "createTime": "2025-10-13T08:37:57.048192",
    "warrantyStatus": 0
}
```

**字段说明:**

| 字段名          | 类型    | 说明                     |
| --------------- | ------- | ------------------------ |
| id              | string  | 设备唯一标识符           |
| deviceType      | integer | 设备类型 (1=UPS)         |
| model           | string  | 设备型号                 |
| alias           | string  | 设备别名                 |
| protocolId      | integer | 通信协议ID               |
| connectType     | integer | 连接类型                 |
| comPort         | string  | 串口号                   |
| baudRate        | integer | 波特率                   |
| areaId          | string  | 所属区域ID               |
| isActive        | boolean | 设备是否激活             |
| firmwareVersion | string  | 固件版本                 |
| createTime      | string  | 创建时间 (ISO 8601 格式) |
| warrantyStatus  | integer | 保修状态                 |

### 2. realtime (实时数据)

```json
{
    "inputVolt1": "236.8",
    "outputVoltageType": "0",
    "loadPercent": "6",
    "isCharging": "1",
    "inputTransformerType": "0",
    "batVoltP": "81.4",
    "outputCurrent1": "1.1",
    "upsTemperature": "27.0",
    "mode": "3",
    "batRemainTime": "6723",
    "loadVa1": "198",
    "faultCode": "",
    "outputVolt1": "220.1",
    "outputFreq": "49.9",
    "inputFreq": "49.9",
    "loadTotalVa": "198",
    "batteryStatus": "2",
    "testStatus": "1",
    "loadTotalWatt": "195",
    "loadWatt1": "195",
    "batCapacity": "90",
    "status": "1"
}
```

**实时数据字段说明:**

| 字段名               | 类型   | 单位 | 说明                      |
| -------------------- | ------ | ---- | ------------------------- |
| inputVolt1           | string | V    | 输入电压 (相1)            |
| outputVoltageType    | string | -    | 输出电压类型              |
| loadPercent          | string | %    | 负载百分比                |
| isCharging           | string | -    | 是否正在充电 (1=是, 0=否) |
| inputTransformerType | string | -    | 输入变压器类型            |
| batVoltP             | string | V    | 电池电压百分比            |
| outputCurrent1       | string | A    | 输出电流 (相1)            |
| upsTemperature       | string | °C   | UPS 温度                  |
| mode                 | string | -    | UPS 工作模式              |
| batRemainTime        | string | 秒   | 电池剩余时间              |
| loadVa1              | string | VA   | 负载视在功率 (相1)        |
| faultCode            | string | -    | 故障代码                  |
| outputVolt1          | string | V    | 输出电压 (相1)            |
| outputFreq           | string | Hz   | 输出频率                  |
| inputFreq            | string | Hz   | 输入频率                  |
| loadTotalVa          | string | VA   | 总负载视在功率            |
| batteryStatus        | string | -    | 电池状态                  |
| testStatus           | string | -    | 测试状态                  |
| loadTotalWatt        | string | W    | 总负载有功功率            |
| loadWatt1            | string | W    | 负载有功功率 (相1)        |
| batCapacity          | string | %    | 电池容量                  |
| status               | string | -    | 设备状态                  |

> 能耗计算输入功率：统一使用 `loadTotalWatt`（总负载有功功率，单位 W）。采集模块将此字段作为能耗累计计算的 `power` 传入。

### 3. config (配置参数)

```json
{
    "inputPhaseNumber": "1",
    "ratingBatUnitNumber": "6",
    "upsModuleNumber": "1",
    "ratingBatVoltPerUnit": "12",
    "ratingVa": "3000",
    "inputSourceNumber": "1",
    "ratingVolt": "220",
    "outputPhaseNumber": "1",
    "ratingFreq": "50",
    "ratingBatVolt": "72"
}
```

**配置参数说明:**

| 字段名               | 类型   | 单位 | 说明           |
| -------------------- | ------ | ---- | -------------- |
| inputPhaseNumber     | string | -    | 输入相数       |
| ratingBatUnitNumber  | string | -    | 额定电池单元数 |
| upsModuleNumber      | string | -    | UPS 模块数     |
| ratingBatVoltPerUnit | string | V    | 每单元额定电压 |
| ratingVa             | string | VA   | 额定视在功率   |
| inputSourceNumber    | string | -    | 输入电源数     |
| ratingVolt           | string | V    | 额定电压       |
| outputPhaseNumber    | string | -    | 输出相数       |
| ratingFreq           | string | Hz   | 额定频率       |
| ratingBatVolt        | string | V    | 额定电池电压   |

### 4. setting (设置信息)

```json
{
    "enableAudible": "1",
    "enableBypassWhenTurnOff": "0",
    "batteryModuleNumber": "1",
    "autoTestPeriod": "90",
    "enableAutoRestart": "1"
}
```

**设置信息说明:**

| 字段名                  | 类型   | 说明               |
| ----------------------- | ------ | ------------------ |
| enableAudible           | string | 是否启用声音告警   |
| enableBypassWhenTurnOff | string | 关机时是否启用旁路 |
| batteryModuleNumber     | string | 电池模块数         |
| autoTestPeriod          | string | 自动测试周期 (天)  |
| enableAutoRestart       | string | 是否启用自动重启   |

### 5. activeAlarms (活动告警)

```json
[]
```

告警数组，当存在活动告警时包含告警对象。

### 6. controlSupported (支持的控制功能)

```json
{
    "supportQuickTest": true,
    "supportDeepTest": true,
    "supportTestDuration": true,
    "supportShutdown": true
}
```

**控制功能说明:**

| 字段名              | 类型    | 说明                     |
| ------------------- | ------- | ------------------------ |
| supportQuickTest    | boolean | 是否支持快速测试         |
| supportDeepTest     | boolean | 是否支持深度测试         |
| supportTestDuration | boolean | 是否支持测试持续时间设置 |
| supportShutdown     | boolean | 是否支持关机控制         |

### 7. connected (连接状态)

```json
true
```

表示设备是否正常连接到系统。

## 数据类型说明

### 设备状态码 (status)

| 值     | 说明             |
| ------ | ---------------- |
| "1"    | 正常运行         |
| 其他值 | 根据具体系统定义 |

### 电池状态码 (batteryStatus)

| 值     | 说明             |
| ------ | ---------------- |
| "2"    | 电池正常         |
| 其他值 | 根据具体系统定义 |

### 工作模式 (mode)

| 值     | 说明             |
| ------ | ---------------- |
| "3"    | 在线模式         |
| 其他值 | 根据具体系统定义 |

## 错误处理

### 常见错误响应

#### 认证失败 (HTTP 401)

```json
{
    "code": "401",
    "msg": "Unauthorized",
    "data": null
}
```

#### 参数错误 (HTTP 400)

```json
{
    "code": "400",
    "msg": "Bad Request",
    "data": null
}
```

#### 服务器错误 (HTTP 500)

```json
{
    "code": "500",
    "msg": "Internal Server Error",
    "data": null
}
```

## 实现建议

### 1. 数据采集策略

- **采集频率**: 建议 5-30 秒采集一次，根据精度需求和系统负载调整
- **分页处理**: 使用适当的分页大小，避免单次请求过多数据
- **缓存机制**: 实现数据缓存，减少 API 调用频率

### 2. 错误处理

- **错误处理**: 完善的错误处理机制
- **连接超时**: 设置合理的连接和读取超时时间
- **Token 刷新**: 监控 Token 过期，及时刷新

### 3. 性能优化

- **连接池**: 使用 HTTP 连接池
- **并发控制**: 限制并发请求数量
- **数据压缩**: 启用 gzip 压缩减少传输量

## 示例代码

### Go 示例

```go
type DeviceDataResponse struct {
    Total       int                    `json:"total"`
    PageSize    int                    `json:"pageSize"`
    CurrentPage int                    `json:"currentPage"`
    Data        []DeviceInfo           `json:"data"`
    Code        string                 `json:"code"`
    Msg         string                 `json:"msg"`
}

type DeviceInfo struct {
    AssetDevice     AssetDevice     `json:"assetDevice"`
    Realtime        RealtimeData    `json:"realtime"`
    Config          ConfigData      `json:"config"`
    Setting         SettingData     `json:"setting"`
    ActiveAlarms    []interface{}   `json:"activeAlarms"`
    ControlSupported ControlSupport  `json:"controlSupported"`
    Connected       bool            `json:"connected"`
}

// 获取设备数据
func GetDeviceData(client *http.Client, host, token string) (*DeviceDataResponse, error) {
    url := fmt.Sprintf("http://%s/api/v1/deviceData/detail/list?current=1&pageSize=100&deviceType=1", host)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Accept", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result DeviceDataResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}
```

### cURL 示例

```bash
# 获取设备数据
curl -X GET "http://192.168.1.100:8081/api/v1/deviceData/detail/list?current=1&pageSize=100&deviceType=1" \
  -H "Authorization: Bearer {jwt_token}" \
  -H "Accept: application/json" \
  -H "Content-language: zh-CN"
```