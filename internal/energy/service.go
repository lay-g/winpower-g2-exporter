package energy

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
)

// EnergyService 电能服务（极简架构）
type EnergyService struct {
	storage storage.StorageManager // 存储接口
	logger  log.Logger             // 日志器
	mutex   sync.RWMutex           // 全局读写锁，确保串行执行
	stats   *Stats                 // 统计信息
}

// NewEnergyService 创建电能服务
func NewEnergyService(storage storage.StorageManager, logger log.Logger) *EnergyService {
	if storage == nil {
		panic("storage manager cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	return &EnergyService{
		storage: storage,
		logger:  logger,
		stats: &Stats{
			LastUpdateTime: time.Now(),
		},
	}
}

// Calculate 计算电能（对外接口，串行执行）
func (es *EnergyService) Calculate(deviceID string, power float64) (float64, error) {
	// 参数验证
	if deviceID == "" {
		return 0, ErrInvalidDeviceID
	}

	// 获取全局写锁（确保串行执行）
	es.mutex.Lock()
	defer es.mutex.Unlock()

	start := time.Now()
	logger := es.logger.With(
		log.String("device_id", deviceID),
		log.Float64("power", power),
	)

	logger.Debug("Starting energy calculation")

	// 加载历史数据
	historyData, err := es.loadHistoryData(deviceID)
	if err != nil {
		es.updateStats(false, time.Since(start))
		logger.Error("Failed to load history data", log.Err(err))
		return 0, fmt.Errorf("%w: %v", ErrStorageRead, err)
	}

	// 计算累计电能
	currentTime := time.Now()
	totalEnergy, err := es.calculateTotalEnergy(historyData, power, currentTime)
	if err != nil {
		es.updateStats(false, time.Since(start))
		logger.Error("Failed to calculate energy", log.Err(err))
		return 0, fmt.Errorf("%w: %v", ErrCalculation, err)
	}

	// 保存数据到storage
	if err := es.saveData(deviceID, totalEnergy); err != nil {
		es.updateStats(false, time.Since(start))
		logger.Error("Failed to save data", log.Err(err))
		return 0, fmt.Errorf("%w: %v", ErrStorageWrite, err)
	}

	// 更新统计信息
	duration := time.Since(start)
	es.updateStats(true, duration)

	logger.Info("Energy calculation completed",
		log.Float64("total_energy", totalEnergy),
		log.Duration("duration", duration))

	return totalEnergy, nil
}

// Get 获取最新电能数据（对外接口）
func (es *EnergyService) Get(deviceID string) (float64, error) {
	// 参数验证
	if deviceID == "" {
		return 0, ErrInvalidDeviceID
	}

	// 获取读锁（允许并发读取）
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	// 从storage读取设备数据
	data, err := es.storage.Read(deviceID)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrStorageRead, err)
	}

	return data.EnergyWH, nil
}

// GetStats 获取统计信息
func (es *EnergyService) GetStats() *Stats {
	return es.stats
}

// calculateTotalEnergy 计算累计电能（内部方法）
func (es *EnergyService) calculateTotalEnergy(historyData *storage.PowerData, currentPower float64, currentTime time.Time) (float64, error) {
	// 首次计算，从0开始
	if historyData == nil {
		return 0, nil
	}

	// 计算时间间隔（小时）
	lastTime := time.UnixMilli(historyData.Timestamp)
	timeIntervalHours := currentTime.Sub(lastTime).Hours()

	// 计算间隔电能 = 功率 × 时间间隔
	intervalEnergy := currentPower * timeIntervalHours

	// 计算新的累计电能 = 历史电能 + 间隔电能
	totalEnergy := historyData.EnergyWH + intervalEnergy

	// 精度控制：保留2位小数（0.01Wh精度）
	totalEnergy = math.Round(totalEnergy*100) / 100

	return totalEnergy, nil
}

// loadHistoryData 加载历史数据（内部方法）
func (es *EnergyService) loadHistoryData(deviceID string) (*storage.PowerData, error) {
	// 调用storage.Read读取历史数据
	data, err := es.storage.Read(deviceID)
	if err != nil {
		// 处理文件不存在等错误情况
		if err == storage.ErrFileNotFound {
			// 首次访问，返回nil表示从0开始
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}

// saveData 保存数据（内部方法）
func (es *EnergyService) saveData(deviceID string, energy float64) error {
	// 创建新的PowerData结构
	data := &storage.PowerData{
		Timestamp: time.Now().UnixMilli(), // 毫秒时间戳
		EnergyWH:  energy,                 // 累计电能(Wh)
	}

	// 调用storage.Write保存数据
	if err := es.storage.Write(deviceID, data); err != nil {
		return err
	}

	return nil
}

// updateStats 更新统计信息（内部方法）
func (es *EnergyService) updateStats(success bool, duration time.Duration) {
	es.stats.mutex.Lock()
	defer es.stats.mutex.Unlock()

	// 更新总计算次数
	es.stats.TotalCalculations++

	// 更新错误次数
	if !success {
		es.stats.TotalErrors++
	}

	// 计算平均执行时间
	if es.stats.AvgCalculationTime == 0 {
		es.stats.AvgCalculationTime = duration
	} else {
		// 移动平均
		es.stats.AvgCalculationTime = (es.stats.AvgCalculationTime + duration) / 2
	}

	// 更新最后更新时间
	es.stats.LastUpdateTime = time.Now()
}
