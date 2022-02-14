// Copyright 2022 Tetrate
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

package scope

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/tetratelabs/telemetry"
	"github.com/tetratelabs/telemetry/function"
)

func TestLogger(t *testing.T) {
	emitter := func(w io.Writer) function.Emit {
		return func(level telemetry.Level, msg string, err error, values function.Values) {
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
			`level=info msg="text" [ctx value scope info lvl info missing (MISSING)]`, 1},
		{"info-with-values", telemetry.LevelInfo, func(l telemetry.Logger) { l.Info("text", "where", "there", 1, "1") },
			`level=info msg="text" [ctx value scope info-with-values lvl info missing (MISSING) where there 1 1]`, 1},
		{"error", telemetry.LevelInfo, func(l telemetry.Logger) { l.Error("text", errors.New("error")) },
			`level=error msg="text" err=error [ctx value scope error lvl info missing (MISSING)]`, 1},
		{"error-with-values", telemetry.LevelInfo, func(l telemetry.Logger) { l.Error("text", errors.New("error"), "where", "there", 1, "1") },
			`level=error msg="text" err=error [ctx value scope error-with-values lvl info missing (MISSING) where there 1 1]`, 1},
		{"debug", telemetry.LevelDebug, func(l telemetry.Logger) { l.Debug("text") },
			`level=debug msg="text" [ctx value scope debug lvl info missing (MISSING)]`, 0},
		{"debug-with-values", telemetry.LevelDebug, func(l telemetry.Logger) { l.Debug("text", "where", "there", 1, "1") },
			`level=debug msg="text" [ctx value scope debug-with-values lvl info missing (MISSING) where there 1 1]`, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(cleanup)

			var out bytes.Buffer
			UseLogger(function.NewLogger(emitter(&out)))

			_ = Register(tt.name, "test logger")
			logger, _ := Find(tt.name)

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

func TestSetLevel(t *testing.T) {
	t.Cleanup(cleanup)

	logger := Register("test-set-level", "test logger")

	withvalues := logger.With("key", "value")
	logger.SetLevel(telemetry.LevelInfo - 1)

	if withvalues.Level() != telemetry.LevelError {
		t.Fatalf("logger.Level()=%v, want: %v", withvalues.Level(), telemetry.LevelError)
	}
}

func TestTwoScopes(t *testing.T) {
	scopeA := Register("a", "Messages from a")
	scopeB := Register("b", "Messages from b")
	UseLogger(function.NewLogger(func(level telemetry.Level, msg string, err error, values function.Values) {
		// Do nothing
	}))

	scopeA.SetLevel(telemetry.LevelDebug)
	scopeB.SetLevel(telemetry.LevelDebug)

	if scopeA.Level() != telemetry.LevelDebug || scopeB.Level() != telemetry.LevelDebug {
		t.Fatalf("logger.Level=%s / logger2.Level=%s, want: %s/%s",
			scopeA.Level().String(), scopeB.Level().String(), telemetry.LevelDebug, telemetry.LevelDebug)
	}

	scopeA.SetLevel(telemetry.LevelInfo)
	if scopeA.Level() != telemetry.LevelInfo || scopeB.Level() != telemetry.LevelDebug {
		t.Fatalf("logger.Level=%s / logger2.Level=%s, want: %s/%s",
			scopeA.Level().String(), scopeB.Level().String(), telemetry.LevelInfo, telemetry.LevelDebug)
	}
}

func TestFind(t *testing.T) {
	s, ok := Find("unexisting")
	if ok {
		t.Fatalf("expected Find to have returned nil, got: %v", s)
	}
}

func cleanup() {
	scopes = make(map[string]*scope)
	uninitialized = make(map[string][]*scope)
	defaultLogger = nil
}

type mockMetric struct {
	telemetry.Metric
	count float64
}

func (m *mockMetric) RecordContext(_ context.Context, value float64) { m.count += value }
