# QuePasa Metrics System - AI Agent Instructions

## Overview
The QuePasa metrics system is designed with a modular, extensible architecture that allows for multiple monitoring backends (Prometheus, InfluxDB, etc.) while maintaining clean separation of concerns between modules.

## Core Principles

### 1. Generic Factory Function- **ALWAYS follow the module ownership principle**: API metrics belong in API module, WhatsMeow metrics in WhatsMeow module**ALWAYS follow the module ownership principle**: API metrics belong in API module, WhatsMeow metrics in WhatsMeow module
- **Central metrics module provides only generic factory functions**: Functions like `CreateCounterRecorder()`, `CreateHistogramVecRecorder()` for creating metrics
- **No module-specific knowledge in central module**: The metrics module doesn't know about API, WhatsMeow, or other modules
- **Direct variable initialization**: Each module initializes its own metrics directly at package level using factory functions

### 2. Module-Specific Metrics
- **Each module owns its metrics**: API metrics belong in the API module, RabbitMQ metrics in RabbitMQ module, etc.
- **Direct initialization**: Metrics are initialized as package-level variables using generic factory functions
- **Clean separation**: Module-specific metrics are defined and managed within their respective modules

### 3. Environment-Based Configuration
- **Centralized enable/disable**: Metrics system is enabled/disabled via `METRICS_ENABLED` environment variable
- **Automatic backend selection**: When enabled, Prometheus backend is used; when disabled, no-op implementations are used
- **No manual configuration**: Modules initialize metrics directly without setup functions

## Architecture

### Generic Factory Functions
```go
// CreateCounterRecorder creates a new counter recorder
func CreateCounterRecorder(name, help string) CounterRecorder

// CreateCounterVecRecorder creates a new counter vector recorder
func CreateCounterVecRecorder(name, help string, labelNames []string) CounterVecRecorder

// CreateHistogramVecRecorder creates a new histogram vector recorder
func CreateHistogramVecRecorder(name, help string, buckets []float64, labelNames []string) HistogramVecRecorder
```

### Module-Specific Implementation Example
```go
package api

import metrics "github.com/nocodeleaks/quepasa/metrics"

// API-specific metrics initialized directly using generic factory functions
var (
	MessagesSent         = metrics.CreateCounterRecorder("quepasa_api_messages_sent_total", "Total messages sent via API")
	MessageSendErrors    = metrics.CreateCounterRecorder("quepasa_api_message_send_errors_total", "Total message send errors via API")
	MessagesReceived     = metrics.CreateCounterRecorder("quepasa_api_messages_received_total", "Total messages received via API")
	MessageReceiveErrors = metrics.CreateCounterRecorder("quepasa_api_message_receive_errors_total", "Total message receive errors via API")
	APIProcessingTime    = metrics.CreateHistogramVecRecorder("quepasa_api_request_duration_seconds", "Time spent processing API requests", []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0}, []string{"method", "endpoint", "status_code"})
)
```

### Interface Definitions
type HistogramVecRecorder interface {
    WithLabelValues(...string) HistogramRecorder
}
```
type CounterRecorder interface {
    Inc()
    Add(float64)
}

// HistogramRecorder - Interface for histogram metrics
type HistogramRecorder interface {
    Observe(float64)
}

// GaugeRecorder - Interface for gauge metrics
type GaugeRecorder interface {
    Set(float64)
    Inc()
    Dec()
    Add(float64)
    Sub(float64)
}
```

### Module Structure
```
metrics/                    # Central metrics module
├── AGENTS.md              # This documentation
├── metric_interfaces.go   # Interface definitions
├── metric_extensions.go   # Backend implementations (Prometheus)
├── metrics_config.go      # Configuration and initialization
└── ...

api/                       # API module
├── api_metrics.go         # API-specific metric definitions
├── api_metrics_internal.go # Internal API metric implementations
└── ...

rabbitmq/                  # RabbitMQ module
├── rabbitmq_metrics.go    # RabbitMQ-specific metrics
└── ...
```

## Implementation Guidelines

### For Module Developers

