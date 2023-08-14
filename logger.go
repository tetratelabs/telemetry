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

package telemetry

import "context"

// Logger provides a simple yet powerful logging abstraction.
type Logger interface {
	// Debug logging with key-value pairs. Don't be shy, use it.
	Debug(msg string, keyValuePairs ...interface{})

	// Info logging with key-value pairs. This is for informational, but not
	// directly actionable conditions. It is highly recommended you attach a
	// Metric to these types of messages. Where a single informational or
	// warning style message might not be reason for action, a change in
	// occurrence does warrant action. By attaching a Metric for these logging
	// situations, you make this easy through histograms, thresholds, etc.
	Info(msg string, keyValuePairs ...interface{})

	// Error logging with key-value pairs. Use this when application state and
	// stability are at risk. These types of conditions are actionable and often
	// alerted on. It is very strongly encouraged to add a Metric to each of
	// these types of messages. Metrics provide the easiest way to coordinate
	// processing of these concerns and triggering alerting systems through your
	// metrics backend.
	Error(msg string, err error, keyValuePairs ...interface{})

	// SetLevel provides the ability to set the desired logging level.
	// This function can be used at runtime and must be safe for concurrent use.
	//
	// Note for Logger implementations, When creating a new Logger with the
	// With, Context, or Metric methods, the level should be set-able for all
	// from any of the Loggers sharing the same root Logger.
	SetLevel(lvl Level)

	// Level returns the currently configured logging level.
	Level() Level

	// With returns a new Logger decorated with the provided key-value pairs.
	With(keyValuePairs ...interface{}) Logger

	// Context returns a new Logger having access to Context for inclusion of
	// registered key-value pairs found in Context. If a Metric is also attached
	// to the Logger, the Metric LabelValue directives found in Context will
	// also be processed.
	Context(ctx context.Context) Logger

	// Metric returns a new Logger which will emit a measurement for the
	// provided Metric when the Log level is either Info or Error.
	// **Note** that in the event the Logger is set to only output Error level
	// messages, Info messages even though silenced from a logging perspective,
	// will still emit their Metric measurements.
	Metric(m Metric) Logger

	// Clone returns a new Logger based on the original implementation.
	Clone() Logger
}

// KeyValuesToContext takes provided Context, retrieves the already stored
// key-value pairs from it, appends the in this function provided key-value
// pairs and stores the result in the returned Context.
func KeyValuesToContext(ctx context.Context, keyValuePairs ...interface{}) context.Context {
	if len(keyValuePairs) == 0 {
		return ctx
	}
	if len(keyValuePairs)%2 != 0 {
		keyValuePairs = append(keyValuePairs, "(MISSING)")
	}
	args := KeyValuesFromContext(ctx)
	args = append(args, keyValuePairs...)
	return context.WithValue(ctx, ctxKVP, args)
}

// KeyValuesFromContext retrieves key-value pairs that might be stored in the
// provided Context. Logging implementations must use this function to retrieve
// the key-value pairs they need to include if a Context object was attached to
// them.
func KeyValuesFromContext(ctx context.Context) (keyValuePairs []interface{}) {
	keyValuePairs, _ = ctx.Value(ctxKVP).([]interface{})
	return
}

// RemoveKeyValuesFromContext returns a Context that copies the provided Context
// but removes the key-value pairs that were stored.
func RemoveKeyValuesFromContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKVP, nil)
}

type tCtxKVP string

var ctxKVP tCtxKVP
