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

package memory

import (
	"context"
	"fmt"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/telemetry"
	"go.opentelemetry.io/otel"
)

// Will not implement
func (da *InMemoryDataAccess) RequestBegin(ctx context.Context, req *data.CommandRequest) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "memory.RequestBegin")
	defer sp.End()

	if req.RequestID != 0 {
		return fmt.Errorf("command request ID already set")
	}

	req.RequestID++

	return nil
}

// Will not implement
func (da *InMemoryDataAccess) RequestUpdate(ctx context.Context, result data.CommandResponse) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "memory.RequestUpdate")
	defer sp.End()

	if result.Command.RequestID == 0 {
		return fmt.Errorf("command request ID unset")
	}

	return nil
}