#### 1. Define Module Metrics
Initialize metrics directly as package-level variables using generic factory functions:

```go
package api

import (
    "github.com/nocodeleaks/quepasa/metrics"
)

// API-specific metrics initialized directly using generic factory functions
var (
    MessagesSent         = metrics.CreateCounterRecorder("quepasa_api_messages_sent_total", "Total messages sent via API")
    MessageSendErrors    = metrics.CreateCounterRecorder("quepasa_api_message_send_errors_total", "Total message send errors via API")
    MessagesReceived     = metrics.CreateCounterRecorder("quepasa_api_messages_received_total", "Total messages received via API")
    MessageReceiveErrors = metrics.CreateCounterRecorder("quepasa_api_message_receive_errors_total", "Total message receive errors via API")
)
```

#### 2. Use Metrics in Code
```go
func SendMessage(w http.ResponseWriter, r *http.Request) {
    startTime := time.Now()
    defer func() {
        ObserveAPIRequestDuration(r.Method, "/send", "200", time.Since(startTime).Seconds())
    }()

    // Your logic here
    if success {
        MessagesSent.Inc()
    } else {
        MessageSendErrors.Inc()
    }
}
```

#### 3. Define Internal Metrics
For module-specific metrics that don't fit the generic interfaces:

```go
package api

import (
    "github.com/nocodeleaks/quepasa/metrics"
)

// Internal metric functions (no direct backend dependencies)
// API module initializes its own metrics directly
var APIProcessingTime = metrics.CreateHistogramVecRecorder(
    "quepasa_api_request_duration_seconds",
    "Time spent processing API requests",
    []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
    []string{"method", "endpoint", "status_code"},
)

func ObserveAPIRequestDuration(method, endpoint, statusCode string, duration float64) {
    if APIProcessingTime != nil {
        histogram := APIProcessingTime.WithLabelValues(method, endpoint, statusCode)
        histogram.Observe(duration)
    }
}
```

### For Backend Implementers

#### 1. Implement Interfaces
```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

// PrometheusCounter implements CounterRecorder for Prometheus
type PrometheusCounter struct {
    counter prometheus.Counter
}

func (p *PrometheusCounter) Inc() {
    go p.counter.Inc() // Async for performance
}

func (p *PrometheusCounter) Add(value float64) {
    go p.counter.Add(value)
}
```

#### 2. Provide Module Implementations
The central metrics module provides implementations for all module-specific metrics through factory functions that return the appropriate interface based on enable/disable state.

## Initialization Flow

1. **Environment Loading**: `environment.Settings.Metrics.Enabled` is set during application startup
2. **Transparent Usage**: Modules use metrics interfaces as if they are always functional
3. **Synchronous Operation**: Metrics module handles enable/disable logic synchronously - enabled metrics go to Prometheus, disabled metrics are no-ops that return success
4. **No Module Awareness**: Individual modules have no knowledge of metrics enable/disable state

## Key Principle: Central Control

**Each module owns its initialization logic**. Modules call generic factory functions directly at package level, with the central metrics module providing only the factory functions that handle enable/disable logic internally.

```go
// Module initializes directly using generic factory functions
var (
    APIProcessingTime = metrics.CreateHistogramVecRecorder(
        "quepasa_api_request_duration_seconds",
        "Time spent processing API requests",
        []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
        []string{"method", "endpoint", "status_code"},
    )
    MessagesSent = metrics.CreateCounterRecorder(
        "quepasa_api_messages_sent_total",
        "Total messages sent via API",
    )
)

// Factory function controls enable/disable logic internally
func CreateHistogramVecRecorder(name, help string, buckets []float64, labelNames []string) HistogramVecRecorder {
    if MetricsEnabled {  // Central control
        vec := promauto.NewHistogramVec(prometheus.HistogramOpts{
            Name:    name,
            Help:    help,
            Buckets: buckets,
        }, labelNames)
        return &PrometheusHistogramVecRecorder{vec: vec}
    }
    return &NoOpHistogramVecRecorder{}  // No-op when disabled
}
```

This ensures modules remain completely unaware of metrics configuration while the central metrics module provides generic factory functions.

