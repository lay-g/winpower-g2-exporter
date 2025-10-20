# Storage Module Error Handling Guide

## 概述

本文档详细说明了存储模块的错误处理机制，包括错误类型、处理策略和最佳实践。

## 错误类型

### 1. StorageError 结构体

```go
type StorageError struct {
    Operation string // 失败的操作类型 (例如: "write", "read")
    DeviceID  string // 设备标识符
    Path      string // 相关文件路径
    Cause     error  // 底层错误原因
}
```

`StorageError` 提供了丰富的上下文信息，便于错误诊断和调试。

### 2. 预定义错误类型

模块提供了以下预定义错误类型：

- `ErrFileNotFound`: 文件不存在
- `ErrInvalidFormat`: 文件格式无效或损坏
- `ErrPermissionDenied`: 权限不足
- `ErrInvalidData`: 数据无效
- `ErrInvalidTimestamp`: 时间戳无效
- `ErrInvalidEnergyValue`: 能量值无效
- `ErrDiskFull`: 磁盘空间不足

## 错误处理模式

### 1. 创建存储错误

使用 `NewStorageError` 函数创建带上下文的错误：

```go
err := NewStorageError("write", deviceID, filePath, underlyingErr)
```

### 2. 错误检查和处理

#### 基本错误检查
```go
data, err := manager.Read(deviceID)
if err != nil {
    // 处理错误
    return fmt.Errorf("failed to read device data: %w", err)
}
```

#### 特定错误类型检查
```go
if errors.Is(err, ErrFileNotFound) {
    // 处理文件不存在的情况
    return initializeDeviceData(deviceID)
}
```

#### StorageError 类型检查
```go
var storageErr *StorageError
if errors.As(err, &storageErr) {
    // 访问存储错误的详细信息
    log.Printf("Storage operation '%s' failed for device '%s': %v",
        storageErr.Operation, storageErr.DeviceID, storageErr.Cause)
}
```

#### 文件不存在错误检查
```go
if IsFileNotFoundError(err) {
    // 处理文件不存在的情况
    return nil, nil // 返回默认值
}
```

## 错误恢复策略

### 1. 文件不存在错误

```go
data, err := reader.ReadAndParse(deviceID)
if err != nil {
    if IsFileNotFoundError(err) {
        // 返回初始化的数据
        return NewPowerData(0), nil
    }
    return nil, err
}
```

### 2. 权限错误

```go
err := writer.WriteDeviceFile(deviceID, data)
if err != nil {
    var storageErr *StorageError
    if errors.As(err, &storageErr) && errors.Is(storageErr.Cause, os.ErrPermission) {
        // 尝试修复权限问题
        if fixPermissions(storageErr.Path) {
            return writer.WriteDeviceFile(deviceID, data)
        }
    }
    return err
}
```

### 3. 磁盘空间不足

```go
err := writer.WriteDeviceFile(deviceID, data)
if err != nil {
    if errors.Is(err, ErrDiskFull) {
        // 清理旧文件
        if cleanupOldFiles() {
            return writer.WriteDeviceFile(deviceID, data)
        }
    }
    return err
}
```

## 日志记录

### 1. 错误日志格式

```go
func (sm *FileStorageManager) Write(deviceID string, data *PowerData) error {
    err := sm.writer.WriteDeviceFile(deviceID, data)
    if err != nil {
        // 记录详细的错误信息
        log.Error("Failed to write device data",
            "device", deviceID,
            "operation", "write",
            "error", err.Error(),
            "cause", err.Cause())
        return err
    }

    log.Info("Successfully wrote device data",
        "device", deviceID,
        "timestamp", data.Timestamp,
        "energy", data.EnergyWH)
    return nil
}
```

### 2. 结构化日志

使用结构化日志记录错误信息，便于后续分析和监控：

```go
log.Error("Storage operation failed",
    "operation", "write",
    "device_id", deviceID,
    "file_path", filePath,
    "error_type", getErrorType(err),
    "error_message", err.Error())
```

## 测试中的错误处理

### 1. 错误模拟

```go
func TestWriteError(t *testing.T) {
    // 创建只读目录模拟权限错误
    readOnlyDir := t.TempDir()
    os.Chmod(readOnlyDir, 0444)

    config := NewConfig()
    config.DataDir = readOnlyDir

    manager, err := NewFileStorageManager(config)
    require.NoError(t, err)

    data := NewPowerData(100.0)
    err = manager.Write("test-device", data)

    // 验证错误类型
    assert.Error(t, err)
    var storageErr *StorageError
    assert.True(t, errors.As(err, &storageErr))
    assert.Equal(t, "write", storageErr.Operation)
}
```

### 2. 错误断言

