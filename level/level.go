// Copyright (c) Tetrate, Inc 2021.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package level provides an interface and wrapper implementation for leveled
// logging.
package level

import (
	"context"
	"sync/atomic"

	"github.com/tetratelabs/telemetry"
)

// Value is an enumeration of the available log levels.
type Value int32

// Available log levels.
const (
	None  Value = 0
	Error Value = 1
	Info  Value = 5
	Debug Value = 10
)

// Logger is an interface for Loggers that support Log Levels.
type Logger interface {
	telemetry.Logger

	// SetLevel provides the ability to set the desired logging level.
	// This function can be used at runtime and must be safe for concurrent use.
	//
	// Note for Logger implementations, When creating a new Logger with the
	// With, Context, or Metric methods, the level should be set-able for all
	// from any of the Loggers sharing the same root Logger.
	SetLevel(lvl Value)

	// Level returns the currently configured logging level.
	Level() Value

	// New returns a new Logger based on the original implementation but with
	// the log level decoupled.
	New() telemetry.Logger
}

// Wrap takes a telemetry.Logger implementation and wraps it with Log Level
// functionality.
func Wrap(l telemetry.Logger) Logger {
	if ll, ok := l.(Logger); ok {
		// already has Level functionality
		return ll
	}
	lvl := int32(Debug)
	return &wrapper{
		logger: l,
		lvl:    &lvl,
	}
}

var _ Logger = (*wrapper)(nil)

type wrapper struct {
	logger telemetry.Logger
	lvl    *int32
}

func (l *wrapper) Debug(msg string, keyValuePairs ...interface{}) {
	if atomic.LoadInt32(l.lvl) >= int32(Debug) {
		keyValuePairs = append(DebugLevel, keyValuePairs...)
		l.logger.Debug(msg, keyValuePairs...)
	}
}

func (l *wrapper) Info(msg string, keyValuePairs ...interface{}) {
	if atomic.LoadInt32(l.lvl) >= int32(Info) {
		keyValuePairs = append(InfoLevel, keyValuePairs...)
		l.logger.Info(msg, keyValuePairs...)
	}
}

func (l *wrapper) Error(msg string, err error, keyValuePairs ...interface{}) {
	if atomic.LoadInt32(l.lvl) >= int32(Error) {
		keyValuePairs = append(ErrorLevel, keyValuePairs...)
		l.logger.Error(msg, err, keyValuePairs...)
	}
}

func (l *wrapper) With(keyValuePairs ...interface{}) telemetry.Logger {
	return &wrapper{
		logger: l.logger.With(keyValuePairs...),
		lvl:    l.lvl,
	}
}

func (l *wrapper) Context(ctx context.Context) telemetry.Logger {
	return &wrapper{
		logger: l.logger.Context(ctx),
		lvl:    l.lvl,
	}
}

func (l *wrapper) Metric(m telemetry.Metric) telemetry.Logger {
	return &wrapper{
		logger: l.logger.Metric(m),
		lvl:    l.lvl,
	}
}

func (l *wrapper) New() telemetry.Logger {
	lvl := atomic.LoadInt32(l.lvl)
	return &wrapper{
		logger: l.logger.Context(context.Background()),
		lvl:    &lvl,
	}
}

func (l *wrapper) SetLevel(lvl Value) {
	if lvl < Info {
		lvl = Error
	} else if lvl < Debug {
		lvl = Info
	} else {
		lvl = Debug
	}
	atomic.StoreInt32(l.lvl, int32(lvl))
}

func (l *wrapper) Level() Value {
	return Value(atomic.LoadInt32(l.lvl))
}

// Adjustable level key-value pairs
var (
	DebugLevel = []interface{}{"level", "debug"}
	InfoLevel  = []interface{}{"level", "info"}
	ErrorLevel = []interface{}{"level", "error"}
)
