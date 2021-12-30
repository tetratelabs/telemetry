package telemetry

import "context"

// NoopLogger returns a no-op logger.
func NoopLogger() Logger {
	return noopLogger{}
}

type noopLogger struct{}

func (noopLogger) Debug(_ string, _ ...interface{}) {}

func (noopLogger) Info(_ string, _ ...interface{}) {}

func (n noopLogger) Error(_ string, _ error, _ ...interface{}) {}

func (n noopLogger) With(_ ...interface{}) Logger { return n }

func (n noopLogger) Context(_ context.Context) Logger { return n }

func (n noopLogger) Metric(_ Metric) Logger { return n }
