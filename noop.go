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

package telemetry

import "context"

// NoopLogger returns a no-op logger.
func NoopLogger() Logger {
	return noopLogger{}
}

type noopLogger struct{}

func (noopLogger) Debug(string, ...interface{})          {}
func (noopLogger) Info(string, ...interface{})           {}
func (n noopLogger) Error(string, error, ...interface{}) {}
func (n noopLogger) SetLevel(Level)                      {}
func (n noopLogger) Level() Level                        { return LevelNone }
func (n noopLogger) With(...interface{}) Logger          { return n }
func (n noopLogger) Context(context.Context) Logger      { return n }
func (n noopLogger) Metric(Metric) Logger                { return n }
func (n noopLogger) Clone() Logger                       { return NoopLogger() }
