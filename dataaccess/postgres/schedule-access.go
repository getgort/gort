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

package postgres

import (
	"context"
	"fmt"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/telemetry"
	"go.opentelemetry.io/otel"
)

//TODO(grebneerg): Assign unique IDs and put in database here.

func (da PostgresDataAccess) ScheduleCreate(ctx context.Context, command *data.ScheduledCommand) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "postgres.ScheduleCreate")
	defer sp.End()

	if command.ScheduleID != 0 {
		return fmt.Errorf("Schedule id already set.")
	}

	command.ScheduleID++

	return nil
}

func (da PostgresDataAccess) ScheduleDelete(ctx context.Context, command data.ScheduledCommand) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "postgres.ScheduleDelete")
	defer sp.End()

	if command.ScheduleID == 0 {
		return fmt.Errorf("Schedule id not set.")
	}

	return nil
}
