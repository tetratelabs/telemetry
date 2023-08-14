package telemetry_test

import (
	"testing"

	"github.com/tetratelabs/telemetry"
)

var _ telemetry.Label = (*label)(nil)

type metricSink struct {
	options telemetry.MetricOptions
}

type label string

func (l label) Insert(string) telemetry.LabelValue { return nil }
func (l label) Update(string) telemetry.LabelValue { return nil }
func (l label) Upsert(string) telemetry.LabelValue { return nil }
func (l label) Delete() telemetry.LabelValue       { return nil }

func newMetricSink(options ...telemetry.MetricOption) *metricSink {
	var ms metricSink
	for _, opt := range options {
		opt(&ms.options)
	}
	return &ms
}

func TestMetricOptions(t *testing.T) {
	label1 := label("label1")
	label2 := label("label2")
	ms := newMetricSink(
		telemetry.WithUnit(telemetry.Milliseconds),
		telemetry.WithEnabled(func() bool { return true }),
		telemetry.WithLabels(label1, label2),
	)

	if ms.options.EnabledCondition == nil {
		t.Fatal("expected EnabledCondition to hold a function")
	}
	if !ms.options.EnabledCondition() {
		t.Errorf("expected EnabledCondition function to return true")
	}
	if ms.options.Unit != telemetry.Milliseconds {
		t.Errorf("expected Unit to be ms (milliseconds)")
	}
	if len(ms.options.Labels) != 2 {
		t.Fatalf("unexpected label count: want: 2, have: %d", len(ms.options.Labels))
	}
	if ms.options.Labels[0] != label1 {
		t.Errorf("[0] unexpected label value: want: %s, have: %s", label1, ms.options.Labels[0].(label))
	}
	if ms.options.Labels[1] != label2 {
		t.Errorf("[1] unexpected label value: want: %s, have: %s", label2, ms.options.Labels[1].(label))
	}
}
