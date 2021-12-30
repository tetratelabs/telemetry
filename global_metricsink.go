package telemetry

import (
	"sync"
)

// OnGlobalMetricSinkFn holds a function signature which can be used to register
// Metric bootstrapping that needs to be called after the GlobalMetricSink has
// been registered.
type OnGlobalMetricSinkFn func(m MetricSink)

var (
	mtx        sync.Mutex
	metricSink MetricSink
	callbacks  []func(MetricSink)
)

// SetGlobalMetricSink allows one to set a global MetricSink, after which all
// registered OnGlobalMetricSinkFn callback functions are executed.
func SetGlobalMetricSink(ms MetricSink) {
	mtx.Lock()
	defer mtx.Unlock()

	metricSink = ms
	for _, callback := range callbacks {
		callback(ms)
	}
	callbacks = nil
}

// ToGlobalMetricSink allows one to set callback functions to bootstrap Metrics
// as soon as the Global MetricSink has been registered. If the MetricSink has
// already been registered, this callback will happen immediately.
func ToGlobalMetricSink(callback OnGlobalMetricSinkFn) {
	mtx.Lock()
	defer mtx.Unlock()

	if metricSink != nil {
		callback(metricSink)
		return
	}

	callbacks = append(callbacks, callback)
}
