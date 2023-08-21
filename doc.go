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

/*
Package telemetry holds observability facades for our services and libraries.

The provided interface here allows for instrumenting libraries and packages
without any dependencies on Logging and Metric instrumentation implementations.
This allows a consistent way of authoring Log lines and Metrics for the
producers of these libraries and packages while providing consumers the ability
to plug in the implementations of their choice.

The following requirements helped shape the form of the interfaces.

  - Simple to use!

Or developers will resort to using `fmt.Printf()`

  - No elaborate amount of logging levels.

Error: something happened that we can't gracefully recover from.
This is a log line that should be actionable by an operator and be
alerted on.

Info: something happened that might be of interest but does not impact
the application stability. E.g. someone gave the wrong credentials and
was therefore denied access, parsing error on external input, etc.

Debug: anything that can help to understand application state during
development.

More levels get tricky to reason about when writing log lines or establishing
the right level of verbosity at runtime. By the above explanations fatal folds
into error, warning folds into info, and trace folds into debug.

We trust more in partitioning loggers per domain, component, etc. and allow them
to be individually addressed to required log levels than controlling a single
logger with more levels.

We also believe that most logs should be metrics. Anything above Debug level
should be able to emit a metric which can be use for dashboards, alerting, etc.

  - Structured logging from the interface side.

We want the ability to rollup / aggregate over the same message while allowing
for contextual data to be added. A logging implementation can make the choice
how to present to provided log data. This can be 100% structured, a single log
line, or a combination.

  - Allow pass through of contextual values.

Allow the Go Context object to be passed and have a registry for values of
interest we want to pull from context. A good example of an item we want to
automatically include in log lines is the `x-request-id` so we can tie log
lines produced in the request path together.

  - Allow each component to have their own "scope".

This allows us to control per component which levels of log lines we want
to output at runtime. The interface design allows for this to be
implemented without having an opinion on it. By providing at each library
or component entry point the ability to provide a Logger implementation,
this can be easily achieved.

  - Zero dependencies.

Look at that lovely very empty go.mod and non-existent go.sum file.
*/
package telemetry
