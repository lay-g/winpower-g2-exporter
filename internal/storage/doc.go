// Package storage provides file-based persistence for power data.
// Each device uses independent files for data storage with atomic write guarantees.
//
// The storage module is responsible for:
// - Persisting cumulative energy data to device-specific files
// - Reading historical data during application startup
// - Providing atomic write operations to prevent data corruption
// - Supporting multi-device scenarios with isolated file storage
//
// Configuration:
//
// The storage module defines its configuration structure (Config) but does not handle
// configuration loading from any sources (files, environment variables, command line, etc.).
// Configuration initialization is the responsibility of external callers, typically the
// config module or the main application.
//
// File format:
//
//	Line 1: Unix timestamp in milliseconds
//	Line 2: Accumulated energy value in Wh (can be negative for net energy)
//
// Example file content:
//
//	1694678400000
//	15000.50
package storage
