package metrics

import (
	events "github.com/nocodeleaks/quepasa/events"
)

var (
	InternalEventsObserved = CreateCounterVecRecorder(
		"quepasa_internal_events_total",
		"Total internal events observed by the metrics subscriber",
		[]string{"event_name", "event_source", "event_status"},
	)
	InternalEventDuration = CreateHistogramVecRecorder(
		"quepasa_internal_event_duration_seconds",
		"Duration observed on internal events when provided by producers",
		[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		[]string{"event_name", "event_source", "event_status"},
	)
)

func init() {
	events.Subscribe(events.SubscribeOptions{
		Name:       "metrics",
		BufferSize: 512,
		Handler:    recordInternalEventMetrics,
	})
}

func recordInternalEventMetrics(event events.Event) {
	name := metricLabelValue(event.Name)
	source := metricLabelValue(event.Source)
	status := metricLabelValue(event.Status)

	InternalEventsObserved.WithLabelValues(name, source, status).Inc()
	if event.Duration > 0 {
		InternalEventDuration.WithLabelValues(name, source, status).Observe(event.Duration.Seconds())
	}
}

func metricLabelValue(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}
