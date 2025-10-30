package energy

import (
	"sync"
	"time"
)

// EnergyInterface 电能模块接口
type EnergyInterface interface {
	// Calculate 计算电能
	// 参数:
	//   - deviceID: 设备ID
	//   - power: 当前功率值(W)
	// 返回:
	//   - 累计电能值(Wh)
	//   - 错误信息
	Calculate(deviceID string, power float64) (float64, error)

	// Get 获取最新电能数据
	// 参数:
	//   - deviceID: 设备ID
	// 返回:
	//   - 累计电能值(Wh)
	//   - 错误信息
	Get(deviceID string) (float64, error)

	// GetStats 获取统计信息
	GetStats() *Stats
}

// Stats 统计信息
type Stats struct {
	TotalCalculations  int64         `json:"total_calculations"`   // 总计算次数
	TotalErrors        int64         `json:"total_errors"`         // 总错误次数
	LastUpdateTime     time.Time     `json:"last_update_time"`     // 最后更新时间
	AvgCalculationTime time.Duration `json:"avg_calculation_time"` // 平均计算时间
	mutex              sync.RWMutex  // 保护统计信息
}

// GetTotalCalculations 获取总计算次数（线程安全）
func (s *Stats) GetTotalCalculations() int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.TotalCalculations
}

// GetTotalErrors 获取总错误次数（线程安全）
func (s *Stats) GetTotalErrors() int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.TotalErrors
}

// GetLastUpdateTime 获取最后更新时间（线程安全）
func (s *Stats) GetLastUpdateTime() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.LastUpdateTime
}

// GetAvgCalculationTime 获取平均计算时间（线程安全）
func (s *Stats) GetAvgCalculationTime() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.AvgCalculationTime
}
