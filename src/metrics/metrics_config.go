package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricRecorder defines a generic interface for all metric types
type MetricRecorder interface {
	// Common methods that all metric types support
}

// CounterRecorder for counter metrics
type CounterRecorder interface {
	Inc()
}

// GaugeRecorder for gauge metrics
type GaugeRecorder interface {
	Add(float64)
}

// HistogramRecorder for histogram metrics
type HistogramRecorder interface {
	Observe(float64)
}

// VectorRecorder for vector metrics (CounterVec, HistogramVec, etc.)
type VectorRecorder interface {
	WithLabelValues(...string) interface{}
}

// CounterVecRecorder for counter vector metrics
type CounterVecRecorder interface {
	WithLabelValues(...string) CounterRecorder
}

// HistogramVecRecorder for histogram vector metrics
type HistogramVecRecorder interface {
	WithLabelValues(...string) HistogramRecorder
}

// AsyncCounterVecRecorder wraps a counter vector with async execution
type AsyncCounterVecRecorder struct {
	recorder *prometheus.CounterVec
	enabled  bool
}

func (a AsyncCounterVecRecorder) WithLabelValues(labels ...string) interface{} {
	if a.enabled {
		counter := a.recorder.WithLabelValues(labels...)
		return &AsyncCounter{recorder: counter.(prometheus.Counter), enabled: a.enabled}
	}
	return &NoOpCounter{}
}

// AsyncHistogramVecRecorder wraps a histogram vector with async execution
type AsyncHistogramVecRecorder struct {
	recorder *prometheus.HistogramVec
	enabled  bool
}

func (a AsyncHistogramVecRecorder) WithLabelValues(labels ...string) HistogramRecorder {
	if a.enabled {
		histogram := a.recorder.WithLabelValues(labels...)
		if hist, ok := histogram.(prometheus.Histogram); ok {
			return &AsyncHistogram{recorder: hist, enabled: a.enabled}
		}
	}
	return &NoOpHistogram{}
}

// AsyncCounter wraps a counter with async execution
type AsyncCounter struct {
	recorder prometheus.Counter
	enabled  bool
}

func (a *AsyncCounter) Inc() {
	if a.enabled {
		go a.recorder.Inc()
	}
}

func (a *AsyncCounter) Add(delta float64) {
	if a.enabled {
		go func() { a.recorder.Add(delta) }()
	}
}

// NoOpCounter is a no-operation counter
type NoOpCounter struct{}

func (n *NoOpCounter) Inc()        {}
func (n *NoOpCounter) Add(float64) {}

// AsyncHistogram wraps a histogram with async execution
type AsyncHistogram struct {
	recorder prometheus.Histogram
	enabled  bool
}

func (a *AsyncHistogram) Observe(value float64) {
	if a.enabled {
		go func() { a.recorder.Observe(value) }()
	}
}

// NoOpHistogram is a no-operation histogram
type NoOpHistogram struct{}

func (n *NoOpHistogram) Observe(float64) {}

// NoOpRecorder is a no-operation recorder that does nothing
type NoOpRecorder struct{}

func (n NoOpRecorder) Inc()                                  {}
func (n NoOpRecorder) Add(float64)                           {}
func (n NoOpRecorder) Observe(float64)                       {}
func (n NoOpRecorder) WithLabelValues(...string) interface{} { return n }

// AsyncCounterRecorder wraps a counter with async execution
type AsyncCounterRecorder struct {
	recorder CounterRecorder
	enabled  bool
}

func (a AsyncCounterRecorder) Inc() {
	if a.enabled {
		go a.recorder.Inc()
	}
}

// AsyncGaugeRecorder wraps a gauge with async execution
type AsyncGaugeRecorder struct {
	recorder GaugeRecorder
	enabled  bool
}

func (a AsyncGaugeRecorder) Add(value float64) {
	if a.enabled {
		go func() { a.recorder.Add(value) }()
	}
}

// AsyncHistogramRecorder wraps a histogram with async execution
type AsyncHistogramRecorder struct {
	recorder HistogramRecorder
	enabled  bool
}

