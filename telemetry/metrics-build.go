package telemetry

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/unit"
)

const (
	ServiceName = "gort-controller"
)

var meterProvider metric.MeterProvider

func BuildPromExporter() (*prometheus.Exporter, error) {
	// Create and configure the Prometheus exporter
	exporter, err := prometheus.NewExportPipeline(prometheus.Config{})
	if err != nil {
		return nil, err
	}

	meterProvider = exporter.MeterProvider()

	if err := buildCounters(); err != nil {
		return nil, err
	}

	if err := buildRuntimeObservers(); err != nil {
		return nil, err
	}

	return exporter, nil
}

func buildCounters() error {
	var err error

	// Retrieve the meter from the meter provider.
	meter := meterProvider.Meter(ServiceName)

	countErrors, err = meter.NewInt64Counter("gort_controller_errors_total",
		metric.WithDescription("Total number of errors recorded by the Gort controller."),
	)
	if err != nil {
		return err
	}

	countTotalRequests, err = meter.NewInt64Counter("gort_controller_requests_total",
		metric.WithDescription("Total number of requests to the Gort controller."),
	)
	if err != nil {
		return err
	}

	countUnauthorizedRequests, err = meter.NewInt64Counter("gort_controller_requests_unauthorized",
		metric.WithDescription("Number of unauthorized requests to the Gort controller."),
	)
	if err != nil {
		return err
	}

	return nil
}

func buildRuntimeObservers() error {
	var err error

	meter := meterProvider.Meter(ServiceName)

	var m runtime.MemStats
	_, err = meter.NewInt64UpDownSumObserver("gort_controller_memory_usage_bytes",
		func(_ context.Context, result metric.Int64ObserverResult) {
			runtime.ReadMemStats(&m)
			result.Observe(int64(m.Sys), defaultLabels...)
		},
		metric.WithDescription("Amount of memory used in bytes."),
		metric.WithUnit(unit.Bytes),
	)
	if err != nil {
		return err
	}

	_, err = meter.NewInt64UpDownSumObserver("gort_controller_num_goroutines",
		func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(int64(runtime.NumGoroutine()), defaultLabels...)
		},
		metric.WithDescription("Number of running goroutines."),
	)
	if err != nil {
		return err
	}

	return nil
}
