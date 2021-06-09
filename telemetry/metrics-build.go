/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package telemetry

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/unit"
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
