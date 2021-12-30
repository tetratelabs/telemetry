package level

import (
	"context"

	"github.com/tetratelabs/telemetry"
)

// NoopLogger returns a no-op Logger.
func NoopLogger() Logger {
	return noopLogger{}
}

// noopLogger is a no-op logger.
type noopLogger struct{}

func (noopLogger) Debug(_ string, _ ...interface{}) {}

func (noopLogger) Info(_ string, _ ...interface{}) {}

func (n noopLogger) Error(_ string, _ error, _ ...interface{}) {}

func (n noopLogger) With(_ ...interface{}) telemetry.Logger { return n }

func (n noopLogger) Context(_ context.Context) telemetry.Logger { return n }

func (n noopLogger) Metric(_ telemetry.Metric) telemetry.Logger { return n }

func (noopLogger) SetLevel(_ Value) {}

func (noopLogger) Level() Value { return None }

func (n noopLogger) New() telemetry.Logger { return n }
