package energy

import "errors"

var (
	// ErrInvalidDeviceID 设备ID无效
	ErrInvalidDeviceID = errors.New("invalid device ID: device ID cannot be empty")

	// ErrInvalidPower 功率值无效
	ErrInvalidPower = errors.New("invalid power value")

	// ErrStorageRead 存储读取失败
	ErrStorageRead = errors.New("failed to read data from storage")

	// ErrStorageWrite 存储写入失败
	ErrStorageWrite = errors.New("failed to write data to storage")

	// ErrCalculation 电能计算失败
	ErrCalculation = errors.New("energy calculation failed")
)