## Migration Guide

### From Direct Prometheus Usage
**Before:**
```go
import "github.com/prometheus/client_golang/prometheus"

var MyCounter = prometheus.NewCounter(...)
```

**After:**
```go
import "github.com/nocodeleaks/quepasa/metrics"

var MyCounter = metrics.CreateCounterRecorder("my_counter_total", "Description of my counter")
```

### From Module-Specific Metrics in Central Module
**Before:** Metrics defined in central `metrics/` module with module-specific knowledge
**After:** Each module defines its own metrics using generic factory functions

## Future Backend Support

The interface-based design allows for easy addition of new backends:

- **InfluxDB**: Implement interfaces using InfluxDB client
- **StatsD**: Implement using StatsD protocol
- **Custom**: Implement for proprietary monitoring systems

Backend selection can be controlled via environment variables:
```
METRICS_BACKEND=prometheus  # or influxdb, statsd, etc.
```

## Best Practices

1. **Always check metrics enabled**: Use environment-based initialization
2. **Use async recording**: For performance, wrap metrics in goroutines
3. **Define clear ownership**: Know which module owns each metric
4. **Document metric usage**: Explain what each metric measures
5. **Test with metrics disabled**: Ensure functionality works without monitoring
6. **Use descriptive names**: Metric names should be self-explanatory

## Common Patterns

### Request Duration Tracking
```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    defer func() {
        duration := time.Since(start).Seconds()
        ObserveRequestDuration(r.Method, r.URL.Path, getStatusCode(w), duration)
    }()
    // Handle request
}
```

### Error Counting
```go
func ProcessMessage(msg *Message) error {
    if err := process(msg); err != nil {
        ProcessingErrors.Inc()
        return err
    }
    ProcessedMessages.Inc()
    return nil
}
```

### Resource Gauges
```go
func ConnectionPool() {
    ActiveConnections.Inc()
    defer ActiveConnections.Dec()
    // Use connection
}
```

This architecture ensures QuePasa can evolve its monitoring capabilities while maintaining clean, maintainable code across all modules.

## Migration Summary

### What Was Migrated
1. **API Module**: Moved all API metrics from central metrics module to API module with direct initialization
2. **WhatsMeow Module**: Moved WhatsMeow metrics to WhatsMeow module with direct initialization
3. **Models Module**: Moved webhook metrics to models module with direct initialization
4. **Generic Factories**: Created `CreateCounterRecorder()`, `CreateCounterVecRecorder()`, `CreateHistogramVecRecorder()` for modules to create their own metrics
5. **Direct Initialization**: Removed `init()` functions and manual initialization, now uses direct variable assignment at package level
6. **Backend Abstraction**: Created interfaces that work with any metrics backend
7. **Clean Separation**: Central metrics module now only provides generic factory functions, no module-specific knowledge

### Key Changes Made
- **api/api_metrics.go**: Direct variable initialization of API counters using generic factories
- **whatsmeow/whatsmeow_metrics.go**: Direct variable initialization of WhatsMeow metrics
- **models/models_metrics.go**: Direct variable initialization of webhook metrics
- **metrics/metrics_config.go**: Added generic factory functions, removed module-specific implementations
- **metrics/metric_extensions.go**: Removed API-specific metric definitions
- **main.go**: Removed manual API metrics initialization
- **Architecture**: Each module now owns and specifies its own metrics

### Implementation Rules for AI Agents
- **ALWAYS use direct variable initialization**: Never use `init()` functions or `_metrics` pattern
- **ALWAYS use generic factory functions**: Call `metrics.CreateCounterRecorder()`, `metrics.CreateHistogramVecRecorder()`, etc. directly
- **NEVER implement module-specific logic in central metrics module**: Keep central module generic
- **ALWAYS declare metrics as package-level variables**: Use `var MyMetric = metrics.CreateCounterRecorder(...)`
- **NEVER import Prometheus directly in modules**: Use only the metrics interfaces
- **ALWAYS follow the module ownership principle**: API metrics belong in API module, WhatsMeow metrics in WhatsMeow module</content>
