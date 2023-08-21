// Copyright (c) Tetrate, Inc 2023.
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

package function

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/tetratelabs/telemetry"
)

func TestLogger(t *testing.T) {
	emitter := func(w io.Writer) Emit {
		return func(level telemetry.Level, msg string, err error, values Values) {
			_, _ = fmt.Fprintf(w, "level=%v msg=%q", level, msg)
			if err != nil {
				_, _ = fmt.Fprintf(w, " err=%v", err)
			}

			all := append(values.FromContext, values.FromLogger...)
			all = append(all, values.FromMethod...)
			_, _ = fmt.Fprintf(w, " %v", all)
		}
	}

	tests := []struct {
		name        string
		level       telemetry.Level
		logfunc     func(telemetry.Logger)
		expected    string
		metricCount float64
	}{
		{"none", telemetry.LevelNone, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, "", 1},
		{"disabled-info", telemetry.LevelNone, func(l telemetry.Logger) { l.Info("text") }, "", 1},
		{"disabled-debug", telemetry.LevelNone, func(l telemetry.Logger) { l.Debug("text") }, "", 0},
		{"disabled-error", telemetry.LevelNone, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, "", 1},
		{"info", telemetry.LevelInfo, func(l telemetry.Logger) { l.Info("text") },
			`level=info msg="text" [ctx value lvl info missing (MISSING)]`, 1},
		{"info-with-values", telemetry.LevelInfo, func(l telemetry.Logger) { l.Info("text", "where", "there", 1, "1") },
			`level=info msg="text" [ctx value lvl info missing (MISSING) where there 1 1]`, 1},
		{"error", telemetry.LevelInfo, func(l telemetry.Logger) { l.Error("text", errors.New("error")) },
			`level=error msg="text" err=error [ctx value lvl info missing (MISSING)]`, 1},
		{"error-with-values", telemetry.LevelInfo, func(l telemetry.Logger) { l.Error("text", errors.New("error"), "where", "there", 1, "1") },
			`level=error msg="text" err=error [ctx value lvl info missing (MISSING) where there 1 1]`, 1},
		{"debug", telemetry.LevelDebug, func(l telemetry.Logger) { l.Debug("text") },
			`level=debug msg="text" [ctx value lvl info missing (MISSING)]`, 0},
		{"debug-with-values", telemetry.LevelDebug, func(l telemetry.Logger) { l.Debug("text", "where", "there", 1, "1") },
			`level=debug msg="text" [ctx value lvl info missing (MISSING) where there 1 1]`, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			logger := NewLogger(emitter(&out))

			logger.SetLevel(tt.level)
			if logger.Level() != tt.level {
				t.Fatalf("loger.Level()=%s, want: %s", logger.Level(), tt.level)
			}

			metric := mockMetric{}
			ctx := telemetry.KeyValuesToContext(context.Background(), "ctx", "value")
			l := logger.Context(ctx).Metric(&metric).With().With(1, "").With("lvl", telemetry.LevelInfo).With("missing")

			tt.logfunc(l)

			str := out.String()
			if str != tt.expected {
				t.Fatalf("expected %s to match %s", str, tt.expected)
			}
			if metric.count != tt.metricCount {
				t.Fatalf("metric.count=%v, want %v", metric.count, tt.metricCount)
			}
		})
	}
}

func TestSetUnexpectedLevel(t *testing.T) {
	logger := NewLogger(nil)
	withvalues := logger.With("key", "value")
	logger.SetLevel(telemetry.LevelInfo - 1)

	if withvalues.Level() != telemetry.LevelError {
		t.Fatalf("Logger.Level()=%v, want: %v", withvalues.Level(), telemetry.LevelError)
	}
}

func TestClone(t *testing.T) {
	logger := NewLogger(nil)

	// Enhancing a logger with values does not alter the logger itself, and setting hte level should
	// affect the enhanced logger and the original one. We're not altering the 'scope' here.
	withvalues := logger.With("key", "value").Context(context.Background()).Metric(nil)

	// Cloning a logger returns a new independent one, with the logging level detached from the original
	cloned := withvalues.Clone()

	// Verify that the level is properly set. The level should affect both, as it is the same logger
	// with just additional info; it is not a clone.
	logger.SetLevel(telemetry.LevelDebug)

	if logger.Level() != telemetry.LevelDebug {
		t.Fatalf("logger.Level()=%v, want: %v", logger.Level(), telemetry.LevelDebug)
	}
	if withvalues.Level() != telemetry.LevelDebug {
		t.Fatalf("withvalues.Level()=%v, want: %v", withvalues.Level(), telemetry.LevelDebug)
	}
	if cloned.Level() != telemetry.LevelInfo {
		t.Fatalf("cloned.Level()=%v, want: %v", cloned.Level(), telemetry.LevelInfo)
	}

	withvalues.SetLevel(telemetry.LevelNone)

	if logger.Level() != telemetry.LevelNone {
		t.Fatalf("logger.Level()=%v, want: %v", logger.Level(), telemetry.LevelNone)
	}
	if withvalues.Level() != telemetry.LevelNone {
		t.Fatalf("withvalues.Level()=%v, want: %v", withvalues.Level(), telemetry.LevelNone)
	}
	if cloned.Level() != telemetry.LevelInfo {
		t.Fatalf("cloned.Level()=%v, want: %v", cloned.Level(), telemetry.LevelInfo)
	}

	cloned.SetLevel(telemetry.LevelError)

	if logger.Level() != telemetry.LevelNone {
		t.Fatalf("logger.Level()=%v, want: %v", logger.Level(), telemetry.LevelNone)
	}
	if withvalues.Level() != telemetry.LevelNone {
		t.Fatalf("withvalues.Level()=%v, want: %v", withvalues.Level(), telemetry.LevelNone)
	}
	if cloned.Level() != telemetry.LevelError {
		t.Fatalf("cloned.Level()=%v, want: %v", cloned.Level(), telemetry.LevelError)
	}
}

type mockMetric struct {
	telemetry.Metric
	count float64
}

func (m *mockMetric) RecordContext(_ context.Context, value float64) { m.count += value }
