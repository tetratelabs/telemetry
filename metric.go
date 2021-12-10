// Copyright (c) Tetrate, Inc 2021.
// Copyright 2019 Istio Authors
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

// Unit encodes the standard name for describing the quantity measured by a
// Metric (if applicable).
type Unit string

// Predefined units for use with the monitoring package.
const (
	None         Unit = "1"
	Bytes        Unit = "By"
	Seconds      Unit = "s"
	Milliseconds Unit = "ms"
)

// Metric collects numerical observations.
type Metric interface {
	// Increment records a value of 1 for the current Metric.
	// For Sums, this is equivalent to adding 1 to the current value.
	// For Gauges, this is equivalent to setting the value to 1.
	// For Distributions, this is equivalent to making an observation of value 1.
	Increment()

	// Decrement records a value of -1 for the current Metric.
	// For Sums, this is equivalent to subtracting -1 to the current value.
	// For Gauges, this is equivalent to setting the value to -1.
	// For Distributions, this is equivalent to making an observation of value -1.
	Decrement()

	// Name returns the name value of a Metric.
	Name() string

	// Record makes an observation of the provided value for the given Metric.
	// LabelValues added through With will be processed in sequence.
	Record(value float64)

	// RecordContext makes an observation of the provided value for the given
	// Metric.
	// If LabelValues for registered Labels are found in context, they will be
	// processed in sequence, after which the LabelValues added through With
	// are handled.
	RecordContext(ctx context.Context, value float64)

	// With returns the Metric with the provided LabelValues encapsulated. This
	// allows creating a set of pre-dimensioned data for recording purposes.
	// It also allows a way to clear out LabelValues found in an attached
	// Context if needing to sanitize.
	With(labelValues ...LabelValue) Metric
}

// LabelValue holds an action to take on a metric dimension's value.
type LabelValue interface{}

// Label holds a metric dimension which can be operated on using the interface
// methods.
type Label interface {
	// Insert will insert the provided value for the Label if not set.
	Insert(value string) LabelValue

	// Update will update the Label with provided value if already set.
	Update(value string) LabelValue

	// Upsert will insert or replace the provided value for the Label.
	Upsert(value string) LabelValue

	// Delete will remove the Label's value.
	Delete() LabelValue
}

// MetricSink bridges libraries bootstrapping metrics from metrics
// instrumentation implementations.
type MetricSink interface {
	// NewSum intents to create a new Metric with an aggregation type of Sum
	// (the values will be cumulative). That means that data collected by the
	// new Metric will be summed before export.
	NewSum(name, description string, opts ...MetricOption) Metric

	// NewGauge intents to creates a new Metric with an aggregation type of
	// LastValue. That means that data collected by the new Metric will export
	// only the last recorded value.
	NewGauge(name, description string, opts ...MetricOption) Metric

	// NewDistribution intents to create a new Metric with an aggregation type
	// of Distribution. This means that the data collected by the Metric will be
	// collected and exported as a histogram, with the specified bounds.
	NewDistribution(name, description string, bounds []float64, opts ...MetricOption) Metric

	// NewLabel creates a new Label to be used as a metrics dimension.
	NewLabel(name string) Label

	// ContextWithLabels takes the existing LabelValue collection found in
	// Context and appends the Label operations as received from the provided
	// values on top, which is then added to the returned Context. The function
	// can return an error in case the provided values contain invalid label
	// names.
	ContextWithLabels(ctx context.Context, values ...LabelValue) (context.Context, error)
}

// MetricOption implements a functional option type for our Metrics.
type MetricOption func(*MetricOptions)

// MetricOptions hold commonly used but optional Metric configuration.
type MetricOptions struct {
	// Unit holds the unit specifier of a Metric.
	Unit Unit
	// Labels holds the registered dimensions for the Metric.
	Labels []Label
}

// WithLabels provides a configuration MetricOption for a new Metric, providing
// the required dimensions for data collection of that Metric.
func WithLabels(labels ...Label) MetricOption {
	return func(opts *MetricOptions) {
		opts.Labels = labels
	}
}

// WithUnit provides a configuration MetricOption for a new Metric, providing
// Unit of measure information for a new Metric.
func WithUnit(unit Unit) MetricOption {
	return func(opts *MetricOptions) {
		opts.Unit = unit
	}
}
