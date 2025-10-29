# 文件持久化模块设计

## 1. 概述

文件持久化模块 (storage) 负责 WinPower G2 Exporter 中电能计算结果的文件存储管理。该模块提供简化的文件读写功能，每个设备使用独立的文件存储电能数据。

### 1.1 设计目标

- **简单性**: 简化的设计，易于理解和维护
- **可靠性**: 基本的文件读写功能，确保数据持久化
- **多设备支持**: 支持多设备分别存储数据
- **原子性**: 文件写入操作具备原子性，避免数据损坏

### 1.2 功能范围

- ✅ **电能数据持久化**: 将累计电能值保存到设备文件
- ✅ **数据读取**: 启动时恢复历史累计电能值
- ✅ **多设备支持**: 为每个设备创建独立的数据文件
- ❌ **数据备份**: 不提供自动备份功能
- ❌ **数据清理**: 不提供自动清理功能
- ❌ **并发访问**: 不支持并发读写，由上层统一控制
- ❌ **重试机制**: 不提供写入重试功能

## 2. 架构设计

### 2.1 模块架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    Storage Module                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐         ┌─────────────────┐            │
│  │  Storage Manager│         │  File Writer    │            │
│  │  (Interface)    │         │  (Write)        │            │
│  └────────┬────────┘         └────────┬────────┘            │
│           │                            │                     │
│           │ Implements                 │ Uses                │
│           ▼                            ▼                     │
│  ┌─────────────────┐         ┌─────────────────┐            │
│  │ File Storage    │         │  File Reader    │            │
│  │ (Core Logic)    │         │ (Read Data)     │            │
│  └─────────────────┘         └─────────────────┘            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ File I/O
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     File System                             │
│  ┌──────────────────┐    ┌──────────────────┐              │
│  │     a1.txt       │    │     a2.txt       │              │
│  │  (Device A1)     │    │  (Device A2)     │              │
│  └──────────────────┘    └──────────────────┘              │
│  ┌──────────────────┐    ┌──────────────────┐              │
│  │     b1.txt       │    │     b2.txt       │              │
│  │  (Device B1)     │    │  (Device B2)     │              │
│  └──────────────────┘    └──────────────────┘              │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 数据流图

```
┌─────────────────┐      Write      ┌─────────────────┐
│ Power Calculator│ ───────────────► │ Storage Manager │
│  (Data Source)  │                  │  (Write Method) │
└─────────────────┘                  └────────┬────────┘
                                             │ Device ID
                                             ▼
                                    ┌─────────────────┐
                                    │  File Writer    │
                                    │  (Write Device) │
                                    └────────┬────────┘
                                             │ Sync Write (固定为同步)
                                             ▼
                                    ┌─────────────────┐
                                    │  File System    │
                                    │ (device_id.txt) │
                                    └─────────────────┘

┌─────────────────┐      Read       ┌─────────────────┐
│   Application   │ ◄────────────── │ Storage Manager │
│  (Startup)      │                  │  (Read Method)  │
└─────────────────┘                  └────────┬────────┘
                                             │ Device ID
                                             ▼
                                    ┌─────────────────┐
                                    │  File Reader    │
                                    │ (Read Device)   │
                                    └────────┬────────┘
                                             │ Return Data
                                             ▼
                                    ┌─────────────────┐
                                    │ Power Calculator │
                                    │ (Restore State)  │
                                    └─────────────────┘
```

### 2.3 文件格式设计

#### 2.3.1 设备文件格式 (device_id.txt)

```
# 行1: 毫秒时间戳 (Unix timestamp in milliseconds)
1694678400000
# 行2: 累计电能值 (Accumulated energy in Wh)
15000.50
```

#### 2.3.2 文件结构说明

| 行号 | 字段名 | 类型 | 说明 | 示例 |
|------|--------|------|------|------|
| 1 | timestamp | int64 | 毫秒时间戳，表示最后更新时间 | 1694678400000 |
| 2 | energy_wh | float64 | 累计电能值（可为负，表示净能量），单位瓦时 | 15000.50 |

#### 2.3.3 设备文件命名规则

- **文件命名**: `{device_id}.txt` - 使用设备ID作为文件名
- **存储目录**: 配置中指定的数据文件目录
- **文件示例**:
  - `a1.txt` - 设备ID为a1的数据文件
  - `b2.txt` - 设备ID为b2的数据文件

## 3. 接口设计

### 3.1 核心接口定义

```go
// StorageManager 存储管理器接口
type StorageManager interface {
    // Write 写入设备电能数据
    Write(deviceID string, data *PowerData) error

    // Read 读取设备电能数据
    Read(deviceID string) (*PowerData, error)
}

// PowerData 电能数据结构
type PowerData struct {
    Timestamp int64   `json:"timestamp"` // 毫秒时间戳
    EnergyWH  float64 `json:"energy_wh"` // 累计电能(Wh)
}

// FileWriter 文件写入器接口
type FileWriter interface {
    Write(deviceID string, data *PowerData) error
}

// FileReader 文件读取器接口
type FileReader interface {
    Read(deviceID string) (*PowerData, error)
}
```

