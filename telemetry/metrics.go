package telemetry

import (
	"go.opentelemetry.io/otel/metric"
)

// The error counter instrument.
var countErrors metric.Int64Counter

// Errors increments the total errors counter.
func Errors() *MetricCounter {
	return newCounter(countErrors)
}

// The requests counter instrument.
var countUnauthorizedRequests metric.Int64Counter

// UnauthorizedRequestCount increments the total requests counter.
func UnauthorizedRequests() *MetricCounter {
	return newCounter(countUnauthorizedRequests)
}

// The requests counter instrument.
var countTotalRequests metric.Int64Counter

// TotalRequestCount increments the unauthorized requests counter.
func TotalRequests() *MetricCounter {
	return newCounter(countTotalRequests)
}
