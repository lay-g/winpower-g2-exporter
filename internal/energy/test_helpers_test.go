package energy

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

// newTestLogger creates a test logger that only emits fatal-level messages.
// This keeps routine test scenarios quiet while still surfacing critical issues.
func newTestLogger(t zaptest.TestingT) *zap.Logger {
	return zaptest.NewLogger(t, zaptest.Level(zapcore.FatalLevel))
}