### 3.2 配置结构

Storage模块定义自己的配置结构体，config模块会主动引用并集成这些配置：

```go
// internal/storage/config.go
package storage

import "os"

type Config struct {
    DataDir         string      `yaml:"data_dir" validate:"required"`
    FilePermissions os.FileMode `yaml:"file_permissions"`
}

func DefaultConfig() *Config {
    return &Config{
        DataDir:         "./data",
        FilePermissions: 0644,
    }
}

// Validate 实现ConfigValidator接口用于配置验证
func (c *Config) Validate() error {
    // 验证数据目录
    if c.DataDir == "" {
        return fmt.Errorf("storage data_dir cannot be empty")
    }

    // 验证数据目录路径是否有效
    // 检查路径中是否包含非法字符
    if strings.ContainsAny(c.DataDir, "\x00") {
        return fmt.Errorf("storage data_dir contains invalid characters")
    }

    // 验证文件权限
    if c.FilePermissions < 0 || c.FilePermissions > 0777 {
        return fmt.Errorf("storage file_permissions must be between 0000 and 0777, got %o", c.FilePermissions)
    }

    // 验证目录是否存在或可以创建
    // 尝试获取目录信息，如果不存在则检查是否可以创建
    if _, err := os.Stat(c.DataDir); os.IsNotExist(err) {
        // 目录不存在，尝试创建以验证权限
        if err := os.MkdirAll(c.DataDir, 0755); err != nil {
            return fmt.Errorf("cannot create storage data_dir '%s': %w", c.DataDir, err)
        }
        // 创建成功后删除，因为这里只是验证
        os.Remove(c.DataDir)
    } else if err != nil {
        return fmt.Errorf("cannot access storage data_dir '%s': %w", c.DataDir, err)
    }

    return nil
}

// Storage模块通过构造函数接收配置参数
func NewFileStorageManager(config *Config, logger Logger) StorageManager
```

**配置结构位置**：
- Config定义在`internal/storage/config.go`中
- 包含DataDir、FilePermissions等字段
- 提供文件存储相关的配置参数

**配置参数说明**：
- `config`: 使用storage模块的Config结构，包含文件存储的所有配置参数（DataDir、FilePermissions）
- `logger`: 日志记录器实例

**配置集成**：
- Config模块会主动引用并集成storage模块的配置结构
- Storage模块的配置由config模块统一管理和加载

## 4. 详细实现

### 4.1 存储管理器实现

```go
// FileStorageManager 文件存储管理器
type FileStorageManager struct {
    config *Config    // 存储配置参数
    logger logger.Logger         // 日志记录器
}

// NewFileStorageManager 创建文件存储管理器
func NewFileStorageManager(config *Config, logger logger.Logger) *FileStorageManager {
    // 验证配置参数，如果为空则使用默认配置
    // 确保数据目录存在
    // 初始化存储管理器结构体
    // 返回配置完成的存储管理器实例
}

// Write 写入设备电能数据
func (fsm *FileStorageManager) Write(deviceID string, data *PowerData) error {
    // 验证输入数据的有效性（检查时间戳、电能值等参数的合法性）
    // 构造设备文件路径
    // 格式化数据内容为字符串（时间戳、电能值各占一行）
    // 同步写入文件
    // 记录操作日志，包括成功信息和错误详情
}

// Read 读取设备电能数据
func (fsm *FileStorageManager) Read(deviceID string) (*PowerData, error) {
    // 构造设备文件路径
    // 尝试读取文件内容
    // 如果文件不存在，返回初始化数据：
    //   &PowerData{Timestamp: time.Now().UnixMilli(), EnergyWH: 0}
    // 如果文件存在，解析数据内容为PowerData结构
    // 验证解析后的数据完整性和有效性
    // 返回最终的有效数据或详细的错误信息
}

// validateData 验证数据有效性（内部方法）
func (fsm *FileStorageManager) validateData(data *PowerData) error {
    // 检查数据指针是否为空，防止空指针异常
    // 验证时间戳是否为正数（Unix时间戳必须大于0）
    // 验证电能值格式（允许负值：表示净能量；需为有限数值）
    // 根据验证结果返回相应的错误信息或nil表示验证通过
}

// getDeviceFilePath 获取设备文件路径（内部方法）
func (fsm *FileStorageManager) getDeviceFilePath(deviceID string) string {
    // 构造文件路径：配置的数据目录 + 设备ID + .txt后缀
    // 返回完整的文件路径
}
```

