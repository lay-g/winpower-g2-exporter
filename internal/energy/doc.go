// Package energy 提供电能计算功能
//
// 该包实现了极简单线程架构的电能计算模块，用于从功率数据计算累计电能消耗。
// 通过全局锁确保所有计算操作串行执行，避免数据竞争，专注于为UPS设备提供精确的电能累计计算功能。
//
// 主要特性：
//   - 极简单线程架构：通过全局锁确保串行执行，彻底避免并发问题
//   - 精确电能计算：基于功率和时间间隔进行积分计算（Wh = W × 时间间隔）
//   - 存储解耦：完全依赖storage模块进行数据持久化
//   - 统计监控：提供简单统计信息用于监控和调试
//
// 使用示例：
//
//	// 创建电能服务
//	energyService := energy.NewEnergyService(storageManager, logger)
//
//	// 计算电能
//	totalEnergy, err := energyService.Calculate("ups-001", 500.0)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 获取电能数据
//	currentEnergy, err := energyService.Get("ups-001")
package energy