```go
func TestReadNotFound(t *testing.T) {
    manager := NewFileStorageManager(NewConfig())

    data, err := manager.Read("non-existent-device")

    // 验证返回了初始化数据而不是错误
    assert.NoError(t, err)
    assert.NotNil(t, data)
    assert.True(t, data.IsZero())
}
```

## 最佳实践

### 1. 错误传播

始终使用 `%w` 动词包装错误以保留错误链：

```go
// 好的做法
return fmt.Errorf("failed to write device data: %w", err)

// 避免的做法
return fmt.Errorf("failed to write device data: %v", err)
```

### 2. 错误上下文

在错误中包含足够的上下文信息：

```go
// 好的做法 - 包含设备ID和操作类型
err := NewStorageError("write", deviceID, filePath, underlyingErr)

// 避免的做法 - 缺少上下文
return underlyingErr
```

### 3. 优雅降级

对于非关键错误，提供优雅降级机制：

```go
func GetDeviceData(deviceID string) (*PowerData, error) {
    data, err := storage.Read(deviceID)
    if err != nil {
        if IsFileNotFoundError(err) {
            // 新设备，返回默认数据
            log.Warn("Device file not found, using default data", "device", deviceID)
            return NewPowerData(0), nil
        }

        // 其他错误需要处理
        return nil, fmt.Errorf("failed to read device data: %w", err)
    }
    return data, nil
}
```

### 4. 错误分类

根据错误的严重程度和类型进行分类处理：

```go
func HandleStorageError(err error, deviceID string) error {
    if err == nil {
        return nil
    }

    var storageErr *StorageError
    if !errors.As(err, &storageErr) {
        return err // 非存储错误，直接返回
    }

    switch {
    case errors.Is(storageErr.Cause, os.ErrPermission):
        log.Error("Permission denied for storage operation",
            "device", deviceID, "path", storageErr.Path)
        return tryFixPermissions(storageErr.Path)

    case errors.Is(storageErr.Cause, os.ErrNoSpace):
        log.Error("Disk full during storage operation",
            "device", deviceID)
        return tryCleanupStorage()

    case IsFileNotFoundError(storageErr):
        log.Info("Device file not found, will be created",
            "device", deviceID)
        return nil // 不是错误，新设备的正常情况

    default:
        log.Error("Unknown storage error",
            "device", deviceID, "error", storageErr.Error())
        return err
    }
}
```

## 监控和告警

### 1. 错误指标

```go
var (
    storageErrorsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "storage_errors_total",
            Help: "Total number of storage errors",
        },
        []string{"operation", "error_type"},
    )

    storageOperationsDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "storage_operation_duration_seconds",
            Help: "Duration of storage operations",
        },
        []string{"operation"},
    )
)
```

### 2. 告警规则

配置告警规则以监控存储错误：

```yaml
groups:
- name: storage
  rules:
  - alert: StorageHighErrorRate
    expr: rate(storage_errors_total[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "Storage module high error rate"

  - alert: StoragePermissionErrors
    expr: increase(storage_errors_total{error_type="permission"}[5m]) > 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Storage permission errors detected"
```

## 调试技巧

### 1. 错误详情打印

```go
func debugStorageError(err error) {
    var storageErr *StorageError
    if errors.As(err, &storageErr) {
        fmt.Printf("Storage Error Details:\n")
        fmt.Printf("  Operation: %s\n", storageErr.Operation)
        fmt.Printf("  Device ID: %s\n", storageErr.DeviceID)
        fmt.Printf("  File Path: %s\n", storageErr.Path)
        fmt.Printf("  Cause: %v\n", storageErr.Cause)
        fmt.Printf("  Error Chain: %+v\n", err)
    }
}
```

### 2. 文件状态检查

```go
func debugFileState(filePath string) {
    if stat, err := os.Stat(filePath); err == nil {
        fmt.Printf("File State:\n")
        fmt.Printf("  Path: %s\n", filePath)
        fmt.Printf("  Size: %d bytes\n", stat.Size())
        fmt.Printf("  Mode: %v\n", stat.Mode())
        fmt.Printf("  ModTime: %v\n", stat.ModTime())
        fmt.Printf("  IsDir: %v\n", stat.IsDir())
    } else {
        fmt.Printf("File not accessible: %v\n", err)
    }
}
```

## 总结

良好的错误处理是存储模块可靠性的关键。通过使用结构化错误类型、提供丰富的上下文信息、实现适当的恢复策略，以及进行充分的测试，可以确保存储模块在各种异常情况下都能正确响应。

关键要点：
1. 使用 `StorageError` 提供丰富的错误上下文
2. 实现适当的错误恢复策略
3. 记录详细的错误日志以便调试
4. 在测试中覆盖各种错误场景
5. 监控错误率并及时告警
6. 提供优雅降级机制