### 4.2 文件写入器实现

```go
// FileWriter 文件写入器
type FileWriter struct {
    config *Config    // 存储配置参数
    logger logger.Logger         // 日志记录器（与StorageManager共享）
}

// NewFileWriter 创建文件写入器
func NewFileWriter(config *Config, logger logger.Logger) *FileWriter {
    // 初始化文件写入器结构体
    // 设置存储配置参数引用
    // 获取日志记录器实例
    // 返回创建完成的文件写入器实例
}

// Write 写入设备文件
func (fw *FileWriter) Write(deviceID string, data *PowerData) error {
    // 构造完整的文件路径
    // 创建文件，使用配置中指定的文件权限
    // 格式化数据内容为字符串（时间戳、电能值各占一行）
    // 将内容同步写入文件
    // 确保文件句柄正确关闭
    // 返回操作结果或错误信息
}
```

### 4.3 文件读取器实现

```go
// FileReader 文件读取器
type FileReader struct {
    config *Config    // 存储配置参数
    logger logger.Logger         // 日志记录器（与StorageManager共享）
}

// NewFileReader 创建文件读取器
func NewFileReader(config *Config, logger logger.Logger) *FileReader {
    // 初始化文件读取器结构体
    // 设置存储配置参数引用
    // 获取日志记录器实例
    // 返回创建完成的文件读取器实例
}

// Read 读取设备文件数据
func (fr *FileReader) Read(deviceID string) (*PowerData, error) {
    // 构造完整的文件路径
    // 使用os.ReadFile读取文件的完整内容到内存
    // 如果文件不存在，返回初始化数据：
    //   &PowerData{Timestamp: time.Now().UnixMilli(), EnergyWH: 0}
    // 如果文件存在，将字节内容转换为字符串并按换行符分割
    // 验证文件格式包含时间戳和电能值两行数据
    // 解析第一行内容为64位整数时间戳，检查格式错误
    // 解析第二行内容为64位浮点数电能值，检查数值范围
    // 处理各种格式错误情况，返回详细的错误信息
    // 返回解析完成的数据结构或解析错误
}
```

## 5. 错误处理

### 5.1 错误类型定义

```go
// StorageError 存储错误类型
type StorageError struct {
    Operation string
    Path      string
    Cause     error
}

func (se *StorageError) Error() string {
    return fmt.Sprintf("storage operation '%s' failed on path '%s': %v", se.Operation, se.Path, se.Cause)
}

func (se *StorageError) Unwrap() error {
    return se.Cause
}

// 预定义错误
var (
    ErrFileNotFound     = errors.New("file not found")
    ErrInvalidFormat    = errors.New("invalid file format")
    ErrPermissionDenied = errors.New("permission denied")
)
```

## 6. 使用示例

### 6.1 基本使用

```go
package main

import (
    "fmt"
    "time"
    "your-project/internal/storage"
    "your-project/internal/pkgs/log"
)

func main() {
    // 创建存储配置
    storageConfig := &storage.Config{
        DataDir:         "./data",
        SyncWrite:       true,
        FilePermissions: 0644,
    }

    // 创建存储管理器
    logger := log.NewLogger(log.Config{Level: "info"})
    manager := storage.NewFileStorageManager(storageConfig, logger)

    // 写入设备电能数据
    deviceID := "a1"
    data := &storage.PowerData{
        Timestamp: time.Now().UnixMilli(),
        EnergyWH:  1500.75,
    }

    // 调用Write方法写入数据
    if err := manager.Write(deviceID, data); err != nil {
        logger.Error("Failed to write power data", "device", deviceID, "error", err)
        return
    }

    // 读取设备电能数据
    readData, err := manager.Read(deviceID)
    if err != nil {
        logger.Error("Failed to read power data", "device", deviceID, "error", err)
        return
    }

    fmt.Printf("Device %s power data: Timestamp=%d, Energy=%.2f Wh\n",
        deviceID, readData.Timestamp, readData.EnergyWH)
}
```

## 7. 注意事项

### 7.1 文件系统限制

- **文件路径长度**: 注意不同文件系统对路径长度的限制
- **文件名大小写**: Windows 系统文件名不区分大小写
- **文件权限**: 确保程序有足够权限读写文件
- **磁盘空间**: 监控磁盘空间，避免写入失败

## 8. 总结

文件持久化模块提供了简化的电能数据存储解决方案，具有以下特点：

- **简单设计**: 去除了复杂的备份、清理、重试机制
- **多设备支持**: 每个设备使用独立文件存储
- **原子写入**: 通过文件写入保证数据一致性
- **易于维护**: 简化的实现，便于理解和维护

该模块为 WinPower G2 Exporter 的电能计算功能提供了基础的持久化支持，专注于简单可靠的数据存储功能。