func (a AsyncHistogramRecorder) Observe(value float64) {
	if a.enabled {
		go func() { a.recorder.Observe(value) }()
	}
}

// AsyncVectorRecorder wraps a vector metric with async execution
type AsyncVectorRecorder struct {
	recorder VectorRecorder
	enabled  bool
}

func (a AsyncVectorRecorder) WithLabelValues(labels ...string) interface{} {
	if a.enabled {
		return a.recorder.WithLabelValues(labels...)
	}
	return NoOpRecorder{}
}

// Global async metric recorders - these are initialized at startup
var (
// Metrics are now async by default, no need for separate async versions
)

// metrics holds metrics specific to the whatsmeow module
type metrics struct {
	MessagesReceived         CounterRecorder
	MessageReceiveErrors     CounterRecorder
	MessageReceiveUnhandled  CounterRecorder
	MessageReceiveSyncEvents CounterRecorder
}

// Getmetrics returns metrics for the whatsmeow module
func Getmetrics() *metrics {
	return &metrics{}
}

// NoOpHistogramVecRecorder provides no-op implementation for histogram vectors
type NoOpHistogramVecRecorder struct{}

func (n *NoOpHistogramVecRecorder) WithLabelValues(labels ...string) HistogramRecorder {
	return &NoOpHistogram{}
}

// PrometheusHistogramVecRecorder wraps prometheus HistogramVec
type PrometheusHistogramVecRecorder struct {
	vec *prometheus.HistogramVec
}

func (p *PrometheusHistogramVecRecorder) WithLabelValues(labels ...string) HistogramRecorder {
	return &AsyncHistogramRecorder{
		recorder: p.vec.WithLabelValues(labels...),
		enabled:  MetricsEnabled,
	}
}

// CreateHistogramVecRecorder creates a new histogram vector recorder
// This is a generic factory function for modules to create their own histogram vectors
func CreateHistogramVecRecorder(name, help string, buckets []float64, labelNames []string) HistogramVecRecorder {
	if MetricsEnabled {
		vec := promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		}, labelNames)
		return &PrometheusHistogramVecRecorder{vec: vec}
	}
	return &NoOpHistogramVecRecorder{}
}

// NoOpCounterVecRecorder provides no-op implementation for counter vectors
type NoOpCounterVecRecorder struct{}

func (n *NoOpCounterVecRecorder) WithLabelValues(labels ...string) CounterRecorder {
	return &NoOpCounter{}
}

// PrometheusCounterVecRecorder wraps prometheus CounterVec
type PrometheusCounterVecRecorder struct {
	vec *prometheus.CounterVec
}

func (p *PrometheusCounterVecRecorder) WithLabelValues(labels ...string) CounterRecorder {
	return &AsyncCounterRecorder{
		recorder: &AsyncCounter{recorder: p.vec.WithLabelValues(labels...), enabled: MetricsEnabled},
		enabled:  MetricsEnabled,
	}
}

// CreateCounterVecRecorder creates a new counter vector recorder
// This is a generic factory function for modules to create their own counter vectors
func CreateCounterVecRecorder(name, help string, labelNames []string) CounterVecRecorder {
	if MetricsEnabled {
		vec := promauto.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: help,
		}, labelNames)
		return &PrometheusCounterVecRecorder{vec: vec}
	}
	return &NoOpCounterVecRecorder{}
}

// NoOpCounterRecorder provides no-op implementation for counters
type NoOpCounterRecorder struct{}

func (n *NoOpCounterRecorder) Inc() {}

// PrometheusCounterRecorder wraps prometheus Counter
type PrometheusCounterRecorder struct {
	counter prometheus.Counter
}

func (p *PrometheusCounterRecorder) Inc() {
	p.counter.Inc()
}

// CreateCounterRecorder creates a new counter recorder
// This is a generic factory function for modules to create their own counters
func CreateCounterRecorder(name, help string) CounterRecorder {
	if MetricsEnabled {
		counter := promauto.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})
		return &AsyncCounterRecorder{
			recorder: &AsyncCounter{recorder: counter, enabled: MetricsEnabled},
			enabled:  MetricsEnabled,
		}
	}
	return &NoOpCounterRecorder{}
}
