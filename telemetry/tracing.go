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
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/getgort/gort/config"
)

func CreateAndRegisterExporters() error {
	// First update is synchronous
	if err := updateTracerProviders(); err != nil {
		return err
	}

	// Subsequent updates are not.
	go func() {
		for state := range config.Updates() {
			if state != config.StateConfigInitialized {
				continue
			}

			err := updateTracerProviders()
			if err != nil {
				otel.SetTracerProvider(sdktrace.NewTracerProvider())
				log.WithError(err).Error("Tracer provider not configured (error)")
			}
		}
	}()

	return nil
}

func updateTracerProviders() error {
	jc := config.GetJaegerConfigs()
	event := log.NewEntry(log.StandardLogger())

	if jc.Endpoint == "" {
		// Set the tracer provider with null set.
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		log.Debug("Tracer provider not configured (no config entry)")
		return nil
	}

	exporters := []sdktrace.SpanExporter{}

	if exporter, err := buildJaegerExporter(); err != nil {
		return err
	} else {
		exporters = append(exporters, exporter)
	}

	tpOptions := []sdktrace.TracerProviderOption{}

	for i, e := range exporters {
		event = event.WithField(fmt.Sprintf("exporter%d", i), fmt.Sprintf("%T", e))
		tpOptions = append(tpOptions, sdktrace.WithSyncer(e))
	}

	tp := sdktrace.NewTracerProvider(tpOptions...)
	otel.SetTracerProvider(tp)

	event.Debug("Tracer provider configured")

	return nil
}

func buildJaegerExporter() (sdktrace.SpanExporter, error) {
	jc := config.GetJaegerConfigs()

	if jc.Endpoint == "" {
		return nil, nil
	}

	event := log.WithField("endpoint", jc.Endpoint)

	endpointOptions := []jaeger.CollectorEndpointOption{jaeger.WithEndpoint(jc.Endpoint)}

	if jc.Username != "" {
		endpointOptions = append(endpointOptions, jaeger.WithUsername(jc.Username), jaeger.WithPassword(jc.Password))
		event = event.WithField("username", jc.Username)
	}

	jaegerExporter, err := jaeger.NewRawExporter(
		jaeger.WithCollectorEndpoint(endpointOptions...),
	)
	if err != nil {
		return nil, err
	}

	event.Trace("Jaeger span exporter built")

	return jaegerExporter, nil
}
